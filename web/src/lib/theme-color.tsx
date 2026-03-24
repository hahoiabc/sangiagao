"use client";

import { createContext, useContext, useEffect, useState } from "react";

export interface ThemeOption {
  key: string;
  label: string;
  hex: string;
  oklch: string;
  ring: string;
}

export const THEME_OPTIONS: ThemeOption[] = [
  { key: "green", label: "Xanh lá", hex: "#2E7D32", oklch: "oklch(0.50 0.17 145)", ring: "oklch(0.50 0.17 145)" },
  { key: "teal", label: "Xanh ngọc", hex: "#339999", oklch: "oklch(0.58 0.10 190)", ring: "oklch(0.58 0.10 190)" },
  { key: "blue", label: "Xanh dương", hex: "#3399FF", oklch: "oklch(0.63 0.16 250)", ring: "oklch(0.63 0.16 250)" },
  { key: "mint", label: "Teal", hex: "#33CC99", oklch: "oklch(0.72 0.14 170)", ring: "oklch(0.72 0.14 170)" },
  { key: "gray", label: "Xám đậm", hex: "#444444", oklch: "oklch(0.35 0 0)", ring: "oklch(0.35 0 0)" },
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
    root.style.setProperty("--primary", theme.oklch);
    root.style.setProperty("--ring", theme.ring);
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
