import { Badge } from "@/components/ui/badge";
import { Sparkles } from "lucide-react";

export function HeroBanner() {
  return (
    <section className="relative rounded-2xl overflow-hidden bg-gradient-to-r from-amber-500/15 via-orange-500/10 to-red-500/5 p-8 sm:p-12 border border-primary/20">
      <div className="max-w-2xl space-y-4">
        <Badge variant="default" className="flex items-center gap-1.5 w-fit">
          <Sparkles className="h-3 w-3" />
          <span>Fast & Reliable Delivery</span>
        </Badge>
        <h1 className="text-3xl sm:text-5xl font-extrabold tracking-tight text-foreground">
          Flavor delivered to your doorstep.
        </h1>
        <p className="text-muted-foreground text-sm sm:text-base leading-relaxed">
          Order delicious artisan meals, marketplace groceries, and street food crafted by verified local merchants.
        </p>
      </div>
    </section>
  );
}
