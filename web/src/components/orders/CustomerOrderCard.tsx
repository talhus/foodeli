"use client";

import { Order } from "@/types/api";
import { Card, CardHeader, CardDescription, CardContent } from "@/components/ui/card";
import { formatCurrency, formatDate } from "@/lib/utils";
import { OrderStatusBadge } from "./OrderStatusBadge";

export function CustomerOrderCard({ order }: { order: Order }) {
  const shortId = order?.id ? order.id.slice(0, 8) : "--------";

  return (
    <Card className="overflow-hidden">
      <CardHeader className="p-5 pb-3 bg-muted/20 border-b flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs font-bold">#{shortId}</span>
            <span className="text-xs text-muted-foreground">&bull;</span>
            <span className="text-xs text-muted-foreground" suppressHydrationWarning>
              {formatDate(order?.created_at)}
            </span>
          </div>
          <CardDescription className="text-xs mt-1">
            Delivering to: <span className="font-medium text-foreground">{order?.delivery_address || "Standard Address"}</span>
          </CardDescription>
        </div>
        <div>
          <OrderStatusBadge status={order?.status} />
        </div>
      </CardHeader>

      <CardContent className="p-5">
        <div className="space-y-2">
          {Array.isArray(order?.items) && order.items.length > 0 ? (
            <div className="divide-y text-sm">
              {order.items.map((item, idx) => (
                <div key={item?.id || idx} className="py-2 flex items-center justify-between">
                  <span className="text-foreground">
                    {item?.quantity || 1}x {item?.product_name || "Item"}
                  </span>
                  <span className="text-muted-foreground font-mono text-xs">
                    {formatCurrency(item?.subtotal || 0)}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-xs text-muted-foreground italic">Order snapshot preserved.</p>
          )}

          <div className="flex items-center justify-between pt-3 border-t font-semibold text-sm">
            <span>Total Amount</span>
            <span className="text-base text-primary font-bold">
              {formatCurrency(order?.total_amount || 0)}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
