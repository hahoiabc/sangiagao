"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { ArrowLeft } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { getListingDetail, deleteListing, type ListingDetail } from "@/services/api";

export default function ListingDetailPage() {
  const { user } = useAuth();
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [listing, setListing] = useState<ListingDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteDialog, setDeleteDialog] = useState(false);

  const fetchDetail = useCallback(async () => {
    if (!user || !id) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getListingDetail("", id);
      setListing(data);
    } catch {
      setError("Không tìm thấy tin đăng.");
    } finally {
      setLoading(false);
    }
  }, [user, id]);

  useEffect(() => {
    fetchDetail();
  }, [fetchDetail]);

  async function handleDelete() {
    if (!user || !listing) return;
    try {
      await deleteListing("", listing.id);
      router.push("/listings");
    } catch (err) {
      console.error(err);
    }
  }

  function formatPrice(price: number) {
    return new Intl.NumberFormat("vi-VN").format(price) + "đ/kg";
  }

  function formatQty(qty: number) {
    return new Intl.NumberFormat("vi-VN").format(qty) + " kg";
  }

  function statusLabel(status: string) {
    switch (status) {
      case "active": return "Đang hiển thị";
      case "sold": return "Đã bán";
      case "expired": return "Hết hạn";
      default: return status;
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-muted-foreground">
        Đang tải...
      </div>
    );
  }

  if (error || !listing) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" size="sm" onClick={() => router.push("/listings")}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Quay lại
        </Button>
        <p className="text-center py-20 text-muted-foreground">{error || "Không tìm thấy tin đăng."}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" onClick={() => router.push("/listings")}>
            <ArrowLeft className="h-4 w-4 mr-2" />
            Quay lại
          </Button>
          <h1 className="text-xl font-semibold">Chi tiết tin đăng</h1>
        </div>
        <Button variant="destructive" size="sm" onClick={() => setDeleteDialog(true)}>
          Xóa tin đăng
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Images */}
          {listing.images && listing.images.length > 0 && (
            <div className="rounded-lg border shadow-sm bg-card p-4">
              <h2 className="text-sm font-medium text-muted-foreground mb-3">Hình ảnh ({listing.images.length})</h2>
              <div className="flex gap-3 flex-wrap">
                {listing.images.map((img, i) => (
                  <img
                    key={i}
                    src={img}
                    alt={`${listing.title} - ${i + 1}`}
                    className="h-40 w-40 rounded-lg object-cover border"
                  />
                ))}
              </div>
            </div>
          )}

          {/* Info */}
          <div className="rounded-lg border shadow-sm bg-card p-4 space-y-4">
            <div className="flex items-start justify-between">
              <div>
                <h2 className="text-xl font-semibold">{listing.title}</h2>
                <p className="text-sm text-muted-foreground mt-1">
                  Đăng ngày {new Date(listing.created_at).toLocaleString("vi-VN")}
                </p>
              </div>
              <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border ${listing.status === "active" ? "bg-emerald-50 text-emerald-700 border-emerald-200" : listing.status === "sold" ? "bg-blue-50 text-blue-700 border-blue-200" : "bg-gray-50 text-gray-600 border-gray-200"}`}>
                {statusLabel(listing.status)}
              </span>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 text-sm">
              <div className="rounded-lg bg-muted/50 p-3">
                <span className="text-muted-foreground block">Loại gạo</span>
                <span className="font-medium">{listing.rice_type || "-"}</span>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <span className="text-muted-foreground block">Số lượng</span>
                <span className="font-medium">{formatQty(listing.quantity_kg)}</span>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <span className="text-muted-foreground block">Giá</span>
                <span className="font-medium">{formatPrice(listing.price_per_kg)}</span>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <span className="text-muted-foreground block">Tỉnh/TP</span>
                <span className="font-medium">{listing.province || "-"}</span>
              </div>
              <div className="rounded-lg bg-muted/50 p-3">
                <span className="text-muted-foreground block">Lượt xem</span>
                <span className="font-medium">{listing.view_count}</span>
              </div>
              {listing.certifications && (
                <div className="rounded-lg bg-muted/50 p-3">
                  <span className="text-muted-foreground block">Chứng nhận</span>
                  <span className="font-medium">{listing.certifications}</span>
                </div>
              )}
            </div>

            {/* Description */}
            {listing.description && (
              <div>
                <h3 className="text-sm font-medium text-muted-foreground mb-2">Mô tả</h3>
                <p className="text-sm whitespace-pre-wrap bg-muted/50 rounded-md p-3">
                  {listing.description}
                </p>
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-4">
          {/* Seller info */}
          {listing.seller && (
            <div className="rounded-lg border shadow-sm bg-card p-4 space-y-3">
              <h3 className="text-sm font-medium text-muted-foreground">Người đăng</h3>
              <div className="flex items-center gap-3">
                <Avatar className="h-10 w-10">
                  {listing.seller.avatar_url ? (
                    <AvatarImage src={listing.seller.avatar_url} alt={listing.seller.name || listing.seller.phone} />
                  ) : null}
                  <AvatarFallback>
                    {(listing.seller.name || listing.seller.phone || "?").charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium text-sm">{listing.seller.name || "-"}</p>
                  <p className="text-xs text-muted-foreground font-mono">{listing.seller.phone}</p>
                </div>
              </div>
              {listing.seller.org_name && (
                <div className="text-sm">
                  <span className="text-muted-foreground">Tổ chức: </span>
                  <span>{listing.seller.org_name}</span>
                </div>
              )}
              {listing.seller.province && (
                <div className="text-sm">
                  <span className="text-muted-foreground">Khu vực: </span>
                  <span>{listing.seller.province}</span>
                </div>
              )}
              <Button
                variant="outline"
                size="sm"
                className="w-full"
                onClick={() => router.push(`/users/${listing.user_id}`)}
              >
                Xem trong quản lý người dùng
              </Button>
            </div>
          )}

          {/* Metadata */}
          <div className="rounded-lg border shadow-sm bg-card p-4 space-y-2 text-xs text-muted-foreground">
            <h3 className="text-sm font-medium">Thông tin hệ thống</h3>
            <div className="space-y-1">
              <p>ID: <span className="font-mono">{listing.id}</span></p>
              <p>User ID: <span className="font-mono">{listing.user_id}</span></p>
              <p>Tạo lúc: {new Date(listing.created_at).toLocaleString("vi-VN")}</p>
              {listing.updated_at && (
                <p>Cập nhật: {new Date(listing.updated_at).toLocaleString("vi-VN")}</p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Delete dialog */}
      <Dialog open={deleteDialog} onOpenChange={setDeleteDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Xóa tin đăng</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Bạn có chắc muốn xóa &quot;{listing.title}&quot;? Hành động này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteDialog(false)}>Hủy</Button>
            <Button variant="destructive" onClick={handleDelete}>Xóa</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
