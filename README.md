# 🌾 Granja

**Multi-project AI agent orchestrator** — Un sistema para coordinar agentes AI que trabajan en múltiples proyectos.

## Concepto

Granja es el "project manager" de tus agentes AI. Recibe PRDs, los parsea en tareas ejecutables, y distribuye el trabajo a agentes disponibles. Todo con visibilidad en tiempo real.

## Flujo de Trabajo

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              FLUJO GRANJA                                │
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
  │ completa │            │ actualiza│            │ (trabajo)│
  └──────────┘            └──────────┘            └──────────┘
        │                       │
        │                       ▼
        │                 ┌──────────┐
        └────────────────►│DASHBOARD │
           next task      │ (Kanban) │
                          └──────────┘
```

## Paso a Paso

### 1️⃣ Submit — Enviar PRD
```bash
granja submit tasks/prd-feature-x.md --project hippo
```
El PM envía un PRD en formato markdown. Granja lo recibe y lo encola para procesamiento.

### 2️⃣ Parse — Granja Procesa
Granja (que es un agente inteligente) lee el PRD y:
- Extrae el **Epic** (título, descripción, contexto)
- Genera **Tasks** individuales con:
  - Título y descripción clara
  - Effort estimado (S/M/L/XL)
  - Archivos relevantes
  - Dependencias (si las hay)

### 3️⃣ Assign — Asignación Inteligente
Granja busca un agente disponible considerando:
- **Rol**: ¿Es tarea de DEV, SUPPORT, QA?
- **Proyecto**: ¿El agente está asignado a este proyecto?
- **Estado**: ¿Está IDLE?

Si no hay agente con proyecto asignado, toma del pool general.

La tarea se envía via **WebSocket** (push, no polling).

### 4️⃣ Execute — El Agente Trabaja
El agente recibe la tarea y ejecuta su loop:
```
recibir tarea → setup repo → trabajar → commit → PR → reportar
```

Durante la ejecución, el agente reporta:
- Progreso (commits, archivos tocados)
- Blockers (si se traba)
- Preguntas (si necesita clarificación)

### 5️⃣ Report — Actualización y Siguiente
Cuando el agente completa:
1. Envía señal de **COMPLETE** + PR URL
2. Granja marca la tarea como **REVIEW** o **DONE**
3. El agente pasa a **IDLE**
4. Granja le asigna la siguiente tarea (si hay)

### 6️⃣ Dashboard — Visibilidad Total
El dashboard muestra en tiempo real:
- **Kanban por proyecto**: Backlog → In Progress → Review → Done
- **Estado de agentes**: Quién trabaja en qué
- **Activity feed**: Stream de eventos

---

## Roles de Agentes

| Rol | Descripción | Tareas típicas |
|-----|-------------|----------------|
| **DEV** | Desarrollador | Código, features, bugfixes |
| **SUPPORT** | Soporte | Emails, preguntas, escalaciones |
| **QA** | Testing | Tests, validación, reportes |
| **DESIGN** | Diseño | Assets, mockups, UI review |

## Asignación a Proyectos

Los agentes pueden estar:
- **Asignados a un proyecto**: Solo reciben tareas de ese proyecto
- **En el pool general**: Reciben cualquier tarea disponible

Cuando un proyecto vacía su backlog, el agente se **libera automáticamente** al pool.

---

## Estructura del Repo

```
granja/
├── README.md           # Este archivo
├── tasks/              # PRDs y specs
│   └── prd-granja.md   # PRD principal
├── src/                # Código fuente (próximamente)
│   ├── parser/         # Parser de PRDs (AI)
│   ├── scheduler/      # Asignación de tareas
│   ├── hub/            # WebSocket hub
│   └── dashboard/      # UI Next.js
└── agents/             # Configs de agentes (próximamente)
```

---

## Tech Stack

- **Backend**: Next.js API routes / Node.js
- **Database**: SQLite (MVP) → Postgres (escala)
- **Real-time**: WebSocket nativo
- **AI Parser**: Claude 3.5 Haiku via OpenRouter
- **Dashboard**: Next.js + React + Tailwind

---

## Status

🚧 **En desarrollo** — Definiendo arquitectura y PRD inicial.

Ver [tasks/prd-granja.md](tasks/prd-granja.md) para el PRD completo con User Stories.
