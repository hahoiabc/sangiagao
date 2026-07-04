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
            <Link href="/dieu-khoan-su-dung" className="hover:text-foreground transition-colors">
              Điều khoản sử dụng
            </Link>
            <Link href="/chinh-sach-bao-mat" className="hover:text-foreground transition-colors">
              Chính sách bảo mật
            </Link>
            <Link href="/huong-dan" className="hover:text-foreground transition-colors">
              Hướng dẫn sử dụng
            </Link>
          </div>
        </div>
        <div className="mt-4 pt-4 border-t border-muted text-center leading-relaxed">
          <div className="mx-auto w-fit rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white">
            Trường Sơn Connect
          </div>
          <p className="mt-3 text-xs text-muted-foreground">&copy; {new Date().getFullYear()} SanGiaGao.Vn — Sàn Giá Gạo</p>
        </div>
      </div>
    </footer>
  );
}
