import { Product } from "@/types/api";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { UtensilsCrossed, Plus, Search } from "lucide-react";
import { ProductItemCard } from "./ProductItemCard";

interface MenuInventoryViewProps {
  products: Product[];
  search: string;
  onSearchChange: (s: string) => void;
  onAddDish: () => void;
  onEditDish: (product: Product) => void;
  onDeleteDish: (product: Product) => void;
}

export function MenuInventoryView({
  products,
  search,
  onSearchChange,
  onAddDish,
  onEditDish,
  onDeleteDish,
}: MenuInventoryViewProps) {
  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-bold">Menu &amp; Dish Catalog</h2>
          <p className="text-xs text-muted-foreground">Manage prices, descriptions, and portion inventory</p>
        </div>

        <div className="flex items-center gap-2">
          <div className="relative w-64">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => onSearchChange(e.target.value)}
              placeholder="Search dishes..."
              className="pl-9 h-9 text-xs"
            />
          </div>
          <Button size="sm" onClick={onAddDish} className="h-9 flex items-center gap-1.5 text-xs">
            <Plus className="h-4 w-4" />
            <span>Add Dish</span>
          </Button>
        </div>
      </div>

      {products.length === 0 ? (
        <Card className="p-12 text-center text-muted-foreground space-y-3">
          <UtensilsCrossed className="h-10 w-10 mx-auto opacity-40 text-primary" />
          <h3 className="font-semibold text-foreground">No dishes found in menu</h3>
          <p className="text-xs max-w-sm mx-auto">
            {search ? "No dishes matched your search." : "Start by adding your first signature dish to the menu!"}
          </p>
          {!search && (
            <Button size="sm" onClick={onAddDish} className="flex items-center gap-1 mx-auto">
              <Plus className="h-4 w-4" />
              <span>Create Dish</span>
            </Button>
          )}
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {products.map((product) => (
            <ProductItemCard
              key={product.id}
              product={product}
              onEdit={onEditDish}
              onDelete={onDeleteDish}
            />
          ))}
        </div>
      )}
    </div>
  );
}
