import { useState, useEffect } from "react";
import { Product } from "@/types/api";
import { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

interface ProductFormModalProps {
  open: boolean;
  editingProduct: Product | null;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: { name: string; description: string; price: number; stock: number }) => Promise<void>;
}

export function ProductFormModal({
  open,
  editingProduct,
  onOpenChange,
  onSubmit,
}: ProductFormModalProps) {
  const [name, setName] = useState("");
  const [desc, setDesc] = useState("");
  const [price, setPrice] = useState("");
  const [stock, setStock] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (editingProduct) {
      setName(editingProduct.name);
      setDesc(editingProduct.description);
      setPrice((editingProduct.price / 100).toFixed(2));
      setStock(editingProduct.stock.toString());
    } else {
      setName("");
      setDesc("");
      setPrice("");
      setStock("20");
    }
  }, [editingProduct, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const priceFloat = parseFloat(price);
    const stockInt = parseInt(stock, 10);

    if (isNaN(priceFloat) || priceFloat <= 0) {
      toast.error("Please enter a valid price greater than 0");
      return;
    }
    if (isNaN(stockInt) || stockInt < 0) {
      toast.error("Please enter a valid stock quantity");
      return;
    }

    setIsSubmitting(true);
    try {
      await onSubmit({
        name: name.trim(),
        description: desc.trim(),
        price: Math.round(priceFloat * 100),
        stock: stockInt,
      });
      onOpenChange(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogHeader>
        <DialogTitle>{editingProduct ? "Edit Menu Dish" : "Add New Dish to Menu"}</DialogTitle>
        <DialogDescription>
          {editingProduct ? "Update dish details, inventory, and pricing." : "Create a new dish to showcase in your restaurant menu."}
        </DialogDescription>
      </DialogHeader>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Dish Name</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Truffle Mushroom Risotto"
            required
            minLength={2}
            maxLength={100}
            disabled={isSubmitting}
          />
        </div>

        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Description</label>
          <textarea
            value={desc}
            onChange={(e) => setDesc(e.target.value)}
            placeholder="e.g. Creamy arborio rice with wild mushrooms, truffle oil, and aged parmesan"
            required
            minLength={2}
            maxLength={1000}
            rows={3}
            disabled={isSubmitting}
            className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Price (&euro;)</label>
            <Input
              type="number"
              step="0.01"
              min="0.10"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              placeholder="14.50"
              required
              disabled={isSubmitting}
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Portions Stock</label>
            <Input
              type="number"
              step="1"
              min="0"
              value={stock}
              onChange={(e) => setStock(e.target.value)}
              placeholder="25"
              required
              disabled={isSubmitting}
            />
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? (
              <span className="flex items-center gap-1.5">
                <Loader2 className="h-4 w-4 animate-spin" />
                Saving...
              </span>
            ) : editingProduct ? (
              "Update Dish"
            ) : (
              "Add Dish"
            )}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
