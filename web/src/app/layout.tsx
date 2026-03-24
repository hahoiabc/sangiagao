import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Toaster } from "sonner";
import "./globals.css";
import { AuthProvider } from "@/lib/auth";
import { ThemeColorProvider } from "@/lib/theme-color";

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] });

export const metadata: Metadata = {
  title: "SanGiaGao.Vn - Sàn Giá Gạo",
  description: "Sàn giao dịch giá gạo trực tuyến",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="vi">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <ThemeColorProvider>
          <AuthProvider>{children}</AuthProvider>
        </ThemeColorProvider>
        <Toaster position="top-right" richColors closeButton />
      </body>
    </html>
  );
}
