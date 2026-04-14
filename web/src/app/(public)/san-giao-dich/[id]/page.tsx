import type { Metadata } from "next";
import ListingDetailPage from "./client";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

interface ListingData {
  id: string;
  title?: string;
  rice_type?: string;
  description?: string;
  price_per_kg?: number;
  quantity_kg?: number;
  province?: string;
  ward?: string;
  images?: string[];
  created_at?: string;
  seller?: { name?: string };
}

async function fetchListing(id: string): Promise<ListingData | null> {
  try {
    const res = await fetch(`${API_BASE}/marketplace/${id}`, { next: { revalidate: 60 } });
    if (res.ok) return await res.json();
  } catch {}
  return null;
}

function formatPrice(price?: number): string {
  if (!price) return "";
  return new Intl.NumberFormat("vi-VN").format(price);
}

function buildTitle(listing: ListingData): string {
  const parts = [listing.title || listing.rice_type || "Tin đăng"];
  if (listing.price_per_kg) parts.push(`${formatPrice(listing.price_per_kg)}đ/kg`);
  if (listing.province) parts.push(listing.province);
  return parts.join(" - ");
}

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params;
  const listing = await fetchListing(id);
  if (!listing) {
    return {
      title: "Chi tiết tin đăng | SanGiaGao.Vn",
      description: "Xem chi tiết tin đăng trên Sàn Giá Gạo",
    };
  }
  const title = buildTitle(listing);
  const desc = listing.description || `${listing.rice_type || ""} - ${listing.quantity_kg ? `${formatPrice(listing.quantity_kg)}kg` : ""} ${listing.province ? `tại ${listing.province}` : ""}`.trim() || "Xem chi tiết trên Sàn Giá Gạo";
  return {
    title: `${title} | SanGiaGao.Vn`,
    description: desc,
    openGraph: {
      title,
      description: desc,
      url: `https://sangiagao.vn/san-giao-dich/${id}`,
      siteName: "SanGiaGao.Vn",
      images: listing.images?.length ? [listing.images[0]] : [],
      type: "website",
      locale: "vi_VN",
    },
    twitter: {
      card: "summary_large_image",
      title,
      description: desc,
      images: listing.images?.length ? [listing.images[0]] : [],
    },
    alternates: {
      canonical: `https://sangiagao.vn/san-giao-dich/${id}`,
    },
  };
}

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const listing = await fetchListing(id);
  const jsonLd = listing ? {
    "@context": "https://schema.org",
    "@type": "Product",
    "name": listing.title || listing.rice_type,
    "description": listing.description,
    "image": listing.images?.length ? listing.images[0] : undefined,
    "brand": { "@type": "Brand", "name": "SanGiaGao.Vn" },
    "offers": {
      "@type": "Offer",
      "price": listing.price_per_kg,
      "priceCurrency": "VND",
      "availability": "https://schema.org/InStock",
      "seller": listing.seller?.name ? { "@type": "Person", "name": listing.seller.name } : undefined,
      "areaServed": listing.province,
    },
  } : null;

  return (
    <>
      {jsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      )}
      <ListingDetailPage />
    </>
  );
}
