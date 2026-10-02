import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { ArrowRight, RotateCcw, UtensilsCrossed } from "lucide-react";

interface CallbackActionsProps {
  isSuccess: boolean;
  isSettling: boolean;
}

export function CallbackActions({ isSuccess, isSettling }: CallbackActionsProps) {
  const router = useRouter();

  if (isSuccess) {
    return (
      <>
        <Button
          onClick={() => router.push("/orders")}
          disabled={isSettling}
          className="w-full flex items-center justify-center gap-1.5"
        >
          <span>Track My Orders</span>
          <ArrowRight className="h-4 w-4" />
        </Button>
        <Button
          variant="outline"
          onClick={() => router.push("/")}
          className="w-full flex items-center justify-center gap-1.5"
        >
          <UtensilsCrossed className="h-4 w-4" />
          <span>Marketplace</span>
        </Button>
      </>
    );
  }

  return (
    <>
      <Button
        variant="default"
        onClick={() => router.push("/")}
        className="w-full flex items-center justify-center gap-1.5"
      >
        <RotateCcw className="h-4 w-4" />
        <span>Retry Order</span>
      </Button>
      <Button variant="outline" onClick={() => router.push("/")} className="w-full">
        Return Home
      </Button>
    </>
  );
}
