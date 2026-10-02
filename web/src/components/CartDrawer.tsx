"use client";

import { useState } from "react";
import { useCartStore } from "@/store/useCartStore";
import { useAuthStore } from "@/store/useAuthStore";
import { Button } from "@/components/ui/button";
import { formatCurrency } from "@/lib/utils";
import { ShoppingBag, X, Plus, Minus, Trash2, ArrowRight } from "lucide-react";
import { CheckoutModal } from "@/components/CheckoutModal";
import { AuthModal } from "@/components/AuthModal";

export function CartDrawer() {
  const { cart, isOpen, toggleCart, updateQuantity, removeItem, clearCart, isLoading } = useCartStore();
  const { isAuthenticated } = useAuthStore();

  const [checkoutOpen, setCheckoutOpen] = useState(false);
  const [authOpen, setAuthOpen] = useState(false);

  if (!isOpen) return null;

  const items = cart?.items || [];
  const isEmpty = items.length === 0;

  const handleProceedCheckout = () => {
    if (!isAuthenticated) {
      setAuthOpen(true);
      return;
    }
    setCheckoutOpen(true);
  };

  return (
    <>
      <div className="fixed inset-0 z-50 overflow-hidden">
        {/* Backdrop */}
        <div
          className="fixed inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
          onClick={() => toggleCart(false)}
        />

        <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
          <div className="w-screen max-w-md bg-background border-l shadow-2xl flex flex-col">
            {/* Header */}
            <div className="flex items-center justify-between px-6 py-4 border-b">
              <div className="flex items-center space-x-2">
                <ShoppingBag className="h-5 w-5 text-primary" />
                <h2 className="text-lg font-semibold">Your Cart</h2>
                {cart?.item_count ? (
                  <span className="text-xs bg-muted text-muted-foreground px-2 py-0.5 rounded-full font-bold">
                    {cart.item_count} items
                  </span>
                ) : null}
              </div>
              <button
                onClick={() => toggleCart(false)}
                className="text-muted-foreground hover:text-foreground transition-colors p-1 rounded-lg"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            {/* Store Name Indicator */}
            {cart?.store_name && !isEmpty && (
              <div className="px-6 py-2.5 bg-muted/40 border-b flex justify-between items-center text-xs">
                <span className="text-muted-foreground">
                  Ordering from: <strong className="text-foreground">{cart.store_name}</strong>
                </span>
                <button
                  onClick={clearCart}
                  className="text-destructive hover:underline text-[11px] font-medium"
                >
                  Clear all
                </button>
              </div>
            )}

            {/* Cart Items List */}
            <div className="flex-1 overflow-y-auto px-6 py-4 divide-y">
              {isEmpty ? (
                <div className="flex flex-col items-center justify-center h-full text-center space-y-3 py-16">
                  <div className="h-16 w-16 rounded-full bg-muted flex items-center justify-center text-muted-foreground">
                    <ShoppingBag className="h-8 w-8" />
                  </div>
                  <h3 className="font-semibold text-lg">Your cart is empty</h3>
                  <p className="text-sm text-muted-foreground max-w-xs">
                    Explore restaurants in the marketplace and add tasty dishes to your order.
                  </p>
                  <Button variant="outline" size="sm" onClick={() => toggleCart(false)}>
                    Browse Marketplace
                  </Button>
                </div>
              ) : (
                items.map((item) => (
                  <div key={item.product_id} className="py-4 flex items-center justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <h4 className="text-sm font-semibold truncate">{item.product_name}</h4>
                      <p className="text-xs text-muted-foreground">
                        {formatCurrency(item.unit_price)} each
                      </p>
                      <p className="text-xs font-semibold text-foreground mt-1">
                        Subtotal: {formatCurrency(item.subtotal)}
                      </p>
                    </div>

                    <div className="flex items-center space-x-2">
                      <div className="flex items-center border rounded-lg overflow-hidden bg-background">
                        <button
                          onClick={() => updateQuantity(item.product_id, item.quantity - 1)}
                          disabled={isLoading}
                          className="p-1.5 hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                        >
                          <Minus className="h-3 w-3" />
                        </button>
                        <span className="w-8 text-center text-xs font-semibold select-none">
                          {item.quantity}
                        </span>
                        <button
                          onClick={() => updateQuantity(item.product_id, item.quantity + 1)}
                          disabled={isLoading}
                          className="p-1.5 hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                        >
                          <Plus className="h-3 w-3" />
                        </button>
                      </div>

                      <button
                        onClick={() => removeItem(item.product_id)}
                        disabled={isLoading}
                        className="p-1.5 text-muted-foreground hover:text-destructive transition-colors"
                        title="Remove"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* Footer Summary & Checkout */}
            {!isEmpty && cart && (
              <div className="border-t p-6 bg-muted/20 space-y-4">
                <div className="space-y-1.5 text-sm">
                  <div className="flex justify-between text-muted-foreground">
                    <span>Subtotal</span>
                    <span>{formatCurrency(cart.total_amount)}</span>
                  </div>
                  <div className="flex justify-between text-muted-foreground">
                    <span>Estimated Delivery</span>
                    <span className="text-emerald-600 font-medium">Free</span>
                  </div>
                  <div className="flex justify-between text-base font-bold text-foreground pt-2 border-t">
                    <span>Total</span>
                    <span>{formatCurrency(cart.total_amount)}</span>
                  </div>
                </div>

                <Button
                  onClick={handleProceedCheckout}
                  className="w-full h-11 text-base font-semibold flex items-center justify-center gap-2"
                >
                  <span>Proceed to Checkout</span>
                  <ArrowRight className="h-4 w-4" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Checkout and Auth Modals */}
      <CheckoutModal open={checkoutOpen} onOpenChange={setCheckoutOpen} />
      <AuthModal open={authOpen} onOpenChange={setAuthOpen} />
    </>
  );
}
