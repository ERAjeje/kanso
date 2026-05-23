import asyncio
import uvicorn
import threading
from src.server import serve_grpc
from src.health import app


def run_http():
    uvicorn.run(app, host="0.0.0.0", port=8000, log_level="info")


async def main():
    http_thread = threading.Thread(target=run_http, daemon=True)
    http_thread.start()
    await serve_grpc()


if __name__ == "__main__":
    asyncio.run(main())
