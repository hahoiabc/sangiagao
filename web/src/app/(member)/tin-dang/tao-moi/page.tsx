"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, ImagePlus, X } from "lucide-react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createListing, uploadImage, addListingImage, getProductCatalog, type RiceCategory } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

const MAX_IMAGES = 3;

export default function CreateListingPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<RiceCategory[]>([]);
  const [images, setImages] = useState<{ file: File; preview: string }[]>([]);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [form, setForm] = useState({
    title: "",
    category: "",
    rice_type: "",
    province: "",
    district: "",
    quantity_kg: "",
    price_per_kg: "",
    harvest_season: "",
    description: "",
  });

  useEffect(() => {
    getProductCatalog().then(setCategories).catch(() => {});
  }, []);

  useEffect(() => {
    return () => {
      images.forEach((img) => URL.revokeObjectURL(img.preview));
    };
  }, [images]);

  const selectedCat = categories.find((c) => c.key === form.category);

  function update(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }));
    if (field === "category") {
      setForm((f) => ({ ...f, rice_type: "" }));
    }
  }

  function handleImageSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files;
    if (!files) return;

    const remaining = MAX_IMAGES - images.length;
    const selected = Array.from(files).slice(0, remaining);

    const newImages = selected.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));

    setImages((prev) => [...prev, ...newImages]);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  function removeImage(index: number) {
    setImages((prev) => {
      URL.revokeObjectURL(prev[index].preview);
      return prev.filter((_, i) => i !== index);
    });
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setLoading(true);
    try {
      // 1. Create listing
      const listing = await createListing(token, {
        ...form,
        quantity_kg: Number(form.quantity_kg),
        price_per_kg: Number(form.price_per_kg),
      });

      // 2. Upload images and attach to listing
      if (images.length > 0) {
        setUploading(true);
        for (const img of images) {
          try {
            const { url } = await uploadImage(token, img.file, "listings");
            await addListingImage(token, listing.id, url);
          } catch {
            // Continue with other images if one fails
          }
        }
        setUploading(false);
      }

      toast.success("Tạo tin đăng thành công!");
      router.push("/tin-dang");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tạo tin đăng thất bại");
    } finally {
      setLoading(false);
      setUploading(false);
    }
  }

  return (
    <div className="max-w-2xl mx-auto">
      <Link href="/tin-dang" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="h-4 w-4" />
        Quay lại
      </Link>

      <Card>
        <CardHeader>
          <CardTitle>Tạo Tin Đăng Mới</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Tiêu đề *</label>
              <Input
                value={form.title}
                onChange={(e) => update("title", e.target.value)}
                placeholder="VD: Bán gạo ST25 Long An"
                required
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-1 block">Phân loại *</label>
                <select
                  value={form.category}
                  onChange={(e) => update("category", e.target.value)}
                  className="w-full h-10 rounded-md border border-input bg-background px-3 text-sm"
                  required
                >
                  <option value="">Chọn phân loại</option>
                  {categories.map((c) => (
                    <option key={c.key} value={c.key}>{c.label}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Loại gạo *</label>
                <select
                  value={form.rice_type}
                  onChange={(e) => update("rice_type", e.target.value)}
                  className="w-full h-10 rounded-md border border-input bg-background px-3 text-sm"
                  required
                  disabled={!form.category}
                >
                  <option value="">Chọn loại gạo</option>
                  {selectedCat?.products.map((p) => (
                    <option key={p.key} value={p.key}>{p.label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-1 block">Số lượng (kg) *</label>
                <Input
                  type="number"
                  value={form.quantity_kg}
                  onChange={(e) => update("quantity_kg", e.target.value)}
                  placeholder="VD: 1000"
                  required
                  min="1"
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Giá (đ/kg) *</label>
                <Input
                  type="number"
                  value={form.price_per_kg}
                  onChange={(e) => update("price_per_kg", e.target.value)}
                  placeholder="VD: 15000"
                  required
                  min="1"
                />
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-sm font-medium mb-1 block">Tỉnh/Thành</label>
                <Input
                  value={form.province}
                  onChange={(e) => update("province", e.target.value)}
                  placeholder="VD: Long An"
                />
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Quận/Huyện</label>
                <Input
                  value={form.district}
                  onChange={(e) => update("district", e.target.value)}
                  placeholder="VD: Tân An"
                />
              </div>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Vụ mùa</label>
              <Input
                value={form.harvest_season}
                onChange={(e) => update("harvest_season", e.target.value)}
                placeholder="VD: Đông Xuân 2025-2026"
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Mô tả</label>
              <textarea
                value={form.description}
                onChange={(e) => update("description", e.target.value)}
                placeholder="Mô tả chi tiết về sản phẩm..."
                className="w-full min-h-24 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </div>

            {/* Image upload */}
            <div>
              <label className="text-sm font-medium mb-2 block">
                Hình ảnh ({images.length}/{MAX_IMAGES})
              </label>
              <div className="flex flex-wrap gap-3">
                {images.map((img, i) => (
                  <div key={i} className="relative w-24 h-24 rounded-lg overflow-hidden border">
                    <img src={img.preview} alt="" className="w-full h-full object-cover" />
                    <button
                      type="button"
                      onClick={() => removeImage(i)}
                      className="absolute top-1 right-1 bg-black/60 text-white rounded-full p-0.5 hover:bg-black/80"
                    >
                      <X className="h-3.5 w-3.5" />
                    </button>
                  </div>
                ))}
                {images.length < MAX_IMAGES && (
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="w-24 h-24 rounded-lg border-2 border-dashed border-muted-foreground/30 flex flex-col items-center justify-center text-muted-foreground hover:border-primary hover:text-primary transition-colors"
                  >
                    <ImagePlus className="h-6 w-6" />
                    <span className="text-xs mt-1">Thêm ảnh</span>
                  </button>
                )}
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept="image/jpeg,image/png,image/webp"
                multiple
                onChange={handleImageSelect}
                className="hidden"
              />
            </div>

            <Button type="submit" className="w-full" disabled={loading || uploading}>
              {uploading ? "Đang tải ảnh..." : loading ? "Đang tạo..." : "Đăng Tin"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
