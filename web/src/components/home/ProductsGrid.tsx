import { Store, Product } from "@/types/api";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Utensils, ShoppingBag, Loader2 } from "lucide-react";
import { ProductCatalogCard } from "./ProductCatalogCard";

interface ProductsGridProps {
  store: Store;
  products: Product[];
  isLoading: boolean;
  isCartLoading: boolean;
  onAddToCart: (product: Product) => void;
}

export function ProductsGrid({ store, products, isLoading, isCartLoading, onAddToCart }: ProductsGridProps) {
  return (
    <section className="space-y-5 pt-4 border-t">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <div className="flex items-center gap-2">
            <Utensils className="h-5 w-5 text-primary" />
            <h2 className="text-2xl font-bold tracking-tight">{store.name} &bull; Catalog</h2>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            Freshly prepared items from {store.address}
          </p>
        </div>
        <Badge variant="secondary" className="w-fit text-xs">
          {products.length} Products
        </Badge>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center p-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : products.length === 0 ? (
        <Card className="p-12 text-center text-muted-foreground space-y-2">
          <ShoppingBag className="h-10 w-10 mx-auto opacity-50" />
          <h3 className="font-semibold text-foreground">Menu is currently being updated</h3>
          <p className="text-xs">This merchant has not listed any dishes yet.</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
          {products.map((product) => (
            <ProductCatalogCard
              key={product.id}
              product={product}
              isCartLoading={isCartLoading}
              onAddToCart={onAddToCart}
            />
          ))}
        </div>
      )}
    </section>
  );
}
