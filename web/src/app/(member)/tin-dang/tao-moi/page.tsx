"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft, ImagePlus, X, Plus, Trash2 } from "lucide-react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createListing, batchCreateListings, uploadImage, addListingImage, getProductCatalog, type RiceCategory } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

const MAX_IMAGES = 3;

interface ListingForm {
  category: string;
  rice_type: string;
  quantity_kg: string;
  price_per_kg: string;
  harvest_season: string;
  description: string;
  images: { file: File; preview: string }[];
}

function emptyForm(): ListingForm {
  return {
    category: "",
    rice_type: "",
    quantity_kg: "",
    price_per_kg: "",
    harvest_season: "",
    description: "",
    images: [],
  };
}

export default function CreateListingPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<RiceCategory[]>([]);
  const [forms, setForms] = useState<ListingForm[]>([emptyForm()]);
  const fileInputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    getProductCatalog().then(setCategories).catch(() => {});
  }, []);

  useEffect(() => {
    return () => {
      forms.forEach((f) => f.images.forEach((img) => URL.revokeObjectURL(img.preview)));
    };
  }, []);

  function updateForm(index: number, field: string, value: string) {
    setForms((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], [field]: value };
      if (field === "category") {
        next[index].rice_type = "";
      }
      return next;
    });
  }

  function addForm() {
    setForms((prev) => [...prev, emptyForm()]);
  }

  function removeForm(index: number) {
    setForms((prev) => {
      prev[index].images.forEach((img) => URL.revokeObjectURL(img.preview));
      return prev.filter((_, i) => i !== index);
    });
  }

  function handleImageSelect(formIndex: number, e: React.ChangeEvent<HTMLInputElement>) {
    const files = e.target.files;
    if (!files) return;
    const form = forms[formIndex];
    const remaining = MAX_IMAGES - form.images.length;
    const selected = Array.from(files).slice(0, remaining);
    const newImages = selected.map((file) => ({
      file,
      preview: URL.createObjectURL(file),
    }));
    setForms((prev) => {
      const next = [...prev];
      next[formIndex] = { ...next[formIndex], images: [...next[formIndex].images, ...newImages] };
      return next;
    });
    const ref = fileInputRefs.current[formIndex];
    if (ref) ref.value = "";
  }

  function removeImage(formIndex: number, imgIndex: number) {
    setForms((prev) => {
      const next = [...prev];
      const imgs = [...next[formIndex].images];
      URL.revokeObjectURL(imgs[imgIndex].preview);
      imgs.splice(imgIndex, 1);
      next[formIndex] = { ...next[formIndex], images: imgs };
      return next;
    });
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;

    // Validate all forms
    for (let i = 0; i < forms.length; i++) {
      const f = forms[i];
      if (!f.category || !f.rice_type || !f.quantity_kg || !f.price_per_kg) {
        toast.error(`Sản phẩm ${i + 1}: Vui lòng điền đầy đủ thông tin bắt buộc`);
        return;
      }
    }

    setLoading(true);
    try {
      if (forms.length === 1) {
        // Single listing
        const f = forms[0];
        const listing = await createListing(token, {
          category: f.category,
          rice_type: f.rice_type,
          quantity_kg: Number(f.quantity_kg),
          price_per_kg: Number(f.price_per_kg),
          harvest_season: f.harvest_season,
          description: f.description,
        });
        // Upload images
        for (const img of f.images) {
          try {
            const { url } = await uploadImage(token, img.file, "listings");
            await addListingImage(token, listing.id, url);
          } catch {
            // continue
          }
        }
      } else {
        // Batch create
        const items = forms.map((f) => ({
          category: f.category,
          rice_type: f.rice_type,
          quantity_kg: Number(f.quantity_kg),
          price_per_kg: Number(f.price_per_kg),
          harvest_season: f.harvest_season,
          description: f.description,
        }));
        const result = await batchCreateListings(token, items);
        // Upload images for each listing
        for (let i = 0; i < forms.length; i++) {
          const listing = result.listings[i];
          if (!listing) continue;
          for (const img of forms[i].images) {
            try {
              const { url } = await uploadImage(token, img.file, "listings");
              await addListingImage(token, listing.id, url);
            } catch {
              // continue
            }
          }
        }
      }

      toast.success(`Tạo ${forms.length} tin đăng thành công!`);
      router.push("/tin-dang");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tạo tin đăng thất bại");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="max-w-3xl mx-auto">
      <Link href="/tin-dang" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="h-4 w-4" />
        Quay lại
      </Link>

      <div className="flex items-center justify-between mb-4">
        <h1 className="text-xl font-bold">Đăng tin</h1>
        <Button type="button" variant="outline" className="gap-1.5" onClick={addForm}>
          <Plus className="h-4 w-4" />
          Thêm sản phẩm
        </Button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {forms.map((form, fi) => {
          const selectedCat = categories.find((c) => c.key === form.category);
          return (
            <Card key={fi}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">
                    Sản phẩm {forms.length > 1 ? `#${fi + 1}` : ""}
                  </CardTitle>
                  {forms.length > 1 && (
                    <Button type="button" variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => removeForm(fi)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium mb-1 block">Phân loại *</label>
                    <select
                      value={form.category}
                      onChange={(e) => updateForm(fi, "category", e.target.value)}
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
                      onChange={(e) => updateForm(fi, "rice_type", e.target.value)}
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
                      onChange={(e) => updateForm(fi, "quantity_kg", e.target.value)}
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
                      onChange={(e) => updateForm(fi, "price_per_kg", e.target.value)}
                      placeholder="VD: 15000"
                      required
                      min="1"
                    />
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Vụ mùa</label>
                  <Input
                    value={form.harvest_season}
                    onChange={(e) => updateForm(fi, "harvest_season", e.target.value)}
                    placeholder="VD: Đông Xuân 2025-2026"
                  />
                </div>

                <div>
                  <label className="text-sm font-medium mb-1 block">Mô tả</label>
                  <textarea
                    value={form.description}
                    onChange={(e) => updateForm(fi, "description", e.target.value)}
                    placeholder="Mô tả chi tiết về sản phẩm..."
                    className="w-full min-h-20 rounded-md border border-input bg-background px-3 py-2 text-sm"
                  />
                </div>

                {/* Image upload */}
                <div>
                  <label className="text-sm font-medium mb-2 block">
                    Hình ảnh ({form.images.length}/{MAX_IMAGES})
                  </label>
                  <div className="flex flex-wrap gap-3">
                    {form.images.map((img, i) => (
                      <div key={i} className="relative w-24 h-24 rounded-lg overflow-hidden border">
                        <img src={img.preview} alt="" className="w-full h-full object-cover" />
                        <button
                          type="button"
                          onClick={() => removeImage(fi, i)}
                          className="absolute top-1 right-1 bg-black/60 text-white rounded-full p-0.5 hover:bg-black/80"
                        >
                          <X className="h-3.5 w-3.5" />
                        </button>
                      </div>
                    ))}
                    {form.images.length < MAX_IMAGES && (
                      <button
                        type="button"
                        onClick={() => fileInputRefs.current[fi]?.click()}
                        className="w-24 h-24 rounded-lg border-2 border-dashed border-muted-foreground/30 flex flex-col items-center justify-center text-muted-foreground hover:border-primary hover:text-primary transition-colors"
                      >
                        <ImagePlus className="h-6 w-6" />
                        <span className="text-xs mt-1">Thêm ảnh</span>
                      </button>
                    )}
                  </div>
                  <input
                    ref={(el) => { fileInputRefs.current[fi] = el; }}
                    type="file"
                    accept="image/jpeg,image/png,image/webp"
                    multiple
                    onChange={(e) => handleImageSelect(fi, e)}
                    className="hidden"
                  />
                </div>
              </CardContent>
            </Card>
          );
        })}

        <div className="flex gap-3">
          <Button type="submit" className="flex-1" disabled={loading}>
            {loading ? "Đang tạo..." : `Đăng ${forms.length > 1 ? `${forms.length} tin` : "tin"}`}
          </Button>
          {forms.length < 10 && (
            <Button type="button" variant="outline" onClick={addForm} className="gap-1.5">
              <Plus className="h-4 w-4" />
              Thêm
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
