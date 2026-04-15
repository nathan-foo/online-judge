import type { LucideIcon } from "lucide-react";
import {
  Users,
  BookOpen,
  Timer,
  BarChart3,
  Trophy,
  ShieldCheck,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import t from "@/lib/locales/en_US.json";

const FEATURE_ICONS: LucideIcon[] = [
  Users,
  BookOpen,
  Timer,
  BarChart3,
  Trophy,
  ShieldCheck,
];

export function FeaturesSection() {
  return (
    <section className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-4 py-16 lg:py-24">
      <div className="flex flex-col items-center gap-3 text-center">
        <Badge variant="outline">{t.features.badge}</Badge>
        <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {t.features.title}
        </h2>
        <p className="max-w-lg text-sm text-muted-foreground">
          {t.features.description}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {t.features.items.map((feature, i) => {
          const Icon = FEATURE_ICONS[i];

          return (
            <Card key={feature.title}>
              <CardHeader>
                <div className="flex size-9 items-center justify-center rounded-lg border bg-muted/50">
                  <Icon className="size-4 text-foreground" />
                </div>
                <CardTitle className="text-sm">{feature.title}</CardTitle>
                <CardDescription>{feature.description}</CardDescription>
              </CardHeader>
            </Card>
          );
        })}
      </div>
    </section>
  );
}
