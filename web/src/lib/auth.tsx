"use client";

import { createContext, useContext, useEffect, useState, useCallback, type ReactNode } from "react";
import { getMyPermissions, getGuestPermissions, type PermissionMap } from "@/services/api";

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
  permissions: PermissionMap;
  hasPermission: (key: string) => boolean;
  login: (user: AuthUser, accessToken: string, refreshToken: string) => void;
  logout: () => void;
  isLoading: boolean;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [permissions, setPermissions] = useState<PermissionMap>({});
  const [isLoading, setIsLoading] = useState(true);

  // Fetch permissions and refresh every 60s so admin changes apply quickly
  useEffect(() => {
    function fetchPermissions() {
      if (!user) {
        getGuestPermissions()
          .then((res) => setPermissions(res.permissions))
          .catch(() => setPermissions({}));
      } else {
        getMyPermissions()
          .then((res) => setPermissions(res.permissions))
          .catch(() => setPermissions({}));
      }
    }
    fetchPermissions();
    const interval = setInterval(fetchPermissions, 60000);
    return () => clearInterval(interval);
  }, [user]);

  useEffect(() => {
    // Migrate: clean up old token storage
    localStorage.removeItem("web_token");
    localStorage.removeItem("web_refresh_token");

    const savedUser = localStorage.getItem("web_user");
    if (savedUser) {
      try {
        setUser(JSON.parse(savedUser));
      } catch {
        localStorage.removeItem("web_user");
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(
    (user: AuthUser, _accessToken: string, _refreshToken: string) => {
      // Tokens are now stored in httpOnly cookies by the backend.
      // We only store user info in localStorage for UI state.
      setUser(user);
      localStorage.setItem("web_user", JSON.stringify(user));
    },
    []
  );

  const logout = useCallback(() => {
    // Call backend to clear httpOnly cookies
    fetch(`${API_BASE}/auth/logout`, { method: "POST", credentials: "include" }).catch(() => {});
    localStorage.removeItem("web_user");
    setUser(null);
    setPermissions({});
    window.location.href = "/dang-nhap";
  }, []);

  const hasPermission = useCallback(
    (key: string) => {
      // Owner always has all permissions
      if (user?.role === "owner") return true;
      // Check permission matrix (works for both logged-in users and guests)
      return permissions[key] === true;
    },
    [user, permissions]
  );

  // token is kept in the interface for backward compatibility but always null
  // (auth is handled via httpOnly cookies now)
  return (
    <AuthContext.Provider value={{ user, token: null, permissions, hasPermission, login, logout, isLoading }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}

export function clearAuth() {
  fetch(`${API_BASE}/auth/logout`, { method: "POST", credentials: "include" }).catch(() => {});
  localStorage.removeItem("web_user");
  window.location.href = "/dang-nhap";
}
