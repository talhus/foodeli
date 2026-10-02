import { Card } from "@/components/ui/card";
import { formatCurrency } from "@/lib/utils";
import { DollarSign, ChefHat, Package, AlertTriangle } from "lucide-react";

interface MerchantKPIsProps {
  totalRevenue: number;
  ordersCount: number;
  activeKitchenOrders: number;
  totalDishes: number;
  lowStockCount: number;
}

export function MerchantKPIs({
  totalRevenue,
  ordersCount,
  activeKitchenOrders,
  totalDishes,
  lowStockCount,
}: MerchantKPIsProps) {
  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <Card className="p-4 shadow-sm border-primary/10">
        <div className="flex items-center justify-between text-muted-foreground text-xs mb-1">
          <span>Total Sales Volume</span>
          <DollarSign className="h-4 w-4 text-emerald-600" />
        </div>
        <div className="text-xl font-bold text-foreground">{formatCurrency(totalRevenue)}</div>
        <div className="text-[11px] text-muted-foreground mt-0.5">{ordersCount} total orders placed</div>
      </Card>

      <Card className="p-4 shadow-sm border-primary/10">
        <div className="flex items-center justify-between text-muted-foreground text-xs mb-1">
          <span>Active Kitchen Orders</span>
          <ChefHat className="h-4 w-4 text-primary" />
        </div>
        <div className="text-xl font-bold text-primary">{activeKitchenOrders}</div>
        <div className="text-[11px] text-muted-foreground mt-0.5">Need immediate attention</div>
      </Card>

      <Card className="p-4 shadow-sm border-primary/10">
        <div className="flex items-center justify-between text-muted-foreground text-xs mb-1">
          <span>Menu Dishes</span>
          <Package className="h-4 w-4 text-foreground" />
        </div>
        <div className="text-xl font-bold text-foreground">{totalDishes}</div>
        <div className="text-[11px] text-muted-foreground mt-0.5">Dishes in active catalog</div>
      </Card>

      <Card className="p-4 shadow-sm border-primary/10">
        <div className="flex items-center justify-between text-muted-foreground text-xs mb-1">
          <span>Low Stock Alerts</span>
          <AlertTriangle className="h-4 w-4 text-amber-500" />
        </div>
        <div className="text-xl font-bold text-amber-600">{lowStockCount}</div>
        <div className="text-[11px] text-muted-foreground mt-0.5">Dishes with &le; 3 portions left</div>
      </Card>
    </div>
  );
}
