import { Task } from "@/lib/types";
import { cn, statusTone } from "@/lib/utils";

interface TaskCardProps {
  task: Task;
  active?: boolean;
  onSelect?: (task: Task) => void;
}

export function TaskCard({ task, active = false, onSelect }: TaskCardProps) {
  return (
    <button
      type="button"
      onClick={() => onSelect?.(task)}
      className={cn(
        "w-full rounded-lg border border-app-border/70 bg-app-panelAlt p-3 text-left transition",
        "hover:border-app-border hover:bg-app-panel",
        active && "border-status-growing/60 bg-app-panel"
      )}
    >
      <div className="mb-2 flex items-center justify-between gap-2">
        <h4 className="line-clamp-2 text-sm font-medium text-app-text">{task.title}</h4>
        <span className={cn("rounded-full border px-2 py-0.5 text-[11px] uppercase tracking-wide", statusTone(task.status))}>
          {task.status.replace("_", " ")}
        </span>
      </div>
      {task.description ? <p className="text-xs text-app-muted line-clamp-2">{task.description}</p> : null}
    </button>
  );
}
