"use client";

import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/lib/auth";
import { getSlogan, updateSlogan } from "@/services/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { Save, Eye } from "lucide-react";

export default function SloganPage() {
  const { token } = useAuth();
  const [slogan, setSlogan] = useState("");
  const [savedSlogan, setSavedSlogan] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const fetchSlogan = useCallback(async () => {
    try {
      const data = await getSlogan();
      setSlogan(data.value);
      setSavedSlogan(data.value);
    } catch {
      toast.error("Không thể tải slogan");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSlogan();
  }, [fetchSlogan]);

  const handleSave = async () => {
    if (!slogan.trim()) {
      toast.error("Slogan không được để trống");
      return;
    }
    setSaving(true);
    try {
      const data = await updateSlogan(slogan.trim());
      setSavedSlogan(data.value);
      setSlogan(data.value);
      toast.success("Cập nhật slogan thành công");
    } catch {
      toast.error("Cập nhật slogan thất bại");
    } finally {
      setSaving(false);
    }
  };

  const hasChanges = slogan !== savedSlogan;

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
            Nội dung slogan sẽ chạy ngang trên app mobile và trang web, thay thế cho dòng chữ cố định.
            Tối đa 500 ký tự.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
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
          <div className="overflow-hidden rounded-lg bg-gradient-to-r from-indigo-600 to-indigo-500 py-3 px-4">
            <div className="animate-marquee whitespace-nowrap text-white font-medium">
              {slogan || "Nhập slogan để xem trước..."}
              <span className="mx-16 opacity-50">|</span>
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
