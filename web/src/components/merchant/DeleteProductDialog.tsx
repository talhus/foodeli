import { Product } from "@/types/api";
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";

interface DeleteProductDialogProps {
  product: Product | null;
  isDeleting: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function DeleteProductDialog({
  product,
  isDeleting,
  onConfirm,
  onCancel,
}: DeleteProductDialogProps) {
  return (
    <Dialog open={!!product} onOpenChange={(open) => !open && onCancel()}>
      <DialogHeader>
        <DialogTitle>Remove Dish from Menu?</DialogTitle>
        <DialogDescription>
          Are you sure you want to remove <strong>{product?.name}</strong> from your active catalog? Customers will no longer be able to order this item.
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" onClick={onCancel} disabled={isDeleting}>
          Cancel
        </Button>
        <Button variant="destructive" onClick={onConfirm} disabled={isDeleting}>
          {isDeleting ? (
            <span className="flex items-center gap-1.5">
              <Loader2 className="h-4 w-4 animate-spin" />
              Deleting...
            </span>
          ) : (
            "Yes, Delete Dish"
          )}
        </Button>
      </DialogFooter>
    </Dialog>
  );
}
