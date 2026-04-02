"use client";

import { Suspense, useEffect, useState, useCallback } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { Search, MapPin, Package, ChevronLeft, ChevronRight, Filter, X, ArrowUpDown } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { searchMarketplace, getProductCatalog, toThumbnailUrl, type Listing, type PaginatedResponse, type RiceCategory } from "@/services/api";
import { formatPrice, formatQuantity, timeAgo } from "@/lib/utils";

interface LocationItem {
  code: string;
  name: string;
}

function removeDiacritics(str: string): string {
  const withDiacritics = "àáảãạăắằẳẵặâấầẩẫậèéẻẽẹêếềểễệìíỉĩịòóỏõọôốồổỗộơớờởỡợùúủũụưứừửữựỳýỷỹỵđ";
  const withoutDiacritics = "aaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd";
  let result = str.toLowerCase();
  for (let i = 0; i < withDiacritics.length; i++) {
    result = result.replaceAll(withDiacritics[i], withoutDiacritics[i]);
  }
  return result;
}

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
  const [category, setCategory] = useState(searchParams.get("category") || "");
  const [riceType, setRiceType] = useState(searchParams.get("rice_type") || "");
  const [province, setProvince] = useState(searchParams.get("province") || "");
  const [ward, setWard] = useState(searchParams.get("ward") || "");
  const [sort, setSort] = useState(searchParams.get("sort") || "price_asc");
  const [minPrice, setMinPrice] = useState(searchParams.get("min_price") || "");
  const [maxPrice, setMaxPrice] = useState(searchParams.get("max_price") || "");
  const [showFilter, setShowFilter] = useState(false);
  const [categories, setCategories] = useState<RiceCategory[]>([]);
  const [result, setResult] = useState<PaginatedResponse<Listing> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);

  // Location data from CSV (like mobile)
  const [provinces, setProvinces] = useState<LocationItem[]>([]);
  const [wards, setWards] = useState<LocationItem[]>([]);
  const [allWards, setAllWards] = useState<{ code: string; name: string; provinceCode: string }[]>([]);

  useEffect(() => {
    getProductCatalog().then(setCategories).catch(() => {});
    // Load location data from CSV
    fetch("/vietnam_divisions.csv")
      .then((res) => res.text())
      .then((text) => {
        const lines = text.trim().split("\n").slice(1);
        const provMap = new Map<string, string>();
        const wardList: { code: string; name: string; provinceCode: string }[] = [];
        for (const line of lines) {
          const cols = line.split(",");
          if (cols.length < 8) continue;
          const pCode = cols[0].trim();
          const pName = cols[1].trim();
          const wCode = cols[6]?.trim();
          const wName = cols[7]?.trim();
          if (pCode && pName) provMap.set(pCode, pName);
          if (wCode && wName && pCode) wardList.push({ code: wCode, name: wName, provinceCode: pCode });
        }
        const provList = Array.from(provMap.entries()).map(([code, name]) => ({ code, name }));
        setProvinces(provList);
        setAllWards(wardList);
      })
      .catch(() => {});
  }, []);

  // Update wards when province changes
  useEffect(() => {
    if (!province) {
      setWards([]);
      setWard("");
      return;
    }
    const provItem = provinces.find((p) => p.name === province);
    if (provItem) {
      const filtered = allWards.filter((w) => w.provinceCode === provItem.code);
      setWards(filtered.map((w) => ({ code: w.code, name: w.name })));
    }
  }, [province, provinces, allWards]);

  const selectedCat = categories.find((c) => c.key === category);

  const fetchListings = useCallback(
    async (p: number) => {
      setLoading(true);
      try {
        const res = await searchMarketplace({
          q: query || undefined,
          category: category || undefined,
          rice_type: riceType || undefined,
          province: province || undefined,
          ward: ward || undefined,
          sort: sort || undefined,
          min_price: minPrice ? Number(minPrice) : undefined,
          max_price: maxPrice ? Number(maxPrice) : undefined,
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
    [query, category, riceType, province, ward, sort, minPrice, maxPrice]
  );

  useEffect(() => {
    fetchListings(page);
  }, [page, fetchListings]);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    setPage(1);
    fetchListings(1);
  }

  function clearFilters() {
    setCategory("");
    setRiceType("");
    setProvince("");
    setWard("");
    setSort("");
    setMinPrice("");
    setMaxPrice("");
    setQuery("");
    setPage(1);
  }

  const hasFilters = category || riceType || province || ward || sort || minPrice || maxPrice;
  const totalPages = result ? Math.ceil(result.total / 20) : 0;

  return (
    <div className="mx-auto max-w-7xl px-4 py-6">
      <form onSubmit={handleSearch} className="flex gap-2 mb-4">
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
        <Button
          type="button"
          variant={showFilter ? "secondary" : "outline"}
          className="gap-1.5"
          onClick={() => setShowFilter(!showFilter)}
        >
          <Filter className="h-4 w-4" />
          Lọc
          {hasFilters && <span className="h-2 w-2 rounded-full bg-primary" />}
        </Button>
      </form>

      {/* Filter panel */}
      {showFilter && (
        <Card className="mb-4">
          <CardContent className="p-4 space-y-4">
            <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Phân loại</label>
                <select
                  value={category}
                  onChange={(e) => { setCategory(e.target.value); setRiceType(""); }}
                  className="w-full h-9 rounded-md border border-input bg-background px-2 text-sm"
                >
                  <option value="">Tất cả</option>
                  {categories.map((c) => (
                    <option key={c.key} value={c.key}>{c.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Loại gạo</label>
                <select
                  value={riceType}
                  onChange={(e) => setRiceType(e.target.value)}
                  className="w-full h-9 rounded-md border border-input bg-background px-2 text-sm"
                  disabled={!category}
                >
                  <option value="">Tất cả</option>
                  {selectedCat?.products.map((p) => (
                    <option key={p.key} value={p.key}>{p.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Tỉnh/Thành phố</label>
                <select
                  value={province}
                  onChange={(e) => { setProvince(e.target.value); setWard(""); }}
                  className="w-full h-9 rounded-md border border-input bg-background px-2 text-sm"
                >
                  <option value="">Tất cả</option>
                  {provinces.map((p) => (
                    <option key={p.code} value={p.name}>{p.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Xã/Phường</label>
                <select
                  value={ward}
                  onChange={(e) => setWard(e.target.value)}
                  className="w-full h-9 rounded-md border border-input bg-background px-2 text-sm"
                  disabled={!province}
                >
                  <option value="">Tất cả</option>
                  {wards.map((w) => (
                    <option key={w.code} value={w.name}>{w.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Giá từ (đ/kg)</label>
                <Input
                  type="number"
                  value={minPrice}
                  onChange={(e) => setMinPrice(e.target.value)}
                  placeholder="Tối thiểu"
                  className="h-9"
                  min="0"
                />
              </div>
              <div>
                <label className="text-xs font-medium mb-1 block text-muted-foreground">Giá đến (đ/kg)</label>
                <Input
                  type="number"
                  value={maxPrice}
                  onChange={(e) => setMaxPrice(e.target.value)}
                  placeholder="Tối đa"
                  className="h-9"
                  min="0"
                />
              </div>
            </div>
            <div className="flex items-center gap-2">
              <label className="text-xs font-medium text-muted-foreground">Sắp xếp:</label>
              <select
                value={sort}
                onChange={(e) => setSort(e.target.value)}
                className="h-9 rounded-md border border-input bg-background px-2 text-sm"
              >
                <option value="">Mới nhất</option>
                <option value="name_asc">Tên gạo (A-Z)</option>
                <option value="name_desc">Tên gạo (Z-A)</option>
                <option value="price_asc">Giá tăng dần</option>
                <option value="price_desc">Giá giảm dần</option>
                <option value="quantity_desc">Số lượng nhiều nhất</option>
              </select>
              <div className="flex-1" />
              <Button type="button" variant="outline" size="sm" onClick={() => { setPage(1); fetchListings(1); }} className="gap-1">
                <Search className="h-3.5 w-3.5" />
                Áp dụng
              </Button>
              {hasFilters && (
                <Button type="button" variant="ghost" size="sm" onClick={clearFilters} className="gap-1">
                  <X className="h-3.5 w-3.5" />
                  Xóa bộ lọc
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Active filters */}
      {hasFilters && !showFilter && (
        <div className="flex flex-wrap gap-2 mb-4">
          {category && <Badge variant="secondary">Loại: {categories.find(c => c.key === category)?.label || category}</Badge>}
          {riceType && <Badge variant="secondary">Gạo: {riceType}</Badge>}
          {province && (
            <Badge variant="secondary">
              <MapPin className="h-3 w-3 mr-1" />{province}
              {ward && ` - ${ward}`}
            </Badge>
          )}
          {(minPrice || maxPrice) && (
            <Badge variant="secondary">
              Giá: {minPrice || "0"} - {maxPrice || "..."}đ/kg
            </Badge>
          )}
          {sort && <Badge variant="secondary"><ArrowUpDown className="h-3 w-3 mr-1" />{sort === "price_asc" ? "Giá tăng" : sort === "price_desc" ? "Giá giảm" : sort === "name_asc" ? "Tên A-Z" : sort === "name_desc" ? "Tên Z-A" : "SL nhiều"}</Badge>}
          <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={clearFilters}>
            Xóa bộ lọc
          </Button>
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
                  <div className="h-56 bg-muted overflow-hidden relative">
                    {listing.images.length > 0 ? (
                      <>
                        <Image
                          src={toThumbnailUrl(listing.images[0])}
                          alt=""
                          fill
                          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
                          className="object-cover scale-110 blur-xl opacity-60"
                        />
                        <Image
                          src={toThumbnailUrl(listing.images[0])}
                          alt={listing.title}
                          fill
                          sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 25vw"
                          className="object-contain relative z-10"
                        />
                      </>
                    ) : (
                      <div className="flex items-center justify-center h-full">
                        <Package className="h-10 w-10 text-muted-foreground/40" />
                      </div>
                    )}
                  </div>
                  <CardContent className="px-3 py-2.5">
                    <h3 className="font-semibold text-sm line-clamp-1 mb-1">
                      {listing.title}
                    </h3>
                    <div className="flex items-center justify-between">
                      <p className="text-base font-bold text-primary">
                        {formatPrice(listing.price_per_kg)}
                      </p>
                      <span className="text-xs text-muted-foreground">{formatQuantity(listing.quantity_kg)}</span>
                    </div>
                    <div className="flex items-center justify-between mt-1 text-xs text-muted-foreground">
                      {listing.province ? (
                        <span className="flex items-center gap-1 truncate">
                          <MapPin className="h-3 w-3 flex-shrink-0" />
                          {listing.province}
                        </span>
                      ) : <span />}
                      <span className="flex-shrink-0">{timeAgo(listing.created_at)}</span>
                    </div>
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
