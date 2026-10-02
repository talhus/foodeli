import { Order, OrderStatus } from "@/types/api";
import { Card, CardHeader, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { formatCurrency, formatDate } from "@/lib/utils";
import { OrderStatusBadge } from "@/components/orders/OrderStatusBadge";
import { ChefHat, Truck } from "lucide-react";

interface KitchenOrderCardProps {
  order: Order;
  isUpdating: boolean;
  onUpdateStatus: (orderId: string, status: OrderStatus) => void;
}

export function KitchenOrderCard({ order, isUpdating, onUpdateStatus }: KitchenOrderCardProps) {
  return (
    <Card className="flex flex-col justify-between overflow-hidden shadow-sm">
      <CardHeader className="p-5 pb-3 bg-muted/20 border-b">
        <div className="flex items-center justify-between">
          <span className="font-mono text-xs font-bold">#{order.id.slice(0, 8)}</span>
          <OrderStatusBadge status={order.status} />
        </div>
        <CardDescription className="text-xs mt-1">Placed: {formatDate(order.created_at)}</CardDescription>
        <p className="text-xs font-medium text-foreground mt-0.5 truncate">📍 {order.delivery_address}</p>
      </CardHeader>

      <CardContent className="p-5 flex-1">
        <div className="space-y-2">
          {order.items && order.items.length > 0 && (
            <div className="divide-y text-xs">
              {order.items.map((item) => (
                <div key={item.id} className="py-1.5 flex justify-between">
                  <span className="font-medium text-foreground">
                    {item.quantity}x {item.product_name}
                  </span>
                  <span className="font-mono text-muted-foreground">{formatCurrency(item.subtotal)}</span>
                </div>
              ))}
            </div>
          )}
          <div className="flex justify-between items-center pt-2 border-t font-semibold text-sm">
            <span>Order Total</span>
            <span className="text-primary font-bold">{formatCurrency(order.total_amount)}</span>
          </div>
        </div>
      </CardContent>

      <CardFooter className="p-5 pt-0 border-t bg-muted/10 flex flex-wrap gap-2 justify-end">
        {order.status === "PENDING" && (
          <Button
            size="sm"
            variant="default"
            disabled={isUpdating}
            onClick={() => onUpdateStatus(order.id, "CONFIRMED")}
            className="text-xs h-8"
          >
            Confirm Order
          </Button>
        )}
        {order.status === "CONFIRMED" && (
          <Button
            size="sm"
            variant="default"
            disabled={isUpdating}
            onClick={() => onUpdateStatus(order.id, "PREPARING")}
            className="text-xs h-8 flex items-center gap-1"
          >
            <ChefHat className="h-3.5 w-3.5" />
            <span>Start Preparing</span>
          </Button>
        )}
        {order.status === "PREPARING" && (
          <Button
            size="sm"
            variant="default"
            disabled={isUpdating}
            onClick={() => onUpdateStatus(order.id, "DELIVERED")}
            className="text-xs h-8 flex items-center gap-1 bg-emerald-600 hover:bg-emerald-700"
          >
            <Truck className="h-3.5 w-3.5" />
            <span>Mark Delivered</span>
          </Button>
        )}
        {order.status !== "DELIVERED" && order.status !== "CANCELLED" && (
          <Button
            size="sm"
            variant="ghost"
            disabled={isUpdating}
            onClick={() => onUpdateStatus(order.id, "CANCELLED")}
            className="text-xs h-8 text-destructive hover:bg-destructive/10"
          >
            Cancel
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
