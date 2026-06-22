// Format an integer amount of Indonesian rupiah.
export function formatIDR(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    maximumFractionDigits: 0,
  }).format(amount);
}

// Format an ISO timestamp as a readable local date-time.
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-GB", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

// Format a timestamp in UTC. Violation timestamps are entered as wall-clock time
// and priced as such, so they are displayed in UTC to stay consistent with the
// time multiplier that was applied (e.g. 23:30 → night).
export function formatDateTimeUTC(iso: string): string {
  return (
    new Date(iso).toLocaleString("en-GB", {
      dateStyle: "medium",
      timeStyle: "short",
      timeZone: "UTC",
    }) + " UTC"
  );
}
