"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";

import AuthShell from "@/components/auth-shell";
import { useAuth } from "@/hooks/useAuth";

export default function RegisterPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [verificationToken, setVerificationToken] = useState("");

  return (
    <AuthShell mode="signup" title="Sign up" subtitle="Create your account.">
      <div className="mx-auto w-full max-w-md">
        <h1 className="text-2xl font-semibold">Register</h1>
        <p className="mt-1 text-sm opacity-70">Create your account.</p>
        <form
          className="mt-4 space-y-3"
          onSubmit={async (event) => {
            event.preventDefault();
            setLoading(true);
            setError("");
            try {
              const result = await fetch(
                `${process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"}/api/auth/register`,
                {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ name, email, password }),
                },
              );
              const data = (await result.json()) as {
                verificationToken?: string;
                error?: string;
              };
              if (!result.ok) {
                throw new Error(data.error ?? "Registration failed");
              }
              setVerificationToken(data.verificationToken ?? "");
              await login({ email, password, remember: true });
              router.replace("/dashboard");
            } catch (err) {
              setError(err instanceof Error ? err.message : "Registration failed");
            } finally {
              setLoading(false);
            }
          }}
        >
          <input
            required
            minLength={2}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Name"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <input
            type="email"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="Email"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <input
            type="password"
            required
            minLength={8}
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            placeholder="Password (min 8 chars)"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          {error && <p className="text-sm text-red-500">{error}</p>}
          <button
            disabled={loading}
            className="w-full rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white shadow disabled:opacity-60"
          >
            {loading ? "Creating..." : "Create Account"}
          </button>
        </form>
        <p className="mt-5 text-sm opacity-80">
          Already registered?{" "}
          <Link href="/login" className="text-[var(--primary)]">
            Sign in
          </Link>
        </p>
        {verificationToken && (
          <p className="mt-2 break-all text-xs text-amber-500">
            Email verification token (dev): {verificationToken}
          </p>
        )}
      </div>
    </AuthShell>
  );
}
