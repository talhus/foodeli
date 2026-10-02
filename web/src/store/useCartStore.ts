import { create } from "zustand";
import { Cart } from "@/types/api";
import { api, ApiError } from "@/lib/api";

interface ConflictState {
  isOpen: boolean;
  pendingStoreId?: string;
  pendingProductId?: string;
  pendingQuantity?: number;
}

interface CartState {
  cart: Cart | null;
  isOpen: boolean;
  isLoading: boolean;
  conflict: ConflictState;

  toggleCart: (open?: boolean) => void;
  fetchCart: () => Promise<void>;
  addItem: (storeId: string, productId: string, quantity: number, force?: boolean) => Promise<boolean>;
  updateQuantity: (productId: string, quantity: number) => Promise<void>;
  removeItem: (productId: string) => Promise<void>;
  clearCart: () => Promise<void>;
  resolveConflict: (force: boolean) => Promise<void>;
  closeConflict: () => void;
}

export const useCartStore = create<CartState>((set, get) => ({
  cart: null,
  isOpen: false,
  isLoading: false,
  conflict: { isOpen: false },

  toggleCart: (open?: boolean) => {
    set((state) => ({ isOpen: open !== undefined ? open : !state.isOpen }));
  },

  fetchCart: async () => {
    try {
      const cart = await api.cart.get();
      set({ cart });
    } catch {
      // User might be logged out or cart empty
      set({ cart: null });
    }
  },

  addItem: async (storeId: string, productId: string, quantity: number, force: boolean = false) => {
    set({ isLoading: true });
    try {
      const updatedCart = await api.cart.addItem(storeId, productId, quantity, force);
      set({ cart: updatedCart, isLoading: false, isOpen: true });
      return true;
    } catch (err: unknown) {
      set({ isLoading: false });
      if (err instanceof ApiError && err.code === "CONFLICT") {
        // Trigger multi-store conflict confirmation modal
        set({
          conflict: {
            isOpen: true,
            pendingStoreId: storeId,
            pendingProductId: productId,
            pendingQuantity: quantity,
          },
        });
        return false;
      }
      throw err;
    }
  },

  updateQuantity: async (productId: string, quantity: number) => {
    if (quantity <= 0) {
      await get().removeItem(productId);
      return;
    }
    set({ isLoading: true });
    try {
      const updatedCart = await api.cart.updateQuantity(productId, quantity);
      set({ cart: updatedCart, isLoading: false });
    } catch (err) {
      set({ isLoading: false });
      throw err;
    }
  },

  removeItem: async (productId: string) => {
    set({ isLoading: true });
    try {
      const updatedCart = await api.cart.removeItem(productId);
      set({ cart: updatedCart, isLoading: false });
    } catch (err) {
      set({ isLoading: false });
      throw err;
    }
  },

  clearCart: async () => {
    set({ isLoading: true });
    try {
      await api.cart.clear();
      set({ cart: null, isLoading: false });
    } catch (err) {
      set({ isLoading: false });
      throw err;
    }
  },

  resolveConflict: async (force: boolean) => {
    const { pendingStoreId, pendingProductId, pendingQuantity } = get().conflict;
    if (force && pendingStoreId && pendingProductId && pendingQuantity) {
      set({ conflict: { isOpen: false } });
      await get().addItem(pendingStoreId, pendingProductId, pendingQuantity, true);
    } else {
      set({ conflict: { isOpen: false } });
    }
  },

  closeConflict: () => {
    set({ conflict: { isOpen: false } });
  },
}));
