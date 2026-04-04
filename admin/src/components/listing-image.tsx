"use client";

import { useState } from "react";
import Image from "next/image";

function toThumbUrl(url: string): string {
  const lastSlash = url.lastIndexOf("/");
  if (lastSlash < 0) return url;
  const base = url.substring(0, lastSlash + 1);
  const filename = url.substring(lastSlash + 1);
  const dotIdx = filename.lastIndexOf(".");
  if (dotIdx < 0) return `${base}thumb_${filename}.jpg`;
  const name = filename.substring(0, dotIdx);
  return `${base}thumb_${name}.jpg`;
}

interface Props {
  src: string;
  alt: string;
  fill?: boolean;
  sizes?: string;
  className?: string;
}

export function ListingImage({ src, alt, fill, sizes, className }: Props) {
  const thumbUrl = toThumbUrl(src);
  const [imgSrc, setImgSrc] = useState(thumbUrl);

  return (
    <Image
      src={imgSrc}
      alt={alt}
      fill={fill}
      sizes={sizes}
      className={className}
      onError={() => {
        if (imgSrc === thumbUrl && thumbUrl !== src) {
          setImgSrc(src);
        }
      }}
      unoptimized={imgSrc !== thumbUrl}
    />
  );
}
