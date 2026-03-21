export type EpicStatus = "planted" | "growing" | "ready" | "harvested" | "blocked";
export type TaskStatus = "todo" | "in_progress" | "done" | "blocked";
export type WorkerStatus = "starting" | "working" | "committing" | "done" | "error";

export interface Epic {
  id: string;
  project_id: string;
  title: string;
  status: EpicStatus;
  branch_name: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  epic_id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  effort?: string;
  container_id?: string;
  worker_logs?: string;
  created_at: string;
}

export interface Worker {
  id: string;
  task_id: string;
  container_id: string;
  status: WorkerStatus;
  started_at: string;
  last_heartbeat?: string;
}

export interface EpicDetailsResponse {
  epic: Epic;
  tasks: Task[];
}

export interface WorkerLogsResponse {
  logs: string;
}
