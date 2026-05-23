FROM python:3.12-slim AS build

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Download model during build
COPY download_model.py .
RUN python download_model.py

FROM python:3.12-slim AS runtime

WORKDIR /app
COPY --from=build /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=build /model /model

COPY src/ ./src/
COPY proto/ ./proto/

ENV MODEL_PATH=/model

EXPOSE 8000 50051

HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD python -c "import urllib.request; exit(0 if urllib.request.urlopen('http://localhost:8000/health').status == 200 else 1)"

CMD ["python", "-m", "src"]
