"use client";

import { useQuery } from "@tanstack/react-query";

import { getEpic, getEpics } from "@/lib/api";

export function useEpics() {
  return useQuery({
    queryKey: ["epics"],
    queryFn: getEpics,
    refetchInterval: 5_000
  });
}

export function useEpicDetails(epicId?: string, enabled = true) {
  return useQuery({
    queryKey: ["epics", epicId],
    queryFn: () => getEpic(epicId ?? ""),
    enabled: Boolean(epicId) && enabled,
    refetchInterval: 5_000
  });
}
