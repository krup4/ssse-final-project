import { useState } from "react";
import { CloudSun, KeyRound } from "lucide-react";
import { useAuth } from "../features/auth/AuthContext";

const demoAccounts = [
  "admin@weather.local",
  "analyst@weather.local",
  "operator@weather.local",
  "viewer@weather.local"
];

export function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState(demoAccounts[0]);
  const [password, setPassword] = useState("demo");
  const [isLoading, setIsLoading] = useState(false);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);
    try {
      await login({ email, password });
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
              <p>Use a demo role to inspect dashboards and access controls.</p>
            </div>
          </div>
          <form onSubmit={onSubmit}>
            <label>
              Account
              <select value={email} onChange={(event) => setEmail(event.target.value)}>
                {demoAccounts.map((account) => (
                  <option key={account} value={account}>
                    {account}
                  </option>
                ))}
              </select>
            </label>
            <label>
              Password
              <input value={password} onChange={(event) => setPassword(event.target.value)} type="password" />
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
