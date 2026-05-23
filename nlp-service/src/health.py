from fastapi import FastAPI

app = FastAPI(title="Kanso NLP Service")


@app.get("/health")
async def health():
    return {"status": "ok", "service": "nlp", "grpc_port": 50051}
