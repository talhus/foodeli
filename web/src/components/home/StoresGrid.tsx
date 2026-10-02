import { Store } from "@/types/api";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Store as StoreIcon, Loader2 } from "lucide-react";
import { StoreCard } from "./StoreCard";

interface StoresGridProps {
  stores: Store[];
  selectedStore: Store | null;
  isLoading: boolean;
  onSelectStore: (store: Store) => void;
}

export function StoresGrid({ stores, selectedStore, isLoading, onSelectStore }: StoresGridProps) {
  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-bold tracking-tight">Partner Restaurants</h2>
          <p className="text-xs text-muted-foreground">Select a store to view their full menu</p>
        </div>
        <Badge variant="outline" className="text-xs">
          {stores.length} Available
        </Badge>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center p-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : stores.length === 0 ? (
        <Card className="p-8 text-center text-muted-foreground">
          <StoreIcon className="h-10 w-10 mx-auto mb-2 opacity-50" />
          <p>No stores registered yet. Sign in as a Merchant to open the first storefront!</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {stores.map((store) => (
            <StoreCard
              key={store.id}
              store={store}
              isSelected={selectedStore?.id === store.id}
              onSelect={onSelectStore}
            />
          ))}
        </div>
      )}
    </section>
  );
}
