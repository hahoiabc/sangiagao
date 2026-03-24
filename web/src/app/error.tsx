"use client";

import { Button } from "@/components/ui/button";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] px-4">
      <h2 className="text-xl font-bold mb-2">Đã xảy ra lỗi</h2>
      <p className="text-muted-foreground mb-6 text-center max-w-md">
        {error.message || "Có gì đó không đúng. Vui lòng thử lại."}
      </p>
      <Button onClick={reset}>Thử lại</Button>
    </div>
  );
}
