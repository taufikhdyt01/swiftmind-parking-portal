"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { AppHeader } from "@/components/app-header";
import { PageLoader } from "@/components/page-loader";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useRequireAuth } from "@/hooks/use-require-auth";
import { formatIDR } from "@/lib/format";
import { listInvoices, payInvoice } from "@/lib/invoices";
import { labelForType, type PaymentScenario } from "@/lib/types";

export default function PayPage() {
  const router = useRouter();
  const { ready } = useRequireAuth("member");
  const queryClient = useQueryClient();

  const invoices = useQuery({
    queryKey: ["invoices"],
    queryFn: listInvoices,
    enabled: ready,
  });

  // Per-invoice chosen scenario (defaults to success).
  const [scenarios, setScenarios] = useState<Record<string, PaymentScenario>>({});

  const pay = useMutation({
    mutationFn: ({ id, scenario }: { id: string; scenario: PaymentScenario }) =>
      payInvoice(id, scenario),
    onSuccess: (res) => {
      if (res.status === "paid") {
        toast.success(`Payment successful — ${res.transaction_id}`);
      } else {
        toast.error(`Payment failed — ${res.transaction_id}`);
      }
      queryClient.invalidateQueries({ queryKey: ["invoices"] });
      queryClient.invalidateQueries({ queryKey: ["violations"] });
    },
    onError: (e) =>
      toast.error(e instanceof Error ? e.message : "Could not process payment"),
  });

  if (!ready) return <PageLoader />;

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-4xl flex-1 px-6 py-8">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-semibold">Pay a fine</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Choose a payment outcome to exercise both paths (the provider is mocked).
            </p>
          </div>
          <Button variant="ghost" size="sm" onClick={() => router.push("/")}>
            ← Dashboard
          </Button>
        </div>

        <Card>
          <CardContent className="pt-6">
            {invoices.isPending ? (
              <p className="text-muted-foreground py-8 text-center text-sm">
                Loading…
              </p>
            ) : invoices.data && invoices.data.length === 0 ? (
              <p className="text-muted-foreground py-8 text-center text-sm">
                No invoices yet.
              </p>
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Plate</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Amount</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">Payment</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {invoices.data?.map((inv) => {
                      const scenario = scenarios[inv.id] ?? "success";
                      const isPaying =
                        pay.isPending && pay.variables?.id === inv.id;
                      return (
                        <TableRow key={inv.id}>
                          <TableCell className="font-medium">{inv.plate}</TableCell>
                          <TableCell>{labelForType(inv.violation_type)}</TableCell>
                          <TableCell>{formatIDR(inv.amount)}</TableCell>
                          <TableCell>
                            {inv.status === "paid" ? (
                              <Badge>Paid</Badge>
                            ) : (
                              <Badge variant="secondary">Open</Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            {inv.status === "paid" ? (
                              <div className="text-muted-foreground text-right font-mono text-xs whitespace-nowrap">
                                {inv.transaction_id ?? "—"}
                              </div>
                            ) : (
                              <div className="flex items-center justify-end gap-2">
                                <Select
                                  value={scenario}
                                  onValueChange={(v) =>
                                    v &&
                                    setScenarios((s) => ({
                                      ...s,
                                      [inv.id]: v as PaymentScenario,
                                    }))
                                  }
                                >
                                  <SelectTrigger className="w-32" size="sm">
                                    <SelectValue>
                                      {(value: string) =>
                                        value === "success" ? "Success" : "Failed"
                                      }
                                    </SelectValue>
                                  </SelectTrigger>
                                  <SelectContent>
                                    <SelectItem value="success">Success</SelectItem>
                                    <SelectItem value="failed">Failed</SelectItem>
                                  </SelectContent>
                                </Select>
                                <Button
                                  size="sm"
                                  disabled={isPaying}
                                  onClick={() =>
                                    pay.mutate({ id: inv.id, scenario })
                                  }
                                >
                                  {isPaying ? "Paying…" : "Pay"}
                                </Button>
                              </div>
                            )}
                          </TableCell>
                        </TableRow>
                      );
                    })}
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
