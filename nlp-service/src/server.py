import asyncio
import logging
import os
import grpc
from concurrent import futures
from src import analysis_pb2, analysis_pb2_grpc
from src.classifier import get_classifier, get_model_version
from src.model_config import MODEL_PATH

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

_model_path = os.environ.get("MODEL_PATH", MODEL_PATH)
_model_version = get_model_version()
logger.info("Model version: %s", _model_version)
_classifier = None
try:
    _classifier = get_classifier(_model_path)
    logger.info("EmotionClassifier loaded from %s", _model_path)
except Exception as e:
    logger.warning(
        "EmotionClassifier not available at %s (%s). "
        "Will lazy-init on first Analyze() call.",
        _model_path, e,
    )


class AnalysisServicer(analysis_pb2_grpc.AnalysisServiceServicer):

    async def Analyze(self, request, context):
        logger.info("Analyze called for registro %s", request.registro_id)
        text_parts = []
        if request.sensacoes:
            text_parts.append(request.sensacoes)
        if request.contexto:
            text_parts.append(request.contexto)
        if request.pensamentos:
            text_parts.append(request.pensamentos)
        text = " ".join(text_parts)

        if not text.strip():
            return analysis_pb2.AnalyzeResponse(
                emotion_principal="neutro",
                emotions=[analysis_pb2.EmotionScore(emotion="neutro", score=0.95)],
                scores={"neutro": 0.95},
                intensidade=0.95,
                analise_adicional="Texto vazio — classificado como neutro",
                modelo_versao=_model_version
            )

        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(None, self._predict_sync, text)

        return analysis_pb2.AnalyzeResponse(
            emotion_principal=result["emotion_principal"],
            emotions=[
                analysis_pb2.EmotionScore(emotion=e["emotion"], score=e["score"])
                for e in result["emotions"]
            ],
            scores=result["scores"],
            intensidade=result["intensidade"],
            analise_adicional="",
            modelo_versao=_model_version,
        )

    def _predict_sync(self, text: str) -> dict:
        global _classifier
        if _classifier is None:
            try:
                _classifier = get_classifier(_model_path)
            except Exception as e:
                logger.error("Failed to lazy-init classifier: %s", e)
                return {
                    "emotion_principal": "neutro",
                    "emotions": [{"emotion": "neutro", "score": 1.0}],
                    "scores": {"neutro": 1.0},
                    "intensidade": 1.0,
                }
        return _classifier.predict(text)


async def serve_grpc(port: int = 50051):
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
    analysis_pb2_grpc.add_AnalysisServiceServicer_to_server(AnalysisServicer(), server)
    server.add_insecure_port(f"[::]:{port}")
    logger.info("gRPC server starting on port %d", port)
    await server.start()
    await server.wait_for_termination()


if __name__ == "__main__":
    asyncio.run(serve_grpc())
