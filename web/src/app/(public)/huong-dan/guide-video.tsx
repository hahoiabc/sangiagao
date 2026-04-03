"use client";

import { useEffect, useState } from "react";
import { getGuideVideo } from "@/services/api";

function toEmbedUrl(url: string): string | null {
  if (!url) return null;
  // youtube.com/watch?v=ID → youtube.com/embed/ID
  const match = url.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]+)/);
  if (match) return `https://www.youtube.com/embed/${match[1]}`;
  return url; // fallback: use as-is (direct embed URL)
}

export default function GuideVideo() {
  const [embedUrl, setEmbedUrl] = useState<string | null>(null);

  useEffect(() => {
    getGuideVideo()
      .then((r) => setEmbedUrl(toEmbedUrl(r.value)))
      .catch(() => {});
  }, []);

  if (!embedUrl) return null;

  return (
    <div className="mb-8 rounded-xl overflow-hidden border">
      <div className="relative w-full" style={{ paddingBottom: "56.25%" }}>
        <iframe
          src={embedUrl}
          title="Hướng dẫn sử dụng SanGiaGao"
          className="absolute inset-0 w-full h-full"
          allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
          allowFullScreen
        />
      </div>
    </div>
  );
}
