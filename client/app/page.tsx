import { Separator } from "@/components/ui/separator";
import { HeroSection } from "@/components/landing/hero-section";
import { FeaturesSection } from "@/components/landing/features-section";
import { StepsSection } from "@/components/landing/steps-section";
import { AuthoringSection } from "@/components/landing/authoring-section";
import { CtaSection } from "@/components/landing/cta-section";

export default function Home() {
  return (
    <main className="flex-1 pt-24 pb-20">
      <HeroSection />
      <Separator className="mx-auto max-w-5xl" />
      <FeaturesSection />
      <Separator className="mx-auto max-w-5xl" />
      <StepsSection />
      <Separator className="mx-auto max-w-5xl" />
      <AuthoringSection />
      <Separator className="mx-auto max-w-5xl" />
      <CtaSection />
    </main>
  );
}
