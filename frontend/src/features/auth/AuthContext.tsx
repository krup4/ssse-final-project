import { createContext, useContext, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../shared/api/endpoints";
import { apiClient } from "../../shared/api/client";
import type { LoginRequest, User, UserRole } from "../../entities/types";

interface AuthContextValue {
  user: User | null;
  token: string | null;
  isAuthReady: boolean;
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
  const [isAuthReady, setIsAuthReady] = useState(false);
  const navigate = useNavigate();

  function clearAuth() {
    localStorage.removeItem(tokenKey);
    localStorage.removeItem(userKey);
    setToken(null);
    setUser(null);
  }

  useEffect(() => {
    let isMounted = true;

    async function verifyToken() {
      if (!token) {
        clearAuth();
        if (isMounted) {
          setIsAuthReady(true);
        }
        return;
      }

      try {
        const verifiedUser = await api.me();
        if (!isMounted) {
          return;
        }
        localStorage.setItem(userKey, JSON.stringify(verifiedUser));
        setUser(verifiedUser);
      } catch {
        if (!isMounted) {
          return;
        }
        clearAuth();
      } finally {
        if (isMounted) {
          setIsAuthReady(true);
        }
      }
    }

    setIsAuthReady(false);
    void verifyToken();

    return () => {
      isMounted = false;
    };
  }, [token]);

  useEffect(() => {
    const interceptor = apiClient.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error?.response?.status === 401) {
          clearAuth();
          navigate("/login", { replace: true });
        }
        return Promise.reject(error);
      }
    );

    return () => {
      apiClient.interceptors.response.eject(interceptor);
    };
  }, [navigate]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isAuthReady,
      login: async (payload) => {
        const result = await api.login(payload);
        localStorage.setItem(tokenKey, result.token);
        localStorage.setItem(userKey, JSON.stringify(result.user));
        setToken(result.token);
        setUser(result.user);
        navigate("/", { replace: true });
      },
      logout: () => {
        clearAuth();
        navigate("/login", { replace: true });
      },
      hasRole: (roles) => Boolean(token && user && roles.includes(user.role))
    }),
    [isAuthReady, navigate, token, user]
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
