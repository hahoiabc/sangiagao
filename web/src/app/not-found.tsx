import Link from "next/link";
import { Button } from "@/components/ui/button";

export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh] px-4">
      <h2 className="text-xl font-bold mb-2">Không tìm thấy trang</h2>
      <p className="text-muted-foreground mb-6">
        Trang bạn tìm không tồn tại hoặc đã bị xóa.
      </p>
      <Link href="/bang-gia">
        <Button>Về trang chủ</Button>
      </Link>
    </div>
  );
}
