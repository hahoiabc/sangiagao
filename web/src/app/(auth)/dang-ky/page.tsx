"use client";

import { useState, useRef } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Wheat, Eye, EyeOff, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/lib/auth";
import { register, completeRegister } from "@/services/api";
import LocationPicker from "@/components/location-picker";

const PHONE_REGEX = /^0(3[2-9]|5[2689]|7[06-9]|8[1-689]|9[0-46-9])\d{7}$/;

export default function RegisterPage() {
  const [step, setStep] = useState(1);
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [address, setAddress] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [acceptedTOS, setAcceptedTOS] = useState(false);
  const [showTerms, setShowTerms] = useState(false);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  // Location state
  const provinceRef = useRef<string | undefined>(undefined);
  const wardRef = useRef<string | undefined>(undefined);

  function handleLocationChanged(province: string | undefined, ward: string | undefined) {
    provinceRef.current = province;
    wardRef.current = ward;
  }

  // Step 1: Validate phone + send OTP
  async function handleStep1(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (!PHONE_REGEX.test(phone)) {
      setError("Số điện thoại không hợp lệ, vui lòng kiểm tra đầu số");
      return;
    }

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

  // Step 2: OTP + all details
  async function handleStep2(e: React.FormEvent) {
    e.preventDefault();
    setError("");

    if (code.length !== 6) {
      setError("Vui lòng nhập mã OTP 6 số");
      return;
    }
    if (name.trim().length < 4 || name.trim().length > 60) {
      setError("Tên hiển thị phải từ 4 đến 60 ký tự");
      return;
    }
    if (password.length < 6) {
      setError("Mật khẩu phải có ít nhất 6 ký tự");
      return;
    }
    if (!/[A-Z]/.test(password)) {
      setError("Mật khẩu phải có ít nhất 1 chữ hoa");
      return;
    }
    if (!/[a-z]/.test(password)) {
      setError("Mật khẩu phải có ít nhất 1 chữ thường");
      return;
    }
    if (!/[^a-zA-Z0-9]/.test(password)) {
      setError("Mật khẩu phải có ít nhất 1 ký tự đặc biệt");
      return;
    }
    if (password !== confirmPassword) {
      setError("Mật khẩu nhập lại không khớp");
      return;
    }
    const trimmedAddress = address.trim();
    if (trimmedAddress && (trimmedAddress.length < 6 || trimmedAddress.length > 80)) {
      setError("Địa chỉ chi tiết phải từ 6 đến 80 ký tự");
      return;
    }
    if (!acceptedTOS) {
      setError("Vui lòng đồng ý điều khoản sử dụng");
      return;
    }

    setLoading(true);
    try {
      const result = await completeRegister({
        phone,
        code,
        name: name.trim(),
        password,
        province: provinceRef.current || undefined,
        ward: wardRef.current || undefined,
        address: trimmedAddress || undefined,
      });
      login(result.user, result.tokens.access_token, result.tokens.refresh_token);
      router.push("/san-giao-dich");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Đăng ký thất bại");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <Card className="w-full max-w-sm shadow-xl border-0">
        <CardHeader className="text-center pb-2">
          <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-primary/10">
            <Wheat className="h-6 w-6 text-primary" />
          </div>
          <CardTitle className="text-2xl font-bold text-primary">Đăng Ký</CardTitle>
          <CardDescription>
            {step === 1
              ? "Nhập số điện thoại để đăng ký"
              : `Xác minh và hoàn tất đăng ký (${phone})`}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {step === 1 ? (
            <form onSubmit={handleStep1} className="space-y-4">
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
                {loading ? "Đang gửi..." : "Tiếp tục"}
              </Button>
            </form>
          ) : (
            <form onSubmit={handleStep2} className="space-y-3">
              {/* OTP */}
              <Input
                type="text"
                inputMode="numeric"
                placeholder="Mã OTP 6 số"
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                className="h-11 text-center text-lg tracking-widest"
                maxLength={6}
                required
              />

              {/* Name */}
              <Input
                type="text"
                placeholder="Họ và tên (4-60 ký tự)"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="h-11"
                maxLength={60}
                required
              />

              {/* Password - separate toggle */}
              <div className="relative">
                <Input
                  type={showPassword ? "text" : "password"}
                  placeholder="Mật khẩu (chữ hoa, thường, đặc biệt)"
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

              {/* Confirm Password - separate toggle */}
              <div className="relative">
                <Input
                  type={showConfirmPassword ? "text" : "password"}
                  placeholder="Nhập lại mật khẩu"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="h-11 pr-10"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                >
                  {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>

              {/* Location Picker (Province → Ward) */}
              <LocationPicker onChanged={handleLocationChanged} />

              {/* Address */}
              <Input
                placeholder="Địa chỉ chi tiết (6-80 ký tự, không bắt buộc)"
                value={address}
                onChange={(e) => setAddress(e.target.value)}
                className="h-11"
                maxLength={80}
              />

              {/* T&C Checkbox */}
              <div className="flex items-start gap-2">
                <input
                  type="checkbox"
                  id="tos"
                  checked={acceptedTOS}
                  onChange={(e) => setAcceptedTOS(e.target.checked)}
                  className="mt-1 h-4 w-4 rounded border-input accent-primary"
                />
                <label htmlFor="tos" className="text-sm text-muted-foreground">
                  Tôi đã đọc và đồng ý với{" "}
                  <button
                    type="button"
                    onClick={() => setShowTerms(true)}
                    className="text-primary font-medium underline hover:no-underline"
                  >
                    Điều khoản sử dụng
                  </button>
                </label>
              </div>

              {error && <p className="text-sm text-destructive">{error}</p>}

              <Button type="submit" className="w-full h-11" disabled={loading || !acceptedTOS}>
                {loading ? "Đang xử lý..." : "Đăng ký"}
              </Button>

              <button
                type="button"
                onClick={() => { setStep(1); setError(""); setCode(""); }}
                className="w-full text-sm text-muted-foreground hover:text-foreground"
              >
                Quay lại
              </button>
            </form>
          )}
          <div className="mt-4 text-center text-sm">
            <Link href="/dang-nhap" className="text-primary hover:underline">
              Đã có tài khoản? Đăng nhập
            </Link>
          </div>
        </CardContent>
      </Card>

      {/* Terms of Service Modal */}
      {showTerms && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="bg-background rounded-lg max-w-lg w-full max-h-[85vh] flex flex-col shadow-xl">
            <div className="flex items-center justify-between p-4 border-b">
              <h2 className="font-semibold text-lg">Điều khoản sử dụng</h2>
              <button
                type="button"
                onClick={() => setShowTerms(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-5 w-5" />
              </button>
            </div>
            <div className="overflow-y-auto p-4 space-y-4 text-sm leading-relaxed">
              <h3 className="font-bold">ĐIỀU KHOẢN SỬ DỤNG SANGIAGAO.COM</h3>

              <div>
                <h4 className="font-semibold">1. GIỚI THIỆU</h4>
                <p className="mt-1">
                  SanGiaGao.Com là nền tảng công nghệ kết nối người sản xuất, thương nhân
                  và người mua trong ngành gạo Việt Nam. SanGiaGao.Com là công cụ hỗ trợ
                  giúp các thành viên kết nối thuận tiện và nhanh chóng, không trực tiếp
                  tham gia vào các giao dịch mua bán giữa các bên.
                </p>
              </div>

              <div>
                <h4 className="font-semibold">2. TRÁCH NHIỆM CỦA THÀNH VIÊN</h4>
                <p className="mt-1">
                  Mỗi thành viên tham gia SanGiaGao.Com phải tự chịu trách nhiệm hoàn toàn
                  cho mọi quyết định giao dịch của mình, bao gồm nhưng không giới hạn:
                </p>
                <ul className="list-disc ml-5 mt-1 space-y-1">
                  <li>Tính chính xác của thông tin sản phẩm đăng tải</li>
                  <li>Chất lượng hàng hóa và dịch vụ</li>
                  <li>Việc thỏa thuận giá cả, số lượng và điều kiện giao hàng</li>
                  <li>Thanh toán và các nghĩa vụ tài chính phát sinh</li>
                  <li>Tuân thủ pháp luật Việt Nam trong quá trình giao dịch</li>
                </ul>
              </div>

              <div>
                <h4 className="font-semibold">3. VAI TRÒ CỦA SANGIAGAO.COM</h4>
                <p className="mt-1">SanGiaGao.Com cam kết:</p>
                <ul className="list-disc ml-5 mt-1 space-y-1">
                  <li>Cung cấp nền tảng kết nối minh bạch và công bằng</li>
                  <li>Hỗ trợ cung cấp thông tin trong khả năng của sàn khi có yêu cầu từ thành viên</li>
                  <li>Duy trì môi trường giao dịch lành mạnh thông qua hệ thống đánh giá và báo cáo vi phạm</li>
                  <li>Bảo mật thông tin cá nhân của thành viên theo quy định pháp luật</li>
                </ul>
                <p className="mt-2">
                  SanGiaGao.Com không chịu trách nhiệm cho bất kỳ tranh chấp, tổn thất hoặc thiệt hại
                  phát sinh từ giao dịch giữa các thành viên.
                </p>
              </div>

              <div>
                <h4 className="font-semibold">4. GÓI DỊCH VỤ</h4>
                <ul className="list-disc ml-5 mt-1 space-y-1">
                  <li>Thành viên được dùng thử miễn phí 30 ngày kể từ ngày đăng ký</li>
                  <li>Sau thời gian dùng thử, phí dịch vụ sẽ được tính theo các gói đăng ký của thành viên</li>
                  <li>Khi hết hạn gói dịch vụ, tin đăng sẽ bị tạm ẩn cho đến khi gia hạn</li>
                </ul>
              </div>

              <div>
                <h4 className="font-semibold">5. NỘI DUNG BỊ CẤM</h4>
                <ul className="list-disc ml-5 mt-1 space-y-1">
                  <li>Đăng thông tin sai lệch, gian lận</li>
                  <li>Sử dụng sàn cho mục đích bất hợp pháp</li>
                  <li>Quấy rối, đe dọa hoặc xúc phạm thành viên khác</li>
                  <li>Spam, đăng tin trùng lặp hoặc không liên quan đến gạo/nông sản</li>
                </ul>
              </div>

              <div>
                <h4 className="font-semibold">6. XỬ LÝ VI PHẠM</h4>
                <p className="mt-1">
                  SanGiaGao.Com có quyền cảnh cáo, tạm khóa hoặc xóa vĩnh viễn tài khoản vi phạm
                  điều khoản sử dụng mà không cần thông báo trước.
                </p>
              </div>

              <p className="italic text-muted-foreground">
                Bằng việc tích chọn &quot;Đồng ý điều khoản&quot;, bạn xác nhận đã đọc, hiểu và chấp nhận
                toàn bộ các điều khoản trên.
              </p>
            </div>
            <div className="p-4 border-t">
              <Button
                className="w-full"
                onClick={() => {
                  setAcceptedTOS(true);
                  setShowTerms(false);
                }}
              >
                Đồng ý điều khoản
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
