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

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Migrate: clean up old token storage
    localStorage.removeItem("admin_token");
    localStorage.removeItem("admin_refresh_token");

    const savedUser = localStorage.getItem("admin_user");
    if (savedUser) {
      try {
        const parsed = JSON.parse(savedUser);
        if (["owner", "admin", "editor"].includes(parsed.role)) {
          setUser(parsed);
        } else {
          localStorage.removeItem("admin_user");
        }
      } catch {
        localStorage.removeItem("admin_user");
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(
    (user: AuthUser, _accessToken: string, _refreshToken: string) => {
      // Tokens are now stored in httpOnly cookies by the backend.
      // We only store user info in localStorage for UI state.
      setUser(user);
      localStorage.setItem("admin_user", JSON.stringify(user));
    },
    []
  );

  const logout = useCallback(() => {
    // Call backend to clear httpOnly cookies
    fetch(`${API_BASE}/auth/logout`, { method: "POST", credentials: "include" }).catch(() => {});
    localStorage.removeItem("admin_user");
    setUser(null);
    window.location.href = "/login";
  }, []);

  // token is kept in the interface for backward compatibility but always null
  // (auth is handled via httpOnly cookies now)
  return (
    <AuthContext.Provider value={{ user, token: null, login, logout, isLoading }}>
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
  fetch(`${API_BASE}/auth/logout`, { method: "POST", credentials: "include" }).catch(() => {});
  localStorage.removeItem("admin_user");
  window.location.href = "/login";
}
