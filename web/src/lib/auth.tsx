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
    window.location.href = "/dang-nhap";
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

export function clearAuth() {
  localStorage.removeItem("web_user");
  localStorage.removeItem("web_token");
  localStorage.removeItem("web_refresh_token");
  window.location.href = "/dang-nhap";
}
