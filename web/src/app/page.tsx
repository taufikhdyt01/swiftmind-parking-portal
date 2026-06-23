"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { AppHeader } from "@/components/app-header";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/contexts/auth-context";

type RoleCard = { title: string; desc: string; href?: string };

// Capabilities per role. Cards with an href are live; the rest are placeholders
// for flows built in later phases.
const ROLE_CARDS: Record<string, RoleCard[]> = {
  officer: [
    { title: "Submit violation", desc: "Record a parking violation with photo and location.", href: "/violations/new" },
    { title: "Fine rules", desc: "View and publish new fine-rule versions.", href: "/rules" },
    { title: "All violations", desc: "Browse every issued violation and its applied rule version.", href: "/violations" },
  ],
  member: [
    { title: "My violations", desc: "See violations issued against your plates.", href: "/violations" },
    { title: "Pay a fine", desc: "Settle an outstanding fine via the payment provider.", href: "/pay" },
    { title: "My history", desc: "Review past violations, fines, and payment status.", href: "/violations" },
  ],
};

export default function HomePage() {
  const router = useRouter();
  const { user, loading } = useAuth();

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
          {cards.map((card) => {
            const inner = (
              <Card
                className={
                  card.href
                    ? "hover:border-primary h-full cursor-pointer transition-colors"
                    : "h-full"
                }
              >
                <CardHeader>
                  <CardTitle className="text-base">{card.title}</CardTitle>
                  <CardDescription>{card.desc}</CardDescription>
                </CardHeader>
                <div className="px-6 pb-4">
                  {card.href ? (
                    <Badge>Open</Badge>
                  ) : (
                    <Badge variant="outline">Coming in a later phase</Badge>
                  )}
                </div>
              </Card>
            );
            return card.href ? (
              <Link key={card.title} href={card.href}>
                {inner}
              </Link>
            ) : (
              <div key={card.title}>{inner}</div>
            );
          })}
        </div>
      </main>
    </>
  );
}
