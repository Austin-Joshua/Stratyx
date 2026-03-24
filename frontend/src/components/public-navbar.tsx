import Image from "next/image";
import Link from "next/link";

import ThemeToggle from "@/components/theme-toggle";

export default function PublicNavbar() {
  return (
    <header className="border-b border-[var(--border)] bg-[var(--surface)]/90 backdrop-blur">
      <div className="mx-auto flex w-full max-w-7xl items-center justify-between px-4 py-3 md:px-8">
        <div className="flex items-center gap-3">
          <Image src="/stratyx-mark.png" alt="STRATYX mark" width={32} height={32} />
          <p className="text-xl font-bold tracking-tight">STRATYX</p>
        </div>

        <nav className="hidden items-center gap-8 text-xs font-semibold uppercase tracking-[0.15em] opacity-75 md:flex">
          <Link href="/">Platform</Link>
          <Link href="/">Solutions</Link>
          <Link href="/">Vision</Link>
          <Link href="/">Network</Link>
        </nav>

        <div className="flex items-center gap-3">
          <ThemeToggle />
          <Link href="/login" className="text-sm font-semibold">
            Login
          </Link>
          <Link
            href="/register"
            className="rounded-xl bg-[#00a86b] px-4 py-2 text-sm font-semibold text-white shadow"
          >
            Get Started
          </Link>
        </div>
      </div>
    </header>
  );
}
