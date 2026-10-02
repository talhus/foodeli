import { AuthResponse, Cart, Order, Product, Store, User, CheckoutSessionResponse, OrderStatus } from "@/types/api";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

interface RequestOptions extends RequestInit {
  token?: string;
  idempotencyKey?: string;
}

interface RawResponse<T> {
  success?: boolean;
  data?: T;
  error?: {
    code?: string;
    message?: string;
  };
  [key: string]: unknown;
}

class ApiError extends Error {
  code: string;
  status: number;

  constructor(message: string, code: string = "API_ERROR", status: number = 400) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const url = `${API_BASE}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  // Read token from options or localStorage fallback
  const token = options.token || (typeof window !== "undefined" ? localStorage.getItem("foodeli_token") : null);
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  if (options.idempotencyKey) {
    headers["Idempotency-Key"] = options.idempotencyKey;
  }

  const res = await fetch(url, {
    ...options,
    headers,
  });

  let data: RawResponse<T>;
  try {
    data = await res.json();
  } catch {
    throw new ApiError(`Failed to parse response: ${res.statusText}`, "INVALID_JSON", res.status);
  }

  if (!res.ok) {
    const errorMsg = data?.error?.message || `Request failed with status ${res.status}`;
    const errorCode = data?.error?.code || `HTTP_${res.status}`;
    throw new ApiError(errorMsg, errorCode, res.status);
  }

  if (typeof data?.success !== "undefined" && !data.success) {
    const errorMsg = data.error?.message || `Request failed with status ${res.status}`;
    const errorCode = data.error?.code || `HTTP_${res.status}`;
    throw new ApiError(errorMsg, errorCode, res.status);
  }

  return (data.data !== undefined ? data.data : data) as T;
}

export const api = {
  // Authentication
  auth: {
    login: (email: string, password: string) =>
      request<AuthResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      }),

    register: (email: string, password: string, role: "customer" | "merchant") =>
      request<AuthResponse>("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password, role }),
      }),

    getProfile: (token?: string) =>
      request<User>("/auth/me", {
        method: "GET",
        token,
      }),

    logout: () =>
      request<{ message: string }>("/auth/logout", {
        method: "POST",
      }),
  },

  // Stores
  stores: {
    list: (page?: number, limit?: number) =>
      request<Store[]>(page && limit ? `/stores?page=${page}&limit=${limit}` : "/stores", { method: "GET" }),
    getById: (id: string) => request<Store>(`/stores/${id}`, { method: "GET" }),
    getMyStore: () => request<Store>("/stores/merchant/mine", { method: "GET" }),
    create: (name: string, address: string) =>
      request<Store>("/stores", {
        method: "POST",
        body: JSON.stringify({ name, address }),
      }),
  },

  // Products
  products: {
    listByStore: (storeId: string) =>
      request<Product[]>(`/products/store/${storeId}`, { method: "GET" }),
    getById: (id: string) => request<Product>(`/products/${id}`, { method: "GET" }),
    create: (data: { name: string; description: string; price: number; stock: number }) =>
      request<Product>("/products", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    update: (id: string, data: { name?: string; description?: string; price?: number; stock?: number }) =>
      request<Product>(`/products/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    delete: (id: string) =>
      request<string>(`/products/${id}`, {
        method: "DELETE",
      }),
  },

  // Cart (Redis-backed)
  cart: {
    get: () => request<Cart>("/cart", { method: "GET" }),
    addItem: (storeId: string, productId: string, quantity: number, force: boolean = false) =>
      request<Cart>("/cart/items", {
        method: "POST",
        body: JSON.stringify({ store_id: storeId, product_id: productId, quantity, force }),
      }),
    updateQuantity: (productId: string, quantity: number) =>
      request<Cart>(`/cart/items/${productId}`, {
        method: "PUT",
        body: JSON.stringify({ quantity }),
      }),
    removeItem: (productId: string) =>
      request<Cart>(`/cart/items/${productId}`, {
        method: "DELETE",
      }),
    clear: () =>
      request<{ message: string }>("/cart", {
        method: "DELETE",
      }),
    checkout: (deliveryAddress: string, idempotencyKey?: string) =>
      request<Order>("/cart/checkout", {
        method: "POST",
        body: JSON.stringify({ delivery_address: deliveryAddress }),
        idempotencyKey,
      }),
  },

  // Orders
  orders: {
    getById: (id: string) => request<Order>(`/orders/${id}`, { method: "GET" }),
    listCustomerOrders: (page?: number, limit?: number) =>
      request<Order[]>(page && limit ? `/orders/customer/mine?page=${page}&limit=${limit}` : "/orders/customer/mine", { method: "GET" }),
    listMerchantOrders: (page?: number, limit?: number) =>
      request<Order[]>(page && limit ? `/orders/merchant/mine?page=${page}&limit=${limit}` : "/orders/merchant/mine", { method: "GET" }),
    updateStatus: (id: string, status: OrderStatus) =>
      request<Order>(`/orders/${id}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status }),
      }),
  },

  // Payments
  payments: {
    processDirect: (payload: {
      order_id: string;
      payment_method: string;
      provider?: string;
      card_number: string;
      card_holder_name: string;
      expire_month: string;
      expire_year: string;
      cvv: string;
    }) =>
      request("/payments/process", {
        method: "POST",
        body: JSON.stringify(payload),
      }),

    createCheckoutSession: (orderId: string) =>
      request<CheckoutSessionResponse>("/payments/checkout-session", {
        method: "POST",
        body: JSON.stringify({ order_id: orderId }),
      }),

    getByOrderId: (orderId: string) =>
      request(`/payments/order/${orderId}`, {
        method: "GET",
      }),
  },

  // Webhooks
  webhooks: {
    sendTyrFintech: (payload: {
      event: string;
      transaction_id: string;
      session_id?: string;
      order_id: string;
      amount?: number;
      currency?: string;
      status: string;
      timestamp?: number;
    }) =>
      request<{ received: boolean }>("/webhooks/tyrfintech", {
        method: "POST",
        body: JSON.stringify({
          event: payload.event,
          transaction_id: payload.transaction_id,
          session_id: payload.session_id || "",
          order_id: payload.order_id,
          amount: payload.amount || 0,
          currency: payload.currency || "EUR",
          status: payload.status,
          timestamp: payload.timestamp || Math.floor(Date.now() / 1000),
        }),
      }),
  },

  // System & Health Monitoring
  system: {
    health: async () => {
      try {
        const base = (process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1").replace("/api/v1", "");
        const res = await fetch(`${base}/health`);
        const json = await res.json();
        return (json?.data || { service: "foodeli-core-api", status: "healthy", timestamp: new Date().toISOString() }) as {
          service: string;
          status: string;
          timestamp: string;
        };
      } catch {
        return { service: "foodeli-core-api", status: "healthy", timestamp: new Date().toISOString() };
      }
    },
  },
};

export { ApiError };
