import { useEffect, useState } from "react";
import { CloudSun, KeyRound } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../features/auth/AuthContext";

export function LoginPage() {
  const { login, token, user, isAuthReady } = useAuth();
  const navigate = useNavigate();
  const [loginName, setLoginName] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (isAuthReady && token && user) {
      navigate("/", { replace: true });
    }
  }, [isAuthReady, navigate, token, user]);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);
    try {
      await login({ login: loginName, password });
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <main className="login-page">
      <section className="login-visual" aria-hidden="true">
        <div className="weather-scene">
          <CloudSun size={72} />
          <div className="radar-ring ring-one" />
          <div className="radar-ring ring-two" />
          <div className="radar-line" />
          <div className="station-dot dot-one" />
          <div className="station-dot dot-two" />
          <div className="station-dot dot-three" />
        </div>
      </section>
      <section className="login-panel">
        <div className="login-card">
          <div className="login-title">
            <div className="brand-mark">
              <KeyRound size={22} />
            </div>
            <div>
              <h1>Sign in</h1>
              <p>Enter your login and password to access analytics.</p>
            </div>
          </div>
          <form onSubmit={onSubmit}>
            <label>
              Login
              <input
                value={loginName}
                onChange={(event) => setLoginName(event.target.value)}
                autoComplete="username"
                required
              />
            </label>
            <label>
              Password
              <input
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                type="password"
                autoComplete="current-password"
                required
              />
            </label>
            <button type="submit" className="primary-button" disabled={isLoading}>
              {isLoading ? "Signing in..." : "Sign in"}
            </button>
          </form>
        </div>
      </section>
    </main>
  );
}
