"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Wheat, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/lib/auth";
import { register, completeRegister } from "@/services/api";

export default function RegisterPage() {
  const [step, setStep] = useState(1);
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  async function handleStep1(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      await register(phone);
      setStep(2);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Đăng ký thất bại");
    } finally {
      setLoading(false);
    }
  }

  async function handleStep2(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const result = await completeRegister({ phone, code, name, password });
      login(result.user, result.tokens.access_token, result.tokens.refresh_token);
      router.push("/san-giao-dich");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Đăng ký thất bại");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card className="w-full max-w-sm shadow-xl border-0">
      <CardHeader className="text-center pb-2">
        <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
          <Wheat className="h-6 w-6 text-primary" />
        </div>
        <CardTitle className="text-2xl font-bold text-primary">Đăng Ký</CardTitle>
        <CardDescription>
          {step === 1 ? "Nhập số điện thoại để đăng ký" : "Hoàn tất đăng ký"}
        </CardDescription>
      </CardHeader>
      <CardContent>
        {step === 1 ? (
          <form onSubmit={handleStep1} className="space-y-4">
            <Input
              type="tel"
              placeholder="Số điện thoại"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="h-11"
              required
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full h-11" disabled={loading}>
              {loading ? "Đang gửi..." : "Tiếp tục"}
            </Button>
          </form>
        ) : (
          <form onSubmit={handleStep2} className="space-y-4">
            <Input
              type="text"
              placeholder="Mã xác thực"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              className="h-11"
              required
            />
            <Input
              type="text"
              placeholder="Tên hiển thị"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="h-11"
              required
            />
            <div className="relative">
              <Input
                type={showPassword ? "text" : "password"}
                placeholder="Mật khẩu"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
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
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full h-11" disabled={loading}>
              {loading ? "Đang xử lý..." : "Đăng ký"}
            </Button>
          </form>
        )}
        <div className="mt-4 text-center text-sm">
          <Link href="/dang-nhap" className="text-primary hover:underline">
            Đã có tài khoản? Đăng nhập
          </Link>
        </div>
      </CardContent>
    </Card>
  );
}
