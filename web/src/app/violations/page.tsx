"use client";

import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { AppHeader } from "@/components/app-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuth } from "@/contexts/auth-context";
import { formatDateTimeUTC, formatIDR } from "@/lib/format";
import { listViolations } from "@/lib/violations";
import { VIOLATION_TYPE_LABELS, type ViolationType } from "@/lib/types";

export default function ViolationsPage() {
  const router = useRouter();
  const { user, loading } = useAuth();

  const violations = useQuery({
    queryKey: ["violations"],
    queryFn: listViolations,
    enabled: !!user,
  });

  useEffect(() => {
    if (!loading && !user) router.replace("/login");
  }, [loading, user, router]);

  if (loading || !user) {
    return (
      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </main>
    );
  }

  const isOfficer = user.role === "officer";

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">
              {isOfficer ? "All violations" : "My violations"}
            </h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Each fine shows the rule version applied when it was issued.
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" onClick={() => router.push("/")}>
              ← Dashboard
            </Button>
            {isOfficer && (
              <Button size="sm" onClick={() => router.push("/violations/new")}>
                Submit violation
              </Button>
            )}
          </div>
        </div>

        <Card>
          <CardContent className="pt-6">
            {violations.data && violations.data.length === 0 ? (
              <p className="text-muted-foreground py-8 text-center text-sm">
                No violations yet.
              </p>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Plate</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>When</TableHead>
                      <TableHead>Fine</TableHead>
                      <TableHead>Rule</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Photo</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {violations.data?.map((v) => (
                      <TableRow key={v.id}>
                        <TableCell className="font-medium">{v.plate}</TableCell>
                        <TableCell>
                          {VIOLATION_TYPE_LABELS[v.violation_type as ViolationType] ??
                            v.violation_type}
                          <div className="text-muted-foreground text-xs">
                            {v.location}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm">
                          {formatDateTimeUTC(v.occurred_at)}
                        </TableCell>
                        <TableCell>
                          <div className="font-medium">
                            {formatIDR(v.final_amount)}
                          </div>
                          <div className="text-muted-foreground text-xs">
                            {formatIDR(v.base_amount)} × {v.time_multiplier} ×{" "}
                            {v.repeat_multiplier}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">v{v.rule_version}</Badge>
                        </TableCell>
                        <TableCell>
                          {v.payment_status === "paid" ? (
                            <Badge>Paid</Badge>
                          ) : (
                            <Badge variant="secondary">Unpaid</Badge>
                          )}
                        </TableCell>
                        <TableCell>
                          {v.photo_url ? (
                            <a
                              href={v.photo_url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-sm underline"
                            >
                              View
                            </a>
                          ) : (
                            <span className="text-muted-foreground text-sm">—</span>
                          )}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </>
  );
}
