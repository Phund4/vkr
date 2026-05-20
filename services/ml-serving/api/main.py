"""ML serving API: incident classification and congestion score."""

from __future__ import annotations

import asyncio
import io
import json
import logging
import os
import sys
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from contextlib import asynccontextmanager
from datetime import datetime, timezone
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

import httpx
import torch
import torch.nn as nn
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from fastapi.responses import JSONResponse, Response
from prometheus_client import CONTENT_TYPE_LATEST, Counter, Histogram, generate_latest
from PIL import Image

sys.path.insert(0, str(ROOT))
from inference_core import load_checkpoint_auto, make_transform

_log = logging.getLogger("ml-serving")

# Переиспользуемый HTTP-клиент для push в analytics (меньше TLS/handshake под нагрузкой).
_analytics_httpx: httpx.AsyncClient | None = None
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

ML_REQUESTS = Counter(
    "ml_process_requests_total",
    "POST /v1/process outcomes",
    ["result"],
)
ML_DURATION = Histogram(
    "ml_process_duration_seconds",
    "Wall time for /v1/process (read image, inference, optional analytics push)",
    buckets=(0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0),
)
ML_IMAGE_BYTES = Counter(
    "ml_request_image_bytes_total",
    "Total bytes read from uploaded image field",
)
ML_ANALYTICS_PUSH_BYTES = Counter(
    "ml_analytics_ingest_bytes_total",
    "JSON bytes sent to analytics POST /v1/ingest",
)


def _congestion_interval_sec() -> float:
    v = os.environ.get("CONGESTION_INTERVAL_SEC", "2").strip()
    try:
        x = float(v)
        return x if x > 0 else 2.0
    except ValueError:
        return 2.0


def _resolve_checkpoint_path(raw: str) -> Path:
    """Absolute path, or relative to SERVING_ROOT."""
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

    ingest = os.environ.get("ANALYTICS_INGEST_URL", "").strip()
    if not ingest:
        _log.warning(
            "ANALYTICS_INGEST_URL не задан: POST /v1/process без push в analytics "
            "(только JSON-ответ, если не задан режим ingest)."
        )


def _configure_torch_threading() -> None:
    """При параллельных запросах через thread pool не раздуваем BLAS внутри каждого forward."""
    intra = int(os.environ.get("TORCH_NUM_THREADS", "1"))
    intra = max(1, intra)
    torch.set_num_threads(intra)
    inter = os.environ.get("TORCH_NUM_INTEROP_THREADS", "").strip()
    if inter:
        try:
            torch.set_num_interop_threads(max(1, int(inter)))
        except RuntimeError:
            pass


