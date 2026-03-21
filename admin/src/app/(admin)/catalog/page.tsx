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
import { Trash2, Plus, Pencil, GripVertical } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import {
  listCatalogCategories, createCatalogCategory, updateCatalogCategory, deleteCatalogCategory,
  listCatalogProducts, createCatalogProduct, updateCatalogProduct, deleteCatalogProduct,
  type CatalogCategory, type CatalogProduct,
} from "@/services/api";
import {
  DndContext, closestCenter, KeyboardSensor, PointerSensor, useSensor, useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  arrayMove, SortableContext, sortableKeyboardCoordinates, useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";

function SortableCategoryRow({ cat, products, openEditCat, handleDeleteCat }: {
  cat: CatalogCategory;
  products: CatalogProduct[];
  openEditCat: (cat: CatalogCategory) => void;
  handleDeleteCat: (id: string) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: cat.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <TableRow ref={setNodeRef} style={style}>
      <TableCell className="w-8">
        <button {...attributes} {...listeners} className="cursor-grab active:cursor-grabbing p-1 text-muted-foreground hover:text-foreground">
          <GripVertical className="h-4 w-4" />
        </button>
      </TableCell>
      <TableCell className="text-sm font-medium">{cat.label}</TableCell>
      <TableCell className="text-sm">{cat.sort_order}</TableCell>
      <TableCell>
        <Badge variant={cat.is_active ? "default" : "secondary"}>
          {cat.is_active ? "Hoạt động" : "Tắt"}
        </Badge>
      </TableCell>
      <TableCell className="text-sm">{products.filter(p => p.category_id === cat.id).length}</TableCell>
      <TableCell className="text-right">
        <div className="flex justify-end gap-1">
          <Button size="sm" variant="ghost" onClick={() => openEditCat(cat)}><Pencil className="h-4 w-4" /></Button>
          <Button size="sm" variant="ghost" className="text-destructive" onClick={() => handleDeleteCat(cat.id)}><Trash2 className="h-4 w-4" /></Button>
        </div>
      </TableCell>
    </TableRow>
  );
}

function SortableProductRow({ prod, openEditProd, handleDeleteProd }: {
  prod: CatalogProduct;
  openEditProd: (prod: CatalogProduct) => void;
  handleDeleteProd: (id: string) => void;
}) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: prod.id });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <TableRow ref={setNodeRef} style={style}>
      <TableCell className="w-8">
        <button {...attributes} {...listeners} className="cursor-grab active:cursor-grabbing p-1 text-muted-foreground hover:text-foreground">
          <GripVertical className="h-4 w-4" />
        </button>
      </TableCell>
      <TableCell className="text-sm font-medium">{prod.label}</TableCell>
      <TableCell className="text-sm">{prod.sort_order}</TableCell>
      <TableCell>
        <Badge variant={prod.is_active ? "default" : "secondary"}>
          {prod.is_active ? "Hoạt động" : "Tắt"}
        </Badge>
      </TableCell>
      <TableCell className="text-right">
        <div className="flex justify-end gap-1">
          <Button size="sm" variant="ghost" onClick={() => openEditProd(prod)}><Pencil className="h-4 w-4" /></Button>
          <Button size="sm" variant="ghost" className="text-destructive" onClick={() => handleDeleteProd(prod.id)}><Trash2 className="h-4 w-4" /></Button>
        </div>
      </TableCell>
    </TableRow>
  );
}

