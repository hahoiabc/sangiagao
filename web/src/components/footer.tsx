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
          </div>
        </div>
        <div className="mt-4 pt-4 border-t border-muted text-center text-xs text-muted-foreground leading-relaxed">
          <p className="font-medium">Công ty TNHH MTV GẠO HÀ ÂN</p>
          <p>MST: 3602984885</p>
          <p>Đường Trần Phú, Tổ 18, Ấp Bến Cam, Xã Phước Thiền, Huyện Nhơn Trạch, Tỉnh Đồng Nai, Việt Nam</p>
          <p className="mt-1">&copy; {new Date().getFullYear()} SanGiaGao.Vn — Sàn Giá Gạo</p>
        </div>
      </div>
    </footer>
  );
}
