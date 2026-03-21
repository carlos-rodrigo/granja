"use client";

import { useQuery } from "@tanstack/react-query";

import { getWorkers } from "@/lib/api";

export function useWorkers() {
  return useQuery({
    queryKey: ["workers"],
    queryFn: getWorkers,
    refetchInterval: 5_000
  });
}
