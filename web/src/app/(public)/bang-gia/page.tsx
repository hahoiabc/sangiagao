"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { Wheat, Eye, Users } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getPriceBoard, getSlogan, getSloganColor, type PriceBoardResponse } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { useThemeColor } from "@/lib/theme-color";
import { formatPrice, timeAgo } from "@/lib/utils";

export default function PriceBoardPage() {
  const { user } = useAuth();
  const { currentTheme } = useThemeColor();
  const [data, setData] = useState<PriceBoardResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [slogan, setSlogan] = useState("Kết nối ngành gạo");
  const [sloganColor, setSloganColor] = useState("#FFFFFF");

  useEffect(() => {
    getPriceBoard()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false));
    getSlogan().then((s) => setSlogan(s.value)).catch(() => {});
    getSloganColor().then((s) => setSloganColor(s.value)).catch(() => {});
  }, []);

  return (
    <div>
      {/* Hero */}
      <section className="relative py-12 sm:py-16 overflow-hidden">
        <div
          className="absolute inset-0 bg-cover bg-center"
          style={{ backgroundImage: "url('/rice-field-bg.jpg')" }}
        />
        <div className="absolute inset-0 bg-black/40" />
        <div className="relative mx-auto max-w-7xl px-4 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
            <Wheat className="h-8 w-8 text-white" />
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold text-white mb-3 drop-shadow-lg">
            Sàn Giá Gạo
          </h1>
          <div className="mb-6 overflow-hidden max-w-xl mx-auto">
            <p className="whitespace-nowrap text-lg drop-shadow animate-[marquee_15s_linear_infinite] inline-block" style={{ color: sloganColor }}>
              {slogan}
              <span className="mx-16">&nbsp;</span>
              {slogan}
            </p>
          </div>
          <div className="flex items-center justify-center gap-3">
            {!user && (
              <Link href="/dang-ky">
                <Button size="lg" className="gap-2 bg-yellow-400 text-gray-900 hover:bg-yellow-300 border-0">
                  <Users className="h-4 w-4" />
                  Đăng Ký Miễn Phí
                </Button>
              </Link>
            )}
          </div>
        </div>
      </section>

      {/* Price Board */}
      <section className="mx-auto max-w-7xl px-4 py-8">
        {data && (
          <p className="text-sm text-muted-foreground mb-4">
            Cập nhật: {timeAgo(data.updated_at)}
          </p>
        )}

        {loading ? (
          <div className="space-y-6">
            {[1, 2, 3].map((i) => (
              <Card key={i}>
                <CardHeader><Skeleton className="h-6 w-40" /></CardHeader>
                <CardContent className="space-y-3">
                  {[1, 2, 3].map((j) => <Skeleton key={j} className="h-10 w-full" />)}
                </CardContent>
              </Card>
            ))}
          </div>
        ) : data ? (
          <div className="space-y-6">
            {data.categories.map((cat) => (
              <Card key={cat.category_key} className="py-0 gap-0 overflow-hidden">
                <CardHeader className="py-4 px-5" style={{ background: `linear-gradient(to right, ${currentTheme.hexDark}, ${currentTheme.hex})` }}>
                  <CardTitle className="text-base flex items-center gap-3 text-white">
                    <span className="flex items-center justify-center h-8 w-8 rounded-lg bg-white/20">
                      <Wheat className="h-4.5 w-4.5 text-white" />
                    </span>
                    <span className="flex-1 tracking-wide">{cat.category_label}</span>
                    <span className="text-xs font-normal text-white/70">{cat.products.length} SP</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <tbody>
                        {cat.products.map((p) => (
                          <tr key={p.product_key} className="border-b last:border-0 hover:bg-muted/50">
                            <td className="py-3 pl-5 pr-4">
                              <div className="flex items-center gap-2">
                                <span className="font-medium">{p.product_label}</span>
                                {p.sponsor_logo && (
                                  <Image src={p.sponsor_logo} alt="Logo tài trợ" width={20} height={20} className="h-5 w-auto" />
                                )}
                              </div>
                            </td>
                            <td className="py-3 px-4 text-right">
                              {p.min_price ? (
                                <span className="font-semibold text-primary">
                                  {formatPrice(p.min_price)}
                                </span>
                              ) : (
                                <span className="text-muted-foreground">Chưa có giá</span>
                              )}
                            </td>
                            <td className="py-3 pl-4 pr-5 text-center">
                              {p.listing_count > 0 && (
                                <Link
                                  href={`/san-giao-dich?category=${cat.category_key}&rice_type=${p.product_key}&sort=price_asc`}
                                >
                                  <Button variant="ghost" size="sm" className="gap-1">
                                    <Eye className="h-3.5 w-3.5" />
                                    Xem
                                  </Button>
                                </Link>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          <p className="text-center text-muted-foreground py-12">
            Không thể tải bảng giá. Vui lòng thử lại sau.
          </p>
        )}
      </section>
    </div>
  );
}
