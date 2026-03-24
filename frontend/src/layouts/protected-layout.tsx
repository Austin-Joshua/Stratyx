"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import Image from "next/image";

import ThemeToggle from "@/components/theme-toggle";
import { useAuth } from "@/hooks/useAuth";

export default function ProtectedLayout({ children }: { children: React.ReactNode }) {
  const { loading, isAuthenticated, logout, user } = useAuth();
  const nav = [
    { label: "Dashboard", href: "/dashboard" },
    { label: "Items", href: "/items" },
    { label: "Notifications", href: "/notifications" },
    { label: "Profile", href: "/profile" },
    { label: "Settings", href: "/settings" },
    ...(user?.role === "ADMIN" || user?.role === "SUPER_ADMIN" || user?.role === "MODERATOR"
      ? [{ label: "Admin", href: "/admin" }]
      : []),
  ];

  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      router.replace("/login");
    }
  }, [loading, isAuthenticated, router]);

  if (loading) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm opacity-70">Loading session...</p>
      </main>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen px-4 py-4 md:px-6">
      <div className="mx-auto grid w-full max-w-7xl gap-4 lg:grid-cols-[220px_1fr]">
        <aside className="stratyx-panel p-4">
          <Image
            src="/stratyx-mark.png"
            alt="STRATYX"
            width={64}
            height={64}
            className="h-16 w-16 rounded-md object-cover"
          />
          <p className="mt-3 text-sm font-semibold">{user?.name}</p>
          <p className="text-xs opacity-65">{user?.email}</p>
          <nav className="mt-5 space-y-2">
            {nav.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={`block rounded-md px-3 py-2 text-sm ${
                  pathname === item.href
                    ? "bg-[var(--primary)] text-white"
                    : "hover:bg-[var(--surface-muted)]"
                }`}
              >
                {item.label}
              </Link>
            ))}
          </nav>
          <button
            onClick={async () => {
              await logout();
              router.replace("/login");
            }}
            className="mt-6 w-full rounded-md border border-[var(--border)] px-3 py-2 text-sm"
          >
            Logout
          </button>
        </aside>
        <section className="space-y-3">
          <header className="stratyx-panel flex items-center justify-between px-4 py-3">
            <p className="text-sm font-semibold opacity-80">Workspace</p>
            <ThemeToggle />
          </header>
          {children}
        </section>
      </div>
    </div>
  );
}
