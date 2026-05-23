import pytest
import grpc
import os
from concurrent import futures
from src.server import AnalysisServicer, _model_version
from src import analysis_pb2, analysis_pb2_grpc
from src.model_config import MODEL_PATH


@pytest.fixture
def servicer():
    return AnalysisServicer()


@pytest.mark.asyncio
async def test_analyze_returns_emotion_principal(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-1",
        sensacoes="coração acelerado",
        contexto="reunião importante",
        pensamentos="preciso me preparar melhor",
        data_hora="2026-05-23T14:00:00Z",
    )
    response = await servicer.Analyze(request, None)
    assert response.emotion_principal != "pendente"
    assert response.emotion_principal != ""
    assert len(response.emotions) > 0


@pytest.mark.asyncio
async def test_analyze_returns_multiple_emotions(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-2",
        sensacoes="Estou muito feliz e grato por tudo",
        contexto="",
        pensamentos="",
        data_hora="",
    )
    response = await servicer.Analyze(request, None)
    assert len(response.emotions) >= 1
    for e in response.emotions:
        assert e.score >= 0


@pytest.mark.asyncio
async def test_analyze_empty_input_returns_neutro(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-3",
        sensacoes="",
        contexto="",
        pensamentos="",
        data_hora="",
    )
    response = await servicer.Analyze(request, None)
    assert response.emotion_principal == "neutro"
    assert response.intensidade > 0.5


@pytest.mark.asyncio
async def test_modelo_versao_populated(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-4",
        sensacoes="teste",
        contexto="",
        pensamentos="",
        data_hora="",
    )
    response = await servicer.Analyze(request, None)
    assert response.modelo_versao != ""
    assert response.modelo_versao != "0.0.0"


@pytest.mark.asyncio
async def test_analyze_scores_contains_all_labels(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-5",
        sensacoes="Estou bem",
        contexto="",
        pensamentos="",
        data_hora="",
    )
    response = await servicer.Analyze(request, None)
    from src.model_config import LABELS
    for label in LABELS:
        assert label in response.scores, f"Missing score for {label}"


@pytest.mark.asyncio
async def test_grpc_server_starts():
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=1))
    analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
    port = server.add_insecure_port("[::]:0")
    assert port > 0
    await server.stop(None)
