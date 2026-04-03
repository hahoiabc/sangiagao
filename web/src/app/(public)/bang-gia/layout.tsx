import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Bảng Giá Gạo Hôm Nay | SanGiaGao.Vn",
  description: "Cập nhật giá gạo mới nhất theo từng loại: ST25, Jasmine, Nếp, Tấm. Giá thấp nhất từ các nhà cung cấp trên toàn quốc.",
  openGraph: {
    title: "Bảng Giá Gạo Hôm Nay | SanGiaGao.Vn",
    description: "Cập nhật giá gạo mới nhất theo từng loại gạo trên toàn quốc.",
    url: "https://sangiagao.vn/bang-gia",
    siteName: "SanGiaGao.Vn",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "Bảng Giá Gạo Hôm Nay | SanGiaGao.Vn",
    description: "Cập nhật giá gạo mới nhất theo từng loại gạo trên toàn quốc.",
  },
  alternates: {
    canonical: "https://sangiagao.vn/bang-gia",
  },
};

export default function Layout({ children }: { children: React.ReactNode }) {
  return children;
}
