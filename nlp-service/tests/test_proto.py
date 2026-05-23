import math
from src import analysis_pb2


def test_analyze_request_message():
    msg = analysis_pb2.AnalyzeRequest(
        registro_id="abc-123",
        sensacoes="coração batendo forte",
        contexto="discussão no trabalho",
        pensamentos="me sentindo injustiçado",
        data_hora="2026-05-23T10:30:00-03:00",
    )
    assert msg.registro_id == "abc-123"
    assert msg.sensacoes == "coração batendo forte"
    assert msg.contexto == "discussão no trabalho"
    assert msg.pensamentos == "me sentindo injustiçado"
    assert msg.data_hora == "2026-05-23T10:30:00-03:00"


def test_emotion_score_message():
    msg = analysis_pb2.EmotionScore(emotion="alegria", score=0.95)
    assert msg.emotion == "alegria"
    assert math.isclose(msg.score, 0.95, rel_tol=1e-3)


def test_analyze_response_message():
    emotion = analysis_pb2.EmotionScore(emotion="neutro", score=1.0)
    msg = analysis_pb2.AnalyzeResponse(
        emotion_principal="neutro",
        emotions=[emotion],
        scores={"neutro": 1.0},
        intensidade=0.5,
        analise_adicional="placeholder",
        modelo_versao="0.0.0",
    )
    assert msg.emotion_principal == "neutro"
    assert len(msg.emotions) == 1
    assert msg.emotions[0].emotion == "neutro"
    assert msg.scores["neutro"] == 1.0
    assert msg.intensidade == 0.5
