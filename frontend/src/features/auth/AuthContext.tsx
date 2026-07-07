import { createContext, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../shared/api/endpoints";
import type { LoginRequest, User, UserRole } from "../../entities/types";

interface AuthContextValue {
  user: User | null;
  token: string | null;
  login: (payload: LoginRequest) => Promise<void>;
  logout: () => void;
  hasRole: (roles: UserRole[]) => boolean;
}

const tokenKey = "weather_accuracy_token";
const userKey = "weather_accuracy_user";

const AuthContext = createContext<AuthContextValue | null>(null);

function loadUser() {
  const raw = localStorage.getItem(userKey);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as User;
  } catch {
    localStorage.removeItem(userKey);
    return null;
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(tokenKey));
  const [user, setUser] = useState<User | null>(() => loadUser());
  const navigate = useNavigate();

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      login: async (payload) => {
        const result = await api.login(payload);
        localStorage.setItem(tokenKey, result.token);
        localStorage.setItem(userKey, JSON.stringify(result.user));
        setToken(result.token);
        setUser(result.user);
        navigate("/", { replace: true });
      },
      logout: () => {
        localStorage.removeItem(tokenKey);
        localStorage.removeItem(userKey);
        setToken(null);
        setUser(null);
        navigate("/login", { replace: true });
      },
      hasRole: (roles) => Boolean(user && roles.includes(user.role))
    }),
    [navigate, token, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used inside AuthProvider");
  }
  return context;
}
