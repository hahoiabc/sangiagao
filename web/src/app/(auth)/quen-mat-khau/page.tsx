"use client";

import { useState } from "react";
import Link from "next/link";
import { KeyRound, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { sendOTP, resetPassword } from "@/services/api";

export default function ForgotPasswordPage() {
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [step, setStep] = useState<"phone" | "reset">("phone");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSendOTP(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await sendOTP(phone);
      setCode("123456");
      setStep("reset");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gửi mã OTP thất bại");
    } finally {
      setLoading(false);
    }
  }

  async function handleReset(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (newPassword.length < 6) {
      setError("Mật khẩu phải có ít nhất 6 ký tự");
      return;
    }
    if (!/[A-Z]/.test(newPassword) || !/[a-z]/.test(newPassword) || !/[^a-zA-Z0-9]/.test(newPassword)) {
      setError("Mật khẩu phải có chữ hoa, chữ thường và ký tự đặc biệt");
      return;
    }
    if (newPassword !== confirmPassword) {
      setError("Mật khẩu nhập lại không khớp");
      return;
    }

    setLoading(true);
    try {
      await resetPassword(phone, code, newPassword);
      setSuccess("Đặt lại mật khẩu thành công! Bạn có thể đăng nhập.");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Đặt lại mật khẩu thất bại");
    } finally {
      setLoading(false);
    }
  }

  if (success) {
    return (
      <Card className="w-full max-w-sm shadow-xl border-0">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-100">
            <KeyRound className="h-6 w-6 text-emerald-600" />
          </div>
          <CardTitle className="text-xl font-bold">Thành công</CardTitle>
          <CardDescription>{success}</CardDescription>
        </CardHeader>
        <CardContent>
          <Link href="/dang-nhap">
            <Button className="w-full h-11">Đăng nhập</Button>
          </Link>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="w-full max-w-sm shadow-xl border-0">
      <CardHeader className="text-center pb-2">
        <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
          <KeyRound className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-xl font-bold">
          {step === "phone" ? "Quên mật khẩu" : "Đặt lại mật khẩu"}
        </CardTitle>
        <CardDescription>
          {step === "phone"
            ? "Nhập số điện thoại để nhận mã OTP"
            : "Nhập mã OTP và mật khẩu mới"}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {step === "phone" ? (
          <form onSubmit={handleSendOTP} className="space-y-4">
            <Input
              type="tel"
              placeholder="Số điện thoại (VD: 0901234567)"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="h-11"
              required
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full h-11" disabled={loading}>
              {loading ? "Đang gửi..." : "Gửi mã OTP"}
            </Button>
          </form>
        ) : (
          <form onSubmit={handleReset} className="space-y-4">
            <Input
              type="text"
              placeholder="Mã OTP (6 số)"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              maxLength={6}
              className="h-11 text-center text-lg tracking-widest"
              required
            />
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                placeholder="Mật khẩu mới"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="h-11 pr-10"
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
              </button>
            </div>
            <Input
              type={showPassword ? "text" : "password"}
              placeholder="Nhập lại mật khẩu mới"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              className="h-11"
              required
            />
            <p className="text-xs text-muted-foreground">
              Mật khẩu phải có chữ hoa, chữ thường và ký tự đặc biệt
            </p>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full h-11" disabled={loading}>
              {loading ? "Đang xử lý..." : "Đặt lại mật khẩu"}
            </Button>
            <button
              type="button"
              onClick={() => { setStep("phone"); setError(""); }}
              className="w-full text-sm text-muted-foreground hover:text-foreground"
            >
              Quay lại
            </button>
          </form>
        )}
        <div className="mt-4 text-center text-sm">
          <Link href="/dang-nhap" className="text-primary hover:underline">
            Quay lại đăng nhập
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
