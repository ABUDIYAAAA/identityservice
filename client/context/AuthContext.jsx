"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import { useRouter, usePathname } from "next/navigation";
import api from "@/lib/api";

const AuthContext = createContext({
  user: null,
  loading: true,
  login: async () => {},
  logout: async () => {},
  refreshUser: async () => {},
  forgotPassword: async () => {},
  resetPassword: async () => {},
  getInviteDetails: async () => {},
  acceptInvite: async () => {},
});

const PUBLIC_ROUTES = [
  "/login",
  "/forgot-password",
  "/reset-password",
  "/accept-invite",
];

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();
  const pathname = usePathname();

  const fetchUser = useCallback(async () => {
    try {
      const res = await api.get("/users/me");
      if (res.data?.data) {
        setUser(res.data.data);
      } else {
        setUser(null);
      }
    } catch (err) {
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const login = async (email, password) => {
    const res = await api.post("/auth/login", { email, password });
    await fetchUser();
    return res.data;
  };

  const logout = async () => {
    try {
      await api.post("/auth/logout");
    } catch (err) {
      // Ignore logout API errors
    } finally {
      setUser(null);
      router.push("/login");
    }
  };

  const forgotPassword = async (email) => {
    const res = await api.post("/auth/forgot-password", { email });
    return res.data;
  };

  const resetPassword = async (token, newPassword, confirmPassword) => {
    const res = await api.post("/auth/reset-password", {
      token,
      new_password: newPassword,
      confirm_password: confirmPassword || newPassword,
    });
    return res.data;
  };

  const getInviteDetails = async (token) => {
    const res = await api.get(`/auth/invite/${token}`);
    return res.data?.data;
  };

  const acceptInvite = async (token, password, confirmPassword) => {
    const res = await api.post("/auth/invite/accept", {
      token,
      password,
      confirm_password: confirmPassword || password,
    });
    return res.data;
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        refreshUser: fetchUser,
        forgotPassword,
        resetPassword,
        getInviteDetails,
        acceptInvite,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
