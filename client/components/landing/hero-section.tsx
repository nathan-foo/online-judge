import Link from "next/link";
import { SignUpButton, Show } from "@clerk/nextjs";
import { ArrowRight, Code } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import t from "@/lib/locales/en_US.json";

export function HeroSection() {
  const { codePreview } = t.hero;

  return (
    <section className="mx-auto grid w-full max-w-5xl items-center gap-8 px-4 py-16 lg:grid-cols-2 lg:gap-12 lg:py-24">
      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-2">
          <Badge variant="outline">{t.hero.badge}</Badge>
          <Separator orientation="vertical" className="h-4" />
          <span className="text-sm text-muted-foreground">
            {t.hero.tagline}
          </span>
        </div>

        <h1 className="text-3xl font-bold leading-tight tracking-tight sm:text-4xl lg:text-5xl">
          {t.hero.title}
        </h1>

        <p className="max-w-md text-base leading-relaxed text-muted-foreground">
          {t.hero.description}
        </p>

        <div className="flex items-center gap-3">
          <Show when="signed-out">
            <SignUpButton>
              <Button size="lg">
                {t.hero.getStarted}
                <ArrowRight className="size-4" />
              </Button>
            </SignUpButton>
          </Show>
          <Show when="signed-in">
            <Button size="lg" asChild>
              <Link href="/explore">
                {t.hero.dashboard}
                <ArrowRight className="size-4" />
              </Link>
            </Button>
          </Show>
          <Button variant="outline" size="lg" asChild>
            <Link href="/pricing">{t.hero.viewPricing}</Link>
          </Button>
        </div>
      </div>

      <div className="overflow-hidden rounded-xl border bg-card shadow-sm">
        <div className="flex items-center gap-2 border-b bg-muted/50 px-4 py-2.5">
          <Code className="size-3.5 text-muted-foreground" />
          <span className="font-mono text-xs text-muted-foreground">
            {codePreview.filename}
          </span>
          <Badge variant="secondary" className="ml-auto text-[10px]">
            {codePreview.badge}
          </Badge>
        </div>
        <pre className="overflow-x-auto p-4 font-mono text-[13px] leading-6">
          {codePreview.code}
        </pre>
      </div>
    </section>
  );
}
