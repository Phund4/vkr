"""ML serving: health/metrics HTTP + Kafka its.ml.*.in → its.ml.*.out."""

from __future__ import annotations

import io
import json
import logging
import os
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from contextlib import asynccontextmanager
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SERVING_ROOT = Path(os.environ.get("SERVING_ROOT", str(ROOT))).resolve()

try:
    from dotenv import load_dotenv
except ImportError:
    pass
else:
    _env_file = Path(os.environ.get("ENV_FILE", str(ROOT / ".env")))
    if _env_file.is_file():
        load_dotenv(_env_file)

import torch
import torch.nn as nn
from fastapi import FastAPI, HTTPException
from fastapi.responses import Response
from PIL import Image
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Histogram, generate_latest

sys.path.insert(0, str(ROOT))
from inference_core import load_checkpoint_auto, make_transform

from api import kafka_worker

_log = logging.getLogger("ml-serving")

_infer_executor: ThreadPoolExecutor | None = None

_acc_model: nn.Module | None = None
_acc_img: int = 128
_acc_tf = None
_crash_idx: int = 0
_cong_model: nn.Module | None = None
_cong_img: int = 128
_cong_tf = None

_cong_lock = threading.Lock()
_cong_cache: dict[str, tuple[float, dict]] = {}

_acc_ckpt_used: str = ""
_cong_ckpt_used: str = ""
_winners_json_path: str = ""
_acc_from_winners: bool = False
_cong_from_winners: bool = False

ML_KAFKA_MESSAGES_PROCESSED = Counter(
    "ml_kafka_messages_total",
    "Kafka ML pipeline messages",
    ["branch", "stage"],
)
ML_KAFKA_ERRORS = Counter(
    "ml_kafka_errors_total",
    "Kafka ML worker failures",
    ["branch", "stage"],
)
ML_KAFKA_DURATION = Histogram(
    "ml_kafka_process_duration_seconds",
    "Wall time per consumed Kafka frame",
    ["branch"],
    buckets=(0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0),
)


def _congestion_interval_sec() -> float:
    v = os.environ.get("CONGESTION_INTERVAL_SEC", "2").strip()
    try:
        x = float(v)
        return x if x > 0 else 2.0
    except ValueError:
        return 2.0


def _resolve_checkpoint_path(raw: str) -> Path:
    p = Path(raw.strip())
    if p.is_file():
        return p.resolve()
    return (SERVING_ROOT / raw.strip().lstrip("/")).resolve()


def _winners_default_path() -> Path:
    return Path(os.environ.get("WINNERS_JSON", str(SERVING_ROOT / "models" / "winners.json")))


def _load_winners_checkpoints() -> tuple[Path | None, Path | None]:
    path = _winners_default_path()
    if not path.is_file():
        return None, None
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return None, None
    acc_s = (data.get("accident") or {}).get("checkpoint")
    cong_s = (data.get("congestion") or {}).get("checkpoint")
    acc_p = _resolve_checkpoint_path(acc_s) if isinstance(acc_s, str) and acc_s.strip() else None
    cong_p = _resolve_checkpoint_path(cong_s) if isinstance(cong_s, str) and cong_s.strip() else None
    if acc_p is not None and not acc_p.is_file():
        acc_p = None
    if cong_p is not None and not cong_p.is_file():
        cong_p = None
    return acc_p, cong_p


def _startup_load_models() -> None:
    global _acc_model, _acc_img, _acc_tf, _crash_idx, _cong_model, _cong_img, _cong_tf
    global _acc_ckpt_used, _cong_ckpt_used, _winners_json_path, _acc_from_winners, _cong_from_winners

    _acc_ckpt_used = ""
    _cong_ckpt_used = ""
    _winners_json_path = ""
    _acc_from_winners = False
    _cong_from_winners = False

    default_acc = SERVING_ROOT / "artifacts" / "accident" / "baseline-cnn" / "best.pt"
    default_cong = SERVING_ROOT / "artifacts" / "congestion" / "tiny-cnn" / "best.pt"

    w_acc, w_cong = _load_winners_checkpoints()
    wp = _winners_default_path()
    if wp.is_file():
        _winners_json_path = str(wp.resolve())

    acc_env = os.environ.get("ACCIDENT_CKPT", "").strip()
    cong_env = os.environ.get("CONGESTION_CKPT", "").strip()

    if acc_env:
        acc_path = _resolve_checkpoint_path(acc_env)
    elif w_acc is not None:
        acc_path = w_acc
        _acc_from_winners = True
    else:
        acc_path = default_acc

    if cong_env:
        cong_path = _resolve_checkpoint_path(cong_env)
    elif w_cong is not None:
        cong_path = w_cong
        _cong_from_winners = True
    else:
        cong_path = default_cong

    if acc_path.is_file():
        _acc_model, _acc_img, meta = load_checkpoint_auto(acc_path)
        _acc_model.eval()
        _acc_tf = make_transform(_acc_img)
        _crash_idx = int(meta.get("class_to_idx", {}).get("crash", 0))
        _acc_ckpt_used = str(acc_path.resolve())
    if cong_path.is_file():
        _cong_model, _cong_img, _ = load_checkpoint_auto(cong_path)
        _cong_model.eval()
        _cong_tf = make_transform(_cong_img)
        _cong_ckpt_used = str(cong_path.resolve())

    if not os.environ.get("KAFKA_BOOTSTRAP_SERVERS", "").strip():
        _log.warning("KAFKA_BOOTSTRAP_SERVERS не задан — Kafka ML workers не запустятся")


