"use client";

import { OrderStatus } from "@/types/api";
import { Badge } from "@/components/ui/badge";
import { Clock, CheckCircle2, ChefHat, Truck, XCircle } from "lucide-react";

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  switch (status) {
    case "PENDING":
      return (
        <Badge variant="warning" className="flex items-center gap-1">
          <Clock className="h-3 w-3" />
          <span>Pending Payment</span>
        </Badge>
      );
    case "CONFIRMED":
      return (
        <Badge variant="success" className="flex items-center gap-1">
          <CheckCircle2 className="h-3 w-3" />
          <span>Confirmed</span>
        </Badge>
      );
    case "PREPARING":
      return (
        <Badge variant="default" className="flex items-center gap-1">
          <ChefHat className="h-3 w-3" />
          <span>Preparing</span>
        </Badge>
      );
    case "DELIVERED":
      return (
        <Badge variant="secondary" className="flex items-center gap-1 text-emerald-700 bg-emerald-100">
          <Truck className="h-3 w-3" />
          <span>Delivered</span>
        </Badge>
      );
    case "CANCELLED":
      return (
        <Badge variant="destructive" className="flex items-center gap-1">
          <XCircle className="h-3 w-3" />
          <span>Cancelled</span>
        </Badge>
      );
    default:
      return <Badge variant="outline">{status}</Badge>;
  }
}
