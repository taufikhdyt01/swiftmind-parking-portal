"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams, useRouter } from "next/navigation";

import { PageLoader } from "@/components/page-loader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useRequireAuth } from "@/hooks/use-require-auth";
import { formatIDR } from "@/lib/format";
import { listInvoices } from "@/lib/invoices";
import { labelForType } from "@/lib/types";

export default function ReceiptPage() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const { ready } = useRequireAuth();

  const { data, isPending } = useQuery({
    queryKey: ["invoices"],
    queryFn: listInvoices,
    enabled: ready,
  });

  if (!ready) return <PageLoader />;

  const inv = data?.find((i) => i.id === params.id);

  return (
    <main className="mx-auto w-full max-w-md flex-1 px-6 py-10">
      <div className="mb-4 flex items-center justify-between print:hidden">
        <Button variant="ghost" size="sm" onClick={() => router.push("/pay")}>
          ← Back
        </Button>
        {inv?.status === "paid" && (
          <Button size="sm" onClick={() => window.print()}>
            Print / Save as PDF
          </Button>
        )}
      </div>

      {isPending ? (
        <p className="text-muted-foreground text-sm">Loading…</p>
      ) : !inv ? (
        <p className="text-muted-foreground text-sm">Receipt not found.</p>
      ) : inv.status !== "paid" ? (
        <p className="text-muted-foreground text-sm">
          This invoice has not been paid yet.
        </p>
      ) : (
        <div className="rounded-xl border bg-white p-8 text-slate-900 shadow-sm print:border-0 print:shadow-none">
          <div className="mb-6 text-center">
            <p className="text-lg font-bold tracking-tight">Swiftmind</p>
            <p className="text-xs tracking-widest text-slate-500 uppercase">
              Payment Receipt
            </p>
          </div>

          <div className="mb-6 text-center">
            <p className="text-xs text-slate-500">Amount paid</p>
            <p className="text-3xl font-bold">{formatIDR(inv.amount)}</p>
            <Badge className="mt-2">PAID</Badge>
          </div>

          <dl className="space-y-2 border-t pt-4 text-sm">
            <Line label="Plate" value={inv.plate} />
            <Line label="Violation" value={labelForType(inv.violation_type)} />
            <Line label="Invoice ID" value={inv.id} mono />
            <Line label="Transaction ID" value={inv.transaction_id ?? "—"} mono />
            <Line
              label="Generated"
              value={new Date().toLocaleString("en-GB", {
                dateStyle: "medium",
                timeStyle: "short",
              })}
            />
          </dl>

          <p className="mt-6 border-t pt-4 text-center text-xs text-slate-400">
            Payment provider is mocked for assignment review. Thank you.
          </p>
        </div>
      )}
    </main>
  );
}

function Line({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="text-slate-500">{label}</dt>
      <dd className={`text-right ${mono ? "font-mono text-xs" : ""}`}>{value}</dd>
    </div>
  );
}
