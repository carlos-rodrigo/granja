# PRD: Granja — Multi-Project AI Agent Orchestrator

## Overview

**Granja** (Farm) is an AI-powered project orchestration system that receives PRDs, breaks them into actionable tasks, distributes work to available agents, and tracks progress across multiple projects.

Unlike traditional task queues, Granja itself is an intelligent agent that understands our PRD methodology and can parse, prioritize, and assign work strategically.

## Problem Statement

Managing multiple AI coding agents across different projects is currently manual:
- No central visibility into what agents are working on
- No intelligent task distribution based on agent capabilities/roles
- No unified backlog across projects
- Risk of conflicts when multiple agents touch the same codebase

## Goals

1. **Centralized orchestration** — One system to manage all agents across all projects
2. **Intelligent parsing** — Granja understands PRDs and breaks them into EPICs → Tasks
3. **Role-based assignment** — Agents have specializations (DEV, Support, QA, etc.)
4. **Project affinity** — Agents can be assigned to specific projects or float in the general pool
5. **Real-time visibility** — Dashboard shows status across all projects and agents

## Non-Goals (MVP)

- Agent marketplace / third-party agents
- Billing / cost allocation per agent
- Auto-scaling agent count
- Cross-project dependencies

---

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                         GRANJA                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Parser    │  │  Scheduler  │  │     Task Store      │  │
│  │ (PRD→Tasks) │  │ (Assignment)│  │ (SQLite/Postgres)   │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                    WebSocket Hub                         ││
│  │         (Push tasks to agents, receive updates)          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
         │                    │                    │
         ▼                    ▼                    ▼
    ┌─────────┐          ┌─────────┐          ┌─────────┐
    │ Agent 1 │          │ Agent 2 │          │ Agent 3 │
    │  (DEV)  │          │  (DEV)  │          │(Support)│
    │ Hippo   │          │ (Pool)  │          │ Hippo   │
    └─────────┘          └─────────┘          └─────────┘
```

### Data Model

```
Project
├── id, name, repo_url, default_branch
└── agents[] (assigned agents)

Agent
├── id, name, role (DEV|SUPPORT|QA|DESIGN)
├── status (IDLE|BUSY|OFFLINE)
├── project_id (nullable - if null, takes from pool)
├── current_task_id
└── connection (WebSocket ref)

Epic
├── id, project_id, title, description
├── status (BACKLOG|IN_PROGRESS|DONE)
└── priority, created_at

Task
├── id, epic_id, title, description
├── status (TODO|IN_PROGRESS|REVIEW|DONE|BLOCKED)
├── assigned_agent_id
├── estimated_effort (S|M|L|XL)
└── branch_name, pr_url, completed_at
```

---

## User Stories

### Project Management

#### US-1: Submit PRD
**As** a PM, **I want** to submit a PRD file to Granja **so that** it gets parsed into actionable tasks.

**Acceptance Criteria:**
- CLI command: `granja submit <prd-file.md> --project hippo`
- Granja parses PRD using AI (Claude)
- Creates Epic with extracted title/description
- Creates Tasks with estimated effort
- Returns summary of created items

#### US-2: View Project Backlog
**As** a PM, **I want** to see all pending tasks for a project **so that** I can prioritize work.

**Acceptance Criteria:**
- Dashboard shows Kanban board per project
- Columns: Backlog | In Progress | Review | Done
- Tasks show: title, effort, assigned agent (if any)
- Can drag to reorder priority

#### US-3: Manual Task Creation
**As** a PM, **I want** to create tasks manually **so that** I can add work outside of PRDs.

**Acceptance Criteria:**
- Form: title, description, effort, epic (optional)
- Created tasks go to backlog
- Can assign to specific agent or leave for auto-assignment

### Agent Management

#### US-4: Register Agent
**As** an agent operator, **I want** to register a new agent with Granja **so that** it can receive work.

**Acceptance Criteria:**
- CLI command: `granja agent register --name ralph --role DEV`
- Returns registration link (one-time use)
- Agent connects via link, establishes WebSocket
- Agent appears in dashboard as IDLE

#### US-5: Assign Agent to Project
**As** a PM, **I want** to assign an agent to a specific project **so that** it focuses on that project's tasks.

**Acceptance Criteria:**
- Dashboard: drag agent to project
- Agent only receives tasks from assigned project
- Can unassign to return to general pool

#### US-6: Release Agent from Project
**As** the system, **I want** to auto-release agents when project backlog is empty **so that** they can help elsewhere.

**Acceptance Criteria:**
- When project has no TODO/IN_PROGRESS tasks, release agent
- Agent returns to general pool
- Notification sent to dashboard

### Task Execution

#### US-7: Receive Task Assignment
**As** an agent, **I want** to receive task assignments via push **so that** I don't need to poll.

**Acceptance Criteria:**
- Granja sends task via WebSocket when:
  - Agent is IDLE
  - Task matches agent's role
  - Task is from agent's project (or pool if unassigned)
- Agent receives: task details, repo info, branch name
- Agent confirms receipt, status → IN_PROGRESS

#### US-8: Report Task Progress
**As** an agent, **I want** to report progress updates **so that** the dashboard reflects current state.

**Acceptance Criteria:**
- Agent sends: status updates, commit SHAs, blockers
- Dashboard updates in real-time
- If BLOCKED, reason is visible to PM

#### US-9: Complete Task
**As** an agent, **I want** to mark a task complete **so that** I receive the next assignment.

**Acceptance Criteria:**
- Agent sends: completion signal, PR URL (if applicable)
- Task status → REVIEW (or DONE if no review needed)
- Agent status → IDLE
- Next task assigned automatically

### Support Role

#### US-10: Receive Support Request
**As** a support agent, **I want** to receive incoming support requests **so that** I can respond or escalate.

**Acceptance Criteria:**
- Support requests queued as special task type
- Contains: source (email/chat), content, user info
- Agent can: respond, escalate, or mark resolved

#### US-11: Escalate to Human
**As** a support agent, **I want** to escalate complex issues **so that** humans handle edge cases.

**Acceptance Criteria:**
- Escalate action notifies configured channel (Telegram/email)
- Original request + agent notes included
- Task marked as ESCALATED

### Dashboard

#### US-12: Unified Epic View
**As** a PM, **I want** to see all epics across projects **so that** I understand overall progress.

**Acceptance Criteria:**
- Grid view: Project | Epic | Progress | Priority
- Filter by: project, status, date range
- Sort by priority or progress

#### US-13: Agent Status Overview
**As** a PM, **I want** to see all agents and their current status **so that** I know resource allocation.

**Acceptance Criteria:**
- List: Agent | Role | Status | Current Task | Project
- Real-time status updates
- Click agent → detail view with history

#### US-14: Activity Feed
**As** a PM, **I want** a live activity feed **so that** I see what's happening in real-time.

**Acceptance Criteria:**
- Stream of: task assignments, completions, blockers, escalations
- Filterable by project/agent
- Timestamps in local timezone

---

## Technical Decisions

### Stack
- **Backend:** Next.js API routes (or standalone Node/Bun)
- **Database:** SQLite (MVP) → Postgres (scale)
- **Real-time:** WebSocket (native or Socket.io)
- **AI Parser:** Claude 3.5 Haiku via OpenRouter
- **Dashboard:** Next.js + React + Tailwind

### Agent Protocol

```typescript
// Agent → Granja
interface AgentMessage {
  type: 'READY' | 'PROGRESS' | 'COMPLETE' | 'BLOCKED' | 'HEARTBEAT';
  taskId?: string;
  payload?: {
    status?: string;
    commitSha?: string;
    prUrl?: string;
    blockerReason?: string;
  };
}

