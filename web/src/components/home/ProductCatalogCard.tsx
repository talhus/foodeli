import { Product } from "@/types/api";
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { formatCurrency } from "@/lib/utils";
import { Plus } from "lucide-react";

interface ProductCatalogCardProps {
  product: Product;
  isCartLoading: boolean;
  onAddToCart: (product: Product) => void;
}

export function ProductCatalogCard({ product, isCartLoading, onAddToCart }: ProductCatalogCardProps) {
  const isOutOfStock = product.stock <= 0;

  return (
    <Card className="flex flex-col justify-between hover:shadow-md transition-shadow">
      <CardHeader className="p-5 pb-3">
        <div className="flex items-start justify-between gap-2">
          <CardTitle className="text-base font-semibold">{product.name}</CardTitle>
          <Badge
            variant={isOutOfStock ? "destructive" : product.stock < 5 ? "warning" : "secondary"}
            className="text-[10px] shrink-0"
          >
            {isOutOfStock ? "Out of Stock" : `${product.stock} in stock`}
          </Badge>
        </div>
        <CardDescription className="text-xs line-clamp-2 mt-1">
          {product.description}
        </CardDescription>
      </CardHeader>

      <CardFooter className="p-5 pt-0 flex items-center justify-between border-t border-border/40 mt-3 pt-3">
        <div>
          <span className="text-xs text-muted-foreground block">Price</span>
          <span className="text-lg font-bold text-foreground">
            {formatCurrency(product.price)}
          </span>
        </div>

        <Button
          size="sm"
          disabled={isOutOfStock || isCartLoading}
          onClick={() => onAddToCart(product)}
          className="flex items-center gap-1.5 font-medium"
        >
          <Plus className="h-4 w-4" />
          <span>Add</span>
        </Button>
      </CardFooter>
    </Card>
  );
}
