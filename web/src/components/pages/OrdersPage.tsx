"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Order } from "@/types/api";
import { useAuthStore } from "@/store/useAuthStore";
import { OrdersHeader } from "@/components/orders/OrdersHeader";
import { OrdersEmptyState } from "@/components/orders/OrdersEmptyState";
import { CustomerOrderCard } from "@/components/orders/CustomerOrderCard";
import { Loader2, AlertCircle, RefreshCw, ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { toast } from "sonner";

export function OrdersPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuthStore();
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isMounted, setIsMounted] = useState(false);
  const [page, setPage] = useState(1);
  const PAGE_LIMIT = 5;

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const loadOrders = useCallback(async (pageNum: number = 1) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await api.orders.listCustomerOrders(pageNum, PAGE_LIMIT);
      setOrders(Array.isArray(data) ? data : []);
      setPage(pageNum);
    } catch (err: unknown) {
      console.error("Failed to load customer orders:", err);
      const msg = err instanceof Error ? err.message : "Failed to load orders. Please try again.";
      setError(msg);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isMounted) return;

    const hasStoredToken = typeof window !== "undefined" && !!localStorage.getItem("foodeli_token");

    // If still authenticating/hydrating and token exists, wait for hydration
    if (isAuthLoading && hasStoredToken) {
      return;
    }

    // If not authenticated and no stored token exists, redirect to home
    if (!isAuthenticated && !hasStoredToken) {
      toast.error("Please sign in to view your orders");
      router.push("/");
      return;
    }

    // If logged in as a merchant, redirect to merchant management portal
    if (isAuthenticated && user?.role === "merchant") {
      router.push("/merchant");
      return;
    }

    if (isAuthenticated) {
      loadOrders();
    }
  }, [isMounted, isAuthenticated, isAuthLoading, user, router, loadOrders]);

  // Auto-refresh orders list in real-time when an SSE notification arrives
  useEffect(() => {
    const handleNotification = () => {
      loadOrders(page);
    };
    window.addEventListener("foodeli:notification", handleNotification);
    return () => window.removeEventListener("foodeli:notification", handleNotification);
  }, [loadOrders, page]);

  if (!isMounted || (isLoading && orders.length === 0 && !error)) {
    return (
      <div className="max-w-4xl mx-auto space-y-6">
        <OrdersHeader />
        <div className="flex items-center justify-center p-16">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <OrdersHeader />

      {error ? (
        <Card className="p-8 text-center space-y-4 border-destructive/20 bg-destructive/5">
          <AlertCircle className="h-10 w-10 text-destructive mx-auto" />
          <div className="space-y-1">
            <h3 className="font-semibold text-foreground">Could not load orders</h3>
            <p className="text-sm text-muted-foreground">{error}</p>
          </div>
          <Button onClick={() => loadOrders()} variant="outline" size="sm" className="gap-2">
            <RefreshCw className="h-4 w-4" />
            <span>Try Again</span>
          </Button>
        </Card>
      ) : orders.length === 0 ? (
        <OrdersEmptyState />
      ) : (
        <div className="space-y-4">
          {orders.map((order) => (
            <CustomerOrderCard key={order.id} order={order} />
          ))}

          {/* Pagination Controls */}
          {(orders.length > 0 || page > 1) && (
            <div className="flex items-center justify-between pt-4 border-t border-border/60">
              <span className="text-xs text-muted-foreground">
                Page <span className="font-semibold text-foreground">{page}</span>
              </span>

              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => loadOrders(page - 1)}
                  disabled={page <= 1 || isLoading}
                  className="h-8 gap-1 text-xs"
                >
                  <ChevronLeft className="h-3.5 w-3.5" />
                  <span>Previous</span>
                </Button>

                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => loadOrders(page + 1)}
                  disabled={orders.length < PAGE_LIMIT || isLoading}
                  className="h-8 gap-1 text-xs"
                >
                  <span>Next</span>
                  <ChevronRight className="h-3.5 w-3.5" />
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
