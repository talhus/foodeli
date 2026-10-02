"use client";

import { useRouter } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { ClipboardList } from "lucide-react";

export function OrdersEmptyState() {
  const router = useRouter();

  return (
    <Card className="p-12 text-center text-muted-foreground space-y-3">
      <ClipboardList className="h-12 w-12 mx-auto opacity-40 text-primary" />
      <h3 className="font-semibold text-foreground text-lg">No orders found</h3>
      <p className="text-sm max-w-sm mx-auto">
        You haven&apos;t placed any orders yet. Discover your favorite dishes and place your first order!
      </p>
      <Button onClick={() => router.push("/")} size="sm">
        Browse Restaurants
      </Button>
    </Card>
  );
}
