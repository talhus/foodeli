import { Badge } from "@/components/ui/badge";
import { ChefHat, UtensilsCrossed } from "lucide-react";

interface MerchantTabsNavProps {
  activeTab: "orders" | "menu";
  ordersCount: number;
  productsCount: number;
  onTabChange: (tab: "orders" | "menu") => void;
}

export function MerchantTabsNav({
  activeTab,
  ordersCount,
  productsCount,
  onTabChange,
}: MerchantTabsNavProps) {
  return (
    <div className="flex items-center gap-2 border-b">
      <button
        onClick={() => onTabChange("orders")}
        className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-all ${
          activeTab === "orders"
            ? "border-primary text-primary"
            : "border-transparent text-muted-foreground hover:text-foreground"
        }`}
      >
        <ChefHat className="h-4 w-4" />
        <span>Kitchen Orders</span>
        <Badge variant="secondary" className="ml-1 text-[10px]">
          {ordersCount}
        </Badge>
      </button>

      <button
        onClick={() => onTabChange("menu")}
        className={`flex items-center gap-2 px-4 py-2.5 text-sm font-semibold border-b-2 transition-all ${
          activeTab === "menu"
            ? "border-primary text-primary"
            : "border-transparent text-muted-foreground hover:text-foreground"
        }`}
      >
        <UtensilsCrossed className="h-4 w-4" />
        <span>Menu &amp; Inventory</span>
        <Badge variant="secondary" className="ml-1 text-[10px]">
          {productsCount}
        </Badge>
      </button>
    </div>
  );
}
