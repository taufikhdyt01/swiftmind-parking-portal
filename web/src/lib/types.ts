export type Role = "officer" | "member";

export interface User {
  id: string;
  email: string;
  name: string;
  role: Role;
}

export const VIOLATION_TYPES = [
  "expired_meter",
  "no_parking_zone",
  "blocking_hydrant",
  "disabled_spot",
] as const;

export type ViolationType = (typeof VIOLATION_TYPES)[number];

export const VIOLATION_TYPE_LABELS: Record<ViolationType, string> = {
  expired_meter: "Expired meter",
  no_parking_zone: "No-parking zone",
  blocking_hydrant: "Blocking hydrant",
  disabled_spot: "Disabled spot",
};

export interface Ruleset {
  base_amounts: Record<string, number>;
  time_multiplier: {
    day_start_hour: number;
    night_start_hour: number;
    day_multiplier: number;
    night_multiplier: number;
  };
  repeat_multiplier: {
    tiers: { min_prior_unpaid: number; multiplier: number }[];
  };
}

export interface RuleVersion {
  id: string;
  version: number;
  is_active: boolean;
  ruleset: Ruleset;
  created_by: string;
  created_at: string;
}

export type PaymentStatus = "unpaid" | "paid";

export type InvoiceStatus = "open" | "paid";
export type PaymentScenario = "success" | "failed";

export interface Invoice {
  id: string;
  violation_id: string;
  plate: string;
  violation_type: ViolationType;
  owner_email: string;
  amount: number;
  status: InvoiceStatus;
  created_at: string;
}

export interface PayResult {
  status: "paid" | "failed";
  transaction_id: string;
  invoice: Invoice;
}

export interface Notification {
  id: string;
  kind: string;
  message: string;
  created_at: string;
}

export interface Violation {
  id: string;
  plate: string;
  violation_type: ViolationType;
  location: string;
  occurred_at: string;
  photo_url?: string;
  owner_email: string;
  issued_by_email: string;
  rule_version: number;
  base_amount: number;
  time_multiplier: number;
  repeat_multiplier: number;
  prior_unpaid_count: number;
  final_amount: number;
  payment_status: PaymentStatus;
  created_at: string;
}
