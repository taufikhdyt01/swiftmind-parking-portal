"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { AppHeader } from "@/components/app-header";
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
import { useAuth } from "@/contexts/auth-context";
import { formatIDR } from "@/lib/format";
import { listInvoices, payInvoice } from "@/lib/invoices";
import {
  VIOLATION_TYPE_LABELS,
  type PaymentScenario,
  type ViolationType,
} from "@/lib/types";

export default function PayPage() {
  const router = useRouter();
  const { user, loading } = useAuth();
  const queryClient = useQueryClient();

  const invoices = useQuery({
    queryKey: ["invoices"],
    queryFn: listInvoices,
    enabled: !!user,
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

  useEffect(() => {
    if (!loading && (!user || user.role !== "member")) router.replace("/");
  }, [loading, user, router]);

  if (loading || !user || user.role !== "member") {
    return (
      <main className="flex flex-1 items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </main>
    );
  }

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
            {invoices.data && invoices.data.length === 0 ? (
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
                      <TableHead className="text-right">Pay</TableHead>
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
                          <TableCell>
                            {VIOLATION_TYPE_LABELS[
                              inv.violation_type as ViolationType
                            ] ?? inv.violation_type}
                          </TableCell>
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
                              <span className="text-muted-foreground flex justify-end text-sm">
                                —
                              </span>
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
