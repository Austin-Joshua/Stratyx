"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";

import { authApi, type OAuthAccount } from "@/services/api";

export default function SettingsPage() {
  const params = useSearchParams();
  const [verifyToken, setVerifyToken] = useState(() => params.get("verifyToken") ?? "");
  const [status, setStatus] = useState("");
  const [twoFASecret, setTwoFASecret] = useState("");
  const [twoFAURI, setTwoFAURI] = useState("");
  const [twoFACode, setTwoFACode] = useState("");
  const [connectedAccounts, setConnectedAccounts] = useState<OAuthAccount[]>([]);

  useEffect(() => {
    authApi.connectedAccounts().then(setConnectedAccounts).catch(() => undefined);
  }, []);

  return (
    <main className="space-y-4">
      <section className="stratyx-panel p-5">
        <h1 className="text-2xl font-semibold">Settings</h1>
        <p className="mt-1 text-sm opacity-75">
          Security settings, email verification, and session controls.
        </p>
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Email Verification</h2>
        <div className="mt-3 flex gap-2">
          <input
            value={verifyToken}
            onChange={(event) => setVerifyToken(event.target.value)}
            placeholder="Verification token"
            className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
          />
          <button
            onClick={async () => {
              const response = await authApi.verifyEmail(verifyToken);
              setStatus(response.message);
            }}
            className="rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white"
          >
            Verify
          </button>
        </div>
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Session Control</h2>
        <button
          onClick={async () => {
            const response = await authApi.logoutAll();
            setStatus(response.message);
          }}
          className="mt-2 rounded-md border border-[var(--border)] px-3 py-2 text-sm"
        >
          Logout all devices
        </button>
      </section>

      <section className="stratyx-panel p-5 space-y-3">
        <h2 className="text-lg font-semibold">Two-Factor Authentication (TOTP)</h2>
        <div className="flex flex-wrap gap-2">
          <button
            onClick={async () => {
              const result = await authApi.setup2FA();
              setTwoFASecret(result.secret);
              setTwoFAURI(result.otpAuthURL);
            }}
            className="rounded-md bg-[var(--primary)] px-3 py-2 text-sm font-semibold text-white"
          >
            Setup 2FA
          </button>
          <button
            onClick={async () => {
              const result = await authApi.disable2FA(twoFACode);
              setStatus(result.message);
            }}
            className="rounded-md border border-[var(--border)] px-3 py-2 text-sm"
          >
            Disable 2FA
          </button>
        </div>
        {twoFASecret && (
          <>
            <p className="text-xs break-all opacity-70">Secret: {twoFASecret}</p>
            <p className="text-xs break-all opacity-70">OTP URL: {twoFAURI}</p>
            <div className="flex gap-2">
              <input
                value={twoFACode}
                onChange={(event) => setTwoFACode(event.target.value)}
                placeholder="OTP code"
                className="w-full rounded-md border border-[var(--border)] bg-transparent px-3 py-2 text-sm"
              />
              <button
                onClick={async () => {
                  const result = await authApi.verify2FASetup(twoFACode);
                  setStatus(result.message);
                }}
                className="rounded-md bg-slate-900 px-3 py-2 text-sm font-semibold text-white"
              >
                Verify
              </button>
            </div>
          </>
        )}
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Connected Accounts</h2>
        {connectedAccounts.length === 0 ? (
          <p className="mt-2 text-sm opacity-70">No connected social accounts.</p>
        ) : (
          <div className="mt-3 space-y-2">
            {connectedAccounts.map((acc) => (
              <article key={acc.id} className="rounded-md border border-[var(--border)] p-3">
                <div className="flex items-center justify-between">
                  <p className="text-sm">
                    {acc.provider} - {acc.email}
                  </p>
                  <button
                    onClick={async () => {
                      await authApi.disconnectAccount(acc.provider);
                      setConnectedAccounts((prev) =>
                        prev.filter((candidate) => candidate.id !== acc.id),
                      );
                    }}
                    className="text-xs text-red-500"
                  >
                    Disconnect
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      {status && (
        <section className="stratyx-panel p-4">
          <p className="text-sm text-emerald-500">{status}</p>
        </section>
      )}
    </main>
  );
}
