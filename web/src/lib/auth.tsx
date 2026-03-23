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

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [permissions, setPermissions] = useState<PermissionMap>({});
  const [isLoading, setIsLoading] = useState(true);

  // Fetch permissions when token is available, or guest permissions when logged out
  useEffect(() => {
    if (!token) {
      getGuestPermissions()
        .then((res) => setPermissions(res.permissions))
        .catch(() => setPermissions({}));
      return;
    }
    getMyPermissions(token)
      .then((res) => setPermissions(res.permissions))
      .catch(() => setPermissions({}));
  }, [token]);

  useEffect(() => {
    const savedUser = localStorage.getItem("web_user");
    const savedToken = localStorage.getItem("web_token");
    if (savedUser && savedToken) {
      try {
        setUser(JSON.parse(savedUser));
        setToken(savedToken);
      } catch {
        localStorage.removeItem("web_user");
        localStorage.removeItem("web_token");
        localStorage.removeItem("web_refresh_token");
      }
    }
    setIsLoading(false);
  }, []);

  const login = useCallback(
    (user: AuthUser, accessToken: string, refreshToken: string) => {
      setUser(user);
      setToken(accessToken);
      localStorage.setItem("web_user", JSON.stringify(user));
      localStorage.setItem("web_token", accessToken);
      localStorage.setItem("web_refresh_token", refreshToken);
    },
    []
  );

  const logout = useCallback(() => {
    localStorage.removeItem("web_user");
    localStorage.removeItem("web_token");
    localStorage.removeItem("web_refresh_token");
    setUser(null);
    setToken(null);
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

  return (
    <AuthContext.Provider value={{ user, token, permissions, hasPermission, login, logout, isLoading }}>
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
  localStorage.removeItem("web_user");
  localStorage.removeItem("web_token");
  localStorage.removeItem("web_refresh_token");
  window.location.href = "/dang-nhap";
}
