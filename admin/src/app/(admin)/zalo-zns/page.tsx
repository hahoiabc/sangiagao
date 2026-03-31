"use client";

import { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { RefreshCw, Save, Send, CheckCircle, XCircle, AlertTriangle, Info, Copy } from "lucide-react";
import { toast } from "sonner";
import { getZaloZNSStatus, updateZaloRefreshToken, testZaloZNS, type ZaloZNSStatus } from "@/services/api";

export default function ZaloZNSPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Zalo ZNS OTP</h1>
        <p className="text-muted-foreground text-sm mt-1">
          Quản lý cấu hình gửi OTP qua Zalo Notification Service
        </p>
      </div>

      <Tabs defaultValue="config">
        <TabsList variant="line">
          <TabsTrigger value="config">Cấu hình</TabsTrigger>
          <TabsTrigger value="guide">Hướng dẫn & Mã lỗi</TabsTrigger>
        </TabsList>

        <TabsContent value="config" className="mt-4">
          <ConfigTab />
        </TabsContent>

        <TabsContent value="guide" className="mt-4">
          <GuideTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function StatusBadge({ value }: { value: string }) {
  if (value === "valid") {
    return <span className="inline-flex items-center gap-1 text-xs font-medium text-green-700 bg-green-50 px-2 py-0.5 rounded-full"><CheckCircle className="h-3 w-3" /> Hoạt động</span>;
  }
  if (value === "expired") {
    return <span className="inline-flex items-center gap-1 text-xs font-medium text-orange-700 bg-orange-50 px-2 py-0.5 rounded-full"><AlertTriangle className="h-3 w-3" /> Hết hạn</span>;
  }
  return <span className="inline-flex items-center gap-1 text-xs font-medium text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full"><XCircle className="h-3 w-3" /> Chưa có</span>;
}

function ConfigTab() {
  const [status, setStatus] = useState<ZaloZNSStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [newRefreshToken, setNewRefreshToken] = useState("");
  const [saving, setSaving] = useState(false);
  const [testPhone, setTestPhone] = useState("");
  const [testing, setTesting] = useState(false);

  const fetchStatus = useCallback(async () => {
    setLoading(true);
    try {
      const data = await getZaloZNSStatus();
      setStatus(data);
    } catch (err) {
      toast.error("Không lấy được trạng thái ZNS");
      console.error(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
  }, [fetchStatus]);

  async function handleUpdateToken() {
    if (!newRefreshToken.trim()) return;
    setSaving(true);
    try {
      const res = await updateZaloRefreshToken(newRefreshToken.trim());
      toast.success(res.message);
      setNewRefreshToken("");
      fetchStatus();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  }

  async function handleTest() {
    if (!testPhone.trim()) return;
    setTesting(true);
    try {
      const res = await testZaloZNS(testPhone.trim());
      toast.success(res.message);
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Gửi test thất bại";
      toast.error(msg);
    } finally {
      setTesting(false);
    }
  }

  if (loading) {
    return <div className="text-sm text-muted-foreground py-8 text-center">Đang tải...</div>;
  }

  if (!status) {
    return <div className="text-sm text-red-500 py-8 text-center">Không lấy được trạng thái</div>;
  }

  if (!status.enabled) {
    return (
      <div className="rounded-lg border border-orange-200 bg-orange-50 p-6 text-center">
        <AlertTriangle className="h-8 w-8 text-orange-500 mx-auto mb-2" />
        <p className="font-medium text-orange-800">Zalo ZNS chưa được bật</p>
        <p className="text-sm text-orange-600 mt-1">
          Đặt <code className="bg-orange-100 px-1 rounded">SMS_PROVIDER=zalo+mock</code> trong file <code className="bg-orange-100 px-1 rounded">.env.backend</code> rồi deploy lại backend.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-2xl">
      {/* Current status */}
      <div className="rounded-lg border p-4 space-y-3">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold">Trạng thái hiện tại</h3>
          <Button variant="ghost" size="sm" onClick={fetchStatus} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-1.5 ${loading ? "animate-spin" : ""}`} />
            Làm mới
          </Button>
        </div>

        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-muted-foreground">App ID:</span>
            <span className="ml-2 font-mono">{status.app_id}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Template ID:</span>
            <span className="ml-2 font-mono">{status.template_id}</span>
          </div>
          <div>
            <span className="text-muted-foreground">App Secret:</span>
            <span className="ml-2 font-mono">{status.app_secret}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Refresh Token:</span>
            <span className="ml-2 font-mono">{status.refresh_token}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Access Token:</span>
            <span className="ml-2"><StatusBadge value={status.access_token || "no_token"} /></span>
          </div>
          <div>
            <span className="text-muted-foreground">Token hết hạn:</span>
            <span className="ml-2 font-mono text-xs">
              {status.token_expiry && status.token_expiry !== "0001-01-01T00:00:00Z"
                ? new Date(status.token_expiry).toLocaleString("vi-VN")
                : "—"}
            </span>
          </div>
          <div>
            <span className="text-muted-foreground">Nguồn token:</span>
            <span className="ml-2">{status.refresh_source === "redis" ? "Redis (tự động)" : "ENV file"}</span>
          </div>
          <div>
            <span className="text-muted-foreground">Redis:</span>
            <span className="ml-2">
              {status.redis_connected
                ? <span className="text-green-600">Kết nối</span>
                : <span className="text-red-500">Không kết nối</span>}
            </span>
          </div>
        </div>
      </div>

      {/* Update refresh token */}
      <div className="rounded-lg border p-4 space-y-3">
        <h3 className="font-semibold">Cập nhật Refresh Token</h3>
        <p className="text-xs text-muted-foreground">
          Chỉ cần cập nhật khi token hết hạn (không dùng &gt;90 ngày) hoặc Redis mất data.
          Token mới lấy từ <a href="https://developers.zalo.me" target="_blank" rel="noopener" className="text-blue-600 underline">Zalo Developer Console</a>.
        </p>
        <div className="flex gap-2">
          <Input
            value={newRefreshToken}
            onChange={(e) => setNewRefreshToken(e.target.value)}
            placeholder="Dán refresh token mới..."
            className="font-mono text-xs"
          />
          <Button onClick={handleUpdateToken} disabled={saving || !newRefreshToken.trim()}>
            <Save className="h-4 w-4 mr-1.5" />
            {saving ? "Đang lưu..." : "Lưu"}
          </Button>
        </div>
      </div>

      {/* Test send */}
      <div className="rounded-lg border p-4 space-y-3">
        <h3 className="font-semibold">Gửi OTP thử nghiệm</h3>
        <p className="text-xs text-muted-foreground">
          Gửi mã OTP test (123456) tới SĐT để kiểm tra ZNS hoạt động. Người nhận cần có Zalo.
        </p>
        <div className="flex gap-2">
          <Input
            value={testPhone}
            onChange={(e) => setTestPhone(e.target.value)}
            placeholder="Nhập SĐT (VD: 0912345678)"
            className="max-w-[220px]"
          />
          <Button onClick={handleTest} disabled={testing || !testPhone.trim()} variant="outline">
            <Send className="h-4 w-4 mr-1.5" />
            {testing ? "Đang gửi..." : "Gửi test"}
          </Button>
        </div>
      </div>
    </div>
  );
}

function GuideTab() {
  return (
    <div className="space-y-6 max-w-3xl">
      {/* How it works */}
      <Section title="Cách hoạt động" icon={<Info className="h-4 w-4" />}>
        <ol className="list-decimal list-inside space-y-1.5 text-sm">
          <li>User nhập SĐT &rarr; Backend gọi Zalo ZNS API gửi OTP qua Zalo</li>
          <li>Zalo gửi tin nhắn ZNS tới SĐT (cần có Zalo, nếu không &rarr; fallback mock)</li>
          <li>User nhập mã OTP trong app &rarr; xác thực thành công</li>
        </ol>
        <div className="mt-3 rounded bg-blue-50 p-3 text-xs text-blue-800 space-y-1">
          <p><strong>Chi phí:</strong> ~200-300 VND/tin ZNS</p>
          <p><strong>Provider hiện tại:</strong> <code>zalo+mock</code> (Zalo trước, mock fallback nếu lỗi)</p>
          <p><strong>Refresh token:</strong> Tự động xoay vòng qua Redis, không cần thay đổi thủ công nếu hệ thống chạy liên tục</p>
        </div>
      </Section>

      {/* Token lifecycle */}
      <Section title="Vòng đời Token">
        <div className="text-sm space-y-2">
          <div className="rounded border p-3 bg-gray-50">
            <p className="font-mono text-xs mb-2">Refresh Token (90 ngày) &rarr; Access Token (~1 giờ) &rarr; Gửi ZNS</p>
            <ol className="list-decimal list-inside space-y-1 text-xs text-muted-foreground">
              <li>Mỗi lần gửi OTP, nếu Access Token hết hạn &rarr; dùng Refresh Token lấy Access Token mới</li>
              <li>Zalo trả về Access Token mới + Refresh Token mới</li>
              <li>Refresh Token mới tự động lưu vào Redis (key: <code>zalo:refresh_token</code>)</li>
              <li>Lần sau container restart &rarr; load Refresh Token từ Redis (không dùng token cũ trong .env)</li>
            </ol>
          </div>
        </div>
      </Section>

      {/* When to intervene */}
      <Section title="Khi nào cần can thiệp thủ công">
        <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left py-2 px-3 font-medium">Tình huống</th>
                <th className="text-left py-2 px-3 font-medium">Nguyên nhân</th>
                <th className="text-left py-2 px-3 font-medium">Cách xử lý</th>
              </tr>
            </thead>
            <tbody className="text-xs">
              <tr className="border-b">
                <td className="py-2 px-3">Redis mất data</td>
                <td className="py-2 px-3">Redis crash / mất volume</td>
                <td className="py-2 px-3">Lấy Refresh Token mới từ Zalo Console &rarr; dán vào tab Cấu hình</td>
              </tr>
              <tr className="border-b">
                <td className="py-2 px-3">Token hết hạn</td>
                <td className="py-2 px-3">Không có OTP nào gửi trong ~90 ngày</td>
                <td className="py-2 px-3">Tương tự: lấy token mới từ Console</td>
              </tr>
              <tr className="border-b">
                <td className="py-2 px-3">Zalo thu hồi token</td>
                <td className="py-2 px-3">Vi phạm policy / đổi App Secret</td>
                <td className="py-2 px-3">Tạo lại Refresh Token từ Console</td>
              </tr>
              <tr className="border-b">
                <td className="py-2 px-3">Đổi App ID / Secret</td>
                <td className="py-2 px-3">Tạo app Zalo mới</td>
                <td className="py-2 px-3">Cập nhật <code>.env.backend</code> &rarr; deploy lại backend</td>
              </tr>
              <tr>
                <td className="py-2 px-3">Đổi Template ID</td>
                <td className="py-2 px-3">Tạo template ZNS mới</td>
                <td className="py-2 px-3">Cập nhật <code>.env.backend</code> &rarr; deploy lại backend</td>
              </tr>
            </tbody>
          </table>
        </div>
      </Section>

      {/* Error codes */}
      <Section title="Mã lỗi Zalo ZNS">
        <div className="overflow-x-auto">
          <table className="w-full text-sm border-collapse">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left py-2 px-3 font-medium w-24">Mã lỗi</th>
                <th className="text-left py-2 px-3 font-medium">Mô tả</th>
                <th className="text-left py-2 px-3 font-medium">Cách xử lý</th>
              </tr>
            </thead>
            <tbody className="text-xs">
              <ErrorRow code="-14001" desc="App ID không hợp lệ" fix="Kiểm tra ZALO_APP_ID trong .env.backend" />
              <ErrorRow code="-14002" desc="App Secret không hợp lệ" fix="Kiểm tra ZALO_APP_SECRET trong .env.backend" />
              <ErrorRow code="-14004" desc="App chưa xác thực hoặc bị khóa" fix="Vào Zalo Developer Console kiểm tra trạng thái app" />
              <ErrorRow code="-14008" desc="Template ID không tồn tại" fix="Kiểm tra ZALO_ZNS_TEMPLATE_ID, tạo lại template nếu cần" />
              <ErrorRow code="-14014" desc="Refresh Token không hợp lệ / hết hạn" fix="Lấy Refresh Token mới từ Console → dán vào tab Cấu hình" />
              <ErrorRow code="-14015" desc="Access Token hết hạn" fix="Tự động refresh, nếu vẫn lỗi → cập nhật Refresh Token" />
              <ErrorRow code="-14016" desc="Không có quyền gửi ZNS" fix="Vào Console bật quyền ZNS cho app" />
              <ErrorRow code="-14029" desc="Người nhận không có Zalo" fix="Không cần xử lý — hệ thống tự fallback sang mock" />
              <ErrorRow code="-14032" desc="Vượt quá quota gửi tin" fix="Đợi reset quota hoặc nâng cấp gói Zalo OA" />
              <ErrorRow code="-14041" desc="Template chưa được duyệt" fix="Vào Console kiểm tra trạng thái duyệt template" />
              <ErrorRow code="-14050" desc="SĐT không hợp lệ" fix="Kiểm tra format SĐT (0xxx hoặc 84xxx)" />
            </tbody>
          </table>
        </div>
      </Section>

      {/* Steps to get new refresh token */}
      <Section title="Cách lấy Refresh Token mới">
        <ol className="list-decimal list-inside space-y-2 text-sm">
          <li>
            Truy cập <a href="https://developers.zalo.me/app" target="_blank" rel="noopener" className="text-blue-600 underline">Zalo Developer Console</a> &rarr; đăng nhập
          </li>
          <li>Chọn app <strong>Sàn Giá Gạo</strong> (ID: 1300786818201714060)</li>
          <li>Vào mục <strong>Cài đặt</strong> hoặc <strong>API Explorer</strong></li>
          <li>Tìm phần <strong>OAuth</strong> &rarr; tạo Authorization Code &rarr; đổi lấy Refresh Token</li>
          <li>Copy Refresh Token &rarr; quay lại tab <strong>Cấu hình</strong> &rarr; dán vào ô &quot;Refresh Token mới&quot; &rarr; bấm <strong>Lưu</strong></li>
          <li>Bấm <strong>Gửi test</strong> để kiểm tra hoạt động</li>
        </ol>
      </Section>

      {/* Env vars reference */}
      <Section title="Biến môi trường (.env.backend)">
        <div className="space-y-2 text-xs font-mono">
          <EnvLine name="SMS_PROVIDER" value="zalo+mock" desc="zalo = chỉ Zalo, zalo+mock = Zalo + fallback, mock = chỉ mock" />
          <EnvLine name="ZALO_APP_ID" value="1300786818201714060" desc="App ID từ Zalo Console" />
          <EnvLine name="ZALO_APP_SECRET" value="IYXk44BOi..." desc="App Secret từ Zalo Console" />
          <EnvLine name="ZALO_ZNS_TEMPLATE_ID" value="558826" desc="Template ID đã duyệt" />
          <EnvLine name="ZALO_REFRESH_TOKEN" value="hmWj5_ysL..." desc="Refresh Token ban đầu (Redis sẽ ghi đè)" />
        </div>
        <p className="text-xs text-muted-foreground mt-2">
          Sau khi thay đổi .env.backend, cần deploy lại: <code className="bg-muted px-1 rounded">bash infras/scripts/quick-deploy.sh backend</code>
        </p>
      </Section>
    </div>
  );
}

function Section({ title, icon, children }: { title: string; icon?: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border p-4 space-y-3">
      <h3 className="font-semibold flex items-center gap-2">{icon}{title}</h3>
      {children}
    </div>
  );
}

function ErrorRow({ code, desc, fix }: { code: string; desc: string; fix: string }) {
  return (
    <tr className="border-b last:border-0">
      <td className="py-2 px-3 font-mono text-red-600">{code}</td>
      <td className="py-2 px-3">{desc}</td>
      <td className="py-2 px-3 text-muted-foreground">{fix}</td>
    </tr>
  );
}

function EnvLine({ name, value, desc }: { name: string; value: string; desc: string }) {
  return (
    <div className="flex items-start gap-2 rounded bg-muted/50 p-2">
      <div className="flex-1 min-w-0">
        <span className="text-blue-600">{name}</span>=<span className="text-green-600">{value}</span>
        <p className="text-[10px] text-muted-foreground mt-0.5 font-sans">{desc}</p>
      </div>
      <button
        onClick={() => { navigator.clipboard.writeText(`${name}=${value}`); toast.success("Đã copy"); }}
        className="shrink-0 p-1 rounded hover:bg-muted transition-colors"
        title="Copy"
      >
        <Copy className="h-3 w-3 text-muted-foreground" />
      </button>
    </div>
  );
}
