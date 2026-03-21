PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    repo_url TEXT NOT NULL,
    default_branch TEXT DEFAULT 'main',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS epics (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'planted',
    branch_name TEXT,
    prd_content TEXT NOT NULL,
    design_content TEXT,
    review_result TEXT,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    epic_id TEXT NOT NULL REFERENCES epics(id),
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'todo',
    effort TEXT,
    relevant_files TEXT,
    container_id TEXT,
    worker_logs TEXT,
    started_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS task_deps (
    task_id TEXT NOT NULL REFERENCES tasks(id),
    depends_on_id TEXT NOT NULL REFERENCES tasks(id),
    PRIMARY KEY (task_id, depends_on_id)
);

CREATE TABLE IF NOT EXISTS workers (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL REFERENCES tasks(id),
    container_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'starting',
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_heartbeat DATETIME
);

CREATE INDEX IF NOT EXISTS idx_epics_project_status ON epics(project_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_epic_status ON tasks(epic_id, status);
CREATE INDEX IF NOT EXISTS idx_workers_status ON workers(status);
