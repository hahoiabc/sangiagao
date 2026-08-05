"use client";

import { useState, useEffect } from "react";
import {
  Wheat, Users, MessageCircle, ShieldCheck, BarChart3, Smartphone,
  Search, Store, Check, TrendingUp, ArrowRight,
} from "lucide-react";
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

const CAI_APP = "/cai-app";
const PLAY = "https://play.google.com/store/apps/details?id=com.sangiagao.rice_marketplace";
const IOS = "https://apps.apple.com/vn/app/sangiagao-vn/id6761744869";

const defaultContent: AboutContent = {
  hero_title: "Sàn Giá Gạo",
  hero_desc: "Kết nối thương nhân gạo Việt — minh bạch giá, giao dịch trực tiếp, không qua trung gian.",
  problem_title: "Tại sao có Sàn Giá Gạo?",
  problem_desc: "Thị trường gạo Việt Nam lâu nay phụ thuộc vào nhiều tầng trung gian, khiến nông dân bán giá thấp trong khi người mua phải trả giá cao. Thông tin giá cả không minh bạch, người bán và người mua khó tìm đến nhau. Sàn Giá Gạo ra đời để giải quyết vấn đề đó — tạo một nơi ai cũng có thể đăng giá, tìm nguồn, và liên hệ trực tiếp mà không mất phí trung gian.",
  features: [
    { title: "Bảng giá gạo", desc: "Cập nhật giá các loại gạo mỗi ngày, giúp bạn nắm bắt thị trường nhanh chóng." },
    { title: "Đăng tin mua/bán", desc: "Đăng tin nhanh kèm hình ảnh, giá, số lượng — tiếp cận người mua/bán toàn quốc." },
    { title: "Chat trực tiếp", desc: "Nhắn tin, gửi hình ảnh ngay trên ứng dụng — không cần trao đổi qua kênh khác." },
    { title: "Kết nối trực tiếp", desc: "Số điện thoại và địa chỉ người bán hiển thị công khai, dễ liên hệ và giao dịch." },
    { title: "An toàn & Minh bạch", desc: "Thông tin mã hoá, kết nối HTTPS. Hệ thống đánh giá và báo cáo giúp cộng đồng uy tín." },
    { title: "Đa nền tảng", desc: "Dùng trên điện thoại (Android & iOS) hoặc máy tính qua website sangiagao.vn." },
  ],
  targets: [
    { label: "Nông dân", desc: "Bán gạo trực tiếp, không qua trung gian" },
    { label: "Thương lái & Đại lý", desc: "Tìm nguồn gạo đa dạng, giá cập nhật" },
    { label: "Nhà máy xay xát", desc: "Kết nối nguồn nguyên liệu ổn định" },
    { label: "Doanh nghiệp xuất khẩu", desc: "Tiếp cận nguồn cung trên toàn quốc" },
  ],
  cta_title: "Bắt đầu ngay hôm nay",
  cta_desc: "Tạo tài khoản miễn phí và khám phá thị trường gạo trên toàn quốc.",
  contact_phone: "0968 660 799",
};

