import type { Metadata } from "next";
import { AboutPageClient } from "./client";

export const revalidate = 600; // 10 minutes

export const metadata: Metadata = {
  title: "Giới thiệu | SanGiaGao.Vn",
  description: "Sàn Giá Gạo - Nền tảng kết nối mua bán gạo trực tiếp giữa nông dân, thương lái và doanh nghiệp trên toàn quốc",
  alternates: { canonical: "https://sangiagao.vn/gioi-thieu" },
};

export default function AboutPage() {
  return <AboutPageClient />;
}
