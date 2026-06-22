import { api } from "./api";
import type { Invoice, PaymentScenario, PayResult } from "./types";

export function listInvoices(): Promise<Invoice[]> {
  return api<{ invoices: Invoice[] }>("/invoices").then((r) => r.invoices);
}

export function payInvoice(
  id: string,
  scenario: PaymentScenario,
): Promise<PayResult> {
  return api<PayResult>(`/invoices/${id}/pay`, {
    method: "POST",
    body: JSON.stringify({ scenario }),
  });
}