// Granja → Agent
interface GranjaMessage {
  type: 'ASSIGN' | 'CANCEL' | 'PING';
  task?: {
    id: string;
    title: string;
    description: string;
    repo: string;
    branch: string;
    context?: string; // relevant files, prior decisions
  };
}
```

### Agent Loop Pattern

Each agent runs a loop similar to:

```bash
while true; do
  # Wait for task via WebSocket
  task=$(wait_for_task)
  
  # Clone/pull repo, checkout branch
  setup_workspace "$task"
  
  # Execute task (AI-driven)
  execute_task "$task"
  
  # Commit, push, create PR
  finalize_work "$task"
  
  # Report completion
  report_complete "$task"
done
```

---

## Milestones

### M1: Core Infrastructure (Week 1-2)
- [ ] Project/Epic/Task data model
- [ ] Basic CRUD API
- [ ] PRD parser (AI-powered)
- [ ] CLI: `granja submit`

### M2: Agent System (Week 3-4)
- [ ] Agent registration flow
- [ ] WebSocket hub
- [ ] Task assignment logic
- [ ] Agent heartbeat/health

### M3: Dashboard (Week 5-6)
- [ ] Kanban board per project
- [ ] Agent status view
- [ ] Activity feed
- [ ] Project/epic overview

### M4: Polish & Roles (Week 7-8)
- [ ] Support role implementation
- [ ] Escalation flow
- [ ] Project affinity / auto-release
- [ ] CLI improvements

---

## Success Metrics

- **Task throughput:** Tasks completed per day per agent
- **Assignment latency:** Time from task created → assigned
- **Completion rate:** % tasks completed vs blocked/abandoned
- **Agent utilization:** % time agents are BUSY vs IDLE

---

## Open Questions

1. **Auth for agents?** — JWT tokens via registration link? API keys?
2. **Multi-repo epics?** — Can one epic span multiple repos?
3. **Task dependencies?** — Should tasks block on other tasks?
4. **Agent capabilities?** — Beyond roles, should agents declare specific skills?

---

## Appendix: Example PRD Parse

**Input PRD excerpt:**
```markdown
## US-1: User Login
As a user, I want to log in with email/password so that I can access my account.

Acceptance Criteria:
- Login form with email and password fields
- Validation errors shown inline
- Redirect to dashboard on success
```

**Parsed Output:**
```json
{
  "epic": {
    "title": "User Authentication",
    "description": "Core auth flows for the application"
  },
  "tasks": [
    {
      "title": "Create login form component",
      "description": "Build LoginForm with email/password fields, client-side validation",
      "effort": "S"
    },
    {
      "title": "Implement login API endpoint",
      "description": "POST /api/auth/login - validate credentials, return session",
      "effort": "M"
    },
    {
      "title": "Add auth redirect logic",
      "description": "Redirect to /dashboard on successful login, handle errors",
      "effort": "S"
    }
  ]
}
```
