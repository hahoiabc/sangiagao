import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/", "/san-giao-dich", "/bang-gia", "/huong-dan", "/nguoi-ban"],
        disallow: ["/tin-dang", "/tin-nhan", "/goi-thanh-vien", "/thong-bao", "/api/"],
      },
    ],
    sitemap: "https://sangiagao.vn/sitemap.xml",
  };
}
