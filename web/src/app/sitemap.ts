import type { MetadataRoute } from "next";
import { getSEOPriceBoard } from "@/services/api";

// Regenerate sitemap every 1 hour (matches ISR of SEO pages)
export const revalidate = 3600;

const BASE_URL = "https://sangiagao.vn";

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticRoutes: MetadataRoute.Sitemap = [
    { url: BASE_URL, lastModified: new Date(), changeFrequency: "daily", priority: 1.0 },
    { url: `${BASE_URL}/san-giao-dich`, lastModified: new Date(), changeFrequency: "hourly", priority: 0.9 },
    { url: `${BASE_URL}/bang-gia`, lastModified: new Date(), changeFrequency: "hourly", priority: 0.9 },
    { url: `${BASE_URL}/bang-gia-gao`, lastModified: new Date(), changeFrequency: "hourly", priority: 0.9 },
    { url: `${BASE_URL}/huong-dan`, lastModified: new Date(), changeFrequency: "monthly", priority: 0.5 },
    { url: `${BASE_URL}/gioi-thieu`, lastModified: new Date(), changeFrequency: "monthly", priority: 0.5 },
    { url: `${BASE_URL}/dieu-khoan-su-dung`, lastModified: new Date(), changeFrequency: "yearly", priority: 0.3 },
    { url: `${BASE_URL}/chinh-sach-bao-mat`, lastModified: new Date(), changeFrequency: "yearly", priority: 0.3 },
  ];

  // Append SEO landing pages from /bang-gia-gao/...
  let seoRoutes: MetadataRoute.Sitemap = [];
  try {
    const board = await getSEOPriceBoard();
    const provinceSet = new Set<string>();
    const detailRoutes: MetadataRoute.Sitemap = [];
    for (const e of board.data) {
      provinceSet.add(e.province_slug);
      detailRoutes.push({
        url: `${BASE_URL}/bang-gia-gao/${e.province_slug}/${e.rice_type_slug}`,
        lastModified: new Date(e.last_updated),
        changeFrequency: "daily",
        priority: 0.7,
      });
    }
    const provinceRoutes: MetadataRoute.Sitemap = Array.from(provinceSet).map((slug) => ({
      url: `${BASE_URL}/bang-gia-gao/${slug}`,
      lastModified: new Date(),
      changeFrequency: "daily",
      priority: 0.8,
    }));
    seoRoutes = [...provinceRoutes, ...detailRoutes];
  } catch {
    // Fallback to static if API fails — sitemap still valid
  }

  return [...staticRoutes, ...seoRoutes];
}
