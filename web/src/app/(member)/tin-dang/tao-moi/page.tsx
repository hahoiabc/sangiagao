"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { createListing, getProductCatalog, type RiceCategory } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

export default function CreateListingPage() {
  const { token } = useAuth();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState<RiceCategory[]>([]);
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

  const selectedCat = categories.find((c) => c.key === form.category);

  function update(field: string, value: string) {
    setForm((f) => ({ ...f, [field]: value }));
    if (field === "category") {
      setForm((f) => ({ ...f, rice_type: "" }));
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setLoading(true);
    try {
      await createListing(token, {
        ...form,
        quantity_kg: Number(form.quantity_kg),
        price_per_kg: Number(form.price_per_kg),
      });
      toast.success("Tạo tin đăng thành công!");
      router.push("/tin-dang");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tạo tin đăng thất bại");
    } finally {
      setLoading(false);
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

            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Đang tạo..." : "Đăng Tin"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
