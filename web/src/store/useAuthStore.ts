import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { User } from "@/types/api";
import { api } from "@/lib/api";

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  setAuth: (user: User, token: string) => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, role: "customer" | "merchant") => Promise<void>;
  logout: () => void;
  initAuth: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      isLoading: false,

      setAuth: (user: User, token: string) => {
        if (typeof window !== "undefined") {
          localStorage.setItem("foodeli_token", token);
        }
        set({ user, token, isAuthenticated: true });
      },

      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const res = await api.auth.login(email, password);
          if (typeof window !== "undefined") {
            localStorage.setItem("foodeli_token", res.token);
          }
          set({
            user: res.user,
            token: res.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (err) {
          set({ isLoading: false });
          throw err;
        }
      },

      register: async (email: string, password: string, role: "customer" | "merchant") => {
        set({ isLoading: true });
        try {
          const res = await api.auth.register(email, password, role);
          if (typeof window !== "undefined") {
            localStorage.setItem("foodeli_token", res.token);
          }
          set({
            user: res.user,
            token: res.token,
            isAuthenticated: true,
            isLoading: false,
          });
        } catch (err) {
          set({ isLoading: false });
          throw err;
        }
      },

      logout: () => {
        if (typeof window !== "undefined") {
          localStorage.removeItem("foodeli_token");
        }
        api.auth.logout().catch(() => {});
        set({ user: null, token: null, isAuthenticated: false });
      },

      initAuth: async () => {
        const storedToken = typeof window !== "undefined" ? localStorage.getItem("foodeli_token") : null;
        const token = get().token || storedToken;
        if (!token) {
          set({ isAuthenticated: false, user: null, token: null });
          return;
        }

        try {
          const user = await api.auth.getProfile(token);
          set({ user, token, isAuthenticated: true });
        } catch {
          // Token expired or invalid
          get().logout();
        }
      },
    }),
    {
      name: "foodeli_auth",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ token: state.token, user: state.user, isAuthenticated: state.isAuthenticated }),
    }
  )
);
