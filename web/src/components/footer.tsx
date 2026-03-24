import Link from "next/link";
import { Wheat } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t bg-muted/30 mt-auto">
      <div className="mx-auto max-w-7xl px-4 py-8">
        <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
          <Link href="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <Wheat className="h-5 w-5 text-primary" />
            <span className="font-semibold text-primary">SanGiaGao.Vn</span>
          </Link>
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <Link href="/chinh-sach-bao-mat" className="hover:text-foreground transition-colors">
              Chính sách bảo mật
            </Link>
            <span>&copy; {new Date().getFullYear()} SanGiaGao</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
