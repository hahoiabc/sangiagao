import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sàn Giao Dịch Gạo | SanGiaGao.Vn",
  description: "Mua bán gạo trực tuyến - Tìm kiếm, so sánh giá gạo từ các nhà cung cấp uy tín trên toàn quốc. ST25, Jasmine, Nếp, Tấm và nhiều loại gạo khác.",
  openGraph: {
    title: "Sàn Giao Dịch Gạo | SanGiaGao.Vn",
    description: "Mua bán gạo trực tuyến - Tìm kiếm, so sánh giá gạo từ các nhà cung cấp uy tín trên toàn quốc.",
    url: "https://sangiagao.vn/san-giao-dich",
    siteName: "SanGiaGao.Vn",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Sàn Giao Dịch Gạo | SanGiaGao.Vn",
    description: "Mua bán gạo trực tuyến - Tìm kiếm, so sánh giá gạo từ các nhà cung cấp uy tín.",
  },
  alternates: {
    canonical: "https://sangiagao.vn/san-giao-dich",
  },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
