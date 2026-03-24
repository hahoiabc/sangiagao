import type { Metadata } from "next";
import SellerProfilePage from "./client";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params;
  try {
    const res = await fetch(`${API_BASE}/users/${id}/public`, { next: { revalidate: 60 } });
    if (res.ok) {
      const profile = await res.json();
      const name = profile.name || "Người bán";
      return {
        title: `${name} | SanGiaGao.Vn`,
        description: `Xem hồ sơ và đánh giá của ${name} trên Sàn Giá Gạo`,
        openGraph: {
          title: `${name} - Người bán trên SanGiaGao`,
          description: profile.description || `Thành viên tại ${profile.province || "Sàn Giá Gạo"}`,
        },
      };
    }
  } catch {
    // fallback to default metadata
  }
  return {
    title: "Hồ sơ người bán | SanGiaGao.Vn",
    description: "Xem hồ sơ người bán trên Sàn Giá Gạo",
  };
}

export default function Page() {
  return <SellerProfilePage />;
}
