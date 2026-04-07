"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import { getAboutPage, updateAboutPage } from "@/services/api";
import { Save, Plus, Trash2, Eye } from "lucide-react";

interface Feature {
  title: string;
  desc: string;
}

interface Target {
  label: string;
  desc: string;
}

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

export default function AboutPageAdmin() {
  const [content, setContent] = useState<AboutContent>(defaultContent);
  const [savedContent, setSavedContent] = useState<AboutContent>(defaultContent);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const fetchContent = useCallback(async () => {
    try {
      const res = await getAboutPage();
      if (res.value) {
        const parsed = JSON.parse(res.value) as AboutContent;
        setContent(parsed);
        setSavedContent(parsed);
      }
    } catch {
      // Use defaults
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchContent();
  }, [fetchContent]);

  const hasChanges = JSON.stringify(content) !== JSON.stringify(savedContent);

  async function handleSave() {
    setSaving(true);
    try {
      await updateAboutPage(JSON.stringify(content));
      setSavedContent(content);
      toast.success("Đã lưu trang giới thiệu");
    } catch {
      toast.error("Lưu thất bại");
    } finally {
      setSaving(false);
    }
  }

  function updateFeature(index: number, field: keyof Feature, value: string) {
    const updated = [...content.features];
    updated[index] = { ...updated[index], [field]: value };
    setContent({ ...content, features: updated });
  }

  function addFeature() {
    setContent({ ...content, features: [...content.features, { title: "", desc: "" }] });
  }

  function removeFeature(index: number) {
    setContent({ ...content, features: content.features.filter((_, i) => i !== index) });
  }

  function updateTarget(index: number, field: keyof Target, value: string) {
    const updated = [...content.targets];
    updated[index] = { ...updated[index], [field]: value };
    setContent({ ...content, targets: updated });
  }

  function addTarget() {
    setContent({ ...content, targets: [...content.targets, { label: "", desc: "" }] });
  }

  function removeTarget(index: number) {
    setContent({ ...content, targets: content.targets.filter((_, i) => i !== index) });
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-4xl">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Chỉnh sửa trang giới thiệu</h1>
        <div className="flex gap-2">
          <a href="https://sangiagao.vn/gioi-thieu" target="_blank" rel="noreferrer">
            <Button variant="outline" size="sm">
              <Eye className="h-4 w-4 mr-1" /> Xem trang
            </Button>
          </a>
          <Button onClick={handleSave} disabled={!hasChanges || saving} size="sm">
            <Save className="h-4 w-4 mr-1" /> {saving ? "Đang lưu..." : "Lưu thay đổi"}
          </Button>
        </div>
      </div>

      {/* Hero Section */}
      <Card>
        <CardHeader><CardTitle>Phần giới thiệu chính</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label>Tiêu đề</Label>
            <Input value={content.hero_title} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setContent({ ...content, hero_title: e.target.value })} />
          </div>
          <div>
            <Label>Mô tả ngắn</Label>
            <Textarea rows={3} value={content.hero_desc} onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setContent({ ...content, hero_desc: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      {/* Problem Section */}
      <Card>
        <CardHeader><CardTitle>Vấn đề & Giải pháp</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label>Tiêu đề</Label>
            <Input value={content.problem_title} onChange={(e) => setContent({ ...content, problem_title: e.target.value })} />
          </div>
          <div>
            <Label>Nội dung</Label>
            <Textarea rows={5} value={content.problem_desc} onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setContent({ ...content, problem_desc: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      {/* Features */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Tính năng chính</CardTitle>
            <Button variant="outline" size="sm" onClick={addFeature}>
              <Plus className="h-4 w-4 mr-1" /> Thêm
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {content.features.map((f, i) => (
            <div key={i} className="flex gap-3 items-start border rounded-lg p-3">
              <div className="flex-1 space-y-2">
                <Input placeholder="Tiêu đề" value={f.title} onChange={(e) => updateFeature(i, "title", e.target.value)} />
                <Textarea rows={2} placeholder="Mô tả" value={f.desc} onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => updateFeature(i, "desc", e.target.value)} />
              </div>
              <Button variant="ghost" size="icon" className="text-destructive shrink-0" onClick={() => removeFeature(i)}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Targets */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>Đối tượng sử dụng</CardTitle>
            <Button variant="outline" size="sm" onClick={addTarget}>
              <Plus className="h-4 w-4 mr-1" /> Thêm
            </Button>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {content.targets.map((t, i) => (
            <div key={i} className="flex gap-3 items-start border rounded-lg p-3">
              <div className="flex-1 space-y-2">
                <Input placeholder="Tên nhóm" value={t.label} onChange={(e) => updateTarget(i, "label", e.target.value)} />
                <Input placeholder="Mô tả" value={t.desc} onChange={(e) => updateTarget(i, "desc", e.target.value)} />
              </div>
              <Button variant="ghost" size="icon" className="text-destructive shrink-0" onClick={() => removeTarget(i)}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </CardContent>
      </Card>

      {/* CTA Section */}
      <Card>
        <CardHeader><CardTitle>Kêu gọi hành động (CTA)</CardTitle></CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label>Tiêu đề</Label>
            <Input value={content.cta_title} onChange={(e) => setContent({ ...content, cta_title: e.target.value })} />
          </div>
          <div>
            <Label>Mô tả</Label>
            <Input value={content.cta_desc} onChange={(e) => setContent({ ...content, cta_desc: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      {/* Contact */}
      <Card>
        <CardHeader><CardTitle>Liên hệ</CardTitle></CardHeader>
        <CardContent>
          <div>
            <Label>Số điện thoại</Label>
            <Input value={content.contact_phone} onChange={(e) => setContent({ ...content, contact_phone: e.target.value })} />
          </div>
        </CardContent>
      </Card>

      {/* Bottom save */}
      {hasChanges && (
        <div className="sticky bottom-4 flex justify-end">
          <Button onClick={handleSave} disabled={saving} className="shadow-lg">
            <Save className="h-4 w-4 mr-1" /> {saving ? "Đang lưu..." : "Lưu thay đổi"}
          </Button>
        </div>
      )}
    </div>
  );
}
