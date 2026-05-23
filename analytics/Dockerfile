# ==========================================
# STAGE 1: Builder
# ==========================================
FROM python:3.12-slim AS builder

WORKDIR /build

RUN pip install --no-cache-dir poetry==2.0.1

COPY pyproject.toml ./

RUN poetry config virtualenvs.in-project true && \
    poetry install --no-root --no-interaction --no-ansi

# ==========================================
# STAGE 2: Runtime
# ==========================================
FROM python:3.12-slim AS runtime

WORKDIR /app

# Ставим минимальные системные либы, без которых OpenCV не запустится
RUN apt-get update && apt-get install -y --no-install-recommends \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/*

# Забираем готовое окружение
COPY --from=builder /build/.venv /app/.venv

ENV PATH="/app/.venv/bin:$PATH"
ENV PYTHONPATH="/app"
ENV PYTHONDONTWRITEBYTECODE=1

COPY src/ /app/src/

CMD ["python", "src/main.py"]