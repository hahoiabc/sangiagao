"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import Image from "next/image";
import { Button } from "@/components/ui/button";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  browseListings, deleteListing, listCatalogCategories,
  type Listing, type CatalogCategory,
} from "@/services/api";
import { cn } from "@/lib/utils";

export default function ListingsPage() {
  const { user } = useAuth();
  const router = useRouter();
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>("gao_deo_thom");
  const [listings, setListings] = useState<Listing[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [deleteDialog, setDeleteDialog] = useState<Listing | null>(null);

  const limit = 20;

  // Fetch categories once
  useEffect(() => {
    if (!user) return;
    listCatalogCategories("").then(cats => {
      const sorted = cats.sort((a, b) => a.sort_order - b.sort_order);
      setCategories(sorted);
      // Nếu category mặc định không tồn tại, chọn category đầu tiên
      const active = sorted.filter(c => c.is_active);
      if (active.length > 0 && !active.find(c => c.key === selectedCategory)) {
        setSelectedCategory(active[0].key);
      }
    }).catch(console.error);
  }, [user]);

  const fetchListings = useCallback(async () => {
    if (!user) return;
    setLoading(true);
    try {
      const res = await browseListings("", page, limit, selectedCategory);
      setListings(res.data);
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [user, page, selectedCategory]);

  useEffect(() => {
    fetchListings();
  }, [fetchListings]);

  // Reset page when category changes
  function handleCategoryChange(catKey: string) {
    setSelectedCategory(catKey);
    setPage(1);
  }

  async function handleDelete() {
    if (!user || !deleteDialog) return;
    try {
      await deleteListing("", deleteDialog.id);
      toast.success("Đã xóa tin đăng");
      setDeleteDialog(null);
      fetchListings();
    } catch {
      toast.error("Xóa tin đăng thất bại");
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

  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      <h1 className="text-xl font-semibold mb-5">Quản lý tin đăng</h1>

      {/* Category tabs */}
      <div className="flex gap-1.5 mb-4 flex-wrap">
        {categories.filter(c => c.is_active).map(cat => (
          <button
            key={cat.key}
            onClick={() => handleCategoryChange(cat.key)}
            className={cn(
              "inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors border",
              selectedCategory === cat.key
                ? "bg-primary text-primary-foreground border-primary"
                : "bg-card text-muted-foreground border-border hover:bg-muted hover:text-foreground"
            )}
          >
            {cat.label}
          </button>
        ))}
      </div>

      <div className="rounded-lg border shadow-sm bg-card overflow-x-auto">
        <Table className="min-w-[900px]">
          <TableHeader>
            <TableRow>
              <TableHead>Hình ảnh</TableHead>
              <TableHead>Tiêu đề</TableHead>
              <TableHead>Tỉnh/TP</TableHead>
              <TableHead>Số lượng</TableHead>
              <TableHead>Giá</TableHead>
              <TableHead>Lượt xem</TableHead>
              <TableHead>Trạng thái</TableHead>
              <TableHead>Ngày tạo</TableHead>
              <TableHead className="text-right">Thao tác</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={10} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
              </TableRow>
            ) : listings.length === 0 ? (
              <TableRow>
                <TableCell colSpan={10} className="text-center py-8 text-muted-foreground">Không tìm thấy tin đăng</TableCell>
              </TableRow>
            ) : (
              listings.map((listing) => (
                <TableRow
                  key={listing.id}
                  className="cursor-pointer hover:bg-muted/50"
                  onClick={() => router.push(`/listings/${listing.id}`)}
                >
                  <TableCell>
                    {listing.images && listing.images.length > 0 ? (
                      <div className="flex gap-1">
                        {listing.images.slice(0, 3).map((img, i) => (
                          <div key={i} className="relative h-8 w-8">
                            <Image
                              src={img}
                              alt={`${listing.title} - ${i + 1}`}
                              fill
                              sizes="32px"
                              className="rounded object-cover border"
                            />
                          </div>
                        ))}
                      </div>
                    ) : (
                      <span className="text-xs text-muted-foreground">Chưa có ảnh</span>
                    )}
                  </TableCell>
                  <TableCell className="font-medium max-w-[200px] truncate">{listing.title}</TableCell>
                  <TableCell>{listing.province || "-"}</TableCell>
                  <TableCell className="text-sm">{formatQty(listing.quantity_kg)}</TableCell>
                  <TableCell className="text-sm">{formatPrice(listing.price_per_kg)}</TableCell>
                  <TableCell>{listing.view_count}</TableCell>
                  <TableCell>
                    <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium border ${listing.status === "active" ? "bg-emerald-50 text-emerald-700 border-emerald-200" : listing.status === "sold" ? "bg-blue-50 text-blue-700 border-blue-200" : "bg-gray-50 text-gray-600 border-gray-200"}`}>
                      {statusLabel(listing.status)}
                    </span>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(listing.created_at).toLocaleDateString("vi-VN")}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="destructive"
                      onClick={(e) => { e.stopPropagation(); setDeleteDialog(listing); }}
                    >
                      Xóa
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
        <span className="text-sm text-muted-foreground">
          Trang {page} / {totalPages || 1} ({total} tin đăng)
        </span>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage(page - 1)}>
            Trước
          </Button>
          <Button size="sm" variant="outline" disabled={page >= totalPages} onClick={() => setPage(page + 1)}>
            Sau
          </Button>
        </div>
      </div>

      {/* Delete Dialog */}
      <Dialog open={!!deleteDialog} onOpenChange={() => setDeleteDialog(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Xóa tin đăng</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Bạn có chắc muốn xóa &quot;{deleteDialog?.title}&quot;? Hành động này không thể hoàn tác.
          </p>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setDeleteDialog(null)}>Hủy</Button>
            <Button variant="destructive" onClick={handleDelete}>Xóa</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
