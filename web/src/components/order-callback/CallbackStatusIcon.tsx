import { CheckCircle2, XCircle, Loader2 } from "lucide-react";

interface CallbackStatusIconProps {
  isSettling: boolean;
  isSuccess: boolean;
}

export function CallbackStatusIcon({ isSettling, isSuccess }: CallbackStatusIconProps) {
  return (
    <div
      className={`h-16 w-16 rounded-full flex items-center justify-center mb-3 ${
        isSettling
          ? "bg-amber-100 text-amber-600"
          : isSuccess
          ? "bg-emerald-100 text-emerald-600"
          : "bg-red-100 text-red-600"
      }`}
    >
      {isSettling ? (
        <Loader2 className="h-10 w-10 animate-spin" />
      ) : isSuccess ? (
        <CheckCircle2 className="h-10 w-10" />
      ) : (
        <XCircle className="h-10 w-10" />
      )}
    </div>
  );
}
