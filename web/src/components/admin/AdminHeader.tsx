"use client";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { RefreshCw, ShieldAlert, Activity } from "lucide-react";

interface AdminHeaderProps {
  onRefresh: () => void;
  isLoading: boolean;
  adminEmail?: string;
}

export function AdminHeader({ onRefresh, isLoading, adminEmail }: AdminHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b">
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <div className="p-1.5 rounded-lg bg-purple-500/10 text-purple-600 dark:text-purple-400">
            <ShieldAlert className="h-5 w-5" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight">Platform Administration</h1>
          <Badge variant="outline" className="text-purple-600 border-purple-500/30 bg-purple-500/5 text-[10px]">
            Superadmin
          </Badge>
        </div>
        <p className="text-xs text-muted-foreground">
          System overview, multi-store registry & architecture health • Signed in as <span className="font-semibold text-foreground">{adminEmail || "admin@foodeli.fi"}</span>
        </p>
      </div>

      <div className="flex items-center gap-2">
        <Badge variant="secondary" className="gap-1.5 py-1 px-2.5 text-xs text-emerald-600 bg-emerald-500/10 border border-emerald-500/20">
          <Activity className="h-3.5 w-3.5" />
          <span>Finland Core API Online</span>
        </Badge>
        <Button onClick={onRefresh} variant="outline" size="sm" disabled={isLoading} className="gap-1.5 h-8">
          <RefreshCw className={`h-3.5 w-3.5 ${isLoading ? "animate-spin" : ""}`} />
          <span>Sync</span>
        </Button>
      </div>
    </div>
  );
}
