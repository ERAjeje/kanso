from src.health import app
from fastapi.testclient import TestClient

client = TestClient(app)


def test_health_endpoint():
    response = client.get("/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"
    assert data["service"] == "nlp"
    assert data["grpc_port"] == 50051


def test_health_method():
    response = client.get("/health")
    assert response.headers["content-type"] == "application/json"


# --- Training endpoints ---

def test_model_version_endpoint():
    response = client.get("/model/version")
    assert response.status_code == 200
    data = response.json()
    assert "version" in data


def test_train_endpoint_returns_json():
    response = client.post("/train")
    assert response.status_code == 200
    data = response.json()
    assert "status" in data
