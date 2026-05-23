import pytest
import grpc
from src.server import AnalysisServicer
from src import analysis_pb2, analysis_pb2_grpc


@pytest.fixture
def servicer():
    return AnalysisServicer()


@pytest.mark.asyncio
async def test_analyze_returns_placeholder(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-1",
        sensacoes="coração acelerado",
        contexto="reunião importante",
        pensamentos="preciso me preparar melhor",
        data_hora="2026-05-23T14:00:00Z",
    )
    response = await servicer.Analyze(request, None)
    assert response.emotion_principal == "pendente"
    assert len(response.emotions) > 0
    assert response.emotions[0].emotion == "neutro"
    assert response.emotions[0].score == 1.0
    assert "neutro" in response.scores
    assert response.intensidade == 0.5
    assert "07-02" in response.analise_adicional


@pytest.mark.asyncio
async def test_analyze_empty_fields(servicer):
    request = analysis_pb2.AnalyzeRequest(
        registro_id="test-2",
        sensacoes="",
        contexto="",
        pensamentos="",
        data_hora="",
    )
    response = await servicer.Analyze(request, None)
    assert response.emotion_principal == "pendente"
    assert response.modelo_versao == "0.0.0"


@pytest.mark.asyncio
async def test_grpc_server_starts():
    import asyncio
    from concurrent import futures

    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=1))
    analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
    port = server.add_insecure_port("[::]:0")
    assert port > 0
    await server.stop(None)
