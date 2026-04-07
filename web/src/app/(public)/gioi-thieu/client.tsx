"use client";

import { useState, useEffect } from "react";
import { Wheat, Users, MessageCircle, ShieldCheck, BarChart3, Smartphone } from "lucide-react";
import { getAboutPage } from "@/services/api";

interface Feature { title: string; desc: string; }
interface Target { label: string; desc: string; }
interface AboutContent {
  hero_title: string;
  hero_desc: string;
  problem_title: string;
  problem_desc: string;
  features: Feature[];
  targets: Target[];
  cta_title: string;
  cta_desc: string;
  contact_phone: string;
}

const featureIcons = [BarChart3, Wheat, MessageCircle, Users, ShieldCheck, Smartphone];

const defaultContent: AboutContent = {
  hero_title: "Sàn Giá Gạo",
  hero_desc: "Nền tảng kết nối trực tiếp người mua và người bán gạo trên toàn quốc.\nMinh bạch giá cả — Giao dịch nhanh chóng — Không trung gian.",
  problem_title: "Tại sao có Sàn Giá Gạo?",
  problem_desc: "Thị trường gạo Việt Nam lâu nay phụ thuộc vào nhiều tầng trung gian, khiến nông dân bán giá thấp trong khi người mua phải trả giá cao. Thông tin giá cả không minh bạch, người bán và người mua khó tìm đến nhau. Sàn Giá Gạo ra đời để giải quyết vấn đề đó — tạo một nơi mà ai cũng có thể đăng giá, tìm nguồn, và liên hệ trực tiếp mà không mất phí trung gian.",
  features: [
    { title: "Bảng giá gạo", desc: "Cập nhật giá các loại gạo theo thời gian thực, giúp bạn nắm bắt thị trường nhanh chóng" },
    { title: "Đăng tin mua/bán", desc: "Đăng tin nhanh kèm hình ảnh, giá, số lượng — tiếp cận người mua/bán trên toàn quốc" },
    { title: "Chat trực tiếp", desc: "Nhắn tin, gửi hình ảnh, tin nhắn thoại ngay trên ứng dụng — không cần trao đổi qua kênh khác" },
    { title: "Kết nối trực tiếp", desc: "Số điện thoại và địa chỉ người bán hiển thị công khai, dễ dàng liên hệ và giao dịch" },
    { title: "An toàn & Minh bạch", desc: "Thông tin được mã hóa, kết nối HTTPS. Hệ thống đánh giá và báo cáo giúp cộng đồng uy tín hơn" },
    { title: "Đa nền tảng", desc: "Sử dụng trên điện thoại (Android & iOS) hoặc máy tính qua website sangiagao.vn" },
  ],
  targets: [
    { label: "Nông dân", desc: "Bán gạo trực tiếp, không qua trung gian" },
    { label: "Thương lái & Đại lý", desc: "Tìm nguồn gạo đa dạng, giá cập nhật" },
    { label: "Nhà máy xay xát", desc: "Kết nối nguồn nguyên liệu ổn định" },
    { label: "Doanh nghiệp xuất khẩu", desc: "Tiếp cận nguồn cung trên toàn quốc" },
  ],
  cta_title: "Bắt đầu ngay hôm nay",
  cta_desc: "Tạo tài khoản miễn phí và khám phá thị trường gạo trên toàn quốc",
  contact_phone: "0968 660 799",
};

export function AboutPageClient() {
  const [content, setContent] = useState<AboutContent>(defaultContent);

  useEffect(() => {
    getAboutPage()
      .then((res) => {
        if (res.value) {
          try {
            setContent(JSON.parse(res.value) as AboutContent);
          } catch { /* use defaults */ }
        }
      })
      .catch(() => { /* use defaults */ });
  }, []);

  return (
    <div className="mx-auto max-w-4xl px-4 py-10">
      {/* Hero */}
      <div className="text-center mb-12">
        <div className="flex justify-center mb-4">
          <div className="rounded-full bg-primary/10 p-4">
            <Wheat className="h-12 w-12 text-primary" />
          </div>
        </div>
        <h1 className="text-3xl font-bold mb-3">{content.hero_title}</h1>
        <p className="text-lg text-muted-foreground max-w-2xl mx-auto whitespace-pre-line">
          {content.hero_desc}
        </p>
      </div>

      {/* Vấn đề & Giải pháp */}
      <section className="mb-12">
        <div className="rounded-lg border bg-muted/30 p-6">
          <h2 className="text-xl font-semibold mb-3">{content.problem_title}</h2>
          <p className="text-muted-foreground leading-relaxed whitespace-pre-line">{content.problem_desc}</p>
        </div>
      </section>

      {/* Tính năng */}
      {content.features.length > 0 && (
        <section className="mb-12">
          <h2 className="text-xl font-semibold mb-6 text-center">Tính năng chính</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {content.features.map((f, i) => {
              const Icon = featureIcons[i % featureIcons.length];
              return (
                <div key={i} className="rounded-lg border p-4 hover:shadow-sm transition-shadow">
                  <Icon className="h-8 w-8 text-primary mb-3" />
                  <h3 className="font-semibold mb-1">{f.title}</h3>
                  <p className="text-sm text-muted-foreground">{f.desc}</p>
                </div>
              );
            })}
          </div>
        </section>
      )}

      {/* Đối tượng */}
      {content.targets.length > 0 && (
        <section className="mb-12">
          <h2 className="text-xl font-semibold mb-6 text-center">Dành cho ai?</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {content.targets.map((t, i) => (
              <div key={i} className="flex items-start gap-3 rounded-lg border p-4">
                <div className="rounded-full bg-primary/10 p-2 mt-0.5">
                  <Users className="h-4 w-4 text-primary" />
                </div>
                <div>
                  <h3 className="font-semibold">{t.label}</h3>
                  <p className="text-sm text-muted-foreground">{t.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      {/* CTA */}
      <section className="text-center rounded-lg border bg-primary/5 p-8">
        <h2 className="text-xl font-semibold mb-2">{content.cta_title}</h2>
        <p className="text-muted-foreground mb-4">{content.cta_desc}</p>
        <div className="flex flex-wrap justify-center gap-3">
          <a
            href="/dang-ky"
            className="inline-flex items-center justify-center rounded-md bg-primary px-6 py-2.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            Đăng ký miễn phí
          </a>
          <a
            href="/bang-gia"
            className="inline-flex items-center justify-center rounded-md border px-6 py-2.5 text-sm font-medium hover:bg-muted transition-colors"
          >
            Xem bảng giá
          </a>
        </div>
      </section>

      {/* Liên hệ */}
      <section className="mt-8 text-center text-sm text-muted-foreground">
        <p>
          Liên hệ hỗ trợ: <strong>{content.contact_phone}</strong> | Website: <strong>sangiagao.vn</strong>
        </p>
      </section>
    </div>
  );
}
