"use client";

import { useEffect, useState, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Trash2, Plus, Pencil } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  listSponsors, createSponsor, updateSponsor, deleteSponsor,
  getProductCatalog,
  type ProductSponsor, type RiceCategory,
} from "@/services/api";

export default function SponsorsPage() {
  const { token } = useAuth();
  const [sponsors, setSponsors] = useState<ProductSponsor[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [catalog, setCatalog] = useState<RiceCategory[]>([]);

  // Form state
  const [showForm, setShowForm] = useState(false);
  const [editingSponsor, setEditingSponsor] = useState<ProductSponsor | null>(null);
  const [formProductKey, setFormProductKey] = useState("");
  const [formLogoUrl, setFormLogoUrl] = useState("");
  const [formSponsorName, setFormSponsorName] = useState("");
  const [formIsActive, setFormIsActive] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const limit = 20;

  const fetchSponsors = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const res = await listSponsors(token, page, limit);
      setSponsors(res.data);
      setTotal(res.total);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token, page]);

  const fetchCatalog = useCallback(async () => {
    if (!token) return;
    try {
      const data = await getProductCatalog(token);
      setCatalog(data);
    } catch (err) {
      console.error(err);
    }
  }, [token]);

  useEffect(() => {
    fetchSponsors();
    fetchCatalog();
  }, [fetchSponsors, fetchCatalog]);

  function openCreateForm() {
    setEditingSponsor(null);
    setFormProductKey("");
    setFormLogoUrl("");
    setFormSponsorName("");
    setFormIsActive(true);
    setShowForm(true);
  }

  function openEditForm(sponsor: ProductSponsor) {
    setEditingSponsor(sponsor);
    setFormProductKey(sponsor.product_key);
    setFormLogoUrl(sponsor.logo_url);
    setFormSponsorName(sponsor.sponsor_name);
    setFormIsActive(sponsor.is_active);
    setShowForm(true);
  }

  async function handleSubmit() {
    if (!token || !formProductKey || !formLogoUrl || !formSponsorName) return;
    setSubmitting(true);
    try {
      if (editingSponsor) {
        await updateSponsor(token, editingSponsor.id, {
          logo_url: formLogoUrl,
          sponsor_name: formSponsorName,
          is_active: formIsActive,
        });
      } else {
        await createSponsor(token, {
          product_key: formProductKey,
          logo_url: formLogoUrl,
          sponsor_name: formSponsorName,
        });
      }
      toast.success(editingSponsor ? "Đã cập nhật tài trợ" : "Đã thêm tài trợ");
      setShowForm(false);
      fetchSponsors();
    } catch {
      toast.error("Lưu tài trợ thất bại");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(id: string) {
    if (!token || !confirm("Xóa tài trợ này?")) return;
    try {
      await deleteSponsor(token, id);
      toast.success("Đã xóa tài trợ");
      fetchSponsors();
    } catch {
      toast.error("Xóa tài trợ thất bại");
    }
  }

  // Build a flat list of all products for dropdown
  const allProducts = catalog.flatMap((cat) =>
    cat.products.map((p) => ({ key: p.key, label: `${cat.label} — ${p.label}` }))
  );

  function productLabel(key: string) {
    return allProducts.find((p) => p.key === key)?.label || key;
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold">Quản lý tài trợ</h1>
        <Button size="sm" onClick={openCreateForm}>
          <Plus className="h-4 w-4 mr-1" /> Thêm tài trợ
        </Button>
      </div>

      <div className="rounded-lg border shadow-sm bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Sản phẩm</TableHead>
              <TableHead>Nhà tài trợ</TableHead>
              <TableHead>Logo</TableHead>
              <TableHead>Trạng thái</TableHead>
              <TableHead>Ngày tạo</TableHead>
              <TableHead className="text-right">Thao tác</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell>
              </TableRow>
            ) : sponsors.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-8 text-muted-foreground">Chưa có tài trợ nào</TableCell>
              </TableRow>
            ) : (
              sponsors.map((sponsor) => (
                <TableRow key={sponsor.id}>
                  <TableCell className="text-sm">{productLabel(sponsor.product_key)}</TableCell>
                  <TableCell className="text-sm font-medium">{sponsor.sponsor_name}</TableCell>
                  <TableCell>
                    {sponsor.logo_url ? (
                      <img src={sponsor.logo_url} alt={sponsor.sponsor_name} className="h-8 w-8 object-contain rounded" />
                    ) : "-"}
                  </TableCell>
                  <TableCell>
                    <Badge variant={sponsor.is_active ? "default" : "secondary"}>
                      {sponsor.is_active ? "Hoạt động" : "Tắt"}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {new Date(sponsor.created_at).toLocaleDateString("vi-VN")}
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex justify-end gap-1">
                      <Button size="sm" variant="ghost" onClick={() => openEditForm(sponsor)}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button size="sm" variant="ghost" className="text-destructive" onClick={() => handleDelete(sponsor.id)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-between bg-muted/30 rounded-b-lg px-4 py-3 border border-t-0 shadow-sm">
          <span className="text-sm text-muted-foreground">
            Trang {page} / {totalPages} ({total} tài trợ)
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
      )}

      <Dialog open={showForm} onOpenChange={() => setShowForm(false)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingSponsor ? "Sửa tài trợ" : "Thêm tài trợ"}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Sản phẩm *</label>
              <select
                className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formProductKey}
                onChange={(e) => setFormProductKey(e.target.value)}
                disabled={!!editingSponsor}
              >
                <option value="">Chọn sản phẩm...</option>
                {catalog.map((cat) => (
                  <optgroup key={cat.key} label={cat.label}>
                    {cat.products.map((p) => (
                      <option key={p.key} value={p.key}>{p.label}</option>
                    ))}
                  </optgroup>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm font-medium">Tên nhà tài trợ *</label>
              <input
                className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formSponsorName}
                onChange={(e) => setFormSponsorName(e.target.value)}
                placeholder="Ví dụ: Công ty ABC"
              />
            </div>
            <div>
              <label className="text-sm font-medium">URL logo *</label>
              <input
                className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={formLogoUrl}
                onChange={(e) => setFormLogoUrl(e.target.value)}
                placeholder="https://..."
              />
            </div>
            {editingSponsor && (
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="is_active"
                  checked={formIsActive}
                  onChange={(e) => setFormIsActive(e.target.checked)}
                />
                <label htmlFor="is_active" className="text-sm">Hoạt động</label>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowForm(false)}>Hủy</Button>
            <Button onClick={handleSubmit} disabled={submitting || !formProductKey || !formLogoUrl || !formSponsorName}>
              {submitting ? "Đang lưu..." : editingSponsor ? "Cập nhật" : "Thêm mới"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
