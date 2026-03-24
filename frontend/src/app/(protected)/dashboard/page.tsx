"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { dashboardApi, type Activity, type DashboardSummary } from "@/services/api";

export default function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [activity, setActivity] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      setLoading(true);
      try {
        const [summaryData, activityData] = await Promise.all([
          dashboardApi.summary(),
          dashboardApi.activity(),
        ]);
        if (!mounted) {
          return;
        }
        setSummary(summaryData);
        setActivity(activityData.slice(0, 5));
      } finally {
        if (mounted) {
          setLoading(false);
        }
      }
    };
    void load();
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <main className="space-y-4">
      <header className="stratyx-panel p-5">
        <h1 className="text-2xl font-semibold">Executive Dashboard</h1>
        <p className="mt-1 text-sm opacity-75">
          Connected end-to-end product flow: UI → API → MongoDB.
        </p>
      </header>

      <section className="grid gap-4 md:grid-cols-3">
        <article className="stratyx-panel p-4">
          <p className="text-xs opacity-70">Data module</p>
          <h2 className="mt-1 text-xl font-semibold">
            Items: {loading ? "..." : summary?.itemsCount ?? 0}
          </h2>
          <p className="mt-2 text-sm opacity-80">
            Create, read, update, and delete real records with owner isolation.
          </p>
          <Link href="/items" className="mt-4 inline-block text-sm text-[var(--primary)]">
            Open items →
          </Link>
        </article>
        <article className="stratyx-panel p-4">
          <p className="text-xs opacity-70">Security</p>
          <h2 className="mt-1 text-xl font-semibold">
            Unread Alerts: {loading ? "..." : summary?.unreadNotifications ?? 0}
          </h2>
          <p className="mt-2 text-sm opacity-80">
            Access tokens auto-refresh through Axios interceptors.
          </p>
        </article>
        <article className="stratyx-panel p-4">
          <p className="text-xs opacity-70">Environment</p>
          <h2 className="mt-1 text-xl font-semibold">
            Your Comments: {loading ? "..." : summary?.commentsCount ?? 0}
          </h2>
          <p className="mt-2 text-sm opacity-80">
            Runtime configured with environment variables for frontend/backend.
          </p>
        </article>
      </section>
      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Recent Activity</h2>
        {activity.length === 0 ? (
          <p className="mt-2 text-sm opacity-70">No activity yet.</p>
        ) : (
          <div className="mt-3 space-y-2">
            {activity.map((entry) => (
              <article key={entry.id} className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
                <p>{entry.message}</p>
                <p className="mt-1 text-xs opacity-60">
                  {new Date(entry.createdAt).toLocaleString()}
                </p>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
