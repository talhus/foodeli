import { Order, OrderStatus } from "@/types/api";
import { Card } from "@/components/ui/card";
import { ChefHat } from "lucide-react";
import { KitchenOrderCard } from "./KitchenOrderCard";

interface KitchenOrdersViewProps {
  orders: Order[];
  filter: "ALL" | "ACTIVE" | "DELIVERED" | "CANCELLED";
  isUpdating: boolean;
  onFilterChange: (f: "ALL" | "ACTIVE" | "DELIVERED" | "CANCELLED") => void;
  onUpdateStatus: (orderId: string, status: OrderStatus) => void;
}

export function KitchenOrdersView({
  orders,
  filter,
  isUpdating,
  onFilterChange,
  onUpdateStatus,
}: KitchenOrdersViewProps) {
  return (
    <div className="space-y-4">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <h2 className="text-lg font-bold">Kitchen Display &amp; Fulfillment</h2>
        <div className="flex items-center gap-1 bg-muted/40 p-1 rounded-lg border">
          {(["ALL", "ACTIVE", "DELIVERED", "CANCELLED"] as const).map((f) => (
            <button
              key={f}
              onClick={() => onFilterChange(f)}
              className={`px-3 py-1 rounded-md text-xs font-medium transition-all ${
                filter === f
                  ? "bg-background text-foreground shadow-sm font-semibold"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      {orders.length === 0 ? (
        <Card className="p-12 text-center text-muted-foreground space-y-2">
          <ChefHat className="h-10 w-10 mx-auto opacity-40 text-primary" />
          <h3 className="font-semibold text-foreground">No orders matching filter</h3>
          <p className="text-xs">Incoming customer orders will appear here in real-time.</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {orders.map((order) => (
            <KitchenOrderCard
              key={order.id}
              order={order}
              isUpdating={isUpdating}
              onUpdateStatus={onUpdateStatus}
            />
          ))}
        </div>
      )}
    </div>
  );
}
