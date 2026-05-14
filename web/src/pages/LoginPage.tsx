import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useLogin } from "@/api/auth";
import { Button } from "@/components/shared/Button";
import { useDocumentTitle } from "@/hooks/useDocumentTitle";

export function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const login = useLogin();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  useDocumentTitle(t("auth.login"));

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    login.mutate(
      { username, password },
      {
        onSuccess: () => navigate("/tenants"),
        onError: (err) =>
          setError(err instanceof Error ? err.message : t("auth.loginFailed")),
      },
    );
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-950">
      <div className="w-full max-w-sm rounded-xl border border-zinc-800 bg-zinc-900 p-8 shadow-2xl">
        <div className="mb-8 text-center">
          <h1 className="text-xl font-bold text-white">{t("auth.loginTitle")}</h1>
          <p className="mt-1 text-sm text-zinc-400">{t("auth.loginSubtitle")}</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="username" className="mb-1 block text-sm font-medium text-zinc-300">
              {t("auth.username")}
            </label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              autoFocus
              className="w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-white placeholder-zinc-500 outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500/30"
              placeholder={t("auth.usernamePlaceholder")}
            />
          </div>

          <div>
            <label htmlFor="password" className="mb-1 block text-sm font-medium text-zinc-300">
              {t("auth.password")}
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              className="w-full rounded-lg border border-zinc-700 bg-zinc-800 px-3 py-2 text-sm text-white placeholder-zinc-500 outline-none focus:border-teal-500 focus:ring-1 focus:ring-teal-500/30"
              placeholder={t("auth.passwordPlaceholder")}
            />
          </div>

          {error && (
            <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400">
              {error}
            </div>
          )}

          <Button type="submit" className="w-full" disabled={login.isPending}>
            {login.isPending ? t("auth.loggingIn") : t("auth.login")}
          </Button>
        </form>
      </div>
    </div>
  );
}
