import asyncio
import logging
import grpc
from concurrent import futures
from src import analysis_pb2, analysis_pb2_grpc

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class AnalysisServicer(analysis_pb2_grpc.AnalysisServiceServicer):

    async def Analyze(self, request, context):
        logger.info("Analyze called for registro %s", request.registro_id)
        return analysis_pb2.AnalyzeResponse(
            emotion_principal="pendente",
            emotions=[analysis_pb2.EmotionScore(emotion="neutro", score=1.0)],
            scores={"neutro": 1.0},
            intensidade=0.5,
            analise_adicional="Modelo não treinado — sub-fase 07-02 pendente",
            modelo_versao="0.0.0"
        )


async def serve_grpc(port: int = 50051):
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
    server.add_insecure_port(f"[::]:{port}")
    logger.info("gRPC server starting on port %d", port)
    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    asyncio.run(serve_grpc())
