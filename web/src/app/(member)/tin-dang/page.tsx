"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import Image from "next/image";
import { Plus, Package, Edit, Trash2, Eye, Calendar, Zap, ChevronLeft, ChevronRight, Check } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getMyListings, deleteListing, batchDeleteOwnListings, toThumbnailUrl, type Listing, type PaginatedResponse } from "@/services/api";
import { formatPrice, formatQuantity, timeAgo } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

// Match mobile status labels exactly
const statusLabels: Record<string, { label: string; color: "default" | "secondary" | "destructive" }> = {
  active: { label: "Đang hiển thị", color: "default" },
  hidden: { label: "Đã ẩn", color: "secondary" },
  hidden_subscription: { label: "Đã ẩn", color: "secondary" },
  deleted: { label: "Đã xóa", color: "destructive" },
};

const PAGE_SIZE = 20;

export default function MyListingsPage() {
  const { user } = useAuth();
  const [result, setResult] = useState<PaginatedResponse<Listing> | null>(null);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [batchDeleting, setBatchDeleting] = useState(false);

  const fetchPage = useCallback((p: number) => {
    setLoading(true);
    getMyListings("", p, PAGE_SIZE)
      .then((r) => { setResult(r); setPage(p); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (user) fetchPage(1);
  }, [user, fetchPage]);

  async function handleDelete(id: string) {
    if (!confirm("Bạn có chắc muốn xóa tin đăng này?")) return;
    try {
      await deleteListing("", id);
      // Remove from UI immediately, then refresh
      setResult((prev) => prev ? { ...prev, data: prev.data.filter((l) => l.id !== id), total: prev.total - 1 } : prev);
      toast.success("Đã xóa tin đăng");
      fetchPage(page);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa thất bại");
    }
  }

  function toggleSelect(id: string) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  function toggleSelectAll() {
    if (!result) return;
    if (selected.size === result.data.length) {
      setSelected(new Set());
    } else {
      setSelected(new Set(result.data.map((l) => l.id)));
    }
  }

  async function handleBatchDelete() {
    if (selected.size === 0) return;
    if (!confirm(`Bạn có chắc muốn xóa ${selected.size} tin đăng?`)) return;
    setBatchDeleting(true);
    try {
      await batchDeleteOwnListings("", Array.from(selected));
      setSelected(new Set());
      toast.success(`Đã xóa ${selected.size} tin đăng`);
      fetchPage(page);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa hàng loạt thất bại");
    } finally {
      setBatchDeleting(false);
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Tin đăng của tôi</h1>
        <div className="flex gap-2">
          {selected.size > 0 && (
            <Button variant="destructive" className="gap-2" onClick={handleBatchDelete} disabled={batchDeleting}>
              <Trash2 className="h-4 w-4" />
              Xóa {selected.size} tin
            </Button>
          )}
          <Link href="/tin-dang/dang-nhieu">
            <Button variant="outline" className="gap-2">
              <Zap className="h-4 w-4" />
              Đăng nhanh
            </Button>
          </Link>
          <Link href="/tin-dang/tao-moi">
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Đăng tin
            </Button>
          </Link>
        </div>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <div className="space-y-3">
          {result.data.length > 1 && (
            <button onClick={toggleSelectAll} className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-2">
              <div className={`w-4 h-4 rounded border flex items-center justify-center ${selected.size === result.data.length ? "bg-primary border-primary text-white" : "border-muted-foreground/40"}`}>
                {selected.size === result.data.length && <Check className="h-3 w-3" />}
              </div>
              {selected.size === result.data.length ? "Bỏ chọn tất cả" : "Chọn tất cả"}
            </button>
          )}
          {result.data.map((listing) => {
            const st = statusLabels[listing.status] || { label: listing.status, color: "secondary" as const };
            const isSelected = selected.has(listing.id);
            return (
              <Card key={listing.id} className={isSelected ? "border-primary/50" : ""}>
                <CardContent className="p-4 space-y-3">
                  <div className="flex gap-3">
                    {/* Checkbox */}
                    <button onClick={() => toggleSelect(listing.id)} className="flex-shrink-0 mt-1">
                      <div className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-colors ${isSelected ? "bg-primary border-primary text-white" : "border-muted-foreground/40"}`}>
                        {isSelected && <Check className="h-3 w-3" />}
                      </div>
                    </button>
                    {/* Images row */}
                    <div className="flex gap-2 overflow-x-auto flex-1">
                    {listing.images.length > 0 ? (
                      listing.images.slice(0, 4).map((img, i) => (
                        <div key={i} className="h-20 w-20 rounded-md bg-muted overflow-hidden flex-shrink-0 relative">
                          <Image src={img} alt={`${listing.title} - ${i + 1}`} fill sizes="80px" className="object-cover" />
                        </div>
                      ))
                    ) : (
                      <div className="h-20 w-20 rounded-md bg-muted flex items-center justify-center flex-shrink-0">
                        <Package className="h-6 w-6 text-muted-foreground/40" />
                      </div>
                    )}
                    </div>
                  </div>
                  {/* Info + Actions */}
                  <div className="flex items-start gap-3">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-semibold text-sm truncate">{listing.title}</h3>
                        <Badge variant={st.color} className="text-xs flex-shrink-0">
                          {st.label}
                        </Badge>
                      </div>
                      <p className="text-sm text-primary font-bold">
                        {formatPrice(listing.price_per_kg)}
                      </p>
                      <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground flex-wrap">
                        <span className="flex items-center gap-1">
                          <Package className="h-3 w-3" />
                          {formatQuantity(listing.quantity_kg)}
                        </span>
                        {listing.harvest_season && (
                          <span className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {listing.harvest_season}
                          </span>
                        )}
                        <span className="flex items-center gap-1">
                          <Eye className="h-3 w-3" />
                          {listing.view_count}
                        </span>
                        <span>{timeAgo(listing.created_at)}</span>
                      </div>
                    </div>
                    {/* Actions */}
                    <div className="flex items-center gap-1 flex-shrink-0">
                      <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
                        <Link href={`/tin-dang/sua/${listing.id}`}>
                          <Edit className="h-4 w-4" />
                        </Link>
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-destructive"
                        onClick={() => handleDelete(listing.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
          {result.total > PAGE_SIZE && (
            <div className="flex items-center justify-center gap-3 mt-4">
              <Button variant="outline" size="sm" disabled={page <= 1} onClick={() => fetchPage(page - 1)}>
                <ChevronLeft className="h-4 w-4 mr-1" /> Trước
              </Button>
              <span className="text-sm text-muted-foreground">
                Trang {page} / {Math.ceil(result.total / PAGE_SIZE)}
              </span>
              <Button variant="outline" size="sm" disabled={page >= Math.ceil(result.total / PAGE_SIZE)} onClick={() => fetchPage(page + 1)}>
                Sau <ChevronRight className="h-4 w-4 ml-1" />
              </Button>
            </div>
          )}
        </div>
      ) : (
        <div className="text-center py-12">
          <Package className="h-12 w-12 text-muted-foreground/40 mx-auto mb-3" />
          <p className="text-muted-foreground mb-4">Bạn chưa có tin đăng nào</p>
          <Link href="/tin-dang/tao-moi">
            <Button className="gap-2">
              <Plus className="h-4 w-4" />
              Đăng tin đầu tiên
            </Button>
          </Link>
        </div>
      )}
    </div>
  );
}
