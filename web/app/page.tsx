import { KanbanBoard } from "@/components/KanbanBoard";

export default function DashboardPage() {
  return (
    <main className="mx-auto max-w-[1500px] p-5 md:p-8">
      <header className="mb-6">
        <p className="text-xs uppercase tracking-[0.25em] text-app-muted">Granja</p>
        <h1 className="mt-2 text-3xl font-semibold text-app-text md:text-4xl">Kanban Dashboard</h1>
        <p className="mt-2 text-sm text-app-muted">Live epic orchestration board. Auto-refresh every 5 seconds.</p>
      </header>

      <KanbanBoard />
    </main>
  );
}
