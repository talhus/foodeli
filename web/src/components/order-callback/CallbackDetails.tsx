interface CallbackDetailsProps {
  orderId: string | null;
  transactionRef: string | null;
  status: string | null;
  isSettling: boolean;
  isSuccess: boolean;
}

export function CallbackDetails({
  orderId,
  transactionRef,
  status,
  isSettling,
  isSuccess,
}: CallbackDetailsProps) {
  return (
    <div className="rounded-lg bg-muted/50 p-4 text-left text-xs space-y-2 border">
      {orderId && (
        <div className="flex justify-between">
          <span className="text-muted-foreground">Order ID:</span>
          <span className="font-mono font-bold text-foreground">#{orderId.slice(0, 8)}</span>
        </div>
      )}
      {transactionRef && (
        <div className="flex justify-between">
          <span className="text-muted-foreground">Transaction Ref:</span>
          <span className="font-mono text-foreground">{transactionRef}</span>
        </div>
      )}
      <div className="flex justify-between">
        <span className="text-muted-foreground">Provider:</span>
        <span className="font-semibold text-foreground">TyrFintech Gateway</span>
      </div>
      <div className="flex justify-between">
        <span className="text-muted-foreground">Payment Status:</span>
        <span
          className={`font-bold uppercase ${
            isSuccess ? "text-emerald-600" : "text-destructive"
          }`}
        >
          {isSettling ? "CONFIRMING" : status || "UNKNOWN"}
        </span>
      </div>
    </div>
  );
}
