# Task: Build Granja MVP

You are building **Granja**, an AI agent orchestration system. Read the PRD and Technical Design carefully.

## Files to read first:
- `prd.md` - Product requirements
- `design.md` - Technical design with architecture details

## Your mission:
Build the **Go backend** first (the core orchestrator). Start with:

1. **Project structure** - Create the directory layout from the design
2. **Database schema** - SQLite setup with migrations
3. **Domain models** - Epic, Task, Project, Worker entities
4. **API layer** - HTTP handlers for the REST endpoints
5. **Orchestrator** - The main loop that polls for tasks and spawns workers

Use standard Go patterns:
- chi router for HTTP
- sqlx or database/sql for SQLite
- Docker SDK for Go

Start building! Create real, working code. Commit your progress.
