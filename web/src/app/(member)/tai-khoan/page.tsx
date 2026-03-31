"use client";

import { useEffect, useState, useRef } from "react";
import {
  User, Save, KeyRound, Phone, Eye, EyeOff, Crown, MessageSquareText,
  ChevronRight, LogOut, Camera, Pencil, X, Trash2,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  getMe, updateMe, changePassword, changePhone, sendOTP, uploadImage, updateMyAvatar, deleteAccount,
  type User as UserType,
} from "@/services/api";
import { useAuth } from "@/lib/auth";
import { useThemeColor, THEME_OPTIONS } from "@/lib/theme-color";
import { toast } from "sonner";
import { formatDate } from "@/lib/utils";
import Link from "next/link";
import LocationPicker from "@/components/location-picker";

export default function ProfilePage() {
  const { logout, user: authUser } = useAuth();
  const { themeKey, setThemeKey } = useThemeColor();
  const [profile, setProfile] = useState<UserType | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [editing, setEditing] = useState(false);
  const [uploadingAvatar, setUploadingAvatar] = useState(false);
  const avatarInputRef = useRef<HTMLInputElement>(null);

  // Edit form state
  const [formName, setFormName] = useState("");
  const [formOrgName, setFormOrgName] = useState("");
  const [formAddress, setFormAddress] = useState("");
  const [formDescription, setFormDescription] = useState("");
  const formProvinceRef = useRef<string | undefined>(undefined);
  const formWardRef = useRef<string | undefined>(undefined);

  // Change password state
  const [pwCurrent, setPwCurrent] = useState("");
  const [pwNew, setPwNew] = useState("");
  const [pwConfirm, setPwConfirm] = useState("");
  const [pwSaving, setPwSaving] = useState(false);
  const [showPwCurrent, setShowPwCurrent] = useState(false);
  const [showPwNew, setShowPwNew] = useState(false);
  const [showPwConfirm, setShowPwConfirm] = useState(false);

  // Change phone state
  const [newPhone, setNewPhone] = useState("");
  const [phoneOtp, setPhoneOtp] = useState("");
  const [phoneStep, setPhoneStep] = useState<"input" | "otp">("input");
  const [phoneSaving, setPhoneSaving] = useState(false);

  // Delete account state
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [deletePassword, setDeletePassword] = useState("");
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (authUser) {
      getMe("")
        .then((u) => {
          setProfile(u);
          initForm(u);
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [authUser]);

  function initForm(u: UserType) {
    setFormName(u.name || "");
    setFormOrgName(u.org_name || "");
    setFormAddress(u.address || "");
    setFormDescription(u.description || "");
    formProvinceRef.current = u.province || undefined;
    formWardRef.current = u.ward || undefined;
  }

  function startEditing() {
    if (profile) initForm(profile);
    setEditing(true);
  }

  function cancelEditing() {
    setEditing(false);
  }

  // Avatar upload
  async function handleAvatarUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file || !authUser) return;

    // Validate file
    if (!file.type.startsWith("image/")) {
      toast.error("Vui lòng chọn file ảnh");
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      toast.error("Ảnh không được quá 5MB");
      return;
    }

    setUploadingAvatar(true);
    try {
      const { url } = await uploadImage("", file, "avatars");
      const updated = await updateMyAvatar("", url);
      setProfile(updated);
      toast.success("Cập nhật ảnh đại diện thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Tải ảnh thất bại");
    } finally {
      setUploadingAvatar(false);
      if (avatarInputRef.current) avatarInputRef.current.value = "";
    }
  }

  // Save profile
  async function handleSave() {
    if (!authUser) return;

    const trimmedName = formName.trim();
    if (trimmedName.length < 4 || trimmedName.length > 60) {
      toast.error("Tên phải có từ 4 đến 60 ký tự");
      return;
    }
    const trimmedAddress = formAddress.trim();
    if (trimmedAddress && (trimmedAddress.length < 6 || trimmedAddress.length > 80)) {
      toast.error("Địa chỉ chi tiết phải từ 6 đến 80 ký tự");
      return;
    }

    setSaving(true);
    try {
      const updated = await updateMe("", {
        name: trimmedName,
        province: formProvinceRef.current || "",
        ward: formWardRef.current || "",
        address: trimmedAddress,
        description: formDescription.trim(),
        org_name: formOrgName.trim(),
      });
      setProfile(updated);
      setEditing(false);
      toast.success("Cập nhật thành công");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  }

  // Change password
  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    if (!authUser) return;

    if (pwNew.length < 6) {
      toast.error("Mật khẩu phải có ít nhất 6 ký tự");
      return;
    }
    if (!/[A-Z]/.test(pwNew)) {
      toast.error("Mật khẩu phải có ít nhất 1 chữ hoa");
      return;
    }
    if (!/[a-z]/.test(pwNew)) {
      toast.error("Mật khẩu phải có ít nhất 1 chữ thường");
      return;
    }
    if (!/[^a-zA-Z0-9]/.test(pwNew)) {
      toast.error("Mật khẩu phải có ít nhất 1 ký tự đặc biệt");
      return;
    }
    if (pwNew !== pwConfirm) {
      toast.error("Mật khẩu nhập lại không khớp");
      return;
    }

    setPwSaving(true);
    try {
      await changePassword("", pwCurrent, pwNew);
      toast.success("Đổi mật khẩu thành công");
      setPwCurrent("");
      setPwNew("");
      setPwConfirm("");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Đổi mật khẩu thất bại");
    } finally {
      setPwSaving(false);
    }
  }

  // Change phone
  async function handleSendPhoneOTP() {
    if (!newPhone) {
      toast.error("Vui lòng nhập số điện thoại mới");
      return;
    }
    setPhoneSaving(true);
    try {
      await sendOTP(newPhone);
      setPhoneStep("otp");
      toast.success("Đã gửi mã OTP đến số mới");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Gửi OTP thất bại");
    } finally {
      setPhoneSaving(false);
    }
  }

  async function handleChangePhone(e: React.FormEvent) {
    e.preventDefault();
    if (!authUser) return;
    setPhoneSaving(true);
    try {
      const updated = await changePhone("", newPhone, phoneOtp);
      setProfile(updated);
      toast.success("Đổi số điện thoại thành công");
      setNewPhone("");
      setPhoneOtp("");
      setPhoneStep("input");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Đổi số điện thoại thất bại");
    } finally {
      setPhoneSaving(false);
    }
  }

  // Delete account
  async function handleDeleteAccount(e: React.FormEvent) {
    e.preventDefault();
    if (!authUser || !deletePassword) return;
    setDeleting(true);
    try {
      await deleteAccount("", deletePassword);
      toast.success("Tài khoản đã được xóa");
      logout();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Xóa tài khoản thất bại");
    } finally {
      setDeleting(false);
    }
  }

  // Logout
  function handleLogout() {
    if (confirm("Bạn có chắc muốn đăng xuất?")) {
      logout();
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

  if (!profile) return null;

  const isAdminRole = ["editor", "admin", "owner"].includes(profile.role);

  return (
    <div className="max-w-2xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Tài khoản</h1>
        {!editing && (
          <Button variant="outline" size="sm" className="gap-2" onClick={startEditing}>
            <Pencil className="h-4 w-4" />
            Chỉnh sửa
          </Button>
        )}
      </div>

      <div className="space-y-4">
        {/* Avatar + Name header */}
        <div className="flex flex-col items-center gap-3 pb-4">
          <div className="relative">
            <Avatar className="h-24 w-24">
              <AvatarImage src={profile.avatar_url} alt={profile.name} />
              <AvatarFallback className="bg-primary/10 text-primary text-2xl font-semibold">
                {(profile.name || "?").charAt(0).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <button
              type="button"
              onClick={() => avatarInputRef.current?.click()}
              disabled={uploadingAvatar}
              className="absolute bottom-0 right-0 h-8 w-8 rounded-full bg-primary text-white flex items-center justify-center border-2 border-background hover:bg-primary/90 transition-colors disabled:opacity-50"
            >
              {uploadingAvatar ? (
                <div className="h-4 w-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <Camera className="h-4 w-4" />
              )}
            </button>
            <input
              ref={avatarInputRef}
              type="file"
              accept="image/*"
              className="hidden"
              onChange={handleAvatarUpload}
            />
          </div>
          <div className="text-center">
            <p className="text-xl font-bold">{profile.name || profile.phone}</p>
            <p className="text-sm text-muted-foreground">Thành viên</p>
          </div>
        </div>

        {editing ? (
          /* ========== EDIT MODE ========== */
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                Chỉnh sửa thông tin
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium mb-1 block">Tên hiển thị (4-60 ký tự)</label>
                <Input
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  maxLength={60}
                />
              </div>

              <div>
                <label className="text-sm font-medium mb-1 block">Tỉnh/Thành phố & Phường/Xã</label>
                <LocationPicker
                  initialProvince={profile.province}
                  initialWard={profile.ward}
                  onChanged={(province, ward) => {
                    formProvinceRef.current = province;
                    formWardRef.current = ward;
                  }}
                />
              </div>

              <div>
                <label className="text-sm font-medium mb-1 block">Địa chỉ chi tiết (6-80 ký tự)</label>
                <Input
                  value={formAddress}
                  onChange={(e) => setFormAddress(e.target.value)}
                  placeholder="VD: 123 Nguyễn Huệ, Quận 1"
                  maxLength={80}
                />
              </div>

              <div>
                <label className="text-sm font-medium mb-1 block">Giới thiệu</label>
                <textarea
                  value={formDescription}
                  onChange={(e) => setFormDescription(e.target.value)}
                  className="w-full min-h-20 rounded-md border border-input bg-background px-3 py-2 text-sm"
                  placeholder="Giới thiệu về bạn hoặc doanh nghiệp"
                />
              </div>

              <div>
                <label className="text-sm font-medium mb-1 block">Tên tổ chức/doanh nghiệp (nếu có)</label>
                <Input
                  value={formOrgName}
                  onChange={(e) => setFormOrgName(e.target.value)}
                />
              </div>

              <div className="flex gap-3">
                <Button variant="outline" onClick={cancelEditing} className="flex-1">
                  Huỷ
                </Button>
                <Button onClick={handleSave} disabled={saving} className="flex-1 gap-2">
                  <Save className="h-4 w-4" />
                  {saving ? "Đang lưu..." : "Lưu"}
                </Button>
              </div>
            </CardContent>
          </Card>
        ) : (
          /* ========== VIEW MODE ========== */
          <>
            {/* Profile info */}
            <Card>
              <CardContent className="p-4 space-y-3">
                <InfoRow icon={<Phone className="h-4 w-4" />} label="Số điện thoại" value={profile.phone} />
                {profile.province && (
                  <InfoRow icon={<User className="h-4 w-4" />} label="Tỉnh/Thành phố" value={profile.province} />
                )}
                {profile.ward && (
                  <InfoRow icon={<User className="h-4 w-4" />} label="Phường/Xã" value={profile.ward} />
                )}
                {profile.address && (
                  <InfoRow icon={<User className="h-4 w-4" />} label="Địa chỉ" value={profile.address} />
                )}
                {profile.org_name && (
                  <InfoRow icon={<User className="h-4 w-4" />} label="Tổ chức" value={profile.org_name} />
                )}
                {profile.description && (
                  <InfoRow icon={<User className="h-4 w-4" />} label="Giới thiệu" value={profile.description} />
                )}
              </CardContent>
            </Card>

            {/* Navigation menu - same as mobile */}
            <Card>
              <CardContent className="p-2">
                {!isAdminRole && (
                  <Link href="/goi-thanh-vien" className="flex items-center justify-between p-3 rounded-lg hover:bg-muted transition-colors">
                    <div className="flex items-center gap-3">
                      <Crown className="h-5 w-5 text-muted-foreground" />
                      <span className="text-sm font-medium">Gói dịch vụ & Gia hạn</span>
                    </div>
                    <ChevronRight className="h-4 w-4 text-muted-foreground" />
                  </Link>
                )}
                <button
                  type="button"
                  onClick={() => document.getElementById("change-password")?.scrollIntoView({ behavior: "smooth" })}
                  className="w-full flex items-center justify-between p-3 rounded-lg hover:bg-muted transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <KeyRound className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">Đổi mật khẩu</span>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
                <button
                  type="button"
                  onClick={() => document.getElementById("change-phone")?.scrollIntoView({ behavior: "smooth" })}
                  className="w-full flex items-center justify-between p-3 rounded-lg hover:bg-muted transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <Phone className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">Đổi số điện thoại</span>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </button>
                <Link href="/phan-hoi" className="flex items-center justify-between p-3 rounded-lg hover:bg-muted transition-colors">
                  <div className="flex items-center gap-3">
                    <MessageSquareText className="h-5 w-5 text-muted-foreground" />
                    <span className="text-sm font-medium">Góp ý cho nhà phát triển</span>
                  </div>
                  <ChevronRight className="h-4 w-4 text-muted-foreground" />
                </Link>
                {/* Theme color picker */}
                <div className="p-3">
                  <p className="text-sm font-medium mb-2 flex items-center gap-2">
                    <span className="h-5 w-5 flex items-center justify-center">🎨</span>
                    Màu chủ đạo
                  </p>
                  <div className="flex gap-2">
                    {THEME_OPTIONS.map((t) => (
                      <button
                        key={t.key}
                        type="button"
                        onClick={() => setThemeKey(t.key)}
                        className={`h-8 w-8 rounded-full border-2 transition-all ${
                          themeKey === t.key ? "border-foreground scale-110" : "border-transparent hover:scale-105"
                        }`}
                        style={{ backgroundColor: t.hex }}
                        title={t.label}
                      />
                    ))}
                  </div>
                </div>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="w-full flex items-center justify-between p-3 rounded-lg hover:bg-muted transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <LogOut className="h-5 w-5 text-destructive" />
                    <span className="text-sm font-medium text-destructive">Đăng xuất</span>
                  </div>
                </button>
              </CardContent>
            </Card>
          </>
        )}

        {/* Change Password */}
        <Card id="change-password">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <KeyRound className="h-4 w-4" />
              Đổi mật khẩu
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleChangePassword} className="space-y-4">
              <div>
                <label className="text-sm font-medium mb-1 block">Mật khẩu hiện tại</label>
                <div className="relative">
                  <Input
                    type={showPwCurrent ? "text" : "password"}
                    value={pwCurrent}
                    onChange={(e) => setPwCurrent(e.target.value)}
                    placeholder="Nhập mật khẩu hiện tại"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPwCurrent(!showPwCurrent)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPwCurrent ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Mật khẩu mới</label>
                <div className="relative">
                  <Input
                    type={showPwNew ? "text" : "password"}
                    value={pwNew}
                    onChange={(e) => setPwNew(e.target.value)}
                    placeholder="Chữ hoa, chữ thường, ký tự đặc biệt"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPwNew(!showPwNew)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPwNew ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium mb-1 block">Nhập lại mật khẩu mới</label>
                <div className="relative">
                  <Input
                    type={showPwConfirm ? "text" : "password"}
                    value={pwConfirm}
                    onChange={(e) => setPwConfirm(e.target.value)}
                    placeholder="Nhập lại mật khẩu mới"
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPwConfirm(!showPwConfirm)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showPwConfirm ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>
              <Button type="submit" variant="outline" className="gap-2" disabled={pwSaving || !pwNew}>
                <KeyRound className="h-4 w-4" />
                {pwSaving ? "Đang xử lý..." : "Đổi mật khẩu"}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Change Phone */}
        <Card id="change-phone">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
              <Phone className="h-4 w-4" />
              Đổi số điện thoại
            </CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground mb-3">
              Số hiện tại: <strong>{profile.phone}</strong>
            </p>
            {phoneStep === "input" ? (
              <div className="space-y-4">
                <div>
                  <label className="text-sm font-medium mb-1 block">Số điện thoại mới</label>
                  <Input
                    type="tel"
                    value={newPhone}
                    onChange={(e) => setNewPhone(e.target.value)}
                    placeholder="VD: 0901234567"
                  />
                </div>
                <Button
                  type="button"
                  variant="outline"
                  className="gap-2"
                  disabled={phoneSaving || !newPhone}
                  onClick={handleSendPhoneOTP}
                >
                  <Phone className="h-4 w-4" />
                  {phoneSaving ? "Đang gửi..." : "Gửi mã OTP"}
                </Button>
              </div>
            ) : (
              <form onSubmit={handleChangePhone} className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Mã OTP đã được gửi đến <strong>{newPhone}</strong>
                </p>
                <div>
                  <label className="text-sm font-medium mb-1 block">Mã OTP</label>
                  <Input
                    type="text"
                    inputMode="numeric"
                    value={phoneOtp}
                    onChange={(e) => setPhoneOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
                    placeholder="Nhập mã OTP 6 số"
                    maxLength={6}
                    className="text-center tracking-widest"
                  />
                </div>
                <div className="flex gap-2">
                  <Button type="submit" variant="outline" className="gap-2" disabled={phoneSaving || !phoneOtp}>
                    {phoneSaving ? "Đang xử lý..." : "Xác nhận đổi SĐT"}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => { setPhoneStep("input"); setPhoneOtp(""); }}
                  >
                    Quay lại
                  </Button>
                </div>
              </form>
            )}
          </CardContent>
        </Card>

        {/* Delete Account */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-destructive flex items-center gap-2">
              <Trash2 className="h-4 w-4" />
              Xoá tài khoản
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!showDeleteConfirm ? (
              <div>
                <p className="text-sm text-muted-foreground mb-3">
                  Xoá tài khoản sẽ xoá toàn bộ dữ liệu của bạn và không thể khôi phục.
                </p>
                <Button variant="destructive" size="sm" onClick={() => setShowDeleteConfirm(true)}>
                  Xoá tài khoản
                </Button>
              </div>
            ) : (
              <form onSubmit={handleDeleteAccount} className="space-y-4">
                <p className="text-sm text-destructive font-medium">
                  Nhập mật khẩu để xác nhận xoá tài khoản. Hành động này không thể hoàn tác.
                </p>
                <div>
                  <label className="text-sm font-medium mb-1 block">Mật khẩu xác nhận</label>
                  <Input
                    type="password"
                    value={deletePassword}
                    onChange={(e) => setDeletePassword(e.target.value)}
                    placeholder="Nhập mật khẩu"
                  />
                </div>
                <div className="flex gap-2">
                  <Button type="submit" variant="destructive" size="sm" disabled={deleting || !deletePassword}>
                    {deleting ? "Đang xoá..." : "Xác nhận xoá"}
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => { setShowDeleteConfirm(false); setDeletePassword(""); }}
                  >
                    Huỷ
                  </Button>
                </div>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function InfoRow({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <div className="flex items-start gap-3">
      <div className="text-muted-foreground mt-0.5">{icon}</div>
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="text-sm">{value}</p>
      </div>
    </div>
  );
}