@asynccontextmanager
async def _lifespan(app: FastAPI):
    global _analytics_httpx, _infer_executor
    _startup_load_models()
    _configure_torch_threading()
    mw = int(os.environ.get("ML_INFER_MAX_THREADS", "0").strip() or "0")
    if mw > 0:
        _infer_executor = ThreadPoolExecutor(max_workers=max(1, mw), thread_name_prefix="ml_infer")
        asyncio.get_running_loop().set_default_executor(_infer_executor)
    ingest = os.environ.get("ANALYTICS_INGEST_URL", "").strip()
    if ingest:
        timeout = float(os.environ.get("ANALYTICS_INGEST_TIMEOUT_SEC", "15"))
        max_conn = int(os.environ.get("ANALYTICS_HTTP_MAX_CONNECTIONS", "32"))
        _analytics_httpx = httpx.AsyncClient(
            timeout=timeout,
            limits=httpx.Limits(max_connections=max_conn, max_keepalive_connections=max(4, max_conn // 2)),
        )
    yield
    if _analytics_httpx is not None:
        await _analytics_httpx.aclose()
        _analytics_httpx = None
    if _infer_executor is not None:
        _infer_executor.shutdown(wait=True, cancel_futures=False)
        _infer_executor = None


app = FastAPI(title="ITS ML Serving", version="0.1", lifespan=_lifespan)


@app.get("/metrics")
def metrics():
    """Минимальный exposition для Prometheus (up/scrape)."""
    return Response(content=generate_latest(), media_type=CONTENT_TYPE_LATEST)


@app.get("/health")
def health():
    return {
        "accident_loaded": _acc_model is not None,
        "congestion_loaded": _cong_model is not None,
        "congestion_interval_sec": _congestion_interval_sec(),
        "analytics_ingest_push": bool(os.environ.get("ANALYTICS_INGEST_URL", "").strip()),
        "accident_checkpoint": _acc_ckpt_used or None,
        "congestion_checkpoint": _cong_ckpt_used or None,
        "winners_json": _winners_json_path or None,
        "accident_from_winners_json": _acc_from_winners,
        "congestion_from_winners_json": _cong_from_winners,
        "analytics_ingest_configured": bool(os.environ.get("ANALYTICS_INGEST_URL", "").strip()),
        "serving_root": str(SERVING_ROOT),
        "torch_num_threads": torch.get_num_threads(),
        "ml_infer_max_threads": int(os.environ.get("ML_INFER_MAX_THREADS", "0").strip() or "0"),
    }


def _bytes_to_tensor(data: bytes, tf, device: torch.device) -> torch.Tensor:
    img = Image.open(io.BytesIO(data)).convert("RGB")
    return tf(img).unsqueeze(0).to(device)


def _predict_incident(raw: bytes) -> dict:
    if _acc_model is None or _acc_tf is None:
        raise HTTPException(503, "accident model not loaded; set ACCIDENT_CKPT")
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
        raise HTTPException(503, "congestion model not loaded; set CONGESTION_CKPT")
    x = _bytes_to_tensor(raw, _cong_tf, torch.device("cpu"))
    with torch.no_grad():
        score = float(_cong_model(x).item())
    return {"congestion_score": score, "note": "proxy [0,1] from lab regressor"}


def _congestion_for_pair(raw: bytes, segment_id: str, camera_id: str) -> dict:
    """Кеш по (segment, camera); инференс вне глобального lock — разные камеры параллелятся."""
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


def _process_payload(raw: bytes, segment_id: str = "", camera_id: str = "") -> dict:
    incident = _predict_incident(raw)
    seg = (segment_id or "").strip()
    cam = (camera_id or "").strip()
    congestion = _congestion_for_pair(raw, seg, cam) if seg and cam else _predict_congestion(raw)
    return {"incident": incident, "congestion": congestion}


async def _push_analytics_ingest(
    ml_payload: dict,
    segment_id: str,
    camera_id: str,
    s3_key: str,
    observed_at: str,
    pipeline_started_at: str = "",
) -> None:
    url = os.environ.get("ANALYTICS_INGEST_URL", "").strip()
    if not url:
        return
    timeout = float(os.environ.get("ANALYTICS_INGEST_TIMEOUT_SEC", "15"))
    body = {
        "segment_id": segment_id,
        "camera_id": camera_id,
        "observed_at": observed_at,
        "s3_key": s3_key,
        "ml": ml_payload,
    }
    ps = (pipeline_started_at or "").strip()
    if ps:
        body["pipeline_started_at"] = ps
    payload_bytes = json.dumps(body, separators=(",", ":")).encode("utf-8")
    ML_ANALYTICS_PUSH_BYTES.inc(len(payload_bytes))
    try:
        client = _analytics_httpx
        if client is not None:
            r = await client.post(
                url,
                content=payload_bytes,
                headers={"Content-Type": "application/json"},
            )
        else:
            async with httpx.AsyncClient(timeout=timeout) as ephemeral:
                r = await ephemeral.post(
                    url,
                    content=payload_bytes,
                    headers={"Content-Type": "application/json"},
                )
    except httpx.RequestError as e:
        raise HTTPException(502, f"analytics ingest unreachable: {e}") from e
    if r.status_code < 200 or r.status_code >= 300:
        raise HTTPException(502, f"analytics ingest HTTP {r.status_code}: {r.text[:512]}")


async def _process_branch(
    image: UploadFile,
    segment_id: str | None,
    camera_id: str | None,
    s3_key: str | None,
    observed_at: str | None,
    pipeline_started_at: str | None,
    branch: str,
):
    """branch: accident → только incident; congestion → только congestion (два ML-пути на схеме)."""
    t0 = time.perf_counter()
    outcome = "success"
    try:
        raw = await image.read()
        ML_IMAGE_BYTES.inc(len(raw))
        ingest_url = os.environ.get("ANALYTICS_INGEST_URL", "").strip()
        seg = (segment_id or "").strip()
        cam = (camera_id or "").strip()
        pipe_at = (pipeline_started_at or "").strip()
        if ingest_url:
            if not seg or not cam:
                outcome = "http_400"
                raise HTTPException(
                    400,
                    "segment_id and camera_id are required when ANALYTICS_INGEST_URL is set",
                )
            if branch == "accident":
                incident = await asyncio.to_thread(_predict_incident, raw)
                ml_payload = {"incident": incident}
            else:
                if seg and cam:
                    cong = await asyncio.to_thread(_congestion_for_pair, raw, seg, cam)
                else:
                    cong = await asyncio.to_thread(_predict_congestion, raw)
                ml_payload = {"congestion": cong}
            obs = (observed_at or "").strip() or datetime.now(timezone.utc).isoformat()
            key = (s3_key or "").strip()
            await _push_analytics_ingest(ml_payload, seg, cam, key, obs, pipe_at)
            return Response(status_code=204)
        if branch == "accident":
            incident = await asyncio.to_thread(_predict_incident, raw)
            return JSONResponse({"incident": incident})
        if seg and cam:
            cong = await asyncio.to_thread(_congestion_for_pair, raw, seg, cam)
            return JSONResponse({"congestion": cong})
        cong = await asyncio.to_thread(_predict_congestion, raw)
        return JSONResponse({"congestion": cong})
    except HTTPException as e:
        if outcome == "success":
            outcome = f"http_{e.status_code}"
        raise
    except Exception:
        outcome = "error"
        raise
    finally:
        ML_DURATION.observe(time.perf_counter() - t0)
        ML_REQUESTS.labels(outcome).inc()


@app.post("/v1/process/accident")
async def process_accident(
    image: UploadFile = File(...),
    segment_id: str | None = Form(None),
    camera_id: str | None = Form(None),
    s3_key: str | None = Form(None),
    observed_at: str | None = Form(None),
    pipeline_started_at: str | None = Form(None),
):
    """ML «важные» задачи (инциденты / ДТП) → analytics."""
    return await _process_branch(
        image,
        segment_id,
        camera_id,
        s3_key,
        observed_at,
        pipeline_started_at,
        "accident",
    )


@app.post("/v1/process/congestion")
async def process_congestion(
    image: UploadFile = File(...),
    segment_id: str | None = Form(None),
    camera_id: str | None = Form(None),
    s3_key: str | None = Form(None),
    observed_at: str | None = Form(None),
    pipeline_started_at: str | None = Form(None),
):
    """ML обычный контур (загруженность) → analytics."""
    return await _process_branch(
        image,
        segment_id,
        camera_id,
        s3_key,
        observed_at,
        pipeline_started_at,
        "congestion",
    )


@app.post("/v1/process")
async def process(
    image: UploadFile = File(...),
    segment_id: str | None = Form(None),
    camera_id: str | None = Form(None),
    s3_key: str | None = Form(None),
    observed_at: str | None = Form(None),
    pipeline_started_at: str | None = Form(None),
):
    """Совместимость: оба вывода в одном запросе (без разделения сервисов)."""
    t0 = time.perf_counter()
    outcome = "success"
    try:
        raw = await image.read()
        ML_IMAGE_BYTES.inc(len(raw))
        ingest_url = os.environ.get("ANALYTICS_INGEST_URL", "").strip()
        seg = (segment_id or "").strip()
        cam = (camera_id or "").strip()
        pipe_at = (pipeline_started_at or "").strip()
        if ingest_url:
            if not seg or not cam:
                outcome = "http_400"
                raise HTTPException(
                    400,
                    "segment_id and camera_id are required when ANALYTICS_INGEST_URL is set",
                )
            payload = await asyncio.to_thread(_process_payload, raw, seg, cam)
            obs = (observed_at or "").strip() or datetime.now(timezone.utc).isoformat()
            key = (s3_key or "").strip()
            await _push_analytics_ingest(payload, seg, cam, key, obs, pipe_at)
            return Response(status_code=204)
        body = await asyncio.to_thread(_process_payload, raw, seg, cam)
        return JSONResponse(body)
    except HTTPException as e:
        if outcome == "success":
            outcome = f"http_{e.status_code}"
        raise
    except Exception:
        outcome = "error"
        raise
    finally:
        ML_DURATION.observe(time.perf_counter() - t0)
        ML_REQUESTS.labels(outcome).inc()
