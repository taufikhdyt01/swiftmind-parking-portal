import { api } from "./api";
import type { Notification } from "./types";

export function listNotifications(): Promise<Notification[]> {
  return api<{ notifications: Notification[] }>("/notifications").then(
    (r) => r.notifications,
  );
}
