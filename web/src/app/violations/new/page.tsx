"use client";

import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { AppHeader } from "@/components/app-header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuth } from "@/contexts/auth-context";
import { formatIDR } from "@/lib/format";
import { createViolation } from "@/lib/violations";
import { VIOLATION_TYPES, VIOLATION_TYPE_LABELS } from "@/lib/types";

// Default the timestamp to "now" in UTC (YYYY-MM-DDTHH:mm). Violation times are
// treated as UTC wall-clock end to end (entry, pricing, and display).
function nowUTC(): string {
  return new Date().toISOString().slice(0, 16);
}

export default function NewViolationPage() {
  const router = useRouter();
  const { user, loading } = useAuth();

  const [plate, setPlate] = useState("B1234ABC");
  const [type, setType] = useState<string>(VIOLATION_TYPES[0]);
  const [location, setLocation] = useState("");
  const [occurredAt, setOccurredAt] = useState(nowUTC());
  const [photo, setPhoto] = useState<File | null>(null);

  const submit = useMutation({
    mutationFn: createViolation,
    onSuccess: (v) => {
      toast.success(
        `Violation issued — fine ${formatIDR(v.final_amount)} (rule v${v.rule_version})`,
      );
      router.push("/violations");
    },
    onError: (e) =>
      toast.error(e instanceof Error ? e.message : "Could not submit violation"),
  });

  useEffect(() => {
    if (!loading && (!user || user.role !== "officer")) router.replace("/");
  }, [loading, user, router]);

  if (loading || !user || user.role !== "officer") {
    return (
      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </main>
    );
  }

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-xl flex-1 px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-xl font-semibold">Submit a violation</h1>
          <Button variant="ghost" size="sm" onClick={() => router.push("/violations")}>
            View all →
          </Button>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Violation details</CardTitle>
            <CardDescription>
              The fine is calculated from the active ruleset and snapshotted now.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                submit.mutate({
                  plate: plate.trim().toUpperCase(),
                  violation_type: type,
                  location: location.trim(),
                  occurred_at: occurredAt,
                  photo,
                });
              }}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="plate">License plate</Label>
                <Input
                  id="plate"
                  required
                  value={plate}
                  onChange={(e) => setPlate(e.target.value)}
                  placeholder="B1234ABC"
                />
                <p className="text-muted-foreground text-xs">
                  Demo: B1234ABC belongs to the member account.
                </p>
              </div>

              <div className="space-y-2">
                <Label>Violation type</Label>
                <Select value={type} onValueChange={(v) => v && setType(v)}>
                  <SelectTrigger>
                    <SelectValue>
                      {(value: string) =>
                        VIOLATION_TYPE_LABELS[
                          value as keyof typeof VIOLATION_TYPE_LABELS
                        ] ?? value
                      }
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {VIOLATION_TYPES.map((t) => (
                      <SelectItem key={t} value={t}>
                        {VIOLATION_TYPE_LABELS[t]}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="location">Location</Label>
                <Input
                  id="location"
                  required
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  placeholder="Jl. Sudirman No. 1"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="occurred_at">Date &amp; time (UTC)</Label>
                <Input
                  id="occurred_at"
                  type="datetime-local"
                  required
                  value={occurredAt}
                  onChange={(e) => setOccurredAt(e.target.value)}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="photo">Photo</Label>
                <Input
                  id="photo"
                  type="file"
                  accept="image/*"
                  onChange={(e) => setPhoto(e.target.files?.[0] ?? null)}
                />
              </div>

              <Button type="submit" disabled={submit.isPending} className="w-full">
                {submit.isPending ? "Submitting…" : "Submit violation"}
              </Button>
            </form>
          </CardContent>
        </Card>
      </main>
    </>
  );
}
