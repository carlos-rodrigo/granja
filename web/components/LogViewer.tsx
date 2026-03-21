"use client";

import { useEffect, useRef } from "react";

import { useWorkerLogs } from "@/hooks/useWorkerLogs";

interface LogViewerProps {
  containerId?: string;
}

export function LogViewer({ containerId }: LogViewerProps) {
  const { logs, mode, error } = useWorkerLogs(containerId);
  const preRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    const node = preRef.current;
    if (!node) {
      return;
    }
    node.scrollTop = node.scrollHeight;
  }, [logs]);

  return (
    <section className="rounded-xl border border-app-border bg-app-panel p-4 shadow-glow">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h3 className="text-sm font-semibold uppercase tracking-wide text-app-muted">Worker Logs</h3>
        <span className="text-xs text-app-muted">Mode: {mode}</span>
      </div>

      {!containerId ? (
        <p className="text-xs text-app-muted">Select an in-progress task to inspect worker output.</p>
      ) : (
        <pre
          ref={preRef}
          className="h-64 overflow-auto rounded-lg border border-app-border/70 bg-black/40 p-3 text-xs leading-relaxed text-slate-200"
        >
          {logs || "Waiting for logs..."}
        </pre>
      )}

      {error ? <p className="mt-2 text-xs text-status-blocked">{error}</p> : null}
    </section>
  );
}
