"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import {
  authApi,
  clearTokens,
  getAccessToken,
  setTokens,
  type User,
} from "@/services/api";

type AuthContextValue = {
  user: User | null;
  loading: boolean;
  login: (payload: { email: string; password: string; remember?: boolean }) => Promise<void>;
  register: (payload: { name: string; email: string; password: string }) => Promise<void>;
  logout: () => Promise<void>;
  refreshMe: () => Promise<void>;
  isAuthenticated: boolean;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const refreshMe = useCallback(async () => {
    const token = getAccessToken();
    if (!token) {
      setUser(null);
      return;
    }
    const me = await authApi.me();
    setUser(me);
  }, []);

  useEffect(() => {
    let active = true;
    const init = async () => {
      setLoading(true);
      try {
        await refreshMe();
      } catch {
        clearTokens();
        if (active) {
          setUser(null);
        }
      } finally {
        if (active) {
          setLoading(false);
        }
      }
    };
    void init();
    return () => {
      active = false;
    };
  }, [refreshMe]);

  const login = useCallback(async (payload: { email: string; password: string; remember?: boolean }) => {
    const data = await authApi.login(payload);
    setTokens(data.accessToken, data.refreshToken);
    setUser(data.user);
  }, []);

  const register = useCallback(
    async (payload: { name: string; email: string; password: string }) => {
      await authApi.register(payload);
      await login({ email: payload.email, password: payload.password });
    },
    [login],
  );

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      clearTokens();
      setUser(null);
    }
  }, []);

  const value = useMemo(
    () => ({
      user,
      loading,
      login,
      register,
      logout,
      refreshMe,
      isAuthenticated: Boolean(user),
    }),
    [user, loading, login, register, logout, refreshMe],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuthContext must be used inside AuthProvider");
  }
  return context;
}
