# Task: Build Next.js Frontend (Kanban Dashboard)

Read `design.md` for the frontend architecture. Build inside `web/` directory.

## Requirements:

### 1. Project Setup
- Next.js 14+ with App Router
- TypeScript
- Tailwind CSS for styling
- React Query (TanStack Query) for data fetching

### 2. Pages
- `/` - Main Kanban dashboard
- `/epics/[id]` - Epic detail page

### 3. Kanban Board (`/`)
- 4 columns: Planted | Growing | Ready | Harvested
- Epic cards showing: title, project badge, progress (3/5 tasks)
- Click epic → expand inline to show tasks
- Real-time updates via polling (every 5s)

### 4. Components
- `KanbanBoard` - main board layout
- `KanbanColumn` - single column
- `EpicCard` - epic display with progress
- `TaskList` - list of tasks for an epic
- `TaskCard` - individual task with status
- `LogViewer` - worker logs (SSE streaming)

### 5. API Integration
- Connect to Go backend at `http://localhost:3000/api`
- Endpoints: GET /epics, GET /epics/:id, GET /workers, GET /workers/:id/logs

### 6. Styling
- Dark theme, modern UI
- Status badges with colors (planted=blue, growing=yellow, ready=green, harvested=purple)
- Responsive layout

Commit your changes when done.
