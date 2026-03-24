import Image from "next/image";
import PublicNavbar from "@/components/public-navbar";

export default function Home() {
  return (
    <main className="min-h-screen bg-[var(--background)] text-[var(--foreground)]">
      <PublicNavbar />
      <div className="mx-auto grid w-full max-w-7xl gap-10 px-6 py-16 md:grid-cols-2 md:items-center md:px-8 md:py-24">
        <section>
          <p className="text-sm font-semibold opacity-70">STRATYX Enterprise Suite</p>
          <h1 className="mt-4 text-5xl font-bold leading-tight tracking-tight md:text-7xl">
            Execute strategy
            <br />
            with measurable
            <br />
            outcomes
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-relaxed opacity-85">
            Align goals, initiatives, projects, and tasks in one AI-powered
            command center. Built for secure multi-tenant organizations that
            need real accountability and performance visibility.
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <a
              href="/dashboard"
              className="rounded-full bg-[var(--primary)] px-6 py-3 text-sm font-semibold text-white"
            >
              Get started
            </a>
            <a
              href="/login"
              className="rounded-full border border-[var(--border)] bg-[var(--surface)] px-6 py-3 text-sm font-semibold"
            >
              See demo
            </a>
          </div>
        </section>

        <section className="relative">
          <div className="rounded-3xl bg-[var(--surface)] p-6 shadow-2xl shadow-slate-400/20 md:p-8">
            <div className="rounded-2xl bg-[var(--surface-muted)] p-4">
              <p className="text-xs font-semibold uppercase tracking-[0.2em] opacity-60">
                Strategic Health
              </p>
              <p className="mt-3 text-4xl font-bold">89%</p>
              <p className="mt-1 text-sm text-emerald-600">+11% this quarter</p>
            </div>
            <div className="mt-4 flex items-end justify-between">
              <div>
                <p className="text-sm font-semibold opacity-75">Portfolio Risk</p>
                <p className="text-2xl font-bold">Low-Medium</p>
              </div>
              <Image
                src="/stratyx-mark.png"
                alt="STRATYX visual"
                width={180}
                height={180}
                className="h-auto w-32 md:w-40"
                priority
              />
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
