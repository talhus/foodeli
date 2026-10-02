"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { useAuthStore } from "@/store/useAuthStore";
import { useCartStore } from "@/store/useCartStore";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ShoppingBag, UtensilsCrossed, User as UserIcon, Store, LogOut, ClipboardList, ShieldAlert } from "lucide-react";
import { AuthModal } from "@/components/AuthModal";

export function Navbar() {
  const { user, isAuthenticated, logout } = useAuthStore();
  const { cart, toggleCart } = useCartStore();
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const pathname = usePathname();

  const isAdmin = user?.role === "admin" || pathname?.startsWith("/admin");
  const isMerchant = user?.role === "merchant" || pathname?.startsWith("/merchant");
  const isCustomer = !isAdmin && !isMerchant;

  const itemCount = cart?.item_count || 0;

  return (
    <>
      <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-16 items-center justify-between">
          {/* Logo & Brand */}
          <Link href="/" className="flex items-center space-x-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary text-primary-foreground shadow-sm">
              <UtensilsCrossed className="h-5 w-5" />
            </div>
            <span className="font-bold text-xl tracking-tight">Foodeli</span>
            <span className="hidden sm:inline-block text-[10px] font-semibold uppercase tracking-wider text-muted-foreground bg-muted px-2 py-0.5 rounded-full">
              Modular Monolith
            </span>
          </Link>

          {/* Navigation Links */}
          <nav className="flex items-center space-x-1 sm:space-x-3">
            <Link href="/" passHref>
              <Button variant="ghost" size="sm">
                Marketplace
              </Button>
            </Link>

            {isAuthenticated && isCustomer && (
              <Link href="/orders" passHref>
                <Button variant="ghost" size="sm" className="flex items-center gap-1.5">
                  <ClipboardList className="h-4 w-4" />
                  <span className="hidden md:inline">Orders</span>
                </Button>
              </Link>
            )}

            {isAuthenticated && user?.role === "merchant" && (
              <Link href="/merchant" passHref>
                <Button variant="outline" size="sm" className="flex items-center gap-1.5 border-primary/40 text-primary hover:bg-primary/5">
                  <Store className="h-4 w-4" />
                  <span className="hidden md:inline">Merchant Portal</span>
                </Button>
              </Link>
            )}

            {isAuthenticated && user?.role === "admin" && (
              <Link href="/admin" passHref>
                <Button variant="outline" size="sm" className="flex items-center gap-1.5 border-purple-500/40 text-purple-600 dark:text-purple-400 hover:bg-purple-500/10">
                  <ShieldAlert className="h-4 w-4" />
                  <span className="hidden md:inline">Admin Panel</span>
                </Button>
              </Link>
            )}

            {/* Cart Trigger (Hidden for Admin and Merchant) */}
            {isCustomer && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => toggleCart(true)}
                className="relative flex items-center gap-2 border-border"
              >
                <ShoppingBag className="h-4 w-4" />
                <span className="hidden sm:inline font-medium">Cart</span>
                {itemCount > 0 && (
                  <Badge variant="default" className="h-5 px-1.5 text-[10px] font-bold">
                    {itemCount}
                  </Badge>
                )}
              </Button>
            )}

            {/* Auth Actions */}
            {isAuthenticated ? (
              <div className="flex items-center space-x-2 pl-2">
                <div className="hidden sm:flex flex-col text-right">
                  <span className="text-xs font-semibold leading-none">{user?.email}</span>
                  <span className="text-[10px] capitalize text-muted-foreground">{user?.role}</span>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={logout}
                  title="Logout"
                  className="h-9 w-9 text-muted-foreground hover:text-destructive"
                >
                  <LogOut className="h-4 w-4" />
                </Button>
              </div>
            ) : (
              <Button
                variant="default"
                size="sm"
                onClick={() => setAuthModalOpen(true)}
                className="flex items-center gap-1.5 font-medium"
              >
                <UserIcon className="h-4 w-4" />
                <span>Sign In</span>
              </Button>
            )}
          </nav>
        </div>
      </header>

      {/* Auth Modal */}
      <AuthModal open={authModalOpen} onOpenChange={setAuthModalOpen} />
    </>
  );
}
