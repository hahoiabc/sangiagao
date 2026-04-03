import type { Metadata } from "next";
import ListingDetailPage from "./client";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params;
  try {
    const res = await fetch(`${API_BASE}/marketplace/${id}`, { next: { revalidate: 60 } });
    if (res.ok) {
      const listing = await res.json();
      const desc = listing.description || `${listing.rice_type} - Xem chi tiết trên Sàn Giá Gạo`;
      return {
        title: `${listing.title || "Chi tiết tin đăng"} | SanGiaGao.Vn`,
        description: desc,
        openGraph: {
          title: listing.title,
          description: desc,
          url: `https://sangiagao.vn/san-giao-dich/${id}`,
          siteName: "SanGiaGao.Vn",
          images: listing.images?.length > 0 ? [listing.images[0]] : [],
          type: "article",
        },
        twitter: {
          card: "summary_large_image",
          title: listing.title,
          description: desc,
          images: listing.images?.length > 0 ? [listing.images[0]] : [],
        },
        alternates: {
          canonical: `https://sangiagao.vn/san-giao-dich/${id}`,
        },
      };
    }
  } catch {
    // fallback to default metadata
  }
  return {
    title: "Chi tiết tin đăng | SanGiaGao.Vn",
    description: "Xem chi tiết tin đăng trên Sàn Giá Gạo",
  };
}

export default function Page() {
  return <ListingDetailPage />;
}
