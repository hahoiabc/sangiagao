"use client";

import { useEffect, useState } from "react";
import { User, Save } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getMe, updateMe, type User as UserType } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";
import { formatDate } from "@/lib/utils";

export default function ProfilePage() {
  const { token } = useAuth();
  const [profile, setProfile] = useState<UserType | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ name: "", province: "", district: "", ward: "", address: "", description: "", org_name: "" });

  useEffect(() => {
    if (token) {
      getMe(token)
        .then((u) => {
          setProfile(u);
          setForm({
            name: u.name || "",
            province: u.province || "",
            district: u.district || "",
            ward: u.ward || "",
            address: u.address || "",
            description: u.description || "",
            org_name: u.org_name || "",
          });
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [token]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;
    setSaving(true);
    try {
      const updated = await updateMe(token, form);
      setProfile(updated);
      toast.success("Cập nhật thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-80 w-full rounded-lg" />
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Tài Khoản</h1>

      {profile && (
        <div className="space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <User className="h-4 w-4" />
                Thông tin tài khoản
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Số điện thoại</span>
                <span className="font-medium">{profile.phone}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Vai trò</span>
                <Badge variant="secondary">{profile.role}</Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Ngày tham gia</span>
                <span>{formatDate(profile.created_at)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Gói thành viên</span>
                {profile.subscription_expires_at ? (
                  <Badge>Hết hạn: {formatDate(profile.subscription_expires_at)}</Badge>
                ) : (
                  <Badge variant="secondary">Chưa kích hoạt</Badge>
                )}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Chỉnh sửa thông tin
              </CardTitle>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSave} className="space-y-4">
                <div>
                  <label className="text-sm font-medium mb-1 block">Tên hiển thị</label>
                  <Input value={form.name} onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))} />
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Tên công ty/đơn vị</label>
                  <Input value={form.org_name} onChange={(e) => setForm((f) => ({ ...f, org_name: e.target.value }))} />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium mb-1 block">Tỉnh/Thành</label>
                    <Input value={form.province} onChange={(e) => setForm((f) => ({ ...f, province: e.target.value }))} />
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-1 block">Quận/Huyện</label>
                    <Input value={form.district} onChange={(e) => setForm((f) => ({ ...f, district: e.target.value }))} />
                  </div>
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Địa chỉ</label>
                  <Input value={form.address} onChange={(e) => setForm((f) => ({ ...f, address: e.target.value }))} />
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Giới thiệu</label>
                  <textarea
                    value={form.description}
                    onChange={(e) => setForm((f) => ({ ...f, description: e.target.value }))}
                    className="w-full min-h-20 rounded-md border border-input bg-background px-3 py-2 text-sm"
                  />
                </div>
                <Button type="submit" className="gap-2" disabled={saving}>
                  <Save className="h-4 w-4" />
                  {saving ? "Đang lưu..." : "Lưu thay đổi"}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
