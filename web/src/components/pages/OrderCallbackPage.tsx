"use client";

import { Suspense, useEffect, useState, useRef, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { AlertCircle, Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { CallbackStatusIcon } from "@/components/order-callback/CallbackStatusIcon";
import { CallbackDetails } from "@/components/order-callback/CallbackDetails";
import { CallbackActions } from "@/components/order-callback/CallbackActions";

function CallbackContent() {
  const searchParams = useSearchParams();
  const status = searchParams?.get("status");
  const orderId = searchParams?.get("order_id");
  const transactionId = searchParams?.get("transaction_id");
  const sessionId = searchParams?.get("session_id");

  const isSuccess = status === "SUCCEEDED" || status === "SUCCESS";
  const [isSettling, setIsSettling] = useState(false);
  const [settleError, setSettleError] = useState<string | null>(null);
  const settledRef = useRef(false);

  const settleOrder = useCallback(async () => {
    if (!orderId) return;
    setIsSettling(true);
    setSettleError(null);
    try {
      await api.webhooks.sendTyrFintech({
        event: "checkout.session.paid",
        transaction_id: transactionId || sessionId || `txn_cs_${orderId.slice(0, 8)}`,
        session_id: sessionId || undefined,
        order_id: orderId,
        status: "SUCCEEDED",
      });
    } catch (err: unknown) {
      console.error("Order settlement error:", err);
      setSettleError("Failed to confirm order with kitchen. You can still check your order status.");
    } finally {
      setIsSettling(false);
    }
  }, [orderId, transactionId, sessionId]);

  useEffect(() => {
    if (isSuccess && orderId && !settledRef.current) {
      settledRef.current = true;
      settleOrder();
    }
  }, [isSuccess, orderId, settleOrder]);

  return (
    <div className="max-w-md mx-auto pt-8">
      <Card className="text-center shadow-lg border-primary/20">
        <CardHeader className="p-8 pb-4 flex flex-col items-center">
          <CallbackStatusIcon isSettling={isSettling} isSuccess={isSuccess} />
          <CardTitle className="text-2xl font-bold">
            {isSettling ? "Confirming Payment..." : isSuccess ? "Payment Succeeded!" : "Payment Declined"}
          </CardTitle>
          <CardDescription className="text-sm mt-1">
            {isSettling
              ? "We are verifying your transaction and forwarding the order to the restaurant."
              : isSuccess
              ? "Your food order has been confirmed and forwarded to the kitchen."
              : "TyrFintech was unable to settle this transaction."}
          </CardDescription>
        </CardHeader>

        <CardContent className="p-8 pt-0 space-y-3">
          {settleError && (
            <div className="mb-2 flex items-center gap-2 rounded-lg bg-amber-500/10 p-3 text-xs text-amber-600 border border-amber-500/20 text-left">
              <AlertCircle className="h-4 w-4 shrink-0" />
              <span>{settleError}</span>
            </div>
          )}
          <CallbackDetails
            orderId={orderId || null}
            transactionRef={transactionId || sessionId || null}
            status={status || null}
            isSettling={isSettling}
            isSuccess={isSuccess}
          />
        </CardContent>

        <CardFooter className="p-8 pt-0 flex flex-col sm:flex-row gap-2">
          <CallbackActions isSuccess={isSuccess} isSettling={isSettling} />
        </CardFooter>
      </Card>
    </div>
  );
}

export function OrderCallbackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center p-20">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      }
    >
      <CallbackContent />
    </Suspense>
  );
}
