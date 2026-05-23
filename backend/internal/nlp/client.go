package nlp

import (
	"context"
	"time"

	pb "github.com/edson/kanso-api/internal/nlp/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.AnalysisServiceClient
}

type AnalyzeRequest struct {
	RegistroID  string
	Sensacoes   string
	Contexto    string
	Pensamentos string
	DataHora    string
}

type EmotionScore struct {
	Emotion string  `json:"emotion"`
	Score   float32 `json:"score"`
}

type AnalyzeResponse struct {
	EmotionPrincipal string            `json:"emotionPrincipal"`
	Emotions         []EmotionScore    `json:"emotions"`
	Scores           map[string]float32 `json:"scores"`
	Intensidade      float32            `json:"intensidade"`
	AnaliseAdicional string            `json:"analiseAdicional"`
	ModeloVersao     string            `json:"modeloVersao"`
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, client: pb.NewAnalysisServiceClient(conn)}, nil
}

func (c *Client) Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	resp, err := c.client.Analyze(ctx, &pb.AnalyzeRequest{
		RegistroId:  req.RegistroID,
		Sensacoes:   req.Sensacoes,
		Contexto:    req.Contexto,
		Pensamentos: req.Pensamentos,
		DataHora:    req.DataHora,
	})
	if err != nil {
		return nil, err
	}
	emotions := make([]EmotionScore, len(resp.Emotions))
	for i, e := range resp.Emotions {
		emotions[i] = EmotionScore{Emotion: e.Emotion, Score: e.Score}
	}
	return &AnalyzeResponse{
		EmotionPrincipal: resp.EmotionPrincipal,
		Emotions:         emotions,
		Scores:           resp.Scores,
		Intensidade:      resp.Intensidade,
		AnaliseAdicional: resp.AnaliseAdicional,
		ModeloVersao:     resp.ModeloVersao,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
