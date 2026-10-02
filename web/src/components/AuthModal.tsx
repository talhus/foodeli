"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Dialog, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAuthStore } from "@/store/useAuthStore";
import { ApiError } from "@/lib/api";
import { AlertCircle, CheckCircle2, Loader2, User, Store, Sparkles, ShieldAlert } from "lucide-react";
import { toast } from "sonner";

interface AuthModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuthModal({ open, onOpenChange }: AuthModalProps) {
  const router = useRouter();
  const [mode, setMode] = useState<"login" | "register">("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState<"customer" | "merchant">("customer");
  const [error, setError] = useState<string | null>(null);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);
  const [quickLoading, setQuickLoading] = useState<"customer" | "merchant" | "admin" | null>(null);

  const { login, register, isLoading } = useAuthStore();

  const resetForm = () => {
    setEmail("");
    setPassword("");
    setError(null);
    setSuccessMsg(null);
    setQuickLoading(null);
  };

  const handleClose = (newOpen: boolean) => {
    if (!newOpen) {
      resetForm();
    }
    onOpenChange(newOpen);
  };

  const handleQuickLogin = async (demoEmail: string, demoPass: string, targetRole: "customer" | "merchant" | "admin") => {
    setError(null);
    setSuccessMsg(null);
    setQuickLoading(targetRole);

    try {
      await login(demoEmail, demoPass);
      toast.success(`Welcome! Logged in as Test ${targetRole.charAt(0).toUpperCase() + targetRole.slice(1)}`);
      resetForm();
      onOpenChange(false);
      if (targetRole === "merchant") {
        router.push("/merchant");
      } else if (targetRole === "admin") {
        router.push("/admin");
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message || "Failed to log in with test credentials.");
      } else {
        setError("Failed to log in with test credentials.");
      }
    } finally {
      setQuickLoading(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccessMsg(null);

    try {
      if (mode === "login") {
        await login(email, password);
        resetForm();
        onOpenChange(false);
      } else {
        await register(email, password, role);
        setSuccessMsg("Account successfully registered! You are now logged in.");
        setTimeout(() => {
          resetForm();
          onOpenChange(false);
        }, 1200);
      }
    } catch (err: unknown) {
      if (err instanceof ApiError && err.status === 429) {
        setError("Rate limit exceeded: Too many attempts. Please slow down and try again later.");
      } else if (err instanceof Error) {
        setError(err.message || "Authentication failed. Please check your credentials.");
      } else {
        setError("Authentication failed. Please check your credentials.");
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogHeader>
        <DialogTitle className="text-xl">
          {mode === "login" ? "Sign in to Foodeli" : "Create an Account"}
        </DialogTitle>
        <DialogDescription>
          {mode === "login"
            ? "Enter your credentials to manage orders and access your cart."
            : "Register as a customer to order food, or as a merchant to manage a storefront."}
        </DialogDescription>
      </DialogHeader>

      {error && (
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-destructive/10 p-3 text-sm text-destructive border border-destructive/20">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {successMsg && (
        <div className="mb-4 flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3 text-sm text-emerald-600 border border-emerald-500/20">
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{successMsg}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Email Address
          </label>
          <Input
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            disabled={isLoading || quickLoading !== null}
          />
        </div>

        <div className="space-y-1.5">
          <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Password
          </label>
          <Input
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            disabled={isLoading || quickLoading !== null}
          />
        </div>

        {mode === "register" && (
          <div className="space-y-1.5">
            <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              Account Role
            </label>
            <div className="grid grid-cols-2 gap-2">
              <Button
                type="button"
                variant={role === "customer" ? "default" : "outline"}
                size="sm"
                onClick={() => setRole("customer")}
                className="w-full"
              >
                Customer
              </Button>
              <Button
                type="button"
                variant={role === "merchant" ? "default" : "outline"}
                size="sm"
                onClick={() => setRole("merchant")}
                className="w-full"
              >
                Merchant (Store)
              </Button>
            </div>
          </div>
        )}

        <Button type="submit" className="w-full mt-2 font-semibold" disabled={isLoading || quickLoading !== null}>
          {isLoading && quickLoading === null ? (
            <span className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" />
              Processing...
            </span>
          ) : mode === "login" ? (
            "Sign In"
          ) : (
            "Create Account"
          )}
        </Button>

        {/* Recruiter & Demo 1-Click Access */}
        {mode === "login" && (
          <div className="pt-3 border-t space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground flex items-center gap-1.5">
                <Sparkles className="h-3.5 w-3.5 text-amber-500" />
                Recruiter 1-Click Access
              </span>
              <span className="text-[10px] text-muted-foreground">Quick Demo</span>
            </div>

            <div className="grid grid-cols-3 gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => handleQuickLogin("customer@foodeli.fi", "password123", "customer")}
                disabled={isLoading || quickLoading !== null}
                className="flex items-center justify-center gap-1 text-[11px] h-9 px-1.5 bg-muted/40 hover:bg-muted hover:border-blue-500/40 transition-all font-medium"
              >
                {quickLoading === "customer" ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <User className="h-3.5 w-3.5 text-blue-500 shrink-0" />
                )}
                <span>Test User</span>
              </Button>

              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => handleQuickLogin("merchant@foodeli.fi", "password123", "merchant")}
                disabled={isLoading || quickLoading !== null}
                className="flex items-center justify-center gap-1 text-[11px] h-9 px-1.5 bg-muted/40 hover:bg-muted hover:border-amber-500/40 transition-all font-medium"
              >
                {quickLoading === "merchant" ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <Store className="h-3.5 w-3.5 text-amber-500 shrink-0" />
                )}
                <span>Test Merchant</span>
              </Button>

              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => handleQuickLogin("admin@foodeli.fi", "password123", "admin")}
                disabled={isLoading || quickLoading !== null}
                className="flex items-center justify-center gap-1 text-[11px] h-9 px-1.5 bg-muted/40 hover:bg-muted hover:border-purple-500/40 transition-all font-medium"
              >
                {quickLoading === "admin" ? (
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                ) : (
                  <ShieldAlert className="h-3.5 w-3.5 text-purple-500 shrink-0" />
                )}
                <span>Test Admin</span>
              </Button>
            </div>
          </div>
        )}
      </form>

      <div className="mt-5 text-center text-sm text-muted-foreground">
        {mode === "login" ? (
          <p>
            Don&apos;t have an account yet?{" "}
            <button
              type="button"
              onClick={() => {
                resetForm();
                setMode("register");
              }}
              className="font-semibold text-primary hover:underline"
            >
              Register here
            </button>
          </p>
        ) : (
          <p>
            Already have an account?{" "}
            <button
              type="button"
              onClick={() => {
                resetForm();
                setMode("login");
              }}
              className="font-semibold text-primary hover:underline"
            >
              Sign in instead
            </button>
          </p>
        )}
      </div>
    </Dialog>
  );
}
