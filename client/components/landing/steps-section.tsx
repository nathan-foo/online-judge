import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import t from "@/lib/locales/en_US.json";

export function StepsSection() {
  return (
    <section className="mx-auto flex w-full max-w-5xl flex-col gap-10 px-4 py-16 lg:py-24">
      <div className="flex flex-col items-center gap-3 text-center">
        <Badge variant="outline">{t.steps.badge}</Badge>
        <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {t.steps.title}
        </h2>
        <p className="max-w-lg text-sm text-muted-foreground">
          {t.steps.description}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        {t.steps.items.map((step) => (
          <Card key={step.step}>
            <CardHeader>
              <span className="font-mono text-xs font-medium text-muted-foreground">
                Step {step.step}
              </span>
              <CardTitle>{step.title}</CardTitle>
              <CardDescription>{step.description}</CardDescription>
            </CardHeader>
          </Card>
        ))}
      </div>
    </section>
  );
}
