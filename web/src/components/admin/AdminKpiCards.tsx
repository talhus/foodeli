"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Store, Package, Server, Globe } from "lucide-react";

interface AdminKpiCardsProps {
  totalStores: number;
  totalProducts: number;
  healthStatus: string;
}

export function AdminKpiCards({ totalStores, totalProducts, healthStatus }: AdminKpiCardsProps) {
  const kpis = [
    {
      title: "Registered Storefronts",
      value: totalStores,
      subtitle: "Active restaurants in Finland",
      icon: Store,
      color: "text-blue-500",
      bg: "bg-blue-500/10",
    },
    {
      title: "Catalog Products",
      value: totalProducts,
      subtitle: "Items across all restaurant menus",
      icon: Package,
      color: "text-amber-500",
      bg: "bg-amber-500/10",
    },
    {
      title: "Core Service Health",
      value: healthStatus.toUpperCase(),
      subtitle: "Go Monolith (gin + pgxpool)",
      icon: Server,
      color: "text-emerald-500",
      bg: "bg-emerald-500/10",
    },
    {
      title: "Regional Locale",
      value: "FI / EUR (€)",
      subtitle: "Helsinki & Tampere operations",
      icon: Globe,
      color: "text-purple-500",
      bg: "bg-purple-500/10",
    },
  ];

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {kpis.map((kpi, idx) => {
        const Icon = kpi.icon;
        return (
          <Card key={idx} className="overflow-hidden">
            <CardContent className="p-4 flex items-center justify-between">
              <div>
                <p className="text-xs font-medium text-muted-foreground">{kpi.title}</p>
                <p className="text-xl font-bold tracking-tight text-foreground mt-0.5">{kpi.value}</p>
                <p className="text-[11px] text-muted-foreground mt-0.5">{kpi.subtitle}</p>
              </div>
              <div className={`p-2.5 rounded-xl ${kpi.bg} ${kpi.color} shrink-0`}>
                <Icon className="h-5 w-5" />
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
