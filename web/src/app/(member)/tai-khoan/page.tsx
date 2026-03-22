"use client";

import { useEffect, useState } from "react";
import { User, Save, KeyRound, Phone, Eye, EyeOff } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { getMe, updateMe, changePassword, changePhone, sendOTP, type User as UserType } from "@/services/api";
import { useAuth } from "@/lib/auth";
import { toast } from "sonner";
import { formatDate } from "@/lib/utils";

export default function ProfilePage() {
  const { token } = useAuth();
  const [profile, setProfile] = useState<UserType | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ name: "", province: "", district: "", ward: "", address: "", description: "", org_name: "" });

  // Change password state
  const [pwCurrent, setPwCurrent] = useState("");
  const [pwNew, setPwNew] = useState("");
  const [pwConfirm, setPwConfirm] = useState("");
  const [pwSaving, setPwSaving] = useState(false);
  const [showPw, setShowPw] = useState(false);

  // Change phone state
  const [newPhone, setNewPhone] = useState("");
  const [phoneOtp, setPhoneOtp] = useState("");
  const [phoneStep, setPhoneStep] = useState<"input" | "otp">("input");
  const [phoneSaving, setPhoneSaving] = useState(false);

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

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    if (!token) return;

    if (pwNew.length < 6) {
      toast.error("Mật khẩu phải có ít nhất 6 ký tự");
      return;
    }
    if (!/[A-Z]/.test(pwNew) || !/[a-z]/.test(pwNew) || !/[^a-zA-Z0-9]/.test(pwNew)) {
      toast.error("Mật khẩu phải có chữ hoa, chữ thường và ký tự đặc biệt");
      return;
    }
    if (pwNew !== pwConfirm) {
      toast.error("Mật khẩu nhập lại không khớp");
      return;
    }

    setPwSaving(true);
    try {
      await changePassword(token, pwCurrent, pwNew);
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
    if (!token) return;
    setPhoneSaving(true);
    try {
      const updated = await changePhone(token, newPhone, phoneOtp);
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

          {/* Change Password */}
          <Card>
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
                      type={showPw ? "text" : "password"}
                      value={pwCurrent}
                      onChange={(e) => setPwCurrent(e.target.value)}
                      placeholder="Nhập mật khẩu hiện tại"
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPw(!showPw)}
                      className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                    >
                      {showPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Mật khẩu mới</label>
                  <Input
                    type={showPw ? "text" : "password"}
                    value={pwNew}
                    onChange={(e) => setPwNew(e.target.value)}
                    placeholder="Chữ hoa, chữ thường, ký tự đặc biệt"
                  />
                </div>
                <div>
                  <label className="text-sm font-medium mb-1 block">Nhập lại mật khẩu mới</label>
                  <Input
                    type={showPw ? "text" : "password"}
                    value={pwConfirm}
                    onChange={(e) => setPwConfirm(e.target.value)}
                    placeholder="Nhập lại mật khẩu mới"
                  />
                </div>
                <Button type="submit" variant="outline" className="gap-2" disabled={pwSaving || !pwNew}>
                  <KeyRound className="h-4 w-4" />
                  {pwSaving ? "Đang xử lý..." : "Đổi mật khẩu"}
                </Button>
              </form>
            </CardContent>
          </Card>

          {/* Change Phone */}
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <Phone className="h-4 w-4" />
                Đổi số điện thoại
              </CardTitle>
            </CardHeader>
            <CardContent>
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
                      value={phoneOtp}
                      onChange={(e) => setPhoneOtp(e.target.value)}
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
        </div>
      )}
    </div>
  );
}
