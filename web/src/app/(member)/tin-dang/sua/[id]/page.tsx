"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Save } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { getListingDetail, updateListing, type ListingDetail } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

export default function EditListingPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const router = useRouter();
  const [listing, setListing] = useState<ListingDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Editable fields (same as mobile: price, quantity, harvest_season, description)
  const [pricePerKg, setPricePerKg] = useState("");
  const [quantityKg, setQuantityKg] = useState("");
  const [harvestSeason, setHarvestSeason] = useState("");
  const [description, setDescription] = useState("");

  useEffect(() => {
    if (id) {
      getListingDetail(id)
        .then((detail) => {
          setListing(detail);
          setPricePerKg(String(Math.round(detail.price_per_kg)));
          setQuantityKg(String(Math.round(detail.quantity_kg)));
          setHarvestSeason(detail.harvest_season || "");
          setDescription(detail.description || "");
        })
        .catch(() => toast.error("Không tìm thấy tin đăng"))
        .finally(() => setLoading(false));
    }
  }, [id]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!id) return;

    const price = Number(pricePerKg);
    const qty = Number(quantityKg);
    if (!price || price <= 5000 || price >= 99000) {
      toast.error("Giá phải từ 5,001 đến 98,999 đ/kg");
      return;
    }
    if (!qty || qty <= 500 || qty >= 100000000) {
      toast.error("Số lượng phải từ 501 đến 99,999,999 kg");
      return;
    }
    if (harvestSeason.trim()) {
      const parts = harvestSeason.split("/");
      if (parts.length === 3) {
        const picked = new Date(Number(parts[2]), Number(parts[1]) - 1, Number(parts[0]));
        if (picked > new Date()) {
          toast.error("Mùa vụ phải trước ngày hiện tại");
          return;
        }
      }
    }

    setSaving(true);
    try {
      const data: Record<string, unknown> = {
        price_per_kg: price,
        quantity_kg: qty,
      };
      if (harvestSeason.trim()) data.harvest_season = harvestSeason.trim();
      if (description.trim()) data.description = description.trim();

      await updateListing("", id, data);
      toast.success("Đã cập nhật tin đăng");
      router.push("/tin-dang");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-6 w-32" />
        <Skeleton className="h-80 w-full rounded-lg" />
      </div>
    );
  }

  if (!listing) {
    return (
      <div className="max-w-2xl mx-auto text-center py-12">
        <p className="text-muted-foreground">Không tìm thấy tin đăng</p>
        <Link href="/tin-dang">
          <Button variant="outline" className="mt-4">Quay lại</Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <Link href="/tin-dang" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4">
        <ArrowLeft className="h-4 w-4" />
        Quay lại
      </Link>

      <h1 className="text-xl font-bold mb-4">Sửa tin đăng</h1>

      <Card>
        <CardContent className="p-6">
          {/* Read-only info */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold">{listing.title}</h2>
            {listing.rice_type && (
              <p className="text-sm text-muted-foreground mt-1">{listing.rice_type}</p>
            )}
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="text-sm font-medium mb-1 block">Giá (đ/kg)</label>
              <Input
                type="number"
                value={pricePerKg}
                onChange={(e) => setPricePerKg(e.target.value)}
                placeholder="VD: 15000"
                min="1"
                required
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Số lượng (kg)</label>
              <Input
                type="number"
                value={quantityKg}
                onChange={(e) => setQuantityKg(e.target.value)}
                placeholder="VD: 1000"
                min="1"
                required
              />
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Vụ mùa</label>
              <div className="flex gap-2">
                <select
                  className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                  value={harvestSeason ? harvestSeason.split("/")[0] : ""}
                  onChange={(e) => {
                    const parts = harvestSeason ? harvestSeason.split("/") : ["", "", ""];
                    parts[0] = e.target.value;
                    setHarvestSeason(parts.join("/"));
                  }}
                >
                  <option value="">Ngày</option>
                  {Array.from({ length: 31 }, (_, k) => k + 1).map((d) => (
                    <option key={d} value={String(d).padStart(2, "0")}>{d}</option>
                  ))}
                </select>
                <select
                  className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                  value={harvestSeason ? harvestSeason.split("/")[1] : ""}
                  onChange={(e) => {
                    const parts = harvestSeason ? harvestSeason.split("/") : ["", "", ""];
                    parts[1] = e.target.value;
                    setHarvestSeason(parts.join("/"));
                  }}
                >
                  <option value="">Tháng</option>
                  {Array.from({ length: 12 }, (_, k) => k + 1).map((m) => (
                    <option key={m} value={String(m).padStart(2, "0")}>{m}</option>
                  ))}
                </select>
                <select
                  className="flex-1 rounded-md border border-input bg-background px-2 py-2 text-sm"
                  value={harvestSeason ? harvestSeason.split("/")[2] : ""}
                  onChange={(e) => {
                    const parts = harvestSeason ? harvestSeason.split("/") : ["", "", ""];
                    parts[2] = e.target.value;
                    setHarvestSeason(parts.join("/"));
                  }}
                >
                  <option value="">Năm</option>
                  {Array.from({ length: new Date().getFullYear() - 2000 + 6 }, (_, k) => 2000 + k).map((y) => (
                    <option key={y} value={String(y)}>{y}</option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="text-sm font-medium mb-1 block">Mô tả thêm</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Mô tả chi tiết về sản phẩm..."
                className="w-full min-h-20 rounded-md border border-input bg-background px-3 py-2 text-sm"
              />
            </div>

            <Button type="submit" className="w-full gap-2" disabled={saving}>
              <Save className="h-4 w-4" />
              {saving ? "Đang lưu..." : "Lưu thay đổi"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