def _configure_torch_threading() -> None:
    intra = int(os.environ.get("TORCH_NUM_THREADS", "1"))
    intra = max(1, intra)
    torch.set_num_threads(intra)
    inter = os.environ.get("TORCH_NUM_INTEROP_THREADS", "").strip()
    if inter:
        try:
            torch.set_num_interop_threads(max(1, int(inter)))
        except RuntimeError:
            pass


def _bytes_to_tensor(data: bytes, tf, device: torch.device) -> torch.Tensor:
    img = Image.open(io.BytesIO(data)).convert("RGB")
    return tf(img).unsqueeze(0).to(device)


def _predict_incident(raw: bytes) -> dict:
    if _acc_model is None or _acc_tf is None:
        raise RuntimeError("accident model not loaded; set ACCIDENT_CKPT")
    x = _bytes_to_tensor(raw, _acc_tf, torch.device("cpu"))
    with torch.no_grad():
        logits = _acc_model(x)
        prob = torch.softmax(logits, dim=1)[0, _crash_idx].item()
        pred = int(logits.argmax(1).item())
    has_incident = pred == _crash_idx
    return {
        "crash_probability": prob,
        "predicted_class_index": pred,
        "crash_class_index": _crash_idx,
        "label": "crash" if has_incident else "normal",
        "has_incident": has_incident,
    }


def _predict_congestion(raw: bytes) -> dict:
    if _cong_model is None or _cong_tf is None:
        raise RuntimeError("congestion model not loaded; set CONGESTION_CKPT")
    x = _bytes_to_tensor(raw, _cong_tf, torch.device("cpu"))
    with torch.no_grad():
        score = float(_cong_model(x).item())
    return {"congestion_score": score, "note": "proxy [0,1] from lab regressor"}


def _congestion_for_pair(raw: bytes, segment_id: str, camera_id: str) -> dict:
    key = f"{segment_id.strip()}|{camera_id.strip()}"
    interval = _congestion_interval_sec()
    now = time.monotonic()
    with _cong_lock:
        if key in _cong_cache:
            last_t, cached = _cong_cache[key]
            if now - last_t < interval:
                return dict(cached)
    out = _predict_congestion(raw)
    now2 = time.monotonic()
    with _cong_lock:
        if key in _cong_cache:
            last_t, cached = _cong_cache[key]
            if now2 - last_t < interval:
                return dict(cached)
        _cong_cache[key] = (now2, dict(out))
        return dict(out)


@asynccontextmanager
async def _lifespan(app: FastAPI):
    global _infer_executor
    _startup_load_models()
    _configure_torch_threading()
    mw = int(os.environ.get("ML_INFER_MAX_THREADS", "0").strip() or "0")
    if mw > 0:
        _infer_executor = ThreadPoolExecutor(max_workers=max(1, mw), thread_name_prefix="ml_infer")
    kafka_worker.start_kafka_workers()
    yield
    kafka_worker.stop_kafka_workers()
    if _infer_executor is not None:
        _infer_executor.shutdown(wait=True, cancel_futures=False)
        _infer_executor = None


app = FastAPI(title="ITS ML Serving", version="0.2", lifespan=_lifespan)


@app.get("/metrics")
def metrics():
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


@app.get("/health")
def health():
    return {
        "accident_loaded": _acc_model is not None,
        "congestion_loaded": _cong_model is not None,
        "congestion_interval_sec": _congestion_interval_sec(),
        "kafka_enabled": bool(os.environ.get("KAFKA_BOOTSTRAP_SERVERS", "").strip()),
        "kafka_topic_accident_in": os.environ.get("KAFKA_TOPIC_ML_ACCIDENT_IN", "its.ml.accident.in"),
        "kafka_topic_accident_out": os.environ.get("KAFKA_TOPIC_ML_ACCIDENT_OUT", "its.ml.accident.out"),
        "kafka_topic_congestion_in": os.environ.get("KAFKA_TOPIC_ML_CONGESTION_IN", "its.ml.congestion.in"),
        "kafka_topic_congestion_out": os.environ.get("KAFKA_TOPIC_ML_CONGESTION_OUT", "its.ml.congestion.out"),
        "accident_checkpoint": _acc_ckpt_used or None,
        "congestion_checkpoint": _cong_ckpt_used or None,
        "winners_json": _winners_json_path or None,
        "serving_root": str(SERVING_ROOT),
        "torch_num_threads": torch.get_num_threads(),
    }
