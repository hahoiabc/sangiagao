import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Toaster } from "sonner";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { ThemeColorProvider } from "@/lib/theme-color";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: { default: "SanGiaGao.Vn - Sàn Giá Gạo", template: "%s | SanGiaGao.Vn" },
  description: "Sàn giao dịch gạo trực tuyến - Mua bán gạo uy tín, so sánh giá gạo từ nhà cung cấp trên toàn quốc",
  metadataBase: new URL("https://sangiagao.vn"),
  openGraph: {
    siteName: "SanGiaGao.Vn",
    type: "website",
    locale: "vi_VN",
  },
  twitter: {
    card: "summary_large_image",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "Organization",
              name: "SanGiaGao.Vn",
              url: "https://sangiagao.vn",
              description: "Sàn giao dịch gạo trực tuyến Việt Nam",
              address: {
                "@type": "PostalAddress",
                streetAddress: "Đường Trần Phú, Tổ 18, Ấp Bến Cam, Xã Phước Thiền",
                addressLocality: "Nhơn Trạch",
                addressRegion: "Đồng Nai",
                addressCountry: "VN",
              },
            }),
          }}
        />
        <ThemeColorProvider>
          <AuthProvider>{children}</AuthProvider>
        </ThemeColorProvider>
        <Toaster position="top-right" richColors closeButton />
      </body>
    </html>
  );
}
