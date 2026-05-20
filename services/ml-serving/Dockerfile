# Полный bookworm (не slim): libgomp и прочие runtime-зависимости PyTorch уже в образе — без apt-get.
FROM python:3.12-bookworm

WORKDIR /app
# CPU-сборки torch/torchvision без CUDA (меньше образ и быстрее сборка).
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir \
      "torch>=2.9.0,<3" "torchvision>=0.24.0,<1" \
      --index-url https://download.pytorch.org/whl/cpu && \
    pip install --no-cache-dir \
      "numpy>=2.1.0" "Pillow>=10.4.0" "fastapi>=0.115.0" "uvicorn[standard]>=0.32.0" \
      "python-dotenv>=1.0.0" "prometheus-client>=0.20.0" "kafka-python>=2.0.2"

COPY api/ ./api/
COPY inference_core.py .
COPY models/ ./models/
COPY artifacts/ ./artifacts/

ENV PYTHONPATH=/app
ENV SERVING_ROOT=/app
EXPOSE 8000
CMD ["uvicorn", "api.main:app", "--host", "0.0.0.0", "--port", "8000"]
