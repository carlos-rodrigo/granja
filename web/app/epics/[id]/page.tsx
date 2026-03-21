"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { LogViewer } from "@/components/LogViewer";
import { TaskList } from "@/components/TaskList";
import { useEpicDetails } from "@/hooks/useEpics";
import { useWorkers } from "@/hooks/useWorkers";
import { Task } from "@/lib/types";
import { cn, statusTone } from "@/lib/utils";
import { useMemo, useState } from "react";

export default function EpicDetailPage() {
  const params = useParams<{ id: string }>();
  const epicId = params?.id;
  const [selectedTask, setSelectedTask] = useState<Task>();

  const { data, isLoading, error } = useEpicDetails(epicId, Boolean(epicId));
  const { data: workers = [] } = useWorkers();

  const selectedContainer = useMemo(() => {
    if (!selectedTask?.container_id) {
      return undefined;
    }
    return workers.find((worker) => worker.container_id === selectedTask.container_id)?.container_id ?? selectedTask.container_id;
  }, [workers, selectedTask]);

  if (isLoading) {
    return <main className="mx-auto max-w-4xl p-6 text-app-muted">Loading epic...</main>;
  }

  if (error || !data) {
    return <main className="mx-auto max-w-4xl p-6 text-status-blocked">Could not load epic.</main>;
  }

  const doneTasks = data.tasks.filter((task) => task.status === "done").length;

  return (
    <main className="mx-auto max-w-6xl space-y-5 p-6">
      <Link href="/" className="text-sm text-app-muted hover:text-app-text">
        ← Back to dashboard
      </Link>

      <section className="rounded-2xl border border-app-border bg-app-panel p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold text-app-text">{data.epic.title}</h1>
            <p className="mt-1 text-sm text-app-muted">Project: {data.epic.project_id}</p>
          </div>
          <span className={cn("rounded-full border px-3 py-1 text-xs uppercase tracking-widest", statusTone(data.epic.status))}>
            {data.epic.status}
          </span>
        </div>

        <p className="mt-3 text-sm text-app-muted">
          Progress: {doneTasks}/{data.tasks.length} tasks completed
        </p>
      </section>

      <section className="grid gap-5 lg:grid-cols-2">
        <div className="rounded-2xl border border-app-border bg-app-panel p-4">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-app-muted">Tasks</h2>
          <TaskList tasks={data.tasks} selectedTaskId={selectedTask?.id} onSelectTask={setSelectedTask} />
        </div>

        <LogViewer containerId={selectedContainer} />
      </section>
    </main>
  );
}
