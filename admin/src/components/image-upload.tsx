"use client";

import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Upload } from "lucide-react";
import { uploadImage } from "@/services/api";

interface ImageUploadProps {
  folder: "avatars" | "listings";
  onUpload: (url: string) => void;
  maxSizeMB?: number;
  className?: string;
}

export function ImageUpload({ folder, onUpload, maxSizeMB = 5, className }: ImageUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  async function handleChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > maxSizeMB * 1024 * 1024) {
      setError(`File không được vượt quá ${maxSizeMB}MB`);
      return;
    }
    if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) {
      setError("Chỉ chấp nhận file JPEG, PNG hoặc WebP");
      return;
    }

    setError(null);
    if (preview) URL.revokeObjectURL(preview);
    setPreview(URL.createObjectURL(file));
    setUploading(true);

    try {
      const result = await uploadImage(file, folder);
      onUpload(result.url);
    } catch {
      setError("Tải ảnh thất bại. Vui lòng thử lại.");
      if (preview) URL.revokeObjectURL(preview);
      setPreview(null);
    } finally {
      setUploading(false);
    }
  }

  useEffect(() => {
    return () => {
      if (preview) URL.revokeObjectURL(preview);
    };
  }, [preview]);

  return (
    <div className={className}>
      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        onChange={handleChange}
        className="hidden"
      />
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => inputRef.current?.click()}
        disabled={uploading}
      >
        <Upload className="h-4 w-4 mr-2" />
        {uploading ? "Đang tải..." : "Tải ảnh lên"}
      </Button>
      {preview && (
        <div className="mt-2 relative inline-block">
          <img src={preview} alt="Xem trước" className="h-16 w-16 rounded object-cover border" />
        </div>
      )}
      {error && <p className="text-xs text-destructive mt-1">{error}</p>}
    </div>
  );
}
