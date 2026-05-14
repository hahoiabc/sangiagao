import Link from "next/link";
import { headers } from "next/headers";
import type { Metadata } from "next";

const PLAY_PACKAGE = "com.sangiagao.rice_marketplace";
const IOS_APP_ID = "6761744869";

export const metadata: Metadata = {
  title: "Cài app Sàn Giá Gạo — Xem giá realtime + Mua bán trực tiếp",
  description:
    "Tải app Sàn Giá Gạo trên Android (Play Store) hoặc iOS (App Store). Xem giá gạo cập nhật theo giờ, kết nối thương lái và nông dân trực tiếp.",
  alternates: { canonical: "https://sangiagao.vn/cai-app" },
};

type Props = { searchParams: Promise<{ ref?: string }> };

export default async function InstallAppPage({ searchParams }: Props) {
  const { ref } = await searchParams;
  const refCode = (ref || "").toUpperCase().replace(/[^A-Z0-9]/g, "");

  // Detect OS server-side from User-Agent so links default correctly.
  const ua = (await headers()).get("user-agent")?.toLowerCase() ?? "";
  const isAndroid = /android/.test(ua);
  const isIOS = /iphone|ipad|ipod/.test(ua);

  // Play Store URL with Install Referrer parameter — mobile app reads this
  // after install to attribute the new signup to the aff partner.
  const playUrl = refCode
    ? `https://play.google.com/store/apps/details?id=${PLAY_PACKAGE}&referrer=${encodeURIComponent(refCode)}`
    : `https://play.google.com/store/apps/details?id=${PLAY_PACKAGE}`;

  // iOS doesn't have a referrer query param the way Play does. Best-effort:
  // user clicks → App Store. Aff attribution recovered via Universal Link
  // (if app is already installed) or by user opening /r/{code} after install.
  const iosUrl = `https://apps.apple.com/vn/app/sangiagao-vn/id${IOS_APP_ID}`;

  const webRegisterUrl = refCode ? `/dang-ky?ref=${refCode}` : "/dang-ky";

  return (
    <main className="min-h-screen bg-gradient-to-b from-amber-50 to-white">
      <div className="max-w-2xl mx-auto px-4 py-12">
        {refCode && (
          <div className="mb-6 p-4 rounded-lg bg-amber-100 border border-amber-300">
            <div className="text-xs text-amber-800">Mã giới thiệu</div>
            <div className="text-2xl font-bold tracking-widest text-amber-900">{refCode}</div>
            <div className="text-xs text-amber-700 mt-1">
              Mã đã được lưu sẵn — bạn không cần nhập lại khi đăng ký.
            </div>
          </div>
        )}

        <h1 className="text-3xl font-bold mb-3">Cài app Sàn Giá Gạo</h1>
        <p className="text-gray-600 mb-8 leading-relaxed">
          Xem giá gạo cập nhật theo giờ, kết nối thương lái và nông dân trực tiếp qua chat.
          Cài app trên điện thoại để dùng nhanh, mượt nhất.
        </p>

        {/* Primary CTA buttons */}
        <div className="space-y-3 mb-8">
          <a
            href={playUrl}
            className={`block w-full text-center px-6 py-4 rounded-lg font-semibold text-white shadow-md transition ${
              isAndroid || !isIOS
                ? "bg-green-700 hover:bg-green-800"
                : "bg-gray-700 hover:bg-gray-800"
            }`}
          >
            <div className="text-xs opacity-80">Android</div>
            <div className="text-lg">📱 Tải Google Play</div>
          </a>

          <a
            href={iosUrl}
            className={`block w-full text-center px-6 py-4 rounded-lg font-semibold text-white shadow-md transition ${
              isIOS ? "bg-blue-700 hover:bg-blue-800" : "bg-gray-700 hover:bg-gray-800"
            }`}
          >
            <div className="text-xs opacity-80">iOS / iPhone</div>
            <div className="text-lg">🍎 Tải App Store</div>
          </a>
        </div>

        {/* Web register fallback */}
        <div className="border-t pt-6 text-center">
          <p className="text-sm text-gray-500 mb-3">Hoặc đăng ký trên web</p>
          <Link
            href={webRegisterUrl}
            className="inline-block px-6 py-2 rounded border border-gray-300 hover:bg-gray-50 text-sm"
          >
            Đăng ký bằng trình duyệt
          </Link>
        </div>

        {refCode && (
          <div className="mt-10 text-xs text-gray-500 text-center leading-relaxed">
            Khi bạn đăng ký + mua gói thành viên, Sàn sẽ ghi nhận hoa hồng cho người giới thiệu bạn.
            Việc này không làm tăng giá gói bạn mua.
          </div>
        )}
      </div>
    </main>
  );
}
