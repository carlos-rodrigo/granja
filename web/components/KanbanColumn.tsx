import { Epic, Task } from "@/lib/types";

import { EpicCard } from "./EpicCard";

interface KanbanColumnProps {
  title: string;
  epics: Epic[];
  expandedEpicId?: string;
  selectedTaskId?: string;
  onToggleEpic: (epicId: string) => void;
  onSelectTask?: (task: Task) => void;
}

export function KanbanColumn({
  title,
  epics,
  expandedEpicId,
  selectedTaskId,
  onToggleEpic,
  onSelectTask
}: KanbanColumnProps) {
  return (
    <section className="min-h-[320px] rounded-2xl border border-app-border bg-app-panel/80 p-3">
      <header className="mb-3 flex items-center justify-between">
        <h2 className="text-xs font-semibold uppercase tracking-[0.18em] text-app-muted">{title}</h2>
        <span className="rounded-full border border-app-border bg-app-panelAlt px-2 py-0.5 text-xs text-app-muted">{epics.length}</span>
      </header>
      <div className="space-y-3">
        {epics.map((epic) => (
          <EpicCard
            key={epic.id}
            epic={epic}
            expanded={expandedEpicId === epic.id}
            onToggle={onToggleEpic}
            selectedTaskId={selectedTaskId}
            onSelectTask={onSelectTask}
          />
        ))}
        {epics.length === 0 ? <p className="rounded-lg border border-dashed border-app-border p-4 text-xs text-app-muted">No epics in this stage.</p> : null}
      </div>
    </section>
  );
}
