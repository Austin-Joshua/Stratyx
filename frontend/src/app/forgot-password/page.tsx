"use client";

import { useState } from "react";

import AuthShell from "@/components/auth-shell";
import { authApi } from "@/services/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [message, setMessage] = useState("");
  const [token, setToken] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [status, setStatus] = useState("");

  return (
    <AuthShell mode="forgot" title="Forgot Password" subtitle="Reset your account access.">
      <div className="mx-auto flex w-full max-w-2xl flex-col gap-4">
      <section className="stratyx-panel p-6">
        <h1 className="text-2xl font-semibold">Forgot Password</h1>
        <p className="mt-1 text-sm opacity-70">Generate and apply a password reset token.</p>
        <div className="mt-4 flex gap-2">
          <input
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            placeholder="Account email"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <button
            onClick={async () => {
              const result = await authApi.forgotPassword(email);
              setMessage(result.message);
              setToken(result.resetToken ?? "");
            }}
            className="rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white"
          >
            Generate
          </button>
        </div>
        {message && <p className="mt-2 text-sm opacity-80">{message}</p>}
        {token && (
          <p className="mt-2 break-all text-xs text-amber-500">
            Reset Token (dev): {token}
          </p>
        )}
      </section>

      <section className="stratyx-panel p-6">
        <h2 className="text-lg font-semibold">Reset Password</h2>
        <div className="mt-3 grid gap-2">
          <input
            value={token}
            onChange={(event) => setToken(event.target.value)}
            placeholder="Reset token"
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <input
            type="password"
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
            placeholder="New password"
            className="rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <button
            onClick={async () => {
              const response = await authApi.resetPassword({ token, newPassword });
              setStatus(response.message);
            }}
            className="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
          >
            Reset Now
          </button>
        </div>
        {status && <p className="mt-2 text-sm text-emerald-500">{status}</p>}
      </section>
      </div>
    </AuthShell>
  );
}
