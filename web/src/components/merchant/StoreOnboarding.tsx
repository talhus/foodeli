import { useState } from "react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Building2, Loader2 } from "lucide-react";
import { toast } from "sonner";

interface StoreOnboardingProps {
  onCreateStore: (name: string, address: string) => Promise<void>;
}

export function StoreOnboarding({ onCreateStore }: StoreOnboardingProps) {
  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || !address.trim()) {
      toast.error("Please provide both store name and address");
      return;
    }
    setIsSubmitting(true);
    try {
      await onCreateStore(name.trim(), address.trim());
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto pt-8">
      <Card className="shadow-lg border-primary/20">
        <CardHeader className="text-center pb-4">
          <div className="h-16 w-16 mx-auto rounded-2xl bg-primary/10 text-primary flex items-center justify-center mb-2">
            <Building2 className="h-8 w-8" />
          </div>
          <CardTitle className="text-2xl font-bold">Open Your Merchant Storefront</CardTitle>
          <CardDescription className="text-sm">
            Register your restaurant to start receiving customer orders and publishing your menu.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Storefront Name
              </label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Bella Italia Trattoria"
                required
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-1.5">
              <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                Physical Address
              </label>
              <Input
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                placeholder="e.g. Singel 380, 1016 AK Amsterdam"
                required
                disabled={isSubmitting}
              />
            </div>
            <Button type="submit" className="w-full mt-2 h-11 text-base font-semibold" disabled={isSubmitting}>
              {isSubmitting ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="h-5 w-5 animate-spin" />
                  Registering Storefront...
                </span>
              ) : (
                "Launch Restaurant Dashboard"
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
