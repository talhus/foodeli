"use client";

import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/navigation";
import { api } from "@/lib/api";
import { Store, Order, Product, OrderStatus } from "@/types/api";
import { useAuthStore } from "@/store/useAuthStore";
import { MerchantHeader } from "@/components/merchant/MerchantHeader";
import { MerchantKPIs } from "@/components/merchant/MerchantKPIs";
import { MerchantTabsNav } from "@/components/merchant/MerchantTabsNav";
import { StoreOnboarding } from "@/components/merchant/StoreOnboarding";
import { KitchenOrdersView } from "@/components/merchant/KitchenOrdersView";
import { MenuInventoryView } from "@/components/merchant/MenuInventoryView";
import { ProductFormModal } from "@/components/merchant/ProductFormModal";
import { DeleteProductDialog } from "@/components/merchant/DeleteProductDialog";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

export function MerchantPage() {
  const router = useRouter();
  const { user, isAuthenticated, isLoading: isAuthLoading } = useAuthStore();

  const [activeTab, setActiveTab] = useState<"orders" | "menu">("orders");
  const [orderFilter, setOrderFilter] = useState<"ALL" | "ACTIVE" | "DELIVERED" | "CANCELLED">("ALL");
  const [menuSearch, setMenuSearch] = useState("");

  const [store, setStore] = useState<Store | null>(null);
  const [orders, setOrders] = useState<Order[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUpdating, setIsUpdating] = useState(false);

  // Modals state
  const [isProductModalOpen, setIsProductModalOpen] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [deletingProduct, setDeletingProduct] = useState<Product | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const myStore = await api.stores.getMyStore().catch(() => null);
      setStore(myStore);
      if (myStore) {
        const [o, p] = await Promise.all([
          api.orders.listMerchantOrders().catch(() => []),
          api.products.listByStore(myStore.id).catch(() => []),
        ]);
        setOrders(o || []);
        setProducts(p || []);
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!isAuthLoading) {
      if (!isAuthenticated || user?.role !== "merchant") {
        toast.error("Access restricted to registered merchants");
        router.push("/");
        return;
      }
      loadData();
    }
  }, [isAuthenticated, isAuthLoading, user, router]);

  // Auto-refresh merchant kitchen orders in real-time when SSE notification arrives
  useEffect(() => {
    const handleNotification = () => {
      loadData();
    };
    window.addEventListener("foodeli:notification", handleNotification);
    return () => window.removeEventListener("foodeli:notification", handleNotification);
  }, []);

  const handleCreateStore = async (name: string, address: string) => {
    const s = await api.stores.create(name, address);
    setStore(s);
    toast.success("Storefront successfully registered!");
    await loadData();
  };

  const handleUpdateStatus = async (orderId: string, status: OrderStatus) => {
    setIsUpdating(true);
    try {
      await api.orders.updateStatus(orderId, status);
      toast.success(`Order status updated to ${status}`);
      await loadData();
    } catch (err: unknown) {
      toast.error(err instanceof Error ? err.message : "Failed to update status");
    } finally {
      setIsUpdating(false);
    }
  };

  const handleSaveProduct = async (data: { name: string; description: string; price: number; stock: number }) => {
    if (editingProduct) {
      await api.products.update(editingProduct.id, data);
      toast.success(`Updated ${data.name}`);
    } else {
      await api.products.create(data);
      toast.success(`Created dish ${data.name}`);
    }
    await loadData();
  };

  const handleDeleteProduct = async () => {
    if (!deletingProduct) return;
    setIsDeleting(true);
    try {
      await api.products.delete(deletingProduct.id);
      toast.success(`Deleted ${deletingProduct.name}`);
      setDeletingProduct(null);
      await loadData();
    } finally {
      setIsDeleting(false);
    }
  };

  const metrics = useMemo(() => {
    const valid = orders.filter((o) => o.status !== "CANCELLED");
    const totalRevenue = valid.reduce((acc, curr) => acc + curr.total_amount, 0);
    const activeKitchenOrders = orders.filter((o) => o.status === "CONFIRMED" || o.status === "PREPARING").length;
    const lowStockCount = products.filter((p) => p.stock <= 3).length;
    return { totalRevenue, activeKitchenOrders, totalDishes: products.length, lowStockCount };
  }, [orders, products]);

  const filteredOrders = useMemo(() => {
    return orders.filter((o) => {
      if (orderFilter === "ACTIVE") return o.status === "CONFIRMED" || o.status === "PREPARING";
      if (orderFilter === "DELIVERED") return o.status === "DELIVERED";
      if (orderFilter === "CANCELLED") return o.status === "CANCELLED";
      return true;
    });
  }, [orders, orderFilter]);

  const filteredProducts = useMemo(() => {
    if (!menuSearch.trim()) return products;
    const q = menuSearch.toLowerCase();
    return products.filter((p) => p.name.toLowerCase().includes(q) || p.description.toLowerCase().includes(q));
  }, [products, menuSearch]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-24">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!store) {
    return <StoreOnboarding onCreateStore={handleCreateStore} />;
  }

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      <MerchantHeader
        store={store}
        onRefresh={loadData}
        onAddDish={() => {
          setEditingProduct(null);
          setIsProductModalOpen(true);
        }}
      />
      <MerchantKPIs
        totalRevenue={metrics.totalRevenue}
        ordersCount={orders.length}
        activeKitchenOrders={metrics.activeKitchenOrders}
        totalDishes={metrics.totalDishes}
        lowStockCount={metrics.lowStockCount}
      />
      <MerchantTabsNav
        activeTab={activeTab}
        ordersCount={orders.length}
        productsCount={products.length}
        onTabChange={setActiveTab}
      />
      {activeTab === "orders" ? (
        <KitchenOrdersView
          orders={filteredOrders}
          filter={orderFilter}
          isUpdating={isUpdating}
          onFilterChange={setOrderFilter}
          onUpdateStatus={handleUpdateStatus}
        />
      ) : (
        <MenuInventoryView
          products={filteredProducts}
          search={menuSearch}
          onSearchChange={setMenuSearch}
          onAddDish={() => {
            setEditingProduct(null);
            setIsProductModalOpen(true);
          }}
          onEditDish={(p) => {
            setEditingProduct(p);
            setIsProductModalOpen(true);
          }}
          onDeleteDish={(p) => setDeletingProduct(p)}
        />
      )}
      <ProductFormModal
        open={isProductModalOpen}
        editingProduct={editingProduct}
        onOpenChange={setIsProductModalOpen}
        onSubmit={handleSaveProduct}
      />
      <DeleteProductDialog
        product={deletingProduct}
        isDeleting={isDeleting}
        onConfirm={handleDeleteProduct}
        onCancel={() => setDeletingProduct(null)}
      />
    </div>
  );
}
