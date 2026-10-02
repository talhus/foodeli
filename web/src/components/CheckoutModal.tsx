"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Dialog, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useCartStore } from "@/store/useCartStore";
import { formatCurrency } from "@/lib/utils";
import { api, ApiError } from "@/lib/api";
import { CreditCard, Wallet, AlertCircle, CheckCircle2, Loader2, ArrowRight } from "lucide-react";

interface CheckoutModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CheckoutModal({ open, onOpenChange }: CheckoutModalProps) {
  const router = useRouter();
  const { cart, clearCart } = useCartStore();

  const [address, setAddress] = useState("Kruisstraat 42, 1012 CR Amsterdam");
  const [paymentMethod, setPaymentMethod] = useState<"TYR_FINTECH" | "CREDIT_CARD">("TYR_FINTECH");

  // Direct Card fields
  const [cardNumber, setCardNumber] = useState("5555444433332222");
  const [cardHolder, setCardHolder] = useState("Jane Doe");
  const [expireMonth, setExpireMonth] = useState("12");
  const [expireYear, setExpireYear] = useState("28");
  const [cvv, setCvv] = useState("123");

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  if (!cart) return null;

  const handleCheckout = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setIsLoading(true);

    try {
      // Step 1: Create Order from Cart
      const idempotencyKey = crypto.randomUUID();
      const order = await api.cart.checkout(address, idempotencyKey);

      // Step 2: Handle Payment Method
      if (paymentMethod === "TYR_FINTECH") {
        // Hosted Checkout Session
        const session = await api.payments.createCheckoutSession(order.id);
        // Clear local cart
        await clearCart();
        // Redirect to TyrFintech checkout page
        window.location.href = session.checkout_url;
      } else {
        // Direct Card Payment
        await api.payments.processDirect({
          order_id: order.id,
          payment_method: "CREDIT_CARD",
          provider: "TYR_FINTECH",
          card_number: cardNumber.replace(/\s+/g, ""),
          card_holder_name: cardHolder,
          expire_month: expireMonth,
          expire_year: expireYear,
          cvv,
        });

        await clearCart();
        setSuccess(`Payment approved! Order #${order.id.slice(0, 8)} confirmed.`);
        setTimeout(() => {
          onOpenChange(false);
          router.push(`/orders`);
        }, 1500);
      }
    } catch (err: unknown) {
      if (err instanceof ApiError && err.status === 402) {
        setError(err.message || "Card declined. Please check card details or use TyrFintech Wallet.");
      } else if (err instanceof Error) {
        setError(err.message || "Checkout failed. Please try again.");
      } else {
        setError("Checkout failed. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogHeader>
        <DialogTitle className="text-xl">Checkout & Payment</DialogTitle>
        <DialogDescription>
          Complete your order with {cart.store_name || "restaurant"}. Total:{" "}
          <strong className="text-foreground">{formatCurrency(cart.total_amount)}</strong>
        </DialogDescription>
      </DialogHeader>

      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive border border-destructive/20">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {success && (
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3 text-sm text-emerald-600 border border-emerald-500/20">
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{success}</span>
        </div>
      )}

      <form onSubmit={handleCheckout} className="space-y-4">
        {/* Delivery Address */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Delivery Address
          </label>
          <Input
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder="Street name, number, postal code, city"
            required
            disabled={isLoading}
          />
        </div>

        {/* Payment Provider Selection */}
        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Select Payment Method
          </label>
          <div className="grid grid-cols-2 gap-2">
            <button
              type="button"
              onClick={() => setPaymentMethod("TYR_FINTECH")}
              className={`flex flex-col items-center justify-center p-3 rounded-lg border text-sm font-medium transition-all ${
                paymentMethod === "TYR_FINTECH"
                  ? "border-primary bg-primary/10 text-primary ring-1 ring-primary"
                  : "border-input hover:bg-accent"
              }`}
            >
              <Wallet className="h-5 w-5 mb-1.5" />
              <span>TyrFintech Wallet</span>
              <span className="text-[10px] text-muted-foreground">Hosted Checkout</span>
            </button>

            <button
              type="button"
              onClick={() => setPaymentMethod("CREDIT_CARD")}
              className={`flex flex-col items-center justify-center p-3 rounded-lg border text-sm font-medium transition-all ${
                paymentMethod === "CREDIT_CARD"
                  ? "border-primary bg-primary/10 text-primary ring-1 ring-primary"
                  : "border-input hover:bg-accent"
              }`}
            >
              <CreditCard className="h-5 w-5 mb-1.5" />
              <span>Direct Card</span>
              <span className="text-[10px] text-muted-foreground">Instant Processing</span>
            </button>
          </div>
        </div>

        {/* Direct Card Form if selected */}
        {paymentMethod === "CREDIT_CARD" && (
          <div className="space-y-3 rounded-lg border bg-muted/30 p-3 animate-in fade-in duration-150">
            <div>
              <label className="text-[11px] font-medium text-muted-foreground">Card Number</label>
              <Input
                value={cardNumber}
                onChange={(e) => setCardNumber(e.target.value)}
                placeholder="16 digits"
                maxLength={16}
                required
                disabled={isLoading}
              />
            </div>
            <div>
              <label className="text-[11px] font-medium text-muted-foreground">Cardholder Name</label>
              <Input
                value={cardHolder}
                onChange={(e) => setCardHolder(e.target.value)}
                placeholder="Name on card"
                required
                disabled={isLoading}
              />
            </div>
            <div className="grid grid-cols-3 gap-2">
              <div>
                <label className="text-[11px] font-medium text-muted-foreground">Month (MM)</label>
                <Input
                  value={expireMonth}
                  onChange={(e) => setExpireMonth(e.target.value)}
                  placeholder="12"
                  maxLength={2}
                  required
                  disabled={isLoading}
                />
              </div>
              <div>
                <label className="text-[11px] font-medium text-muted-foreground">Year (YY)</label>
                <Input
                  value={expireYear}
                  onChange={(e) => setExpireYear(e.target.value)}
                  placeholder="28"
                  maxLength={4}
                  required
                  disabled={isLoading}
                />
              </div>
              <div>
                <label className="text-[11px] font-medium text-muted-foreground">CVV</label>
                <Input
                  value={cvv}
                  onChange={(e) => setCvv(e.target.value)}
                  placeholder="123"
                  maxLength={4}
                  required
                  disabled={isLoading}
                />
              </div>
            </div>
          </div>
        )}

        {/* Submit Button */}
        <Button type="submit" className="w-full mt-4 h-11 text-base font-semibold" disabled={isLoading}>
          {isLoading ? (
            <span className="flex items-center gap-2">
              <Loader2 className="h-5 w-5 animate-spin" />
              Processing Order...
            </span>
          ) : paymentMethod === "TYR_FINTECH" ? (
            <span className="flex items-center gap-2">
              Pay with TyrFintech ({formatCurrency(cart.total_amount)})
              <ArrowRight className="h-4 w-4" />
            </span>
          ) : (
            `Pay ${formatCurrency(cart.total_amount)} Now`
          )}
        </Button>
      </form>
    </Dialog>
  );
}
