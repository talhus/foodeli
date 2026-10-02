"use client";

import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";

export function OrdersHeader() {
  const router = useRouter();

  return (
    <div className="space-y-1">
      <Button
        variant="ghost"
        size="sm"
        onClick={() => router.push("/")}
        className="flex items-center gap-1 text-muted-foreground p-0 h-auto hover:bg-transparent hover:text-foreground mb-1"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Marketplace</span>
      </Button>
      <h1 className="text-2xl font-bold tracking-tight">Your Order History</h1>
      <p className="text-xs text-muted-foreground">Review your past and active food delivery orders</p>
    </div>
  );
}
