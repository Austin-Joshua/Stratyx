"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import AuthShell from "@/components/auth-shell";
import { useAuth } from "@/hooks/useAuth";
import { API_BASE_URL, setTokens } from "@/services/api";

export default function LoginPage() {
  const router = useRouter();
  const { refreshMe } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [remember, setRemember] = useState(true);
  const [challengeToken, setChallengeToken] = useState("");
  const [otpCode, setOtpCode] = useState("");

  return (
    <AuthShell mode="signin" title="Login" subtitle="Sign in to your account.">
      <div className="mx-auto w-full max-w-md">
        <h1 className="text-2xl font-semibold">Login</h1>
        <p className="mt-1 text-sm opacity-70">Sign in to your account.</p>
        <form
          className="mt-4 space-y-3"
          onSubmit={async (event) => {
            event.preventDefault();
            setLoading(true);
            setError("");
            try {
              if (challengeToken) {
                const result = await fetch(`${API_BASE_URL}/api/auth/2fa/complete-login`, {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ challengeToken, code: otpCode }),
                });
                const data = (await result.json()) as { accessToken?: string; refreshToken?: string; error?: string };
                if (!result.ok || !data.accessToken || !data.refreshToken) {
                  throw new Error(data.error ?? "2FA verification failed");
                }
                localStorage.setItem("stratyx_access_token", data.accessToken);
                localStorage.setItem("stratyx_refresh_token", data.refreshToken);
                window.location.href = "/dashboard";
                return;
              }
              const response = await fetch(`${API_BASE_URL}/api/auth/login`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ email, password, remember }),
              });
              const data = (await response.json()) as {
                requires2FA?: boolean;
                challengeToken?: string;
                accessToken?: string;
                refreshToken?: string;
                error?: string;
              };
              if (!response.ok) {
                throw new Error(data.error ?? "Login failed");
              }
              if (data.requires2FA && data.challengeToken) {
                setChallengeToken(data.challengeToken);
                return;
              }
              if (!data.accessToken || !data.refreshToken) {
                throw new Error("Login tokens missing");
              }
              setTokens(data.accessToken, data.refreshToken);
              await refreshMe();
              router.replace("/dashboard");
            } catch (err) {
              setError(err instanceof Error ? err.message : "Login failed");
            } finally {
              setLoading(false);
            }
          }}
        >
          <input
            type="email"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="Email"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          {!challengeToken ? (
            <input
              type="password"
              required
              minLength={8}
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Password"
              className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
            />
          ) : (
            <input
              required
              value={otpCode}
              onChange={(event) => setOtpCode(event.target.value)}
              placeholder="Enter 2FA OTP code"
              className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
            />
          )}
          {error && <p className="text-sm text-red-500">{error}</p>}
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={remember}
              onChange={(event) => setRemember(event.target.checked)}
            />
            Remember me
          </label>
          <button
            disabled={loading}
            className="w-full rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white shadow disabled:opacity-60"
          >
            {loading ? "Signing in..." : challengeToken ? "Verify 2FA" : "Sign In"}
          </button>
        </form>
        <div className="mt-4 grid grid-cols-2 gap-2">
          <a
            href={`${API_BASE_URL}/api/auth/oauth/google`}
            className="rounded-md border border-[var(--border)] px-3 py-2 text-center text-xs"
          >
            Google OAuth
          </a>
          <a
            href={`${API_BASE_URL}/api/auth/oauth/github`}
            className="rounded-md border border-[var(--border)] px-3 py-2 text-center text-xs"
          >
            GitHub OAuth
          </a>
        </div>
        <p className="mt-5 text-sm opacity-80">
          New user?{" "}
          <Link href="/register" className="text-[var(--primary)]">
            Create account
          </Link>
        </p>
        <p className="mt-2 text-sm opacity-80">
          <Link href="/forgot-password" className="text-[var(--primary)]">
            Forgot password?
          </Link>
        </p>
      </div>
    </AuthShell>
  );
}
