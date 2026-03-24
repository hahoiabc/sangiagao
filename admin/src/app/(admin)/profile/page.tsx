"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Camera } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/lib/auth";
import { getMe, updateMe, uploadImage, updateMyAvatar, type User } from "@/services/api";

export default function ProfilePage() {
  const { user: authUser, login } = useAuth();
  const [profile, setProfile] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const [name, setName] = useState("");
  const [address, setAddress] = useState("");
  const [province, setProvince] = useState("");
  const [ward, setWard] = useState("");
  const [description, setDescription] = useState("");
  const [orgName, setOrgName] = useState("");

  const [saving, setSaving] = useState(false);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setLoading(true);
    getMe()
      .then((data) => {
        setProfile(data);
        syncForm(data);
      })
      .catch(() => setError("Không thể tải thông tin tài khoản"))
      .finally(() => setLoading(false));
  }, []);

  function syncForm(data: User) {
    setName(data.name || "");
    setAddress(data.address || "");
    setProvince(data.province || "");
    setWard(data.ward || "");
    setDescription(data.description || "");
    setOrgName(data.org_name || "");
  }

  async function handleAvatarChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith("image/")) {
      setError("Vui lòng chọn file ảnh");
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      setError("Ảnh không được vượt quá 5MB");
      return;
    }

    setUploadingAvatar(true);
    setMessage("");
    setError("");
    try {
      const { url } = await uploadImage(file, "avatars");
      const updated = await updateMyAvatar(url);
      setProfile(updated);
      // Sync avatar to auth context so header/sidebar update immediately
      if (authUser) {
        login({ ...authUser, avatar_url: updated.avatar_url || undefined }, "", "");
      }
      toast.success("Cập nhật ảnh đại diện thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tải ảnh thất bại");
    } finally {
      setUploadingAvatar(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!authUser) return;
    setSaving(true);
    setMessage("");
    setError("");
    try {
      const updated = await updateMe({
        name: name || undefined,
        address: address || undefined,
        province: province || undefined,
        ward: ward || undefined,
        description: description || undefined,
        org_name: orgName || undefined,
      });
      setProfile(updated);
      syncForm(updated);

      // Update auth context so sidebar/header reflect changes immediately
      login({ ...authUser, name: updated.name || undefined, avatar_url: updated.avatar_url || undefined }, "", "");

      toast.success("Cập nhật thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <span className="text-muted-foreground">Đang tải...</span>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-xl font-semibold mb-5">Tài khoản của tôi</h1>

      <Card className="shadow-sm">
        <CardHeader className="pb-4">
          <div className="flex items-center gap-4">
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleAvatarChange}
            />
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploadingAvatar}
              className="relative group"
              title="Thay đổi ảnh đại diện"
            >
              <Avatar className="h-16 w-16">
                <AvatarImage src={profile?.avatar_url} />
                <AvatarFallback className="text-lg font-semibold bg-primary/10 text-primary">
                  {(profile?.name || profile?.phone || "?").slice(0, 1).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity">
                {uploadingAvatar ? (
                  <div className="h-5 w-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <Camera className="h-5 w-5 text-white" />
                )}
              </div>
            </button>
            <div>
              <CardTitle className="text-base">{profile?.name || profile?.phone}</CardTitle>
              <p className="text-sm text-muted-foreground">{profile?.phone}</p>
              <p className="text-xs text-muted-foreground mt-0.5">
                Vai trò: <span className="font-medium capitalize">{profile?.role}</span>
              </p>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="text-sm font-medium">Tên hiển thị</label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Nhập tên hiển thị"
                className="mt-1"
              />
            </div>
            <div>
              <label className="text-sm font-medium">Số điện thoại</label>
              <Input value={profile?.phone || ""} disabled className="mt-1 bg-muted/50" />
              <p className="text-xs text-muted-foreground mt-1">Số điện thoại không thể thay đổi</p>
            </div>

            <div className="pt-2 border-t">
              <p className="text-sm font-semibold mb-3">Địa chỉ</p>
              <div className="space-y-3">
                <div>
                  <label className="text-sm font-medium">Tỉnh/Thành phố</label>
                  <Input
                    value={province}
                    onChange={(e) => setProvince(e.target.value)}
                    placeholder="VD: Hồ Chí Minh"
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Xã/Phường</label>
                  <Input
                    value={ward}
                    onChange={(e) => setWard(e.target.value)}
                    placeholder="VD: Phường Bến Nghé"
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Địa chỉ chi tiết</label>
                  <Input
                    value={address}
                    onChange={(e) => setAddress(e.target.value)}
                    placeholder="Số nhà, tên đường..."
                    className="mt-1"
                  />
                </div>
              </div>
            </div>

            <div className="pt-2 border-t">
              <p className="text-sm font-semibold mb-3">Thông tin thêm</p>
              <div className="space-y-3">
                <div>
                  <label className="text-sm font-medium">Tên tổ chức/Cửa hàng</label>
                  <Input
                    value={orgName}
                    onChange={(e) => setOrgName(e.target.value)}
                    placeholder="VD: Đại lý gạo Minh Tâm"
                    className="mt-1"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium">Mô tả</label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="Giới thiệu về bản thân hoặc cửa hàng..."
                    rows={3}
                    className="mt-1 flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  />
                </div>
              </div>
            </div>

            {message && <p className="text-sm text-emerald-600 font-medium">{message}</p>}
            {error && <p className="text-sm text-destructive">{error}</p>}

            <Button type="submit" disabled={saving} className="w-full">
              {saving ? "Đang lưu..." : "Lưu thay đổi"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
