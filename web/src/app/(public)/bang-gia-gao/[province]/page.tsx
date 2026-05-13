import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getSEOPriceBoard } from "@/services/api";
import { formatPriceVND, formatDateVN } from "@/lib/seo-helpers";

export const revalidate = 3600;

type Props = { params: Promise<{ province: string }> };

async function loadProvinceData(provinceSlug: string) {
  const all = await getSEOPriceBoard();
  const entries = all.data.filter((e) => e.province_slug === provinceSlug);
  if (entries.length === 0) return null;
  return { entries, province: entries[0].province, generated_at: all.generated_at };
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { province } = await params;
  try {
    const result = await loadProvinceData(province);
    if (!result) return { title: "Không tìm thấy" };

    const minPrice = Math.min(...result.entries.map((e) => e.min_price));
    return {
      title: `Giá gạo ${result.province} hôm nay — Bảng giá ${result.entries.length} loại gạo`,
      description: `Giá gạo tại ${result.province} cập nhật hàng giờ. ${result.entries.length} loại gạo (ST 25, Tài Nguyên, Nàng Hoa...). Giá từ ${formatPriceVND(minPrice)}/kg. Liên hệ trực tiếp thương lái qua app.`,
      keywords: [
        `giá gạo ${result.province}`,
        `giá gạo ${result.province} hôm nay`,
        `thương lái gạo ${result.province}`,
        `mua gạo ${result.province}`,
        "bảng giá gạo Việt Nam",
      ],
      openGraph: {
        title: `Giá gạo ${result.province} hôm nay`,
        description: `Bảng giá ${result.entries.length} loại gạo tại ${result.province}. Cập nhật realtime.`,
        url: `https://sangiagao.vn/bang-gia-gao/${province}`,
        locale: "vi_VN",
        type: "website",
      },
      alternates: { canonical: `https://sangiagao.vn/bang-gia-gao/${province}` },
    };
  } catch {
    return { title: "Lỗi" };
  }
}

export default async function ProvincePricePage({ params }: Props) {
  const { province: provinceSlug } = await params;
  const result = await loadProvinceData(provinceSlug);
  if (!result) notFound();

  const sortedTypes = [...result.entries].sort((a, b) => a.rice_type_label.localeCompare(b.rice_type_label, "vi"));

  return (
    <main className="max-w-5xl mx-auto px-4 py-8">
      <nav aria-label="breadcrumb" className="text-sm text-gray-500 mb-3">
        <Link href="/">Trang chủ</Link> /{" "}
        <Link href="/bang-gia-gao">Bảng giá gạo</Link> /{" "}
        <span className="text-gray-700">{result.province}</span>
      </nav>

      <h1 className="text-3xl font-bold mb-3">
        Giá gạo {result.province} hôm nay
      </h1>
      <p className="text-gray-600 mb-6">
        Bảng giá gạo tại <strong>{result.province}</strong> với{" "}
        <strong>{result.entries.length}</strong> loại gạo khác nhau từ tin đăng thực tế của thương lái
        và nông dân địa phương. Cập nhật mỗi giờ.
      </p>

      <div className="border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="text-left px-4 py-2 font-semibold">Loại gạo</th>
              <th className="text-right px-4 py-2 font-semibold">Giá thấp nhất</th>
              <th className="text-right px-4 py-2 font-semibold">Giá TB</th>
              <th className="text-right px-4 py-2 font-semibold">Giá cao nhất</th>
              <th className="text-right px-4 py-2 font-semibold">Tin đăng</th>
            </tr>
          </thead>
          <tbody>
            {sortedTypes.map((e) => (
              <tr key={e.rice_type_slug} className="border-b hover:bg-red-50">
                <td className="px-4 py-2">
                  <Link
                    href={`/bang-gia-gao/${provinceSlug}/${e.rice_type_slug}`}
                    className="text-red-700 hover:underline font-medium"
                  >
                    {e.rice_type_label}
                  </Link>
                  <div className="text-xs text-gray-400">{e.category_label}</div>
                </td>
                <td className="px-4 py-2 text-right text-green-700 font-semibold">
                  {formatPriceVND(e.min_price)}
                </td>
                <td className="px-4 py-2 text-right">{formatPriceVND(e.avg_price)}</td>
                <td className="px-4 py-2 text-right">{formatPriceVND(e.max_price)}</td>
                <td className="px-4 py-2 text-right text-gray-500">{e.listing_count}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-6 bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h2 className="font-semibold text-blue-900 mb-2">Mua/bán gạo tại {result.province}?</h2>
        <p className="text-sm text-blue-800">
          Cài app <strong>SanGiaGao.vn</strong> để liên hệ trực tiếp thương lái, chat realtime,
          xem giá cập nhật từng giờ và đăng tin gạo của bạn.
        </p>
        <div className="mt-3 flex gap-2 flex-wrap">
          <a
            href="https://apps.apple.com/app/sangiagao-vn/id6761744869"
            className="inline-block bg-black text-white text-xs px-3 py-2 rounded"
          >
            📱 App Store iOS
          </a>
          <a
            href="https://play.google.com/store/apps/details?id=com.sangiagao.rice_marketplace"
            className="inline-block bg-green-700 text-white text-xs px-3 py-2 rounded"
          >
            🤖 Google Play Android
          </a>
        </div>
      </div>

      <div className="mt-6 text-xs text-gray-500">
        Cập nhật: {formatDateVN(result.generated_at)}
      </div>

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "BreadcrumbList",
            itemListElement: [
              { "@type": "ListItem", position: 1, name: "Trang chủ", item: "https://sangiagao.vn" },
              {
                "@type": "ListItem",
                position: 2,
                name: "Bảng giá gạo",
                item: "https://sangiagao.vn/bang-gia-gao",
              },
              {
                "@type": "ListItem",
                position: 3,
                name: result.province,
                item: `https://sangiagao.vn/bang-gia-gao/${provinceSlug}`,
              },
            ],
          }),
        }}
      />
    </main>
  );
}
