import { Epic, EpicDetailsResponse, Worker, WorkerLogsResponse } from "@/lib/types";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:3000/api";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    cache: "no-store"
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || `Request failed: ${response.status}`);
  }

  return (await response.json()) as T;
}

export function apiBaseUrl() {
  return API_BASE;
}

export async function getEpics(): Promise<Epic[]> {
  return request<Epic[]>("/epics");
}

export async function getEpic(id: string): Promise<EpicDetailsResponse> {
  return request<EpicDetailsResponse>(`/epics/${id}`);
}

export async function getWorkers(): Promise<Worker[]> {
  return request<Worker[]>("/workers");
}

export async function getWorkerLogs(containerId: string): Promise<WorkerLogsResponse> {
  return request<WorkerLogsResponse>(`/workers/${containerId}/logs`);
}
