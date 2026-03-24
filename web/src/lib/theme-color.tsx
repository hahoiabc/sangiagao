"use client";

import { createContext, useContext, useEffect, useState } from "react";

export interface ThemeOption {
  key: string;
  label: string;
  hex: string;
  hexDark: string;
  hexLight: string;
  /** oklch hue value */
  hue: number;
  /** oklch chroma for primary */
  chroma: number;
  /** primary oklch */
  oklch: string;
}

export const THEME_OPTIONS: ThemeOption[] = [
  { key: "green", label: "Xanh lá", hex: "#2E7D32", hexDark: "#1B5E20", hexLight: "#4CAF50", hue: 145, chroma: 0.17, oklch: "oklch(0.50 0.17 145)" },
  { key: "teal", label: "Xanh ngọc", hex: "#339999", hexDark: "#267373", hexLight: "#4DB3B3", hue: 190, chroma: 0.10, oklch: "oklch(0.58 0.10 190)" },
  { key: "blue", label: "Xanh dương", hex: "#3399FF", hexDark: "#2673BF", hexLight: "#66B3FF", hue: 250, chroma: 0.16, oklch: "oklch(0.63 0.16 250)" },
  { key: "mint", label: "Teal", hex: "#33CC99", hexDark: "#269973", hexLight: "#66D9B3", hue: 170, chroma: 0.14, oklch: "oklch(0.72 0.14 170)" },
  { key: "gray", label: "Xám đậm", hex: "#444444", hexDark: "#333333", hexLight: "#666666", hue: 0, chroma: 0, oklch: "oklch(0.35 0 0)" },
];

const STORAGE_KEY = "sgg_theme_color";
const DEFAULT_KEY = "green";

interface ThemeColorContextType {
  themeKey: string;
  setThemeKey: (key: string) => void;
  currentTheme: ThemeOption;
}

const ThemeColorContext = createContext<ThemeColorContextType>({
  themeKey: DEFAULT_KEY,
  setThemeKey: () => {},
  currentTheme: THEME_OPTIONS[0],
});

export function ThemeColorProvider({ children }: { children: React.ReactNode }) {
  const [themeKey, setThemeKeyState] = useState(DEFAULT_KEY);

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && THEME_OPTIONS.some((t) => t.key === saved)) {
      setThemeKeyState(saved);
    }
  }, []);

  useEffect(() => {
    const theme = THEME_OPTIONS.find((t) => t.key === themeKey) || THEME_OPTIONS[0];
    const root = document.documentElement;
    const h = theme.hue;
    const c = theme.chroma;
    // Scale chroma for subtle tints (background, border, muted, etc.)
    const micro = Math.min(c, 0.005); // very subtle tint
    const tiny = Math.min(c * 0.08, 0.014); // subtle tint
    const small = Math.min(c * 0.06, 0.01); // borders
    const med = Math.min(c * 0.18, 0.03); // muted foreground
    const accent = Math.min(c * 0.18, 0.031); // accent bg
    const secFg = Math.min(c * 0.7, 0.12); // secondary foreground

    // Primary
    root.style.setProperty("--primary", theme.oklch);
    root.style.setProperty("--ring", theme.oklch);

    // Backgrounds & surfaces — tinted with theme hue
    root.style.setProperty("--background", `oklch(0.985 ${micro} ${h})`);
    root.style.setProperty("--foreground", `oklch(0.16 0.02 ${h})`);
    root.style.setProperty("--card-foreground", `oklch(0.16 0.02 ${h})`);
    root.style.setProperty("--popover-foreground", `oklch(0.16 0.02 ${h})`);
    root.style.setProperty("--secondary", `oklch(0.965 ${tiny} ${h})`);
    root.style.setProperty("--secondary-foreground", `oklch(0.33 ${secFg} ${h})`);
    root.style.setProperty("--muted", `oklch(0.965 ${micro} ${h})`);
    root.style.setProperty("--muted-foreground", `oklch(0.50 ${med} ${h})`);
    root.style.setProperty("--accent", `oklch(0.94 ${accent} ${h})`);
    root.style.setProperty("--accent-foreground", `oklch(0.33 ${secFg} ${h})`);
    root.style.setProperty("--border", `oklch(0.91 ${small} ${h})`);
    root.style.setProperty("--input", `oklch(0.91 ${small} ${h})`);
    root.style.setProperty("--chart-1", theme.oklch);
  }, [themeKey]);

  function setThemeKey(key: string) {
    setThemeKeyState(key);
    localStorage.setItem(STORAGE_KEY, key);
  }

  const currentTheme = THEME_OPTIONS.find((t) => t.key === themeKey) || THEME_OPTIONS[0];

  return (
    <ThemeColorContext.Provider value={{ themeKey, setThemeKey, currentTheme }}>
      {children}
    </ThemeColorContext.Provider>
  );
}

export function useThemeColor() {
  return useContext(ThemeColorContext);
}
