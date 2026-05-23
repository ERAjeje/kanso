# NLP Service — Infra NLP (07-01)

Python NLP service for emotion analysis using FastAPI + gRPC.

## Architecture

```
┌──────────────┐    gRPC (internal)    ┌──────────────────┐
│  Go Backend  │ ───────────────────▶ │  NLP Service     │
│  (kanso-api) │                      │  (FastAPI+gRPC)  │
│              │ ◀─────────────────── │                  │
│  nlp.Client  │    AnalyzeResponse   │  BERTimbau model │
└──────────────┘                      │  (in image)      │
                                       └──────────────────┘
                                                │
                                       ┌────────▼───────┐
                                       │  Docker build   │
                                       │  downloads      │
                                       │  model to image │
                                       └─────────────────┘
```

## Status

**Sub-phase 07-01 complete.** The service scaffold is in place:

- gRPC proto definition (`proto/analysis.proto`)
- FastAPI health endpoint (`/health`)
- gRPC server with placeholder `AnalysisServicer`
- Multi-stage Dockerfile with BERTimbau model download at build
- docker-compose integration (service `nlp`)
- Go gRPC client stub (`backend/internal/nlp/client.go`)

## Next Steps

- **07-02:** Fine-tune BERTimbau for emotion classification
- **07-03:** Go _changes listener + CouchDB enrichment + frontend display

## Development

```bash
# Generate Python gRPC stubs
bash gen_proto.sh

# Run tests
python -m pytest tests/ -v

# Start locally
python -m src
```

## Docker

```bash
docker build -t kanso-nlp .
docker compose -f ../infra/docker-compose.yml up nlp
```

See: `.planning/PROJECT.md` | `.planning/ROADMAP.md`
