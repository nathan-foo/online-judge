import Link from "next/link";
import { SignUpButton, Show } from "@clerk/nextjs";
import { ArrowRight } from "lucide-react";

import { Button } from "@/components/ui/button";
import t from "@/lib/locales/en_US.json";

export function CtaSection() {
  return (
    <section className="mx-auto flex w-full max-w-5xl flex-col items-center gap-6 px-4 py-16 text-center lg:py-24">
      <h2 className="text-2xl font-bold tracking-tight sm:text-3xl">
        {t.cta.title}
      </h2>
      <p className="max-w-md text-sm leading-relaxed text-muted-foreground">
        {t.cta.description}
      </p>

      <div className="flex items-center gap-3 pt-2">
        <Show when="signed-out">
          <SignUpButton>
            <Button size="lg">
              {t.cta.signUp}
              <ArrowRight className="size-4" />
            </Button>
          </SignUpButton>
        </Show>
        <Show when="signed-in">
          <Button size="lg" asChild>
            <Link href="/explore">
              {t.cta.dashboard}
              <ArrowRight className="size-4" />
            </Link>
          </Button>
        </Show>
        <Button variant="outline" size="lg" asChild>
          <Link href="/pricing">{t.cta.viewPricing}</Link>
        </Button>
      </div>
    </section>
  );
}
