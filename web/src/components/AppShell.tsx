"use client";

import { useEffect } from "react";
import { useAuthStore } from "@/store/useAuthStore";
import { useCartStore } from "@/store/useCartStore";
import { useNotifications } from "@/hooks/useNotifications";
import { Navbar } from "@/components/Navbar";
import { CartDrawer } from "@/components/CartDrawer";
import { ConflictModal } from "@/components/ConflictModal";
import { Toaster } from "sonner";

export function AppShell({ children }: { children: React.ReactNode }) {
  const { initAuth } = useAuthStore();
  const { fetchCart } = useCartStore();

  // Listen for real-time SSE notifications globally
  useNotifications();

  useEffect(() => {
    initAuth().then(() => {
      fetchCart();
    });
  }, [initAuth, fetchCart]);

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <Navbar />
      <main className="flex-1 container py-8">{children}</main>
      <footer className="border-t py-6 text-center text-xs text-muted-foreground">
        Foodeli Marketplace & Delivery Platform &bull; Modular Monolith &bull; Go + Next.js
      </footer>

      {/* Global Overlays */}
      <CartDrawer />
      <ConflictModal />
      <Toaster position="top-right" richColors />
    </div>
  );
}