export default function CatalogPage() {
  const { token } = useAuth();
  const [tab, setTab] = useState<"categories" | "products">("categories");
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [products, setProducts] = useState<CatalogProduct[]>([]);
  const [loading, setLoading] = useState(false);

  // Category form
  const [showCatForm, setShowCatForm] = useState(false);
  const [editingCat, setEditingCat] = useState<CatalogCategory | null>(null);
  const [catKey, setCatKey] = useState("");
  const [catLabel, setCatLabel] = useState("");
  const [catIcon, setCatIcon] = useState("");
  const [catSortOrder, setCatSortOrder] = useState(0);
  const [catIsActive, setCatIsActive] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  // Product form
  const [showProdForm, setShowProdForm] = useState(false);
  const [editingProd, setEditingProd] = useState<CatalogProduct | null>(null);
  const [prodKey, setProdKey] = useState("");
  const [prodLabel, setProdLabel] = useState("");
  const [prodCategoryId, setProdCategoryId] = useState("");
  const [prodSortOrder, setProdSortOrder] = useState(0);
  const [prodIsActive, setProdIsActive] = useState(true);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const fetchData = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    try {
      const [cats, prods] = await Promise.all([
        listCatalogCategories(token),
        listCatalogProducts(token),
      ]);
      setCategories(cats);
      setProducts(prods);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => { fetchData(); }, [fetchData]);

  // Drag-and-drop handlers
  async function handleCatDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id || !token) return;

    const oldIndex = categories.findIndex(c => c.id === active.id);
    const newIndex = categories.findIndex(c => c.id === over.id);
    const reordered = arrayMove(categories, oldIndex, newIndex);

    // Optimistic update
    setCategories(reordered);

    // Persist new sort_order for changed items
    try {
      const updates = reordered.map((cat, i) => ({ ...cat, sort_order: i + 1 }));
      setCategories(updates);
      await Promise.all(
        updates
          .filter((cat, i) => cat.sort_order !== categories[categories.findIndex(c => c.id === cat.id)]?.sort_order || oldIndex === i || newIndex === i)
          .map(cat => updateCatalogCategory(token, cat.id, { sort_order: cat.sort_order }))
      );
      toast.success("Đã cập nhật thứ tự danh mục");
    } catch {
      toast.error("Cập nhật thứ tự thất bại");
      fetchData();
    }
  }

  async function handleGroupedProdDragEnd(event: DragEndEvent, categoryId: string) {
    const { active, over } = event;
    if (!over || active.id === over.id || !token) return;

    const catProds = products.filter(p => p.category_id === categoryId).sort((a, b) => a.sort_order - b.sort_order);
    const oldIndex = catProds.findIndex(p => p.id === active.id);
    const newIndex = catProds.findIndex(p => p.id === over.id);
    if (oldIndex === -1 || newIndex === -1) return;

    const reordered = arrayMove(catProds, oldIndex, newIndex);
    const updatedProds = reordered.map((prod, i) => ({ ...prod, sort_order: i + 1 }));

    // Optimistic update — replace products in this category
    setProducts(prev => {
      const others = prev.filter(p => p.category_id !== categoryId);
      return [...others, ...updatedProds].sort((a, b) => a.sort_order - b.sort_order);
    });

    try {
      await Promise.all(
        updatedProds
          .filter((prod, i) => prod.sort_order !== catProds[i]?.sort_order)
          .map(prod => updateCatalogProduct(token, prod.id, { sort_order: prod.sort_order }))
      );
      toast.success("Đã cập nhật thứ tự sản phẩm");
    } catch {
      toast.error("Cập nhật thứ tự thất bại");
      fetchData();
    }
  }

  // Category CRUD
  function openCreateCat() {
    setEditingCat(null); setCatKey(""); setCatLabel(""); setCatIcon(""); setCatSortOrder(0); setCatIsActive(true);
    setShowCatForm(true);
  }
  function openEditCat(cat: CatalogCategory) {
    setEditingCat(cat); setCatKey(cat.key); setCatLabel(cat.label); setCatIcon(cat.icon); setCatSortOrder(cat.sort_order); setCatIsActive(cat.is_active);
    setShowCatForm(true);
  }
  async function handleCatSubmit() {
    if (!token || !catKey || !catLabel) return;
    setSubmitting(true);
    try {
      if (editingCat) {
        await updateCatalogCategory(token, editingCat.id, { label: catLabel, icon: catIcon, sort_order: catSortOrder, is_active: catIsActive });
      } else {
        await createCatalogCategory(token, { key: catKey, label: catLabel, icon: catIcon || undefined });
      }
      toast.success(editingCat ? "Đã cập nhật danh mục" : "Đã thêm danh mục");
      setShowCatForm(false); fetchData();
    } catch { toast.error("Lưu danh mục thất bại"); }
    finally { setSubmitting(false); }
  }
  async function handleDeleteCat(id: string) {
    if (!token || !confirm("Xóa danh mục này? Tất cả sản phẩm trong danh mục sẽ bị xóa theo.")) return;
    try { await deleteCatalogCategory(token, id); toast.success("Đã xóa danh mục"); fetchData(); }
    catch { toast.error("Xóa danh mục thất bại"); }
  }

  // Product CRUD
  function openCreateProd() {
    setEditingProd(null); setProdKey(""); setProdLabel(""); setProdCategoryId(categories[0]?.id || ""); setProdSortOrder(0); setProdIsActive(true);
    setShowProdForm(true);
  }
  function openEditProd(prod: CatalogProduct) {
    setEditingProd(prod); setProdKey(prod.key); setProdLabel(prod.label); setProdCategoryId(prod.category_id); setProdSortOrder(prod.sort_order); setProdIsActive(prod.is_active);
    setShowProdForm(true);
  }
  async function handleProdSubmit() {
    if (!token || !prodKey || !prodLabel || !prodCategoryId) return;
    setSubmitting(true);
    try {
      if (editingProd) {
        await updateCatalogProduct(token, editingProd.id, { label: prodLabel, category_id: prodCategoryId, sort_order: prodSortOrder, is_active: prodIsActive });
      } else {
        await createCatalogProduct(token, { key: prodKey, label: prodLabel, category_id: prodCategoryId });
      }
      toast.success(editingProd ? "Đã cập nhật sản phẩm" : "Đã thêm sản phẩm");
      setShowProdForm(false); fetchData();
    } catch { toast.error("Lưu sản phẩm thất bại"); }
    finally { setSubmitting(false); }
  }
  async function handleDeleteProd(id: string) {
    if (!token || !confirm("Xóa sản phẩm này?")) return;
    try { await deleteCatalogProduct(token, id); toast.success("Đã xóa sản phẩm"); fetchData(); }
    catch { toast.error("Xóa sản phẩm thất bại"); }
  }

  // Group products by category
  const productsByCategory = categories.map(cat => ({
    category: cat,
    products: products.filter(p => p.category_id === cat.id).sort((a, b) => a.sort_order - b.sort_order),
  })).filter(g => g.products.length > 0);

  // Products without a valid category
  const orphanProducts = products.filter(p => !categories.find(c => c.id === p.category_id));

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h1 className="text-xl font-semibold">Quản lý danh mục sản phẩm</h1>
      </div>

      {/* Tab buttons */}
      <div className="flex gap-2 mb-4">
        <Button variant={tab === "categories" ? "default" : "outline"} size="sm" onClick={() => setTab("categories")}>
          Danh mục ({categories.length})
        </Button>
        <Button variant={tab === "products" ? "default" : "outline"} size="sm" onClick={() => setTab("products")}>
          Sản phẩm ({products.length})
        </Button>
      </div>

      {tab === "categories" ? (
        <>
          <div className="flex justify-end mb-3">
            <Button size="sm" onClick={openCreateCat}><Plus className="h-4 w-4 mr-1" /> Thêm danh mục</Button>
          </div>
          <div className="rounded-lg border shadow-sm bg-card overflow-hidden">
            <div className="px-4 py-3 bg-gradient-to-r from-indigo-500 to-purple-500 text-white">
              <h3 className="text-sm font-bold">Danh sách danh mục</h3>
            </div>
            <Table>
              <TableHeader>
                <TableRow className="bg-indigo-50/60 dark:bg-indigo-950/20">
                  <TableHead className="w-8"></TableHead>
                  <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Tên danh mục</TableHead>
                  <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Thứ tự</TableHead>
                  <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Trạng thái</TableHead>
                  <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Số SP</TableHead>
                  <TableHead className="text-right font-semibold text-indigo-700 dark:text-indigo-300">Thao tác</TableHead>
                </TableRow>
              </TableHeader>
              <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleCatDragEnd} modifiers={[restrictToVerticalAxis]}>
                <SortableContext items={categories.map(c => c.id)} strategy={verticalListSortingStrategy}>
                  <TableBody>
                    {loading ? (
                      <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Đang tải...</TableCell></TableRow>
                    ) : categories.length === 0 ? (
                      <TableRow><TableCell colSpan={8} className="text-center py-8 text-muted-foreground">Chưa có danh mục nào</TableCell></TableRow>
                    ) : categories.map(cat => (
                      <SortableCategoryRow key={cat.id} cat={cat} products={products} openEditCat={openEditCat} handleDeleteCat={handleDeleteCat} />
                    ))}
                  </TableBody>
                </SortableContext>
              </DndContext>
            </Table>
          </div>
        </>
      ) : (
        <>
          <div className="flex justify-end mb-3">
            <Button size="sm" onClick={openCreateProd}><Plus className="h-4 w-4 mr-1" /> Thêm sản phẩm</Button>
          </div>
          {loading ? (
            <div className="text-center py-8 text-muted-foreground">Đang tải...</div>
          ) : products.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">Chưa có sản phẩm nào</div>
          ) : (
            <div className="space-y-4">
              {productsByCategory.map(({ category: cat, products: catProds }) => (
                <div key={cat.id} className="rounded-lg border shadow-sm bg-card overflow-hidden">
                  <div className="flex items-center gap-2 px-4 py-3 bg-gradient-to-r from-indigo-500 to-purple-500 text-white">
                    <h3 className="text-sm font-bold">{cat.label}</h3>
                    <span className="ml-1 inline-flex items-center rounded-full bg-white/20 px-2 py-0.5 text-[10px] font-medium">{catProds.length} sản phẩm</span>
                    {!cat.is_active && <span className="inline-flex items-center rounded-full bg-orange-400/30 px-2 py-0.5 text-[10px] font-medium">Danh mục tắt</span>}
                  </div>
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-indigo-50/60 dark:bg-indigo-950/20">
                        <TableHead className="w-8"></TableHead>
                        <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Tên sản phẩm</TableHead>
                        <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Thứ tự</TableHead>
                        <TableHead className="font-semibold text-indigo-700 dark:text-indigo-300">Trạng thái</TableHead>
                        <TableHead className="text-right font-semibold text-indigo-700 dark:text-indigo-300">Thao tác</TableHead>
                      </TableRow>
                    </TableHeader>
                    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={(event) => handleGroupedProdDragEnd(event, cat.id)} modifiers={[restrictToVerticalAxis]}>
                      <SortableContext items={catProds.map(p => p.id)} strategy={verticalListSortingStrategy}>
                        <TableBody>
                          {catProds.map(prod => (
                            <SortableProductRow key={prod.id} prod={prod} openEditProd={openEditProd} handleDeleteProd={handleDeleteProd} />
                          ))}
                        </TableBody>
                      </SortableContext>
                    </DndContext>
                  </Table>
                </div>
              ))}
              {orphanProducts.length > 0 && (
                <div className="rounded-lg border shadow-sm bg-card overflow-hidden">
                  <div className="flex items-center gap-2 px-4 py-3 bg-gradient-to-r from-gray-400 to-gray-500 text-white">
                    <h3 className="text-sm font-bold">Chưa phân loại</h3>
                    <span className="ml-1 inline-flex items-center rounded-full bg-white/20 px-2 py-0.5 text-[10px] font-medium">{orphanProducts.length} sản phẩm</span>
                  </div>
                  <Table>
                    <TableHeader>
                      <TableRow className="bg-gray-50/60 dark:bg-gray-950/20">
                        <TableHead className="w-8"></TableHead>
                        <TableHead className="font-semibold">Tên sản phẩm</TableHead>
                        <TableHead className="font-semibold">Thứ tự</TableHead>
                        <TableHead className="font-semibold">Trạng thái</TableHead>
                        <TableHead className="text-right font-semibold">Thao tác</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {orphanProducts.map(prod => (
                        <SortableProductRow key={prod.id} prod={prod} openEditProd={openEditProd} handleDeleteProd={handleDeleteProd} />
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          )}
        </>
      )}

      {/* Category Form Dialog */}
      <Dialog open={showCatForm} onOpenChange={() => setShowCatForm(false)}>
        <DialogContent>
          <DialogHeader><DialogTitle>{editingCat ? "Sửa danh mục" : "Thêm danh mục"}</DialogTitle></DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Mã danh mục *</label>
              <input className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={catKey} onChange={e => setCatKey(e.target.value)} disabled={!!editingCat} placeholder="vd: gao_deo_thom" />
            </div>
            <div>
              <label className="text-sm font-medium">Tên danh mục *</label>
              <input className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={catLabel} onChange={e => setCatLabel(e.target.value)} placeholder="vd: Gạo dẻo thơm" />
            </div>
            <div>
              <label className="text-sm font-medium">Icon</label>
              <input className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={catIcon} onChange={e => setCatIcon(e.target.value)} placeholder="vd: rice_bowl" />
            </div>
            {editingCat && (
              <>
                <div>
                  <label className="text-sm font-medium">Thứ tự</label>
                  <input type="number" className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={catSortOrder} onChange={e => setCatSortOrder(Number(e.target.value))} />
                </div>
                <div className="flex items-center gap-2">
                  <input type="checkbox" id="cat_active" checked={catIsActive} onChange={e => setCatIsActive(e.target.checked)} />
                  <label htmlFor="cat_active" className="text-sm">Hoạt động</label>
                </div>
              </>
            )}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowCatForm(false)}>Hủy</Button>
            <Button onClick={handleCatSubmit} disabled={submitting || !catKey || !catLabel}>
              {submitting ? "Đang lưu..." : editingCat ? "Cập nhật" : "Thêm mới"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Product Form Dialog */}
      <Dialog open={showProdForm} onOpenChange={() => setShowProdForm(false)}>
        <DialogContent>
          <DialogHeader><DialogTitle>{editingProd ? "Sửa sản phẩm" : "Thêm sản phẩm"}</DialogTitle></DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Mã sản phẩm *</label>
              <input className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={prodKey} onChange={e => setProdKey(e.target.value)} disabled={!!editingProd} placeholder="vd: st_25" />
            </div>
            <div>
              <label className="text-sm font-medium">Tên sản phẩm *</label>
              <input className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={prodLabel} onChange={e => setProdLabel(e.target.value)} placeholder="vd: ST 25" />
            </div>
            <div>
              <label className="text-sm font-medium">Danh mục *</label>
              <select className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={prodCategoryId} onChange={e => setProdCategoryId(e.target.value)}>
                <option value="">Chọn danh mục...</option>
                {categories.map(cat => (
                  <option key={cat.id} value={cat.id}>{cat.label}</option>
                ))}
              </select>
            </div>
            {editingProd && (
              <>
                <div>
                  <label className="text-sm font-medium">Thứ tự</label>
                  <input type="number" className="w-full mt-1 rounded-md border border-input bg-background px-3 py-2 text-sm" value={prodSortOrder} onChange={e => setProdSortOrder(Number(e.target.value))} />
                </div>
                <div className="flex items-center gap-2">
                  <input type="checkbox" id="prod_active" checked={prodIsActive} onChange={e => setProdIsActive(e.target.checked)} />
                  <label htmlFor="prod_active" className="text-sm">Hoạt động</label>
                </div>
              </>
            )}
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setShowProdForm(false)}>Hủy</Button>
            <Button onClick={handleProdSubmit} disabled={submitting || !prodKey || !prodLabel || !prodCategoryId}>
              {submitting ? "Đang lưu..." : editingProd ? "Cập nhật" : "Thêm mới"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
