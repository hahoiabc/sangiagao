"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, Zap, ChevronDown, ChevronRight, Check, ImagePlus, X } from "lucide-react";
import Link from "next/link";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { batchCreateListings, getProductCatalog, uploadImage, addListingImage, type RiceCategory, type RiceProduct } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

const MAX_IMAGES = 2;

interface ImageItem {
  file: File;
  preview: string;
}

interface ProductEntry {
  product: RiceProduct;
  selected: boolean;
  price: string;
  quantity: string;
  season: string;
  description: string;
  expanded: boolean;
  images: ImageItem[];
}

function createEntry(product: RiceProduct): ProductEntry {
  return { product, selected: false, price: "", quantity: "", season: "", description: "", expanded: false, images: [] };
}

export default function QuickBatchPage() {
  const { user } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<RiceCategory[]>([]);
  const [loadingCatalog, setLoadingCatalog] = useState(true);

  // null = category grid, non-null = product list
  const [selectedCategory, setSelectedCategory] = useState<RiceCategory | null>(null);
  const [entries, setEntries] = useState<ProductEntry[]>([]);
  const fileInputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    getProductCatalog()
      .then(setCategories)
      .catch(() => {})
      .finally(() => setLoadingCatalog(false));
  }, []);

  function selectCategory(cat: RiceCategory) {
    setSelectedCategory(cat);
    setEntries(cat.products.map(createEntry));
  }

  function backToCategories() {
    setSelectedCategory(null);
    setEntries([]);
  }

  function updateEntry(index: number, field: keyof ProductEntry, value: string | boolean) {
    setEntries((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], [field]: value };
      return next;
    });
  }

  function toggleExpand(index: number) {
    setEntries((prev) => {
      const next = [...prev];
      const wasExpanded = next[index].expanded;
      next[index] = { ...next[index], expanded: !wasExpanded };
      // Auto-select when expanding
      if (!wasExpanded && !next[index].selected) {
        next[index] = { ...next[index], selected: true, expanded: true };
      }
      return next;
    });
  }

  function toggleSelect(index: number) {
    setEntries((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], selected: !next[index].selected };
      return next;
    });
  }

  function handleImageSelect(entryIndex: number, e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files;
    if (!files) return;
    const entry = entries[entryIndex];
    const remaining = MAX_IMAGES - entry.images.length;
    const selected = Array.from(files).slice(0, remaining);
    const newImages: ImageItem[] = selected.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));
    setEntries((prev) => {
      const next = [...prev];
      next[entryIndex] = { ...next[entryIndex], images: [...next[entryIndex].images, ...newImages] };
      return next;
    });
    const ref = fileInputRefs.current[entryIndex];
    if (ref) ref.value = "";
  }

  function removeImage(entryIndex: number, imgIndex: number) {
    setEntries((prev) => {
      const next = [...prev];
      const imgs = [...next[entryIndex].images];
      URL.revokeObjectURL(imgs[imgIndex].preview);
      imgs.splice(imgIndex, 1);
      next[entryIndex] = { ...next[entryIndex], images: imgs };
      return next;
    });
  }

  function clearAll() {
    setEntries((prev) => prev.map((e) => ({ ...e, selected: false, expanded: false })));
  }

  const selectedCount = entries.filter((e) => e.selected).length;

  async function handleSubmit() {
    if (!user) return;

    const selected = entries.filter((e) => e.selected);
    if (selected.length === 0) {
      toast.error("Vui lòng chọn ít nhất 1 sản phẩm");
      return;
    }

    // Validate
    const errors: string[] = [];
    const items: Record<string, unknown>[] = [];
    for (const e of selected) {
      const price = Number(e.price);
      const qty = Number(e.quantity);
      if (!price || price <= 5000 || price >= 99000) {
        errors.push(`${e.product.label}: Giá phải từ 5,001 đến 98,999 đ/kg`);
        continue;
      }
      if (!qty || qty <= 500 || qty >= 100000000) {
        errors.push(`${e.product.label}: Số lượng phải từ 501 đến 99,999,999 kg`);
        continue;
      }
      if (e.season) {
        const parts = e.season.split("/");
        if (parts.length === 3) {
          const picked = new Date(Number(parts[2]), Number(parts[1]) - 1, Number(parts[0]));
          if (picked > new Date()) {
            errors.push(`${e.product.label}: Mùa vụ phải trước ngày hiện tại`);
            continue;
          }
        }
      }
      const item: Record<string, unknown> = {
        category: selectedCategory!.key,
        rice_type: e.product.key,
        price_per_kg: price,
        quantity_kg: qty,
      };
      if (e.season) item.harvest_season = e.season;
      if (e.description) item.description = e.description;
      items.push(item);
    }

    if (errors.length > 0) {
      toast.error(errors.join("\n"));
      return;
    }

    if (items.length > 20) {
      toast.error("Tối đa 20 sản phẩm mỗi lần đăng");
      return;
    }

    setLoading(true);
    try {
      const result = await batchCreateListings("", items);
      const listings = result.listings ?? [];
      const count = listings.length || items.length;

      // Upload images for each listing
      for (let i = 0; i < listings.length && i < selected.length; i++) {
        const listing = listings[i];
        const entry = selected[i];
        if (!listing?.id || entry.images.length === 0) continue;
        for (const img of entry.images) {
          try {
            const { url } = await uploadImage("", img.file, "listings");
            await addListingImage("", listing.id, url);
          } catch {
            // continue
          }
        }
      }

      toast.success(`Đã đăng ${count} tin thành công!`);
      router.push("/tin-dang");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Đăng tin thất bại");
    } finally {
      setLoading(false);
    }
  }

  // ─── Category Grid ───
  if (loadingCatalog) {
    return (
      <div className="max-w-3xl mx-auto flex items-center justify-center py-20">
        <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }

  if (!selectedCategory) {
    return (
      <div className="max-w-3xl mx-auto">
        <Link href="/tin-dang" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
          <ArrowLeft className="h-4 w-4" />
          Quay lại
        </Link>

        <div className="flex items-center gap-2 mb-6">
          <Zap className="h-5 w-5 text-primary" />
          <h1 className="text-xl font-bold">Đăng nhanh theo danh mục</h1>
        </div>

        <p className="text-sm text-muted-foreground mb-4">
          Chọn danh mục để xem và đăng nhiều sản phẩm cùng lúc.
        </p>

        {categories.length === 0 ? (
          <p className="text-muted-foreground text-center py-10">Không có danh mục nào</p>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-4">
            {categories.map((cat) => (
              <Card
                key={cat.key}
                className="cursor-pointer hover:border-primary hover:shadow-md transition-all"
                onClick={() => selectCategory(cat)}
              >
                <CardContent className="flex flex-col items-center justify-center py-6 px-4 text-center">
                  <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-3">
                    <span className="text-xl">🌾</span>
                  </div>
                  <p className="font-semibold text-sm">{cat.label}</p>
                  <p className="text-xs text-muted-foreground mt-1">{cat.products.length} sản phẩm</p>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    );
  }

  // ─── Product List ───
  return (
    <div className="max-w-3xl mx-auto">
      <button
        onClick={backToCategories}
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4"
      >
        <ArrowLeft className="h-4 w-4" />
        Chọn danh mục khác
      </button>

      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-xl font-bold">{selectedCategory.label}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Đã chọn {selectedCount} / {entries.length} sản phẩm
          </p>
        </div>
        {selectedCount > 0 && (
          <Button variant="ghost" size="sm" onClick={clearAll}>
            Bỏ chọn tất cả
          </Button>
        )}
      </div>

      <div className="space-y-2 mb-6">
        {entries.map((entry, i) => (
          <Card key={entry.product.key} className={entry.selected ? "border-primary/50" : ""}>
            <div
              className="flex items-center gap-3 px-4 py-3 cursor-pointer select-none"
              onClick={() => toggleExpand(i)}
            >
              {/* Checkbox */}
              <button
                type="button"
                onClick={(e) => { e.stopPropagation(); toggleSelect(i); }}
                className={`w-5 h-5 rounded border-2 flex items-center justify-center shrink-0 transition-colors ${
                  entry.selected ? "bg-primary border-primary text-white" : "border-muted-foreground/40"
                }`}
              >
                {entry.selected && <Check className="h-3 w-3" />}
              </button>

              {/* Product name */}
              <span className="flex-1 text-sm font-medium">{entry.product.label}</span>

              {/* Price preview */}
              {entry.selected && entry.price && (
                <Badge variant="secondary" className="text-xs">
                  {Number(entry.price).toLocaleString("vi-VN")}đ/kg
                </Badge>
              )}

              {/* Expand icon */}
              {entry.expanded ? (
                <ChevronDown className="h-4 w-4 text-muted-foreground" />
              ) : (
                <ChevronRight className="h-4 w-4 text-muted-foreground" />
              )}
            </div>

            {/* Expanded form */}
            {entry.expanded && (
              <CardContent className="pt-0 pb-4 space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-medium text-muted-foreground mb-1 block">Giá (đ/kg) *</label>
                    <Input
                      type="number"
                      value={entry.price}
                      onChange={(e) => updateEntry(i, "price", e.target.value)}
                      placeholder="VD: 15000"
                      min="1"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-medium text-muted-foreground mb-1 block">Số lượng (kg) *</label>
                    <Input
                      type="number"
                      value={entry.quantity}
                      onChange={(e) => updateEntry(i, "quantity", e.target.value)}
                      placeholder="VD: 1000"
                      min="1"
                    />
                  </div>
                </div>
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1 block">Vụ mùa</label>
                  <div className="flex gap-2">
                    <select
                      className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                      value={entry.season ? entry.season.split("/")[0] : ""}
                      onChange={(e) => {
                        const parts = entry.season ? entry.season.split("/") : ["", "", ""];
                        parts[0] = e.target.value;
                        updateEntry(i, "season", parts.join("/"));
                      }}
                    >
                      <option value="">Ngày</option>
                      {Array.from({ length: 31 }, (_, k) => k + 1).map((d) => (
                        <option key={d} value={String(d).padStart(2, "0")}>{d}</option>
                      ))}
                    </select>
                    <select
                      className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                      value={entry.season ? entry.season.split("/")[1] : ""}
                      onChange={(e) => {
                        const parts = entry.season ? entry.season.split("/") : ["", "", ""];
                        parts[1] = e.target.value;
                        updateEntry(i, "season", parts.join("/"));
                      }}
                    >
                      <option value="">Tháng</option>
                      {Array.from({ length: 12 }, (_, k) => k + 1).map((m) => (
                        <option key={m} value={String(m).padStart(2, "0")}>{m}</option>
                      ))}
                    </select>
                    <select
                      className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                      value={entry.season ? entry.season.split("/")[2] : ""}
                      onChange={(e) => {
                        const parts = entry.season ? entry.season.split("/") : ["", "", ""];
                        parts[2] = e.target.value;
                        updateEntry(i, "season", parts.join("/"));
                      }}
                    >
                      <option value="">Năm</option>
                      {Array.from({ length: 6 }, (_, k) => new Date().getFullYear() - 5 + k).map((y) => (
                        <option key={y} value={String(y)}>{y}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-1 block">Mô tả</label>
                  <textarea
                    value={entry.description}
                    onChange={(e) => updateEntry(i, "description", e.target.value)}
                    placeholder="Mô tả thêm..."
                    className="w-full min-h-16 rounded-md border border-input bg-background px-3 py-2 text-sm"
                  />
                </div>
                {/* Image upload */}
                <div>
                  <label className="text-xs font-medium text-muted-foreground mb-2 block">
                    Hình ảnh ({entry.images.length}/{MAX_IMAGES})
                  </label>
                  <div className="flex flex-wrap gap-2">
                    {entry.images.map((img, imgIdx) => (
                      <div key={imgIdx} className="relative w-20 h-20 rounded-lg overflow-hidden border">
                        <img src={img.preview} alt="Ảnh sản phẩm" className="w-full h-full object-cover" />
                        <button
                          type="button"
                          onClick={() => removeImage(i, imgIdx)}
                          className="absolute top-0.5 right-0.5 bg-black/60 text-white rounded-full p-0.5 hover:bg-black/80"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </div>
                    ))}
                    {entry.images.length < MAX_IMAGES && (
                      <button
                        type="button"
                        onClick={() => fileInputRefs.current[i]?.click()}
                        className="w-20 h-20 rounded-lg border-2 border-dashed border-muted-foreground/30 flex flex-col items-center justify-center text-muted-foreground hover:border-primary hover:text-primary transition-colors"
                      >
                        <ImagePlus className="h-5 w-5" />
                        <span className="text-[10px] mt-0.5">Thêm ảnh</span>
                      </button>
                    )}
                  </div>
                  <input
                    ref={(el) => { fileInputRefs.current[i] = el; }}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    multiple
                    onChange={(e) => handleImageSelect(i, e)}
                    className="hidden"
                  />
                </div>
              </CardContent>
            )}
          </Card>
        ))}
      </div>

      {/* Submit */}
      <div className="sticky bottom-4">
        <Button
          className="w-full gap-2"
          disabled={loading || selectedCount === 0}
          onClick={handleSubmit}
        >
          <Zap className="h-4 w-4" />
          {loading ? "Đang đăng..." : `Đăng ${selectedCount} tin`}
        </Button>
      </div>
    </div>
  );
}
