"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { Store, Product } from "@/types/api";
import { useCartStore } from "@/store/useCartStore";
import { useAuthStore } from "@/store/useAuthStore";
import { AuthModal } from "@/components/AuthModal";
import { StoresGrid } from "@/components/home/StoresGrid";
import { ProductsGrid } from "@/components/home/ProductsGrid";
import { toast } from "sonner";

export function HomePage() {
  const [stores, setStores] = useState<Store[]>([]);
  const [selectedStore, setSelectedStore] = useState<Store | null>(null);
  const [products, setProducts] = useState<Product[]>([]);
  const [isLoadingStores, setIsLoadingStores] = useState(true);
  const [isLoadingProducts, setIsLoadingProducts] = useState(false);
  const [authModalOpen, setAuthModalOpen] = useState(false);

  const { addItem, isLoading: isCartLoading } = useCartStore();
  const { isAuthenticated } = useAuthStore();

  useEffect(() => {
    async function loadStores() {
      try {
        const data = await api.stores.list();
        setStores(data || []);
        if (data && data.length > 0) {
          setSelectedStore(data[0]);
        }
      } catch (err) {
        console.error("Failed to load stores:", err);
      } finally {
        setIsLoadingStores(false);
      }
    }
    loadStores();
  }, []);

  useEffect(() => {
    if (!selectedStore) return;
    const storeId = selectedStore.id;
    async function loadProducts() {
      setIsLoadingProducts(true);
      try {
        const data = await api.products.listByStore(storeId);
        setProducts(data || []);
      } catch (err) {
        console.error("Failed to load products for store:", err);
      } finally {
        setIsLoadingProducts(false);
      }
    }
    loadProducts();
  }, [selectedStore]);

  const handleAddToCart = async (product: Product) => {
    if (!isAuthenticated) {
      toast.info("Please sign in to add items to your cart");
      setAuthModalOpen(true);
      return;
    }
    if (!selectedStore) return;

    try {
      const added = await addItem(selectedStore.id, product.id, 1);
      if (added) {
        toast.success(`Added ${product.name} to cart`);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : "Failed to add item to cart";
      toast.error(msg);
    }
  };

  return (
    <div className="space-y-8">
      <StoresGrid
        stores={stores}
        selectedStore={selectedStore}
        isLoading={isLoadingStores}
        onSelectStore={setSelectedStore}
      />
      {selectedStore && (
        <ProductsGrid
          store={selectedStore}
          products={products}
          isLoading={isLoadingProducts}
          isCartLoading={isCartLoading}
          onAddToCart={handleAddToCart}
        />
      )}
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} />
    </div>
  );
}
