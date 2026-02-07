# 🌾 Granja

**Multi-project AI agent orchestrator** — A system to coordinate AI agents working across multiple projects.

## Concept

Granja is the "project manager" for your AI agents. It receives PRDs, parses them into executable tasks, and distributes work to available agents. All with real-time visibility.

## Workflow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            GRANJA WORKFLOW                               │
└─────────────────────────────────────────────────────────────────────────┘

  1. SUBMIT                2. PARSE                 3. ASSIGN
  ────────                 ─────────                ────────
  
  ┌──────────┐            ┌──────────┐            ┌──────────┐
  │   PRD    │  ───────►  │  GRANJA  │  ───────►  │  AGENT   │
  │  (.md)   │            │ (parser) │            │  (DEV)   │
  └──────────┘            └──────────┘            └──────────┘
                                │                       │
                                ▼                       │
                          ┌──────────┐                  │
                          │  EPIC    │                  │
                          │  + Tasks │                  │
                          └──────────┘                  │
                                                        │
  4. EXECUTE               5. REPORT                    │
  ──────────               ─────────                    │
                                                        ▼
  ┌──────────┐            ┌──────────┐            ┌──────────┐
  │  AGENT   │  ───────►  │ GRANJA   │  ◄──────   │  LOOP    │
  │ completes│            │ updates  │            │  (work)  │
  └──────────┘            └──────────┘            └──────────┘
        │                       │
        │                       ▼
        │                 ┌──────────┐
        └────────────────►│DASHBOARD │
           next task      │ (Kanban) │
                          └──────────┘
```

## Step by Step

### 1️⃣ Submit — Send PRD
```bash
granja submit tasks/prd-feature-x.md --project hippo
```
The PM submits a PRD in markdown format. Granja receives it and queues it for processing.

### 2️⃣ Parse — Granja Processes
Granja (which is an intelligent agent itself) reads the PRD and:
- Extracts the **Epic** (title, description, context)
- Generates individual **Tasks** with:
  - Clear title and description
  - Estimated effort (S/M/L/XL)
  - Relevant files
  - Dependencies (if any)

### 3️⃣ Assign — Smart Assignment
Granja finds an available agent considering:
- **Role**: Is this a DEV, SUPPORT, or QA task?
- **Project**: Is the agent assigned to this project?
- **Status**: Is the agent IDLE?

If no project-assigned agent is available, it pulls from the general pool.

Tasks are sent via **WebSocket** (push, not polling).

### 4️⃣ Execute — Agent Works
The agent receives the task and runs its loop:
```
receive task → setup repo → work → commit → PR → report
```

During execution, the agent reports:
- Progress (commits, files touched)
- Blockers (if stuck)
- Questions (if clarification needed)

### 5️⃣ Report — Update and Next
When the agent completes:
1. Sends **COMPLETE** signal + PR URL
2. Granja marks the task as **REVIEW** or **DONE**
3. Agent goes to **IDLE**
4. Granja assigns the next task (if available)

### 6️⃣ Dashboard — Full Visibility
The dashboard shows in real-time:
- **Kanban per project**: Backlog → In Progress → Review → Done
- **Agent status**: Who's working on what
- **Activity feed**: Event stream

---

## Agent Roles

| Role | Description | Typical Tasks |
|------|-------------|---------------|
| **DEV** | Developer | Code, features, bugfixes |
| **SUPPORT** | Support | Emails, questions, escalations |
| **QA** | Testing | Tests, validation, reports |
| **DESIGN** | Design | Assets, mockups, UI review |

## Project Assignment

Agents can be:
- **Assigned to a project**: Only receive tasks from that project
- **In the general pool**: Receive any available task

When a project empties its backlog, the agent is **automatically released** to the pool.

---

## Repo Structure

```
granja/
├── README.md           # This file
├── tasks/              # PRDs and specs
│   └── prd-granja.md   # Main PRD
├── src/                # Source code (coming soon)
│   ├── parser/         # PRD parser (AI)
│   ├── scheduler/      # Task assignment
│   ├── hub/            # WebSocket hub
│   └── dashboard/      # Next.js UI
└── agents/             # Agent configs (coming soon)
```

---

## Tech Stack

- **Backend**: Next.js API routes / Node.js
- **Database**: SQLite (MVP) → Postgres (scale)
- **Real-time**: Native WebSocket
- **AI Parser**: Claude 3.5 Haiku via OpenRouter
- **Dashboard**: Next.js + React + Tailwind

---

## Status

🚧 **In development** — Defining architecture and initial PRD.

See [tasks/prd-granja.md](tasks/prd-granja.md) for the full PRD with User Stories.
