"use client";

import { useEffect } from "react";
import { toast } from "sonner";
import { useAuthStore } from "@/store/useAuthStore";

export interface NotificationEvent {
  type: "order.created" | "payment.succeeded" | string;
  title: string;
  message: string;
  order_id: string;
  created_at: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export function useNotifications() {
  const { user, token } = useAuthStore();

  useEffect(() => {
    // Only establish SSE stream when user is authenticated with a token
    if (!user || !token) return;

    const streamUrl = `${API_BASE}/notifications/stream?token=${encodeURIComponent(token)}`;
    const eventSource = new EventSource(streamUrl);

    eventSource.onopen = () => {
      console.log("[SSE] Connected to real-time notification stream");
    };

    eventSource.addEventListener("message", (e) => {
      try {
        const data: NotificationEvent = JSON.parse(e.data);

        if (data.type === "order.created") {
          toast.info(data.title || "Order Created", {
            description: data.message,
            duration: 6000,
          });
        } else if (data.type === "payment.succeeded") {
          toast.success(data.title || "Payment Successful", {
            description: data.message,
            duration: 6000,
          });
        } else {
          toast(data.title || "Notification", {
            description: data.message,
            duration: 5000,
          });
        }

        // Broadcast custom window event so active pages (e.g., OrdersPage, KitchenOrdersView) can auto-refresh data
        if (typeof window !== "undefined") {
          window.dispatchEvent(new CustomEvent("foodeli:notification", { detail: data }));
        }
      } catch (err) {
        console.error("[SSE] Failed to parse notification payload:", err);
      }
    });

    eventSource.onerror = (err) => {
      console.warn("[SSE] Notification stream reconnecting...", err);
    };

    return () => {
      eventSource.close();
      console.log("[SSE] Closed notification stream connection");
    };
  }, [user, token]);
}
