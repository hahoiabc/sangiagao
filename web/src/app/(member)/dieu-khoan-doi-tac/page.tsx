"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, CheckCircle2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth";
import {
  getAffTerms,
  acceptAffTerms,
  becomeAffiliate,
  getMe,
  type AffTermsResponse,
} from "@/services/api";
import { toast } from "sonner";

const fmtVND = (n: number) => new Intl.NumberFormat("vi-VN").format(n) + " đ";
const fmtPct = (n: number) => `${Math.round(n * 100)}%`;

export default function AffTermsPage() {
  const { user } = useAuth();
  const [terms, setTerms] = useState<AffTermsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [checked, setChecked] = useState(false);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    getAffTerms()
      .then(setTerms)
      .catch(() => toast.error("Không tải được điều khoản"))
      .finally(() => setLoading(false));
  }, []);

  const isAff = user?.role === "aff";
  const alreadyAccepted = terms?.accepted ?? false;
  const showAccept = !isAff && !alreadyAccepted;

  const handleAccept = async () => {
    if (!checked || !terms) return;
    setSaving(true);
    try {
      await acceptAffTerms(terms.current_version);
      if (!isAff) {
        await becomeAffiliate();
      }
      // Refresh cached user role then go to dashboard
      try {
        const me = await getMe("");
        localStorage.setItem(
          "web_user",
          JSON.stringify({ id: me.id, phone: me.phone, name: me.name, avatar_url: me.avatar_url, role: me.role }),
        );
      } catch {}
      toast.success("Đã kích hoạt vai trò Đối tác Affiliate");
      setTimeout(() => {
        window.location.href = "/gioi-thieu-ban";
      }, 800);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Không thể kích hoạt");
    } finally {
      setSaving(false);
    }
  };

  if (loading || !terms) {
    return <div className="p-8 text-center text-gray-500">Đang tải…</div>;
  }

  const r = terms.rule;

  return (
    <main className="max-w-3xl mx-auto px-4 py-6 space-y-4">
      <Link
        href="/tai-khoan"
        className="inline-flex items-center gap-1 text-sm text-gray-600 hover:text-primary"
      >
        <ArrowLeft className="h-4 w-4" /> Quay lại
      </Link>

      <h1 className="text-2xl font-bold">Điều khoản đối tác Affiliate</h1>
      <p className="text-xs text-gray-500">Phiên bản {terms.current_version}</p>

      {alreadyAccepted && (
        <div className="p-3 rounded bg-green-50 text-green-700 text-sm flex items-center gap-2">
          <CheckCircle2 className="h-4 w-4" />
          Bạn đã đồng ý điều khoản phiên bản hiện hành
          {terms.accepted_at && ` (${new Date(terms.accepted_at).toLocaleDateString("vi-VN")})`}.
        </div>
      )}

      <Card>
        <CardContent className="p-6 space-y-5 text-sm leading-relaxed">
          <Section title="1. Quyền lợi">
            Đối tác (Aff) nhận hoa hồng theo 3 mức dựa trên lần thanh toán của người được giới thiệu (Referee):
            <ul className="list-disc list-inside mt-1 space-y-0.5">
              <li>Lần thanh toán đầu tiên: <strong>{fmtPct(r.stage1_pct)}</strong> doanh thu ròng</li>
              <li>Lần thanh toán thứ 2: <strong>{fmtPct(r.stage2_pct)}</strong> doanh thu ròng</li>
              <li>Từ lần thứ 3 trở đi (vĩnh viễn): <strong>{fmtPct(r.stage3_pct)}</strong> doanh thu ròng</li>
            </ul>
            <p className="mt-2">
              Doanh thu ròng = số tiền Sàn thực nhận sau khi trừ phí nền tảng (Apple 30%, SePay 0%).
            </p>
          </Section>

          <Section title="2. Thanh toán">
            <ul className="list-disc list-inside space-y-0.5">
              <li>Ngưỡng tối thiểu hiện hành: <strong>{fmtVND(r.minimum_payout)}</strong>. Sàn có thể điều chỉnh theo nhu cầu vận hành.</li>
              <li>Thời gian đối soát: <strong>T+45 ngày</strong> sau khi Referee thanh toán.</li>
              <li>Phí chuyển khoản: Đối tác chịu, trừ trực tiếp từ số tiền payout (thực tế từng lần, tuỳ ngân hàng).</li>
              <li>Đối tác phải cập nhật chính xác thông tin tài khoản nhận tiền. Sàn không chịu trách nhiệm nếu chuyển sai do thông tin sai.</li>
            </ul>
          </Section>

          <Section title="3. Thuế thu nhập cá nhân">
            Đối tác tự kê khai và đóng thuế TNCN theo Luật thuế Việt Nam. Sàn không khấu trừ thuế tại nguồn.
            Đối tác chịu hoàn toàn trách nhiệm pháp lý liên quan đến nghĩa vụ thuế.
          </Section>

          <Section title="4. Không hoàn tiền (clawback)">
            Hoa hồng đã ghi nhận và đã thanh toán SẼ KHÔNG bị thu hồi nếu Referee yêu cầu hoàn tiền sau ngày
            Sàn thanh toán cho Đối tác.
            <p className="mt-2">
              Trong thời gian T+45 trước khi thanh toán, nếu Referee được hoàn tiền, hoa hồng tương ứng sẽ bị huỷ.
            </p>
          </Section>

          <Section title="5. Hành vi cấm">
            <ul className="list-disc list-inside space-y-0.5">
              <li>Không tự đăng ký bằng SĐT/email khác để tự nhận hoa hồng (self-referral).</li>
              <li>Không spam, không quảng cáo sai sự thật về Sàn.</li>
              <li>Không hứa hẹn thưởng/quà tặng vượt ngoài chương trình chính thức.</li>
              <li>Vi phạm → tạm khoá tài khoản Đối tác + huỷ hoa hồng chưa thanh toán.</li>
            </ul>
          </Section>

          <Section title="6. Bảo mật người được giới thiệu">
            Thông tin Referee (SĐT, tên) được mask sẵn trong dashboard. Đối tác cam kết không lưu, share, hoặc
            dùng thông tin này cho mục đích khác ngoài chương trình.
          </Section>

          <Section title="7. Thay đổi điều khoản">
            Sàn có thể điều chỉnh % hoa hồng, ngưỡng payout, hoặc thời gian đối soát. Sàn báo trước 30 ngày qua
            app + email trước khi áp dụng. Hoa hồng đã ghi nhận trước khi điều khoản mới có hiệu lực vẫn được tính
            theo điều khoản cũ (snapshot tại thời điểm payment).
          </Section>

          <Section title="8. Chấm dứt chương trình">
            Sàn có thể chấm dứt chương trình Affiliate bất kỳ lúc nào, báo trước 60 ngày. Hoa hồng đã ghi nhận
            đến ngày chấm dứt vẫn được thanh toán đầy đủ.
          </Section>

          <Section title="9. Pháp luật áp dụng">
            Điều khoản này tuân theo Luật Việt Nam. Tranh chấp giải quyết tại Toà án có thẩm quyền nơi đặt
            trụ sở Sàn Giá Gạo.
          </Section>
        </CardContent>
      </Card>

      {showAccept && (
        <Card>
          <CardContent className="p-4 space-y-3">
            <label className="flex items-start gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="mt-1"
                checked={checked}
                onChange={(e) => setChecked(e.target.checked)}
              />
              <span className="text-sm">Tôi đã đọc, hiểu, và đồng ý các điều khoản trên</span>
            </label>
            <Button
              onClick={handleAccept}
              disabled={!checked || saving}
              className="w-full bg-amber-600 hover:bg-amber-700 text-white"
            >
              {saving ? "Đang kích hoạt…" : "Đồng ý & Kích hoạt làm Đối tác"}
            </Button>
          </CardContent>
        </Card>
      )}
    </main>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h2 className="font-bold text-base mb-1">{title}</h2>
      <div className="text-gray-700">{children}</div>
    </div>
  );
}
