"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Store, Product } from "@/types/api";
import { useAuthStore } from "@/store/useAuthStore";
import { AdminHeader } from "@/components/admin/AdminHeader";
import { AdminKpiCards } from "@/components/admin/AdminKpiCards";
import { AdminStoresTable } from "@/components/admin/AdminStoresTable";
import { AdminArchitectureCard } from "@/components/admin/AdminArchitectureCard";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

export function AdminPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuthStore();
  const [stores, setStores] = useState<Store[]>([]);
  const [storeProductCounts, setStoreProductCounts] = useState<Record<string, number>>({});
  const [totalProducts, setTotalProducts] = useState(0);
  const [healthStatus, setHealthStatus] = useState("healthy");
  const [isLoading, setIsLoading] = useState(true);
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [storeList, health] = await Promise.all([
        api.stores.list().catch(() => []),
        api.system.health().catch(() => ({ status: "healthy", service: "foodeli-core-api" })),
      ]);

      const safeStores = Array.isArray(storeList) ? storeList : [];
      setStores(safeStores);
      setHealthStatus(health.status || "healthy");

      // Fetch products count for each store concurrently
      const counts: Record<string, number> = {};
      let total = 0;

      const productPromises = safeStores.map(async (st) => {
        try {
          const prods = await api.products.listByStore(st.id);
          const count = Array.isArray(prods) ? prods.length : 0;
          counts[st.id] = count;
          total += count;
        } catch {
          counts[st.id] = 0;
        }
      });

      await Promise.all(productPromises);
      setStoreProductCounts(counts);
      setTotalProducts(total);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!isMounted) return;

    const hasStoredToken = typeof window !== "undefined" && !!localStorage.getItem("foodeli_token");

    if (isAuthLoading && hasStoredToken) {
      return;
    }

    if (!isAuthenticated && !hasStoredToken) {
      toast.error("Access restricted: Please sign in as Administrator");
      router.push("/");
      return;
    }

    if (isAuthenticated && user?.role !== "admin") {
      toast.error("Unauthorized: Administrator permissions required");
      router.push("/");
      return;
    }

    if (isAuthenticated && user?.role === "admin") {
      loadData();
    }
  }, [isMounted, isAuthenticated, isAuthLoading, user, router, loadData]);

  if (!isMounted || (isLoading && stores.length === 0)) {
    return (
      <div className="flex items-center justify-center p-24">
        <Loader2 className="h-8 w-8 animate-spin text-purple-600" />
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <AdminHeader
        onRefresh={loadData}
        isLoading={isLoading}
        adminEmail={user?.email}
      />

      <AdminKpiCards
        totalStores={stores.length}
        totalProducts={totalProducts}
        healthStatus={healthStatus}
      />

      <div className="grid grid-cols-1 gap-6">
        <AdminStoresTable
          stores={stores}
          storeProductCounts={storeProductCounts}
        />

        <AdminArchitectureCard />
      </div>
    </div>
  );
}
