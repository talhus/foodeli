"use client";

import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useCartStore } from "@/store/useCartStore";
import { AlertTriangle } from "lucide-react";

export function ConflictModal() {
  const { conflict, resolveConflict, closeConflict, cart } = useCartStore();

  if (!conflict.isOpen) return null;

  return (
    <Dialog open={conflict.isOpen} onOpenChange={(open) => !open && closeConflict()}>
      <DialogHeader>
        <div className="flex items-center gap-2 text-amber-600 mb-2">
          <AlertTriangle className="h-5 w-5" />
          <DialogTitle className="text-lg">Different Store in Cart</DialogTitle>
        </div>
        <DialogDescription>
          Your cart currently contains items from{" "}
          <strong className="text-foreground">{cart?.store_name || "another restaurant"}</strong>.
          Foodeli only supports ordering from one restaurant at a time.
        </DialogDescription>
      </DialogHeader>

      <p className="text-sm text-muted-foreground my-2">
        Do you want to clear your current cart and start a new order with this item?
      </p>

      <DialogFooter className="gap-2 sm:gap-0">
        <Button variant="outline" onClick={() => resolveConflict(false)}>
          Keep Existing Cart
        </Button>
        <Button variant="destructive" onClick={() => resolveConflict(true)}>
          Clear & Replace Cart
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
