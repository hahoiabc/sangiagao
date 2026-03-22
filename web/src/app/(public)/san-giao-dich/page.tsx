"use client";

import { Suspense, useEffect, useState, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { Search, MapPin, Package, ChevronLeft, ChevronRight } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { searchMarketplace, type Listing, type PaginatedResponse } from "@/services/api";
import { formatPrice, formatQuantity, timeAgo } from "@/lib/utils";

export default function MarketplacePage() {
  return (
    <Suspense fallback={<div className="mx-auto max-w-7xl px-4 py-6"><Skeleton className="h-80 w-full" /></div>}>
      <MarketplaceContent />
    </Suspense>
  );
}

function MarketplaceContent() {
  const searchParams = useSearchParams();
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [category] = useState(searchParams.get("category") || "");
  const [riceType] = useState(searchParams.get("rice_type") || "");
  const [result, setResult] = useState<PaginatedResponse<Listing> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  const fetchListings = useCallback(
    async (p: number) => {
      setLoading(true);
      try {
        const res = await searchMarketplace({
          q: query || undefined,
          category: category || undefined,
          rice_type: riceType || undefined,
          page: p,
          limit: 20,
        });
        setResult(res);
      } catch {
        // ignore
      } finally {
        setLoading(false);
      }
    },
    [query, category, riceType]
  );

  useEffect(() => {
    fetchListings(page);
  }, [page, fetchListings]);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setPage(1);
    fetchListings(1);
  }

  const totalPages = result ? Math.ceil(result.total / 20) : 0;

  return (
    <div className="mx-auto max-w-7xl px-4 py-6">
      <h1 className="text-2xl font-bold mb-4">Sàn Giao Dịch</h1>

      <form onSubmit={handleSearch} className="flex gap-2 mb-6">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Tìm kiếm gạo, loại gạo, tỉnh..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button type="submit">Tìm</Button>
      </form>

      {(category || riceType) && (
        <div className="flex gap-2 mb-4">
          {category && <Badge variant="secondary">Loại: {category}</Badge>}
          {riceType && <Badge variant="secondary">Gạo: {riceType}</Badge>}
          <Link href="/san-giao-dich">
            <Button variant="ghost" size="sm">Xóa bộ lọc</Button>
          </Link>
        </div>
      )}

      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Card key={i}>
              <Skeleton className="h-40 rounded-t-lg" />
              <CardContent className="p-4 space-y-2">
                <Skeleton className="h-5 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
                <Skeleton className="h-4 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : result && result.data.length > 0 ? (
        <>
          <p className="text-sm text-muted-foreground mb-4">
            {result.total} tin đăng
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {result.data.map((listing) => (
              <Link key={listing.id} href={`/san-giao-dich/${listing.id}`}>
                <Card className="overflow-hidden hover:shadow-md transition-shadow cursor-pointer h-full">
                  <div className="h-40 bg-muted flex items-center justify-center overflow-hidden">
                    {listing.images.length > 0 ? (
                      <img
                        src={listing.images[0]}
                        alt={listing.title}
                        className="h-full w-full object-cover"
                      />
                    ) : (
                      <Package className="h-10 w-10 text-muted-foreground/40" />
                    )}
                  </div>
                  <CardContent className="p-4">
                    <h3 className="font-semibold text-sm line-clamp-2 mb-2">
                      {listing.title}
                    </h3>
                    <p className="text-lg font-bold text-primary mb-1">
                      {formatPrice(listing.price_per_kg)}
                    </p>
                    <div className="flex items-center gap-1 text-xs text-muted-foreground mb-1">
                      <Package className="h-3 w-3" />
                      {formatQuantity(listing.quantity_kg)}
                    </div>
                    {listing.province && (
                      <div className="flex items-center gap-1 text-xs text-muted-foreground mb-1">
                        <MapPin className="h-3 w-3" />
                        {listing.province}
                      </div>
                    )}
                    <p className="text-xs text-muted-foreground mt-2">
                      {timeAgo(listing.created_at)}
                    </p>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-center gap-2 mt-6">
              <Button
                variant="outline"
                size="sm"
                disabled={page <= 1}
                onClick={() => setPage((p) => p - 1)}
              >
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <span className="text-sm text-muted-foreground">
                Trang {page} / {totalPages}
              </span>
              <Button
                variant="outline"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-12">
          <Package className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
          <p className="text-muted-foreground">Không tìm thấy tin đăng nào</p>
        </div>
      )}
    </div>
  );
}
