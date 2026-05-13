import Link from "next/link";
import type { Metadata } from "next";
import { getSEOPriceBoard } from "@/services/api";
import { formatPriceVND, formatDateVN } from "@/lib/seo-helpers";

// ISR — regenerate every 1 hour
export const revalidate = 3600;

export const metadata: Metadata = {
  title: "Bảng giá gạo Việt Nam hôm nay — Cập nhật theo tỉnh",
  description:
    "Bảng giá gạo Việt Nam cập nhật hàng giờ theo từng tỉnh, từng loại gạo. Giá gạo ST 25, Tài Nguyên, Nàng Hoa, Nhật, Đài Loan... So sánh giá thương lái thu mua.",
  keywords: [
    "giá gạo hôm nay",
    "bảng giá gạo Việt Nam",
    "giá gạo theo tỉnh",
    "giá gạo ST 25",
    "giá gạo Tài Nguyên",
    "thương lái thu mua gạo",
    "sàn giá gạo",
  ],
  openGraph: {
    title: "Bảng giá gạo Việt Nam hôm nay theo tỉnh",
    description: "Giá gạo cập nhật realtime theo từng tỉnh và loại gạo. Hơn 175 quốc gia.",
    url: "https://sangiagao.vn/bang-gia-gao",
    siteName: "SanGiaGao.vn",
    locale: "vi_VN",
    type: "website",
  },
  alternates: { canonical: "https://sangiagao.vn/bang-gia-gao" },
};

export default async function PriceBoardIndexPage() {
  let data: Awaited<ReturnType<typeof getSEOPriceBoard>>;
  try {
    data = await getSEOPriceBoard();
  } catch {
    data = { data: [], total: 0, generated_at: new Date().toISOString() };
  }

  // Group by province
  const byProvince = new Map<
    string,
    { slug: string; types: { riceType: string; slug: string; minPrice: number; count: number }[] }
  >();
  for (const e of data.data) {
    if (!byProvince.has(e.province)) {
      byProvince.set(e.province, { slug: e.province_slug, types: [] });
    }
    byProvince.get(e.province)!.types.push({
      riceType: e.rice_type_label,
      slug: e.rice_type_slug,
      minPrice: e.min_price,
      count: e.listing_count,
    });
  }

  const provinces = Array.from(byProvince.entries()).sort(([a], [b]) => a.localeCompare(b, "vi"));

  return (
    <main className="max-w-5xl mx-auto px-4 py-8">
      <nav aria-label="breadcrumb" className="text-sm text-gray-500 mb-3">
        <Link href="/">Trang chủ</Link> / <span className="text-gray-700">Bảng giá gạo</span>
      </nav>

      <h1 className="text-3xl font-bold mb-3">Bảng giá gạo Việt Nam hôm nay</h1>
      <p className="text-gray-600 mb-6">
        Giá gạo cập nhật theo từng tỉnh, từng loại gạo từ tin đăng thực tế của thương lái và nông dân
        trên sàn SanGiaGao.vn. Dữ liệu cập nhật mỗi giờ.
      </p>

      {provinces.length === 0 ? (
        <div className="bg-yellow-50 border border-yellow-200 rounded p-4 text-yellow-800">
          Chưa có dữ liệu giá gạo công khai. Hãy đăng tin gạo đầu tiên trên app SanGiaGao.vn để
          xuất hiện trên trang này!
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {provinces.map(([province, info]) => (
            <Link
              key={province}
              href={`/bang-gia-gao/${info.slug}`}
              className="block border rounded-lg p-4 hover:shadow-md hover:border-red-300 transition"
            >
              <h2 className="text-lg font-semibold text-red-700">{province}</h2>
              <p className="text-sm text-gray-500 mt-1">{info.types.length} loại gạo</p>
              <p className="text-xs text-gray-400 mt-2">
                Từ{" "}
                {formatPriceVND(
                  Math.min(...info.types.map((t) => t.minPrice)),
                )}{" "}
                / kg
              </p>
            </Link>
          ))}
        </div>
      )}

      <div className="mt-10 border-t pt-6 text-sm text-gray-500">
        <p>
          Cập nhật lần gần nhất: {formatDateVN(data.generated_at)}. Cài app{" "}
          <strong>SanGiaGao.vn</strong> để xem giá realtime + đăng tin gạo.
        </p>
      </div>

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "WebSite",
            name: "SanGiaGao.vn",
            url: "https://sangiagao.vn",
            description: "Sàn giá gạo Việt Nam — cập nhật giá realtime theo tỉnh",
          }),
        }}
      />
    </main>
  );
}
