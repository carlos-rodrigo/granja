"use client";

import { Task } from "@/lib/types";

import { TaskCard } from "./TaskCard";

interface TaskListProps {
  tasks: Task[];
  selectedTaskId?: string;
  onSelectTask?: (task: Task) => void;
}

export function TaskList({ tasks, selectedTaskId, onSelectTask }: TaskListProps) {
  if (tasks.length === 0) {
    return <p className="rounded-lg border border-dashed border-app-border p-3 text-xs text-app-muted">No tasks yet.</p>;
  }

  return (
    <div className="space-y-2">
      {tasks.map((task) => (
        <TaskCard key={task.id} task={task} active={task.id === selectedTaskId} onSelect={onSelectTask} />
      ))}
    </div>
  );
}
