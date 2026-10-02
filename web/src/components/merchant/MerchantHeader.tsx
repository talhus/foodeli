import { Store } from "@/types/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Store as StoreIcon, RefreshCw, Plus } from "lucide-react";

interface MerchantHeaderProps {
  store: Store;
  onRefresh: () => void;
  onAddDish: () => void;
}

export function MerchantHeader({ store, onRefresh, onAddDish }: MerchantHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b pb-6">
      <div className="flex items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-sm">
          <StoreIcon className="h-6 w-6" />
        </div>
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold tracking-tight">{store.name}</h1>
            <Badge variant="outline" className="text-[10px] text-emerald-600 border-emerald-500/30 bg-emerald-500/10">
              Live &amp; Open
            </Badge>
          </div>
          <p className="text-xs text-muted-foreground">{store.address} &bull; Merchant Portal</p>
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onRefresh} className="flex items-center gap-1.5">
          <RefreshCw className="h-4 w-4" />
          <span>Refresh</span>
        </Button>
        <Button size="sm" onClick={onAddDish} className="flex items-center gap-1.5 font-medium shadow-sm">
          <Plus className="h-4 w-4" />
          <span>Add New Dish</span>
        </Button>
      </div>
    </div>
  );
}
