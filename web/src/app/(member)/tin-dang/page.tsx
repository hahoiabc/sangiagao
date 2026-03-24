"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Package, Edit, Trash2, Eye, Calendar } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getMyListings, deleteListing, type Listing, type PaginatedResponse } from "@/services/api";
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

export default function MyListingsPage() {
  const { token } = useAuth();
  const [result, setResult] = useState<PaginatedResponse<Listing> | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (token) {
      getMyListings(token, 1, 50)
        .then(setResult)
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [token]);

  async function handleDelete(id: string) {
    if (!token || !confirm("Bạn có chắc muốn xóa tin đăng này?")) return;
    try {
      await deleteListing(token, id);
      setResult((prev) =>
        prev ? { ...prev, data: prev.data.filter((l) => l.id !== id), total: prev.total - 1 } : prev
      );
      toast.success("Đã xóa tin đăng");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa thất bại");
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Tin đăng của tôi</h1>
        <Link href="/tin-dang/tao-moi">
          <Button className="gap-2">
            <Plus className="h-4 w-4" />
            Đăng tin
          </Button>
        </Link>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-28 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <div className="space-y-3">
          {result.data.map((listing) => {
            const st = statusLabels[listing.status] || { label: listing.status, color: "secondary" as const };
            return (
              <Card key={listing.id}>
                <CardContent className="p-4">
                  <div className="flex items-start gap-4">
                    {/* Thumbnail */}
                    <div className="h-20 w-20 rounded-md bg-muted flex items-center justify-center overflow-hidden flex-shrink-0">
                      {listing.images.length > 0 ? (
                        <img src={listing.images[0]} alt={listing.title} loading="lazy" className="h-full w-full object-cover" />
                      ) : (
                        <Package className="h-6 w-6 text-muted-foreground/40" />
                      )}
                    </div>
                    {/* Info */}
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
                      {/* Image strip - show additional images like mobile */}
                      {listing.images.length > 1 && (
                        <div className="flex gap-1.5 mt-2 overflow-x-auto">
                          {listing.images.slice(0, 3).map((img, i) => (
                            <div key={i} className="h-14 w-14 rounded-md overflow-hidden bg-muted flex-shrink-0">
                              <img src={img} alt={listing.title} loading="lazy" className="h-full w-full object-cover" />
                            </div>
                          ))}
                        </div>
                      )}
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
