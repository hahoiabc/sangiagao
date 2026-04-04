"use client";

import { useState } from "react";
import Image from "next/image";
import { toThumbnailUrl } from "@/services/api";

interface ListingImageProps {
  src: string;
  alt: string;
  fill?: boolean;
  width?: number;
  height?: number;
  sizes?: string;
  className?: string;
  priority?: boolean;
}

/** Hiển thị thumbnail, fallback sang ảnh gốc nếu thumbnail 404 */
export function ListingImage({ src, alt, fill, width, height, sizes, className, priority }: ListingImageProps) {
  const thumbUrl = toThumbnailUrl(src);
  const [imgSrc, setImgSrc] = useState(thumbUrl);

  return (
    <Image
      src={imgSrc}
      alt={alt}
      fill={fill}
      width={width}
      height={height}
      sizes={sizes}
      className={className}
      priority={priority}
      onError={() => {
        if (imgSrc === thumbUrl && thumbUrl !== src) {
          setImgSrc(src); // fallback to original
        }
      }}
      unoptimized={imgSrc !== thumbUrl} // original may be external, skip optimization
    />
  );
}
