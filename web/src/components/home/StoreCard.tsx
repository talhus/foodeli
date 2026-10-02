import { Store } from "@/types/api";
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Store as StoreIcon, MapPin } from "lucide-react";

interface StoreCardProps {
  store: Store;
  isSelected: boolean;
  onSelect: (store: Store) => void;
}

export function StoreCard({ store, isSelected, onSelect }: StoreCardProps) {
  return (
    <Card
      onClick={() => onSelect(store)}
      className={`cursor-pointer transition-all hover:shadow-md ${
        isSelected
          ? "border-primary bg-primary/5 ring-2 ring-primary/30"
          : "hover:border-primary/50"
      }`}
    >
      <CardHeader className="p-5 pb-3">
        <div className="flex items-center justify-between">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <StoreIcon className="h-5 w-5" />
          </div>
          {isSelected && (
            <Badge variant="default" className="text-[10px]">
              Selected
            </Badge>
          )}
        </div>
        <CardTitle className="text-base mt-2">{store.name}</CardTitle>
        <CardDescription className="flex items-center gap-1 text-xs">
          <MapPin className="h-3 w-3 shrink-0" />
          <span className="truncate">{store.address}</span>
        </CardDescription>
      </CardHeader>
      <CardFooter className="p-5 pt-0">
        <Button
          variant={isSelected ? "default" : "outline"}
          size="sm"
          className="w-full text-xs"
          onClick={(e) => {
            e.stopPropagation();
            onSelect(store);
          }}
        >
          {isSelected ? "Browsing Menu" : "View Menu"}
        </Button>
      </CardFooter>
    </Card>
  );
}
