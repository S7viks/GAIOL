# DRAFT — RELAY UNIFIED PLAN

> DRAFT — needs real content

This document is a scaffolded placeholder for the active product plan described in docs/README.md. It should be replaced with full content copied from the team's canonical plan.

## One-pipeline architecture

- Overview: single unified pipeline for orchestration, retrieval, and agent workflows
- Components: web client, API gateway, orchestrator, model adapters, storage layer
- Integration notes: merge of Go orchestrator and TypeScript dashboard/orchestrator responsibilities

## HCC (High-Confidence Computing)

- Goals: define budget/latency/quality tradeoffs, safety checks, and guardrails for production
- Key metrics: per-query cost, success rate, latency, model trust scores

## Go + TS orchestrator merge

- Objective: unify the legacy TypeScript orchestrator (archive/orchestrator) with the currently maintained Go orchestrator
- Phases:
  1. Define API contract for model adapters and workflows
  2. Migrate minimal router features to Go
  3. Remove duplicate orchestrator code and keep canonical implementation

---

_This is a scaffold only. Replace with the true product plan authored by the team._
