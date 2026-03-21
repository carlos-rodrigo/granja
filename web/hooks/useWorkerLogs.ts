"use client";

import { useEffect, useState } from "react";

import { apiBaseUrl, getWorkerLogs } from "@/lib/api";

interface WorkerLogsState {
  logs: string;
  mode: "sse" | "polling";
  error: string | null;
}

export function useWorkerLogs(containerId?: string, paused = false) {
  const [state, setState] = useState<WorkerLogsState>({
    logs: "",
    mode: "sse",
    error: null
  });

  useEffect(() => {
    if (!containerId) {
      setState({ logs: "", mode: "sse", error: null });
      return;
    }

    if (paused) {
      return;
    }

    let cleanupPolling: (() => void) | null = null;
    let active = true;

    const startPolling = () => {
      setState((prev) => ({ ...prev, mode: "polling", error: null }));
      const interval = setInterval(async () => {
        try {
          const response = await getWorkerLogs(containerId);
          if (!active) {
            return;
          }
          setState((prev) => ({ ...prev, logs: response.logs }));
        } catch (error) {
          if (!active) {
            return;
          }
          setState((prev) => ({
            ...prev,
            error: error instanceof Error ? error.message : "Failed to fetch logs"
          }));
        }
      }, 2_000);

      cleanupPolling = () => clearInterval(interval);
    };

    const streamUrl = `${apiBaseUrl()}/workers/${containerId}/logs`;
    const source = new EventSource(streamUrl);

    source.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as { logs?: string };
        setState((prev) => ({ ...prev, mode: "sse", logs: parsed.logs ?? prev.logs, error: null }));
      } catch {
        setState((prev) => ({ ...prev, mode: "sse", logs: event.data, error: null }));
      }
    };

    source.onerror = () => {
      source.close();
      startPolling();
    };

    return () => {
      active = false;
      source.close();
      cleanupPolling?.();
    };
  }, [containerId, paused]);

  return state;
}
