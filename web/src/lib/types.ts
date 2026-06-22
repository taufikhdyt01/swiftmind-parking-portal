export type Role = "officer" | "member";

export interface User {
  id: string;
  email: string;
  name: string;
  role: Role;
}
