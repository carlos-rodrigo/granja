import clsx, { type ClassValue } from "clsx";

export function cn(...parts: ClassValue[]) {
  return clsx(parts);
}

export function statusTone(status: string) {
  switch (status) {
    case "planted":
      return "bg-status-planted/20 text-status-planted border-status-planted/40";
    case "growing":
      return "bg-status-growing/20 text-status-growing border-status-growing/40";
    case "ready":
      return "bg-status-ready/20 text-status-ready border-status-ready/40";
    case "harvested":
      return "bg-status-harvested/20 text-status-harvested border-status-harvested/40";
    case "blocked":
      return "bg-status-blocked/20 text-status-blocked border-status-blocked/40";
    case "in_progress":
      return "bg-status-growing/20 text-status-growing border-status-growing/40";
    case "todo":
      return "bg-slate-500/20 text-slate-300 border-slate-500/40";
    case "done":
      return "bg-status-ready/20 text-status-ready border-status-ready/40";
    default:
      return "bg-slate-700/20 text-slate-300 border-slate-500/30";
  }
}
