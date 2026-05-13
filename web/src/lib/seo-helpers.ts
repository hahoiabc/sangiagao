// Helper functions for SEO landing pages (bang-gia-gao/...)

/** Vietnamese diacritic-strip map for slug generation client-side. */
const DIACRITIC_MAP: Record<string, string> = {
  à: "a", á: "a", ả: "a", ã: "a", ạ: "a",
  ă: "a", ằ: "a", ắ: "a", ẳ: "a", ẵ: "a", ặ: "a",
  â: "a", ầ: "a", ấ: "a", ẩ: "a", ẫ: "a", ậ: "a",
  è: "e", é: "e", ẻ: "e", ẽ: "e", ẹ: "e",
  ê: "e", ề: "e", ế: "e", ể: "e", ễ: "e", ệ: "e",
  ì: "i", í: "i", ỉ: "i", ĩ: "i", ị: "i",
  ò: "o", ó: "o", ỏ: "o", õ: "o", ọ: "o",
  ô: "o", ồ: "o", ố: "o", ổ: "o", ỗ: "o", ộ: "o",
  ơ: "o", ờ: "o", ớ: "o", ở: "o", ỡ: "o", ợ: "o",
  ù: "u", ú: "u", ủ: "u", ũ: "u", ụ: "u",
  ư: "u", ừ: "u", ứ: "u", ử: "u", ữ: "u", ự: "u",
  ỳ: "y", ý: "y", ỷ: "y", ỹ: "y", ỵ: "y",
  đ: "d",
};

export function stripDiacritics(s: string): string {
  let out = "";
  for (const ch of s) {
    const lower = ch.toLowerCase();
    out += DIACRITIC_MAP[lower] !== undefined
      ? (ch === lower ? DIACRITIC_MAP[lower] : DIACRITIC_MAP[lower].toUpperCase())
      : ch;
  }
  return out;
}

export function slugify(s: string): string {
  return stripDiacritics(s)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");
}

export function formatPriceVND(value: number): string {
  return new Intl.NumberFormat("vi-VN").format(value) + " ₫";
}

export function formatDateVN(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric" });
}
