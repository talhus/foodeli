"use client";

import { useState } from "react";
import { Store as StoreType } from "@/types/api";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Search, MapPin, Store } from "lucide-react";

interface AdminStoresTableProps {
  stores: StoreType[];
  storeProductCounts: Record<string, number>;
}

export function AdminStoresTable({ stores, storeProductCounts }: AdminStoresTableProps) {
  const [search, setSearch] = useState("");

  const filteredStores = stores.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.address.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Card>
      <CardHeader className="p-5 pb-3 border-b flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <CardTitle className="text-base font-semibold flex items-center gap-2">
            <Store className="h-4 w-4 text-primary" />
            <span>Store Registry ({stores.length})</span>
          </CardTitle>
          <CardDescription className="text-xs">
            All registered restaurants, geographical coordinates & catalog counts
          </CardDescription>
        </div>
        <div className="relative w-full sm:w-64">
          <Search className="h-3.5 w-3.5 absolute left-2.5 top-2.5 text-muted-foreground" />
          <Input
            placeholder="Search stores or addresses..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8 h-8 text-xs"
          />
        </div>
      </CardHeader>

      <CardContent className="p-0">
        <div className="divide-y divide-border">
          {filteredStores.length === 0 ? (
            <div className="p-8 text-center text-sm text-muted-foreground">
              No stores matching &quot;{search}&quot;
            </div>
          ) : (
            filteredStores.map((st) => (
              <div key={st.id} className="p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-muted/20 transition-colors">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-sm text-foreground">{st.name}</span>
                    <Badge variant="success" className="text-[10px] py-0 px-1.5 h-4">
                      Active
                    </Badge>
                  </div>
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                    <MapPin className="h-3 w-3 text-red-500 shrink-0" />
                    <span>{st.address}</span>
                  </div>
                </div>

                <div className="flex items-center gap-4 text-xs">
                  <div className="text-left sm:text-right">
                    <span className="font-semibold text-foreground">{storeProductCounts[st.id] ?? 0}</span>
                    <span className="text-muted-foreground ml-1">items</span>
                  </div>
                  <div className="hidden md:block font-mono text-[11px] text-muted-foreground bg-muted/60 px-2 py-0.5 rounded">
                    ID: {st.id.slice(0, 8)}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </CardContent>
    </Card>
  );
}
