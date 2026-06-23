"use client";

import { useQuery } from "@tanstack/react-query";
import { Bell } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import { formatDateTime } from "@/lib/format";
import { listNotifications } from "@/lib/notifications";

export function NotificationBell() {
  const { data } = useQuery({
    queryKey: ["notifications"],
    queryFn: listNotifications,
    refetchInterval: 15_000,
  });
  const items = data ?? [];

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        aria-label="Notifications"
        className={cn(buttonVariants({ variant: "outline", size: "icon" }), "relative")}
      >
        <Bell className="h-[1.2rem] w-[1.2rem]" />
        {items.length > 0 && (
          <span className="bg-primary text-primary-foreground absolute -top-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-medium">
            {items.length > 9 ? "9+" : items.length}
          </span>
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-80 p-0">
        <div className="border-b px-3 py-2 text-sm font-medium">
          Notifications
        </div>
        {items.length === 0 ? (
          <p className="text-muted-foreground px-3 py-6 text-center text-sm">
            No notifications yet.
          </p>
        ) : (
          <div className="max-h-80 overflow-y-auto">
            {items.slice(0, 15).map((n) => (
              <div key={n.id} className="border-b px-3 py-2 last:border-0">
                <p className="text-sm">{n.message}</p>
                <p className="text-muted-foreground mt-0.5 text-xs">
                  {formatDateTime(n.created_at)}
                </p>
              </div>
            ))}
          </div>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
