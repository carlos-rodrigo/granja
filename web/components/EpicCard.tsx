"use client";

import Link from "next/link";

import { useEpicDetails } from "@/hooks/useEpics";
import { Epic, Task } from "@/lib/types";
import { cn, statusTone } from "@/lib/utils";

import { TaskList } from "./TaskList";

interface EpicCardProps {
  epic: Epic;
  expanded: boolean;
  onToggle: (epicId: string) => void;
  selectedTaskId?: string;
  onSelectTask?: (task: Task) => void;
}

export function EpicCard({ epic, expanded, onToggle, selectedTaskId, onSelectTask }: EpicCardProps) {
  const { data, isLoading } = useEpicDetails(epic.id);
  const tasks = data?.tasks ?? [];
  const doneTasks = tasks.filter((task) => task.status === "done").length;
  const totalTasks = tasks.length;
  const progressPct = totalTasks ? Math.round((doneTasks / totalTasks) * 100) : 0;

  return (
    <article className="rounded-xl border border-app-border bg-app-panel p-4 shadow-glow">
      <button type="button" onClick={() => onToggle(epic.id)} className="w-full text-left">
        <div className="mb-3 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-sm font-semibold text-app-text">{epic.title}</h3>
            <p className="mt-1 text-xs text-app-muted">Project: {epic.project_id}</p>
          </div>
          <span className={cn("rounded-full border px-2 py-0.5 text-[11px] uppercase tracking-wide", statusTone(epic.status))}>
            {epic.status}
          </span>
        </div>

        <div className="mb-1 flex items-center justify-between text-xs text-app-muted">
          <span>Progress</span>
          <span>
            {doneTasks}/{totalTasks} tasks
          </span>
        </div>
        <div className="h-2 rounded-full bg-slate-800">
          <div className="h-full rounded-full bg-status-ready transition-all" style={{ width: `${progressPct}%` }} />
        </div>
      </button>

      <div className="mt-3 flex items-center justify-between text-xs text-app-muted">
        <Link href={`/epics/${epic.id}`} className="hover:text-app-text">
          Open details →
        </Link>
        <span>{isLoading ? "Syncing..." : "Live"}</span>
      </div>

      {expanded ? (
        <div className="mt-3 border-t border-app-border pt-3">
          <TaskList tasks={tasks} selectedTaskId={selectedTaskId} onSelectTask={onSelectTask} />
        </div>
      ) : null}
    </article>
  );
}
