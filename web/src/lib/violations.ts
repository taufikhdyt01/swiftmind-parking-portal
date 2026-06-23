import { api } from "./api";
import type { Violation } from "./types";

export function listViolations(): Promise<Violation[]> {
  return api<{ violations: Violation[] }>("/violations").then((r) => r.violations);
}

export function getViolation(id: string): Promise<Violation> {
  return api<Violation>(`/violations/${id}`);
}

export interface CreateViolationInput {
  plate: string;
  violation_type: string;
  location: string;
  occurred_at: string;
  photo: File | null;
}

export function createViolation(input: CreateViolationInput): Promise<Violation> {
  const form = new FormData();
  form.append("plate", input.plate);
  form.append("violation_type", input.violation_type);
  form.append("location", input.location);
  form.append("occurred_at", input.occurred_at);
  if (input.photo) form.append("photo", input.photo);

  return api<Violation>("/violations", { method: "POST", body: form });
}
