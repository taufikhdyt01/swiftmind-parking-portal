"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";

import { AppHeader } from "@/components/app-header";
import { PageLoader } from "@/components/page-loader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useRequireAuth } from "@/hooks/use-require-auth";
import { formatDateTime, formatDateTimeUTC, formatIDR } from "@/lib/format";
import { labelForType } from "@/lib/types";
import { getViolation } from "@/lib/violations";

export default function ViolationDetailPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const { user, ready } = useRequireAuth();

  const query = useQuery({
    queryKey: ["violations", params.id],
    queryFn: () => getViolation(params.id),
    enabled: ready,
    retry: false,
  });

  if (!ready || !user) return <PageLoader />;

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-3xl flex-1 px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <h1 className="text-xl font-semibold">Violation detail</h1>
          <Button variant="ghost" size="sm" onClick={() => router.push("/violations")}>
            ← Back
          </Button>
        </div>

        {query.isPending ? (
          <p className="text-muted-foreground text-sm">Loading…</p>
        ) : query.isError || !query.data ? (
          <Card>
            <CardContent className="text-muted-foreground py-10 text-center text-sm">
              Violation not found.
            </CardContent>
          </Card>
        ) : (
          (() => {
            const v = query.data;
            return (
              <div className="space-y-6">
                <div className="flex flex-wrap items-center gap-3">
                  <span className="text-lg font-semibold">{v.plate}</span>
                  <Badge variant="secondary">{labelForType(v.violation_type)}</Badge>
                  {v.payment_status === "paid" ? (
                    <Badge>Paid</Badge>
                  ) : (
                    <Badge variant="outline">Unpaid</Badge>
                  )}
                </div>

                {v.photo_url && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img
                    src={v.photo_url}
                    alt={`Violation ${v.plate}`}
                    className="max-h-80 w-full rounded-lg border object-contain"
                  />
                )}

                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Fine calculation</CardTitle>
                    <CardDescription>
                      Snapshotted from rule version v{v.rule_version} at issue time —
                      a later rule change does not affect it.
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    <Row label="Base amount" value={formatIDR(v.base_amount)} />
                    <Row label="Time multiplier" value={`× ${v.time_multiplier}`} />
                    <Row
                      label={`Repeat multiplier (${v.prior_unpaid_count} prior unpaid)`}
                      value={`× ${v.repeat_multiplier}`}
                    />
                    <div className="mt-2 flex items-center justify-between border-t pt-2">
                      <span className="font-medium">Final fine</span>
                      <span className="text-base font-semibold">
                        {formatIDR(v.final_amount)}
                      </span>
                    </div>
                    <p className="text-muted-foreground pt-1 text-xs">
                      Applied rule version: v{v.rule_version}
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Details</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-2 text-sm">
                    <Row label="Location" value={v.location} />
                    <Row label="Occurred at" value={formatDateTimeUTC(v.occurred_at)} />
                    <Row label="Issued by" value={v.issued_by_email} />
                    <Row label="Plate owner" value={v.owner_email || "—"} />
                    <Row label="Recorded at" value={formatDateTime(v.created_at)} />
                  </CardContent>
                </Card>
              </div>
            );
          })()
        )}
      </main>
    </>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <span className="text-muted-foreground">{label}</span>
      <span className="text-right">{value}</span>
    </div>
  );
}
