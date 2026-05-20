"""Kafka-конвейер: its.ml.*.in → инференс → its.ml.*.out (без HTTP к analytics)."""

from __future__ import annotations

import base64
import json
import logging
import os
import threading
import time
from typing import Any

from kafka import KafkaConsumer, KafkaProducer
from kafka.errors import KafkaError, NoBrokersAvailable

_log = logging.getLogger("ml-serving.kafka")

_connect_backoff_sec = 3

# Импорт inference из main после загрузки моделей (lazy).
_main: Any = None


def _brokers() -> list[str]:
    raw = os.environ.get("KAFKA_BOOTSTRAP_SERVERS", "").strip()
    return [p.strip() for p in raw.split(",") if p.strip()]


def _topic(name: str, default: str) -> str:
    v = os.environ.get(name, "").strip()
    return v or default


def _decode_frame(payload: bytes) -> tuple[bytes, dict[str, str]]:
    data = json.loads(payload.decode("utf-8"))
    b64 = (data.get("jpeg_base64") or "").strip()
    if not b64:
        raise ValueError("jpeg_base64 missing")
    raw = base64.b64decode(b64, validate=True)
    meta = {
        "segment_id": str(data.get("segment_id") or "").strip(),
        "camera_id": str(data.get("camera_id") or "").strip(),
        "observed_at": str(data.get("observed_at") or "").strip(),
        "pipeline_started_at": str(data.get("pipeline_started_at") or "").strip(),
        "s3_key": str(data.get("s3_key") or "").strip(),
    }
    if not meta["segment_id"] or not meta["camera_id"] or not meta["observed_at"]:
        raise ValueError("segment_id, camera_id, observed_at required")
    return raw, meta


def _result_envelope(meta: dict[str, str], ml: dict) -> bytes:
    body = {
        "segment_id": meta["segment_id"],
        "camera_id": meta["camera_id"],
        "observed_at": meta["observed_at"],
        "s3_key": meta["s3_key"],
        "pipeline_started_at": meta["pipeline_started_at"],
        "ml": ml,
    }
    return json.dumps(body, separators=(",", ":")).encode("utf-8")


def _connect_kafka(
    *,
    branch: str,
    topic_in: str,
    group_id: str,
    brokers: list[str],
    stop: threading.Event,
) -> tuple[KafkaConsumer, KafkaProducer]:
    """Подключение к брокеру с повтором (при старте pod Kafka может быть ещё недоступна)."""
    while not stop.is_set():
        try:
            consumer = KafkaConsumer(
                topic_in,
                bootstrap_servers=brokers,
                group_id=group_id,
                enable_auto_commit=True,
                auto_offset_reset="latest",
                max_poll_records=1,
                consumer_timeout_ms=1000,
            )
            producer = KafkaProducer(
                bootstrap_servers=brokers,
                acks="all",
                linger_ms=5,
            )
            return consumer, producer
        except (KafkaError, NoBrokersAvailable) as e:
            _log.warning(
                "kafka connect %s (%s): %s — retry in %ss",
                branch,
                ",".join(brokers),
                e,
                _connect_backoff_sec,
            )
            time.sleep(_connect_backoff_sec)
    raise RuntimeError("stop requested before kafka connect")


def _run_branch(
    *,
    branch: str,
    topic_in: str,
    topic_out: str,
    group_id: str,
    stop: threading.Event,
) -> None:
    import api.main as main

    brokers = _brokers()
    if not brokers:
        _log.error("KAFKA_BOOTSTRAP_SERVERS empty, %s worker not started", branch)
        return

    try:
        consumer, producer = _connect_kafka(
            branch=branch,
            topic_in=topic_in,
            group_id=group_id,
            brokers=brokers,
            stop=stop,
        )
    except RuntimeError:
        return

    _log.info(
        "kafka worker started branch=%s in=%s out=%s group=%s",
        branch,
        topic_in,
        topic_out,
        group_id,
    )

    while not stop.is_set():
        try:
            batch = consumer.poll(timeout_ms=1000, max_records=1)
        except KafkaError as e:
            _log.warning("kafka poll %s: %s", branch, e)
            time.sleep(1)
            continue
        if not batch:
            continue
        for _tp, records in batch.items():
            for rec in records:
                if stop.is_set():
                    break
                t0 = time.perf_counter()
                outcome = "success"
                try:
                    raw, meta = _decode_frame(rec.value)
                    key = meta["segment_id"].encode("utf-8")
                    if branch == "accident":
                        incident = main._predict_incident(raw)
                        out = _result_envelope(meta, {"incident": incident})
                    else:
                        cong = main._congestion_for_pair(
                            raw, meta["segment_id"], meta["camera_id"]
                        )
                        out = _result_envelope(meta, {"congestion": cong})
                    fut = producer.send(topic_out, key=key, value=out)
                    fut.get(timeout=30)
                    main.ML_KAFKA_MESSAGES_PROCESSED.labels(branch=branch, stage="publish").inc()
                except Exception as e:
                    outcome = "error"
                    main.ML_KAFKA_ERRORS.labels(branch=branch, stage="process").inc()
                    _log.warning("kafka process %s: %s", branch, e)
                finally:
                    main.ML_KAFKA_DURATION.labels(branch=branch).observe(
                        time.perf_counter() - t0
                    )
                    main.ML_KAFKA_MESSAGES_PROCESSED.labels(
                        branch=branch, stage="consume"
                    ).inc()
                    if outcome == "success":
                        main.ML_KAFKA_MESSAGES_PROCESSED.labels(
                            branch=branch, stage="ok"
                        ).inc()

    try:
        consumer.close()
    except Exception:
        pass
    try:
        producer.flush(5)
        producer.close()
    except Exception:
        pass
    _log.info("kafka worker stopped branch=%s", branch)


_workers: list[threading.Thread] = []
_stop = threading.Event()


def start_kafka_workers() -> None:
    if not os.environ.get("KAFKA_BOOTSTRAP_SERVERS", "").strip():
        _log.warning("KAFKA_BOOTSTRAP_SERVERS not set — Kafka ML workers disabled")
        return
    base_group = os.environ.get("KAFKA_CONSUMER_GROUP", "ml-serving").strip() or "ml-serving"
    pairs = (
        (
            "accident",
            _topic("KAFKA_TOPIC_ML_ACCIDENT_IN", "its.ml.accident.in"),
            _topic("KAFKA_TOPIC_ML_ACCIDENT_OUT", "its.ml.accident.out"),
            f"{base_group}-accident",
        ),
        (
            "congestion",
            _topic("KAFKA_TOPIC_ML_CONGESTION_IN", "its.ml.congestion.in"),
            _topic("KAFKA_TOPIC_ML_CONGESTION_OUT", "its.ml.congestion.out"),
            f"{base_group}-congestion",
        ),
    )
    _stop.clear()
    for branch, tin, tout, gid in pairs:
        th = threading.Thread(
            target=_run_branch,
            kwargs={
                "branch": branch,
                "topic_in": tin,
                "topic_out": tout,
                "group_id": gid,
                "stop": _stop,
            },
            name=f"ml-kafka-{branch}",
            daemon=True,
        )
        th.start()
        _workers.append(th)


def stop_kafka_workers() -> None:
    _stop.set()
    for th in _workers:
        th.join(timeout=15)
    _workers.clear()
