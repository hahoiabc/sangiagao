import Link from "next/link";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getSEOPriceBoard, getSEOListings } from "@/services/api";
import { formatPriceVND, formatDateVN } from "@/lib/seo-helpers";

export const revalidate = 3600;

type Props = { params: Promise<{ province: string; riceType: string }> };

async function loadDetailData(provinceSlug: string, riceTypeSlug: string) {
  const [board, listingsRes] = await Promise.all([
    getSEOPriceBoard(),
    getSEOListings(provinceSlug, riceTypeSlug).catch(() => ({ data: [], total: 0 })),
  ]);
  const entry = board.data.find(
    (e) => e.province_slug === provinceSlug && e.rice_type_slug === riceTypeSlug,
  );
  if (!entry) return null;
  return { entry, listings: listingsRes.data, generated_at: board.generated_at };
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { province, riceType } = await params;
  try {
    const result = await loadDetailData(province, riceType);
    if (!result) return { title: "Không tìm thấy" };

    const e = result.entry;
    const title = `Giá gạo ${e.rice_type_label} tại ${e.province} hôm nay — ${formatPriceVND(e.min_price)}/kg`;
    const description = `Giá gạo ${e.rice_type_label} ở ${e.province} từ ${formatPriceVND(e.min_price)} đến ${formatPriceVND(e.max_price)}/kg. ${e.listing_count} tin đăng từ thương lái và nông dân. Liên hệ trực tiếp qua app SanGiaGao.vn.`;
    return {
      title,
      description,
      keywords: [
        `giá gạo ${e.rice_type_label}`,
        `giá gạo ${e.rice_type_label} ${e.province}`,
        `giá gạo ${e.rice_type_label} hôm nay`,
        `mua gạo ${e.rice_type_label} ${e.province}`,
        `thương lái gạo ${e.rice_type_label}`,
      ],
      openGraph: {
        title,
        description,
        url: `https://sangiagao.vn/bang-gia-gao/${province}/${riceType}`,
        locale: "vi_VN",
        type: "website",
      },
      alternates: { canonical: `https://sangiagao.vn/bang-gia-gao/${province}/${riceType}` },
    };
  } catch {
    return { title: "Lỗi" };
  }
}

export default async function PriceDetailPage({ params }: Props) {
  const { province: provinceSlug, riceType: riceTypeSlug } = await params;
  const result = await loadDetailData(provinceSlug, riceTypeSlug);
  if (!result) notFound();

  const e = result.entry;

  return (
    <main className="max-w-5xl mx-auto px-4 py-8">
      <nav aria-label="breadcrumb" className="text-sm text-gray-500 mb-3">
        <Link href="/">Trang chủ</Link> /{" "}
        <Link href="/bang-gia-gao">Bảng giá gạo</Link> /{" "}
        <Link href={`/bang-gia-gao/${provinceSlug}`}>{e.province}</Link> /{" "}
        <span className="text-gray-700">{e.rice_type_label}</span>
      </nav>

      <h1 className="text-3xl font-bold mb-2">
        Giá gạo {e.rice_type_label} tại {e.province} hôm nay
      </h1>
      <p className="text-gray-600 mb-6">
        Giá gạo <strong>{e.rice_type_label}</strong> ({e.category_label}) tại{" "}
        <strong>{e.province}</strong> dao động từ {formatPriceVND(e.min_price)} đến{" "}
        {formatPriceVND(e.max_price)}/kg, theo <strong>{e.listing_count}</strong> tin đăng từ
        thương lái và nông dân trên sàn SanGiaGao.vn.
      </p>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8">
        <div className="border rounded p-4 bg-green-50">
          <div className="text-xs text-gray-500">Giá thấp nhất</div>
          <div className="text-xl font-bold text-green-700">{formatPriceVND(e.min_price)}</div>
          <div className="text-xs text-gray-400 mt-1">/ kg</div>
        </div>
        <div className="border rounded p-4 bg-blue-50">
          <div className="text-xs text-gray-500">Giá trung bình</div>
          <div className="text-xl font-bold text-blue-700">{formatPriceVND(e.avg_price)}</div>
          <div className="text-xs text-gray-400 mt-1">/ kg</div>
        </div>
        <div className="border rounded p-4 bg-orange-50">
          <div className="text-xs text-gray-500">Giá cao nhất</div>
          <div className="text-xl font-bold text-orange-700">{formatPriceVND(e.max_price)}</div>
          <div className="text-xs text-gray-400 mt-1">/ kg</div>
        </div>
      </div>

      {result.listings.length > 0 && (
        <>
          <h2 className="text-xl font-semibold mb-3">
            Tin đăng gạo {e.rice_type_label} mới nhất tại {e.province}
          </h2>
          <div className="border rounded-lg overflow-hidden mb-6">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="text-left px-4 py-2">Tin đăng</th>
                  <th className="text-right px-4 py-2">Giá</th>
                  <th className="text-right px-4 py-2">Số lượng</th>
                  <th className="text-left px-4 py-2">Khu vực</th>
                </tr>
              </thead>
              <tbody>
                {result.listings.map((l) => (
                  <tr key={l.id} className="border-b hover:bg-red-50">
                    <td className="px-4 py-2">
                      <Link
                        href={`/san-giao-dich/${l.id}`}
                        className="text-red-700 hover:underline"
                      >
                        {l.title}
                      </Link>
                      {l.seller_name && (
                        <div className="text-xs text-gray-400">bởi {l.seller_name}</div>
                      )}
                    </td>
                    <td className="px-4 py-2 text-right font-semibold text-green-700">
                      {formatPriceVND(l.price_per_kg)}
                    </td>
                    <td className="px-4 py-2 text-right">
                      {new Intl.NumberFormat("vi-VN").format(l.quantity_kg)} kg
                    </td>
                    <td className="px-4 py-2 text-xs text-gray-500">
                      {l.ward ?? "—"}, {l.province ?? e.province}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <h2 className="font-semibold text-blue-900 mb-2">
          Liên hệ thương lái mua gạo {e.rice_type_label} tại {e.province}
        </h2>
        <p className="text-sm text-blue-800">
          Cài app <strong>SanGiaGao.vn</strong> để chat trực tiếp với người bán, xem giá realtime
          và đăng tin gạo của bạn.
        </p>
        <div className="mt-3 flex gap-2 flex-wrap">
          <a
            href="https://apps.apple.com/app/sangiagao-vn/id6761744869"
            className="inline-block bg-black text-white text-xs px-3 py-2 rounded"
          >
            📱 Tải iOS
          </a>
          <a
            href="https://play.google.com/store/apps/details?id=com.sangiagao.rice_marketplace"
            className="inline-block bg-green-700 text-white text-xs px-3 py-2 rounded"
          >
            🤖 Tải Android
          </a>
        </div>
      </div>

      <div className="text-xs text-gray-500">
        Cập nhật lần gần nhất: {formatDateVN(result.generated_at)}
      </div>

      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@type": "Product",
            name: `Gạo ${e.rice_type_label} ${e.province}`,
            description: `Gạo ${e.rice_type_label} (${e.category_label}) tại ${e.province}. ${e.listing_count} tin đăng từ thương lái và nông dân.`,
            category: e.category_label,
            offers: {
              "@type": "AggregateOffer",
              priceCurrency: "VND",
              lowPrice: e.min_price,
              highPrice: e.max_price,
              offerCount: e.listing_count,
              availability: "https://schema.org/InStock",
            },
          }),
        }}
      />
    </main>
  );
}
