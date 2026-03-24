"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Wheat, TrendingUp, Eye, Users } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getPriceBoard, type PriceBoardResponse } from "@/services/api";
import { formatPrice, timeAgo } from "@/lib/utils";

export default function PriceBoardPage() {
  const [data, setData] = useState<PriceBoardResponse | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getPriceBoard()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false));
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
          <p className="text-lg text-white/90 mb-6 max-w-xl mx-auto drop-shadow">
            Bảng giá gạo cập nhật liên tục từ các thành viên trên toàn quốc
          </p>
          <div className="flex items-center justify-center gap-3">
            <Link href="/san-giao-dich">
              <Button size="lg" className="gap-2 bg-white text-primary hover:bg-white/90">
                <TrendingUp className="h-4 w-4" />
                Xem tin đăng
              </Button>
            </Link>
            <Link href="/dang-ky">
              <Button variant="outline" size="lg" className="gap-2 border-white text-white hover:bg-white/20">
                <Users className="h-4 w-4" />
                Đăng Ký Miễn Phí
              </Button>
            </Link>
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
              <Card key={cat.category_key}>
                <CardHeader className="py-4 px-5 bg-gradient-to-r from-green-800 to-green-600 rounded-t-lg">
                  <CardTitle className="text-base flex items-center gap-3 text-white">
                    <span className="flex items-center justify-center h-8 w-8 rounded-lg bg-white/20">
                      <Wheat className="h-4.5 w-4.5 text-white" />
                    </span>
                    <span className="flex-1 tracking-wide">{cat.category_label}</span>
                    <span className="text-xs font-normal text-white/70">{cat.products.length} SP</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b text-muted-foreground">
                          <th className="text-left py-2 pr-4 font-medium">Sản phẩm</th>
                          <th className="text-right py-2 px-4 font-medium">Giá thấp nhất</th>
                          <th className="text-center py-2 px-4 font-medium">Tin đăng</th>
                          <th className="text-center py-2 pl-4 font-medium"></th>
                        </tr>
                      </thead>
                      <tbody>
                        {cat.products.map((p) => (
                          <tr key={p.product_key} className="border-b last:border-0 hover:bg-muted/50">
                            <td className="py-3 pr-4">
                              <div className="flex items-center gap-2">
                                <span className="font-medium">{p.product_label}</span>
                                {p.sponsor_logo && (
                                  <img src={p.sponsor_logo} alt="" className="h-5 w-auto" />
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
                            <td className="py-3 px-4 text-center">
                              {p.listing_count > 0 ? (
                                <Badge variant="secondary">{p.listing_count}</Badge>
                              ) : (
                                <span className="text-muted-foreground">0</span>
                              )}
                            </td>
                            <td className="py-3 pl-4 text-center">
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
