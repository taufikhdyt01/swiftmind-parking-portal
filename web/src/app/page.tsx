"use client";

import Link from "next/link";

import { AppHeader } from "@/components/app-header";
import { PageLoader } from "@/components/page-loader";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useRequireAuth } from "@/hooks/use-require-auth";

type RoleCard = { title: string; desc: string; href: string };

// Capabilities per role; each links to its flow.
const ROLE_CARDS: Record<string, RoleCard[]> = {
  officer: [
    { title: "Submit violation", desc: "Record a parking violation with photo and location.", href: "/violations/new" },
    { title: "Fine rules", desc: "View and publish new fine-rule versions.", href: "/rules" },
    { title: "All violations", desc: "Browse every issued violation and its applied rule version.", href: "/violations" },
  ],
  member: [
    { title: "Pay a fine", desc: "Settle an outstanding fine via the payment provider.", href: "/pay" },
    { title: "My violations & history", desc: "Every violation on your plates with its fine, the rule version applied, and payment status.", href: "/violations" },
  ],
};

export default function HomePage() {
  const { user, ready } = useRequireAuth();
  if (!ready || !user) return <PageLoader />;

  const cards = ROLE_CARDS[user.role] ?? [];

  return (
    <>
      <AppHeader />
      <main className="mx-auto w-full max-w-5xl flex-1 px-6 py-8">
        <h1 className="text-xl font-semibold">
          Welcome, {user.name.split(" ")[0]}
        </h1>
        <p className="text-muted-foreground mt-1 text-sm">
          You are signed in as a{" "}
          <span className="text-foreground font-medium capitalize">
            {user.role}
          </span>
          .
        </p>

        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {cards.map((card) => (
            <Link key={card.title} href={card.href}>
              <Card className="hover:border-primary h-full cursor-pointer transition-colors">
                <CardHeader>
                  <CardTitle className="text-base">{card.title}</CardTitle>
                  <CardDescription>{card.desc}</CardDescription>
                </CardHeader>
                <div className="px-6 pb-4">
                  <Badge>Open</Badge>
                </div>
              </Card>
            </Link>
          ))}
        </div>
      </main>
    </>
  );
}
