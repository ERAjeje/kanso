# NLP Service (v2 — Deferred)

This directory is reserved for the NLP analysis service, planned for v2.

## Purpose

Analyze diary entries using natural language processing to detect emotions
in Portuguese (pt-BR) and enrich registrations with detected emotion tags.

## Planned Stack

- **Python 3.12**
- **FastAPI** — REST API
- **transformers** — Pre-trained emotion classification model
- **Portuguese language model** — e.g., BERTimbau or similar

## Architecture

The NLP service will be an internal service (not exposed to the frontend).
The Go backend will call it asynchronously during report generation or
as a background job when new entries are synced.

## Status

Deferred to v2. The MVP focuses on manual emotion registration with
user-defined sentiment fields. Automated emotion detection via NLP
is a future enhancement.

See: [PROJECT.md](../.planning/PROJECT.md) | [ROADMAP.md](../.planning/ROADMAP.md)
