"use client";

import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Cpu, CheckCircle2, Layers } from "lucide-react";

export function AdminArchitectureCard() {
  const subsystems = [
    { name: "Auth & RBAC Service", tech: "JWT + Bcrypt + Redis Limiter", status: "Operational" },
    { name: "Store & Catalog Domain", tech: "PostgreSQL with Multi-Store Indexing", status: "Operational" },
    { name: "Order Processing Engine", tech: "Transactional Outbox Pattern + State Machine", status: "Operational" },
    { name: "Payment Gateway Core", tech: "TyrFintech B2B + Webhook HMAC/Workers", status: "Operational" },
    { name: "Cart & Fast Cache", tech: "Redis Cluster (TTL Session Key-Value)", status: "Operational" },
    { name: "Async Event Dispatcher", tech: "RabbitMQ (foodeli.events exchange)", status: "Operational" },
  ];

  return (
    <Card>
      <CardHeader className="p-5 pb-3 border-b">
        <div className="flex items-center justify-between">
          <div className="space-y-0.5">
            <CardTitle className="text-base font-semibold flex items-center gap-2">
              <Layers className="h-4 w-4 text-purple-500" />
              <span>Modular Monolith Subsystems</span>
            </CardTitle>
            <CardDescription className="text-xs">
              De-coupled domain boundaries with event-driven asynchronous message relay
            </CardDescription>
          </div>
          <Badge variant="outline" className="gap-1 text-emerald-600 border-emerald-500/30 bg-emerald-500/5 text-xs">
            <Cpu className="h-3 w-3" />
            <span>Architecture v2.0</span>
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="p-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {subsystems.map((sub, i) => (
            <div key={i} className="p-3 rounded-lg border bg-muted/20 flex items-center justify-between">
              <div className="space-y-0.5">
                <p className="text-xs font-semibold text-foreground">{sub.name}</p>
                <p className="text-[11px] text-muted-foreground font-mono">{sub.tech}</p>
              </div>
              <div className="flex items-center gap-1 text-[11px] text-emerald-600 font-medium shrink-0">
                <CheckCircle2 className="h-3.5 w-3.5" />
                <span>{sub.status}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
