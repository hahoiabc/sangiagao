"use client";

import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from "react";

interface AuthUser {
  id: string;
  phone: string;
  name?: string;
  avatar_url?: string;
  role: string;
}

interface AuthContextType {
  user: AuthUser | null;
  token: string | null;
  login: (user: AuthUser, accessToken: string, refreshToken: string) => void;
  logout: () => void;
  isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  useEffect(() => {
    const savedUser = localStorage.getItem("admin_user");
    const savedToken = localStorage.getItem("admin_token");
    if (savedUser && savedToken) {
      try {
        const parsed = JSON.parse(savedUser);
        if (["owner", "admin", "editor"].includes(parsed.role)) {
          setUser(parsed);
          setToken(savedToken);
        } else {
          localStorage.removeItem("admin_user");
          localStorage.removeItem("admin_token");
          localStorage.removeItem("admin_refresh_token");
        }
      } catch {
        localStorage.removeItem("admin_user");
        localStorage.removeItem("admin_token");
        localStorage.removeItem("admin_refresh_token");
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(
    (user: AuthUser, accessToken: string, refreshToken: string) => {
      setUser(user);
      setToken(accessToken);
      localStorage.setItem("admin_user", JSON.stringify(user));
      localStorage.setItem("admin_token", accessToken);
      localStorage.setItem("admin_refresh_token", refreshToken);
    },
    []
  );

  const logout = useCallback(() => {
    localStorage.removeItem("admin_user");
    localStorage.removeItem("admin_token");
    localStorage.removeItem("admin_refresh_token");
    setUser(null);
    setToken(null);
    window.location.href = "/login";
  }, []);

  return (
    <AuthContext.Provider value={{ user, token, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

// Standalone function for use outside React components (e.g., api.ts)
export function clearAuth() {
  localStorage.removeItem("admin_user");
  localStorage.removeItem("admin_token");
  localStorage.removeItem("admin_refresh_token");
  window.location.href = "/login";
}
