"use client";

import { useMemo, useState } from "react";

import { useEpics } from "@/hooks/useEpics";
import { useWorkers } from "@/hooks/useWorkers";
import { Epic, Task } from "@/lib/types";

import { KanbanColumn } from "./KanbanColumn";
import { LogViewer } from "./LogViewer";

const COLUMNS: Array<{ title: string; key: Epic["status"] }> = [
  { title: "Planted", key: "planted" },
  { title: "Growing", key: "growing" },
  { title: "Ready", key: "ready" },
  { title: "Harvested", key: "harvested" }
];

export function KanbanBoard() {
  const { data: epics = [], isLoading, error } = useEpics();
  const { data: workers = [] } = useWorkers();
  const [expandedEpicId, setExpandedEpicId] = useState<string>();
  const [selectedTask, setSelectedTask] = useState<Task>();

  const grouped = useMemo(() => {
    return COLUMNS.map((column) => ({
      ...column,
      epics: epics.filter((epic) => epic.status === column.key)
    }));
  }, [epics]);

  const selectedWorker = useMemo(() => {
    if (!selectedTask?.container_id) {
      return undefined;
    }
    return workers.find((worker) => worker.container_id === selectedTask.container_id);
  }, [workers, selectedTask]);

  if (isLoading) {
    return <p className="text-sm text-app-muted">Loading epics...</p>;
  }

  if (error) {
    return <p className="text-sm text-status-blocked">Failed to load epics: {String(error)}</p>;
  }

  return (
    <div className="space-y-5">
      <div className="grid gap-4 lg:grid-cols-4">
        {grouped.map((column) => (
          <KanbanColumn
            key={column.key}
            title={column.title}
            epics={column.epics}
            expandedEpicId={expandedEpicId}
            selectedTaskId={selectedTask?.id}
            onToggleEpic={(epicId) => {
              setExpandedEpicId((current) => (current === epicId ? undefined : epicId));
              setSelectedTask(undefined);
            }}
            onSelectTask={setSelectedTask}
          />
        ))}
      </div>

      <LogViewer containerId={selectedWorker?.container_id ?? selectedTask?.container_id} />
    </div>
  );
}
