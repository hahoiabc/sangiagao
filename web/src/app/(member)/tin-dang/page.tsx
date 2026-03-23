"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Plus, Package, Edit, Trash2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getMyListings, deleteListing, type Listing, type PaginatedResponse } from "@/services/api";
import { formatPrice, formatQuantity, timeAgo } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";

const statusLabels: Record<string, string> = {
  active: "Đang hiển thị",
  hidden_subscription: "Ẩn (hết hạn)",
  deleted: "Đã xóa",
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
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-24 w-full rounded-lg" />)}
        </div>
      ) : result && result.data.length > 0 ? (
        <div className="space-y-3">
          {result.data.map((listing) => (
            <Card key={listing.id}>
              <CardContent className="p-4 flex items-center gap-4">
                <div className="h-16 w-16 rounded-md bg-muted flex items-center justify-center overflow-hidden flex-shrink-0">
                  {listing.images.length > 0 ? (
                    <img src={listing.images[0]} alt="" className="h-full w-full object-cover" />
                  ) : (
                    <Package className="h-6 w-6 text-muted-foreground/40" />
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-sm truncate">{listing.title}</h3>
                  <p className="text-sm text-primary font-medium">
                    {formatPrice(listing.price_per_kg)} &middot; {formatQuantity(listing.quantity_kg)}
                  </p>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge
                      variant={listing.status === "active" ? "default" : "secondary"}
                      className="text-xs"
                    >
                      {statusLabels[listing.status] || listing.status}
                    </Badge>
                    <span className="text-xs text-muted-foreground">{timeAgo(listing.created_at)}</span>
                  </div>
                </div>
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
              </CardContent>
            </Card>
          ))}
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
