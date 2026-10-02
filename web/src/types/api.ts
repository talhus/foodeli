export type Role = "customer" | "merchant" | "admin";

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
}

export interface User {
  id: string;
  email: string;
  role: Role;
  created_at?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Store {
  id: string;
  merchant_id: string;
  name: string;
  address: string;
  created_at: string;
  updated_at: string;
}

export interface MenuItem {
  id: string;
  store_id: string;
  name: string;
  description: string;
  price: number; // floating fallback for menu items
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  store_id: string;
  name: string;
  description: string;
  price: number; // Minor units (cents), e.g. 1250 = €12.50
  stock: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface CartItem {
  product_id: string;
  product_name: string;
  unit_price: number; // Minor units (cents)
  quantity: number;
  subtotal: number; // Minor units (cents)
}

export interface Cart {
  user_id: string;
  store_id?: string;
  store_name?: string;
  items: CartItem[];
  total_amount: number; // Minor units (cents)
  item_count: number;
  updated_at: string;
}

export type OrderStatus =
  | "PENDING"
  | "CONFIRMED"
  | "PREPARING"
  | "DELIVERED"
  | "CANCELLED";

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  product_name: string;
  unit_price: number;
  quantity: number;
  subtotal: number;
  created_at: string;
}

export interface Order {
  id: string;
  user_id: string;
  store_id: string;
  delivery_address: string;
  total_amount: number; // Minor units (cents)
  status: OrderStatus;
  items?: OrderItem[];
  idempotency_key: string;
  created_at: string;
  updated_at: string;
}

export interface Payment {
  id: string;
  order_id: string;
  user_id: string;
  amount: number;
  currency: string;
  provider: string;
  payment_method: string;
  transaction_id: string;
  status: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
}

export interface CheckoutSessionResponse {
  session_id: string;
  checkout_url: string;
  status: string;
  expires_at?: string;
}
