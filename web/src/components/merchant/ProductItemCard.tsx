import { Product } from "@/types/api";
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/utils";
import { Pencil, Trash2 } from "lucide-react";

interface ProductItemCardProps {
  product: Product;
  onEdit: (product: Product) => void;
  onDelete: (product: Product) => void;
}

export function ProductItemCard({ product, onEdit, onDelete }: ProductItemCardProps) {
  const isOutOfStock = product.stock <= 0;
  const isLowStock = product.stock > 0 && product.stock <= 3;

  return (
    <Card className="flex flex-col justify-between overflow-hidden shadow-sm hover:shadow transition-shadow">
      <CardHeader className="p-5 pb-3">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base font-semibold">{product.name}</CardTitle>
          <Badge
            variant={isOutOfStock ? "destructive" : isLowStock ? "warning" : "secondary"}
            className="text-[10px] shrink-0"
          >
            {isOutOfStock ? "Out of Stock" : isLowStock ? `${product.stock} Left` : `${product.stock} in stock`}
          </Badge>
        </div>
        <CardDescription className="text-xs line-clamp-2 mt-1">{product.description}</CardDescription>
      </CardHeader>

      <CardFooter className="p-5 pt-0 border-t bg-muted/10 flex items-center justify-between mt-3 pt-3">
        <div>
          <span className="text-[10px] text-muted-foreground block uppercase font-medium">Price</span>
          <span className="text-base font-bold text-foreground">{formatCurrency(product.price)}</span>
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onEdit(product)}
            className="h-8 w-8 text-muted-foreground hover:text-foreground"
            title="Edit Dish"
          >
            <Pencil className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => onDelete(product)}
            className="h-8 w-8 text-muted-foreground hover:text-destructive"
            title="Delete Dish"
          >
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </CardFooter>
    </Card>
  );
}