export function AboutPageClient() {
  const [content, setContent] = useState<AboutContent>(defaultContent);

  useEffect(() => {
    getAboutPage()
      .then((res) => {
        if (res.value) {
          try { setContent(JSON.parse(res.value) as AboutContent); } catch { /* defaults */ }
        }
      })
      .catch(() => { /* defaults */ });
  }, []);

  return (
    <div className="bg-[#faf7ee] text-[#10261a]">
      {/* ================= HERO ================= */}
      <section className="relative overflow-hidden bg-gradient-to-br from-[#14532d] via-[#0f4526] to-[#0c3b20] text-[#fbf7ec]">
        {/* rice-grain decoration */}
        <svg aria-hidden="true" viewBox="0 0 100 200" className="pointer-events-none absolute -right-8 top-6 h-[115%] w-auto opacity-[0.13] rotate-12">
          <g fill="#e4b443">
            <path d="M50 0 C56 40 56 70 50 110 C44 70 44 40 50 0Z" />
            <ellipse cx="50" cy="40" rx="9" ry="19" /><ellipse cx="34" cy="58" rx="8" ry="17" transform="rotate(-28 34 58)" /><ellipse cx="66" cy="58" rx="8" ry="17" transform="rotate(28 66 58)" /><ellipse cx="34" cy="86" rx="8" ry="17" transform="rotate(-28 34 86)" /><ellipse cx="66" cy="86" rx="8" ry="17" transform="rotate(28 66 86)" /><ellipse cx="34" cy="114" rx="8" ry="16" transform="rotate(-28 34 114)" /><ellipse cx="66" cy="114" rx="8" ry="16" transform="rotate(28 66 114)" />
          </g>
        </svg>

        <div className="relative mx-auto max-w-5xl px-4 py-14 sm:py-20">
          <div className="mb-7 flex items-center gap-3">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src="/logo-sgg.png" alt="Sàn Giá Gạo" width={44} height={44} className="h-11 w-11 rounded-xl bg-white shadow-md" />
            <span className="font-serif text-xl font-bold tracking-tight">Sàn Giá Gạo</span>
          </div>

          <p className="mb-3 text-sm font-semibold uppercase tracking-[0.22em] text-[#f2d48a]">Sàn giao dịch gạo trực tuyến</p>
          <h1 className="font-serif text-[2.6rem] font-bold leading-[1.03] sm:text-6xl">{content.hero_title}</h1>
          <p className="mt-5 max-w-2xl whitespace-pre-line text-lg leading-relaxed text-[#e9e2cf] sm:text-xl">{content.hero_desc}</p>

          <div className="mt-8 flex flex-wrap gap-3">
            <a href={CAI_APP} className="inline-flex items-center gap-2 rounded-full bg-[#e4b443] px-7 py-3 font-bold text-[#10261a] shadow-lg transition hover:brightness-110">
              <Smartphone className="h-5 w-5" /> Tải app miễn phí
            </a>
            <a href="/san-giao-dich" className="inline-flex items-center gap-2 rounded-full border border-[#e4b443]/50 px-7 py-3 font-semibold text-[#fbf7ec] transition hover:bg-white/10">
              Vào sàn giao dịch <ArrowRight className="h-4 w-4" />
            </a>
          </div>
          <p className="mt-5 flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-[#bcd4b6]">
            <span className="inline-flex items-center gap-1"><Check className="h-4 w-4 text-[#e4b443]" /> Android &amp; iOS</span>
            <span className="inline-flex items-center gap-1"><Check className="h-4 w-4 text-[#e4b443]" /> Miễn phí đăng ký</span>
            <span className="inline-flex items-center gap-1"><Check className="h-4 w-4 text-[#e4b443]" /> Không qua trung gian</span>
          </p>
        </div>
      </section>

      <div className="mx-auto max-w-5xl px-4">
        {/* ================= PROBLEM ================= */}
        <section className="py-12 sm:py-16">
          <div className="rounded-2xl border-l-4 border-[#e4b443] bg-white p-6 shadow-sm sm:p-8">
            <h2 className="mb-3 font-serif text-2xl font-bold text-[#14532d]">{content.problem_title}</h2>
            <p className="whitespace-pre-line leading-relaxed text-[#3f4c41]">{content.problem_desc}</p>
          </div>
        </section>

        {/* ================= BUYER vs SELLER ================= */}
        <section className="pb-12 sm:pb-16">
          <h2 className="mb-2 text-center font-serif text-3xl font-bold text-[#14532d]">Sàn cho cả hai đầu</h2>
          <p className="mb-8 text-center text-[#5c6b5c]">Dù bạn mua hay bán gạo — Sàn Giá Gạo đều rút ngắn khoảng cách.</p>
          <div className="grid gap-5 md:grid-cols-2">
            {/* Buyer */}
            <div className="rounded-2xl border border-[#d9e3d3] bg-white p-6 shadow-sm">
              <div className="mb-4 flex items-center gap-3">
                <span className="grid h-11 w-11 place-items-center rounded-xl bg-[#14532d]/10 text-[#14532d]"><Search className="h-6 w-6" /></span>
                <h3 className="font-serif text-xl font-bold text-[#14532d]">Cho người mua</h3>
              </div>
              <ul className="space-y-2.5 text-[#3f4c41]">
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#14532d]" /> Giá thật, minh bạch — không lo bị đội giá.</li>
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#14532d]" /> Chat trực tiếp người bán, xem tận nơi.</li>
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#14532d]" /> Lọc theo tỉnh &amp; loại gạo, tìm nguồn nhanh.</li>
              </ul>
              <a href="/san-giao-dich" className="mt-5 inline-flex items-center gap-1 font-semibold text-[#14532d] hover:underline">Vào sàn tìm gạo <ArrowRight className="h-4 w-4" /></a>
            </div>
            {/* Seller */}
            <div className="rounded-2xl border border-[#e7cf8f] bg-gradient-to-br from-[#fbf3dc] to-[#f6e9c4] p-6 shadow-sm">
              <div className="mb-4 flex items-center gap-3">
                <span className="grid h-11 w-11 place-items-center rounded-xl bg-[#14532d] text-[#f2d48a]"><Store className="h-6 w-6" /></span>
                <h3 className="font-serif text-xl font-bold text-[#14532d]">Cho người bán</h3>
              </div>
              <ul className="space-y-2.5 text-[#3f3211]">
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#b8860b]" /> Đăng tin <b>MIỄN PHÍ</b> — dùng thử 30 ngày.</li>
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#b8860b]" /> Tiếp cận hàng ngàn người mua toàn quốc.</li>
                <li className="flex gap-2"><Check className="mt-0.5 h-5 w-5 flex-none text-[#b8860b]" /> Chốt giá trực tiếp, không mất phí trung gian.</li>
              </ul>
              <a href="/dang-ky" className="mt-5 inline-flex items-center gap-1 font-semibold text-[#14532d] hover:underline">Đăng tin ngay <ArrowRight className="h-4 w-4" /></a>
            </div>
          </div>
        </section>

        {/* ================= FEATURES ================= */}
        {content.features.length > 0 && (
          <section className="pb-12 sm:pb-16">
            <h2 className="mb-8 text-center font-serif text-3xl font-bold text-[#14532d]">Tính năng chính</h2>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {content.features.map((f, i) => {
                const Icon = featureIcons[i % featureIcons.length];
                return (
                  <div key={i} className="rounded-2xl border border-[#e3ddc9] bg-white p-5 shadow-sm transition hover:shadow-md">
                    <span className="mb-3 grid h-11 w-11 place-items-center rounded-xl bg-[#e4b443]/15 text-[#b8860b]"><Icon className="h-6 w-6" /></span>
                    <h3 className="mb-1 font-bold text-[#14532d]">{f.title}</h3>
                    <p className="text-sm leading-relaxed text-[#5c6b5c]">{f.desc}</p>
                  </div>
                );
              })}
            </div>
          </section>
        )}

        {/* ================= TARGETS ================= */}
        {content.targets.length > 0 && (
          <section className="pb-12 sm:pb-16">
            <h2 className="mb-6 text-center font-serif text-2xl font-bold text-[#14532d]">Dành cho ai?</h2>
            <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
              {content.targets.map((t, i) => (
                <div key={i} className="rounded-xl border border-[#e3ddc9] bg-white px-4 py-4 text-center shadow-sm">
                  <div className="mx-auto mb-2 grid h-9 w-9 place-items-center rounded-full bg-[#14532d]/10 text-[#14532d]"><Users className="h-5 w-5" /></div>
                  <div className="font-semibold text-[#14532d]">{t.label}</div>
                  <div className="mt-0.5 text-xs text-[#5c6b5c]">{t.desc}</div>
                </div>
              ))}
            </div>
          </section>
        )}
      </div>

      {/* ================= PRICE TEASER (full-bleed gold) ================= */}
      <section className="bg-gradient-to-r from-[#e4b443] to-[#d4a017]">
        <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 px-4 py-8 sm:flex-row">
          <div className="flex items-center gap-3 text-[#10261a]">
            <TrendingUp className="h-8 w-8 flex-none" />
            <p className="font-serif text-xl font-bold sm:text-2xl">Giá gạo minh bạch, cập nhật mỗi ngày</p>
          </div>
          <a href="/bang-gia-gao" className="inline-flex flex-none items-center gap-2 rounded-full bg-[#14532d] px-6 py-3 font-semibold text-[#fbf7ec] transition hover:bg-[#0c3b20]">
            Xem bảng giá hôm nay <ArrowRight className="h-4 w-4" />
          </a>
        </div>
      </section>

      {/* ================= DOWNLOAD + QR ================= */}
      <section className="bg-[#faf7ee]">
        <div className="mx-auto max-w-5xl px-4 py-14 sm:py-18">
          <div className="grid items-center gap-8 rounded-3xl bg-gradient-to-br from-[#14532d] to-[#0c3b20] p-7 text-[#fbf7ec] shadow-lg sm:grid-cols-[1fr_auto] sm:p-10">
            <div>
              <h2 className="font-serif text-3xl font-bold sm:text-4xl">Tải ngay <span className="text-[#e4b443]">Sàn Giá Gạo</span></h2>
              <p className="mt-3 max-w-md text-[#e9e2cf]">Cả chợ gạo Việt trong túi bạn — niêm yết, xem giá, chat mua bán mọi lúc. Quét mã QR hoặc bấm nút để tải.</p>
              <div className="mt-6 flex flex-wrap gap-3">
                <a href={PLAY} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-3 rounded-xl bg-white px-4 py-2.5 text-[#0b1a10] transition hover:brightness-95">
                  <svg viewBox="0 0 24 24" className="h-6 w-6" aria-hidden="true"><path d="M3 2.5 L14 12 L3 21.5 Z" fill="#00c4b3" /><path d="M3 2.5 L14 12 L18 8.2 Z" fill="#ffd400" /><path d="M3 21.5 L14 12 L18 15.8 Z" fill="#f0403a" /><path d="M18 8.2 L21.5 10.4 Q22.6 11.1 21.5 11.9 L18 14 Z" fill="#3bccff" /></svg>
                  <span className="leading-tight"><span className="block text-[10px] font-medium text-[#4a5a4a]">TẢI TRÊN</span><span className="block text-sm font-extrabold">Google Play</span></span>
                </a>
                <a href={IOS} target="_blank" rel="noopener noreferrer" className="inline-flex items-center gap-3 rounded-xl bg-white px-4 py-2.5 text-[#0b1a10] transition hover:brightness-95">
                  <svg viewBox="0 0 24 24" className="h-6 w-6" aria-hidden="true"><path fill="#0b1a10" d="M16.4 12.6c0-2 1.6-3 1.7-3-.9-1.4-2.4-1.5-2.9-1.6-1.2-.1-2.4.7-3 .7-.6 0-1.6-.7-2.6-.7-1.3 0-2.6.8-3.3 2-1.4 2.4-.4 6 1 8 .7 1 1.4 2 2.4 2 1 0 1.3-.6 2.5-.6s1.5.6 2.5.6 1.6-.9 2.3-1.9c.7-1 .9-2 .9-2.1-.1 0-1.9-.7-1.9-2.9ZM14.6 6.3c.5-.7.9-1.6.8-2.5-.8 0-1.7.5-2.3 1.2-.5.6-.9 1.5-.8 2.4.9.1 1.7-.4 2.3-1.1Z" /></svg>
                  <span className="leading-tight"><span className="block text-[10px] font-medium text-[#4a5a4a]">TẢI TRÊN</span><span className="block text-sm font-extrabold">App Store</span></span>
                </a>
              </div>
            </div>
            <div className="mx-auto rounded-2xl bg-white p-3 text-center shadow-md">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src="/qr-cai-app.png" alt="Mã QR tải Sàn Giá Gạo" width={150} height={150} className="h-[150px] w-[150px]" />
              <span className="mt-1 block text-xs font-bold uppercase tracking-wider text-[#10261a]">Quét để tải</span>
            </div>
          </div>
        </div>
      </section>

      {/* ================= FINAL CTA + CONTACT ================= */}
      <section className="bg-[#faf7ee] pb-16">
        <div className="mx-auto max-w-3xl px-4 text-center">
          <h2 className="font-serif text-2xl font-bold text-[#14532d]">{content.cta_title}</h2>
          <p className="mt-2 text-[#5c6b5c]">{content.cta_desc}</p>
          <div className="mt-5 flex flex-wrap justify-center gap-3">
            <a href="/dang-ky" className="rounded-full bg-[#14532d] px-7 py-3 font-semibold text-[#fbf7ec] transition hover:bg-[#0c3b20]">Đăng ký miễn phí</a>
            <a href="/bang-gia-gao" className="rounded-full border border-[#14532d]/30 px-7 py-3 font-semibold text-[#14532d] transition hover:bg-[#14532d]/5">Xem bảng giá</a>
          </div>
          <p className="mt-8 text-sm text-[#5c6b5c]">
            Liên hệ hỗ trợ: <strong className="text-[#14532d]">{content.contact_phone}</strong> · Website: <strong className="text-[#14532d]">sangiagao.vn</strong>
          </p>
        </div>
      </section>
    </div>
  );
}
