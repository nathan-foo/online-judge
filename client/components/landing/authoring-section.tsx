import Link from "next/link";
import { SignUpButton, Show } from "@clerk/nextjs";
import { ArrowRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "@/components/ui/card";
import t from "@/lib/locales/en_US.json";

export function AuthoringSection() {
  const { sessionConfig } = t.authoring;

  return (
    <section className="mx-auto grid w-full max-w-5xl items-center gap-8 px-4 py-16 lg:grid-cols-2 lg:gap-12 lg:py-24">
      <div className="flex flex-col gap-6">
        <Badge variant="outline" className="w-fit">
          {t.authoring.badge}
        </Badge>
        <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
          {t.authoring.title}
        </h2>
        <p className="text-sm leading-relaxed text-muted-foreground">
          {t.authoring.description}
        </p>
        <div className="flex items-center gap-3">
          <Show when="signed-out">
            <SignUpButton>
              <Button>
                {t.authoring.startAuthoring}
                <ArrowRight className="size-4" />
              </Button>
            </SignUpButton>
          </Show>
          <Show when="signed-in">
            <Button asChild>
              <Link href="/problems">
                {t.authoring.createProblem}
                <ArrowRight className="size-4" />
              </Link>
            </Button>
          </Show>
          <Button variant="outline" asChild>
            <Link href="/explore">{t.authoring.browseExamples}</Link>
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm">{sessionConfig.title}</CardTitle>
          <CardDescription>{sessionConfig.description}</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-3 font-mono text-[13px]">
            {sessionConfig.fields.map((field, i) => (
              <div
                key={field.label}
                className={`flex justify-between${i < sessionConfig.fields.length - 1 ? " border-b pb-2" : ""}`}
              >
                <span className="text-muted-foreground">{field.label}</span>
                <span>{field.value}</span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
