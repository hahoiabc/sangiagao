"use client";

import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/lib/auth";
import { getSlogan, updateSlogan, getSloganColor, updateSloganColor } from "@/services/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { Save, Eye, Palette } from "lucide-react";

const PRESET_COLORS = [
  { label: "Indigo", value: "#4F46E5" },
  { label: "Đỏ", value: "#DC2626" },
  { label: "Xanh lá", value: "#16A34A" },
  { label: "Cam", value: "#EA580C" },
  { label: "Tím", value: "#9333EA" },
  { label: "Hồng", value: "#DB2777" },
  { label: "Xanh dương", value: "#2563EB" },
  { label: "Vàng", value: "#CA8A04" },
  { label: "Trắng", value: "#FFFFFF" },
];

export default function SloganPage() {
  const { token } = useAuth();
  const [slogan, setSlogan] = useState("");
  const [savedSlogan, setSavedSlogan] = useState("");
  const [color, setColor] = useState("#4F46E5");
  const [savedColor, setSavedColor] = useState("#4F46E5");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const [sloganData, colorData] = await Promise.all([getSlogan(), getSloganColor()]);
      setSlogan(sloganData.value);
      setSavedSlogan(sloganData.value);
      setColor(colorData.value);
      setSavedColor(colorData.value);
    } catch {
      toast.error("Không thể tải slogan");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleSave = async () => {
    if (!slogan.trim()) {
      toast.error("Slogan không được để trống");
      return;
    }
    setSaving(true);
    try {
      const promises: Promise<unknown>[] = [];
      if (slogan !== savedSlogan) promises.push(updateSlogan(slogan.trim()).then((d) => { setSavedSlogan(d.value); setSlogan(d.value); }));
      if (color !== savedColor) promises.push(updateSloganColor(color).then((d) => { setSavedColor(d.value); setColor(d.value); }));
      await Promise.all(promises);
      toast.success("Cập nhật thành công");
    } catch {
      toast.error("Cập nhật thất bại");
    } finally {
      setSaving(false);
    }
  };

  const hasChanges = slogan !== savedSlogan || color !== savedColor;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }

  return (
    <div className="p-4 lg:p-6 space-y-6 max-w-2xl">
      <Card>
        <CardHeader>
          <CardTitle>Slogan hiển thị</CardTitle>
          <CardDescription>
            Nội dung slogan sẽ chạy ngang trên app mobile và trang web.
            Tối đa 500 ký tự.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <p className="text-sm font-medium">Nội dung</p>
            <Input
              value={slogan}
              onChange={(e) => setSlogan(e.target.value)}
              placeholder="Nhập slogan..."
              maxLength={500}
              className="text-base"
            />
            <p className="text-xs text-muted-foreground text-right">
              {slogan.length}/500 ký tự
            </p>
          </div>

          <div className="space-y-2">
            <p className="text-sm font-medium flex items-center gap-2">
              <Palette className="h-4 w-4" />
              Màu chữ
            </p>
            <div className="flex items-center gap-3 flex-wrap">
              {PRESET_COLORS.map((c) => (
                <button
                  key={c.value}
                  type="button"
                  onClick={() => setColor(c.value)}
                  className={`w-8 h-8 rounded-full border-2 transition-all ${color === c.value ? "border-primary scale-110 ring-2 ring-primary/30" : "border-muted hover:scale-105"}`}
                  style={{ backgroundColor: c.value }}
                  title={c.label}
                />
              ))}
              <div className="flex items-center gap-2 ml-2">
                <input
                  type="color"
                  value={color}
                  onChange={(e) => setColor(e.target.value)}
                  className="w-8 h-8 rounded cursor-pointer border border-muted"
                />
                <Input
                  value={color}
                  onChange={(e) => setColor(e.target.value)}
                  placeholder="#000000"
                  className="w-28 text-sm font-mono"
                  maxLength={7}
                />
              </div>
            </div>
          </div>

          <Button onClick={handleSave} disabled={saving || !hasChanges} className="gap-2">
            <Save className="h-4 w-4" />
            {saving ? "Đang lưu..." : "Lưu thay đổi"}
          </Button>
        </CardContent>
      </Card>

      {/* Preview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Eye className="h-4 w-4" />
            Xem trước
          </CardTitle>
          <CardDescription>Hiệu ứng chạy ngang như trên app/web</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-hidden rounded-lg bg-gradient-to-r from-gray-800 to-gray-700 py-3 px-4">
            <div className="animate-marquee whitespace-nowrap font-medium" style={{ color }}>
              {slogan || "Nhập slogan để xem trước..."}
              <span className="mx-16">&nbsp;</span>
              {slogan || "Nhập slogan để xem trước..."}
            </div>
          </div>
          <style jsx>{`
            @keyframes marquee {
              0% { transform: translateX(0%); }
              100% { transform: translateX(-50%); }
            }
            .animate-marquee {
              display: inline-block;
              animation: marquee 15s linear infinite;
            }
          `}</style>
        </CardContent>
      </Card>
    </div>
  );
}
