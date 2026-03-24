import Image from "next/image";
import Link from "next/link";

import ThemeToggle from "@/components/theme-toggle";

export default function AuthShell({
  mode,
  title,
  subtitle,
  children,
}: {
  mode: "signin" | "signup" | "forgot";
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  const heroText =
    mode === "signup"
      ? "Create your account and get started."
      : mode === "forgot"
        ? "Reset access securely and quickly."
        : "Manage your strategy in one place.";

  return (
    <main className="min-h-screen bg-[var(--background)] px-4 py-8">
      <div className="mx-auto max-w-6xl">
        <div className="mb-3 flex justify-end">
          <ThemeToggle />
        </div>
        <section className="auth-enter overflow-hidden rounded-3xl border border-[var(--border)] bg-[var(--surface)] shadow-xl">
          <div className="grid md:grid-cols-2">
            <aside className="flex flex-col justify-between bg-gradient-to-br from-[#0aa377] to-[#0b7b63] p-8 text-white">
              <div>
                <Link href="/" className="text-xs font-semibold uppercase tracking-[0.15em] opacity-80">
                  ← Back to Landing
                </Link>
                <Image
                  src="/stratyx-logo.png"
                  alt="STRATYX"
                  width={90}
                  height={90}
                  className="mt-8 h-20 w-20 rounded-2xl object-cover"
                />
                <h2 className="mt-8 text-4xl font-bold leading-tight">{heroText}</h2>
              <p className="mt-4 text-sm font-semibold uppercase tracking-[0.16em] opacity-80">
                {title}
              </p>
              <p className="mt-1 text-sm opacity-90">{subtitle}</p>
                <div className="mt-4 h-1 w-14 rounded bg-emerald-300" />
                <p className="mt-5 max-w-xs text-sm opacity-90">
                  Secure, collaborative, and analytics-driven execution for modern teams.
                </p>
              </div>
            </aside>
            <section className="p-8 md:p-10">{children}</section>
          </div>
        </section>
      </div>
    </main>
  );
}
