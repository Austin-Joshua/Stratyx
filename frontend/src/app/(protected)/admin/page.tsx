"use client";

import { useEffect, useState } from "react";

import { adminApi, type Activity, type AdminMetrics, type ModerationReport, type User } from "@/services/api";

export default function AdminPage() {
  const [metrics, setMetrics] = useState<AdminMetrics | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [reports, setReports] = useState<ModerationReport[]>([]);
  const [auditLogs, setAuditLogs] = useState<Activity[]>([]);
  const [error, setError] = useState("");

  async function load() {
    try {
      const [m, u, r, a] = await Promise.all([
        adminApi.metrics(),
        adminApi.users(),
        adminApi.moderationQueue(),
        adminApi.auditLogs(),
      ]);
      setMetrics(m);
      setUsers(u);
      setReports(r);
      setAuditLogs(a);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load admin data");
    }
  }

  useEffect(() => {
    let mounted = true;
    Promise.all([adminApi.metrics(), adminApi.users(), adminApi.moderationQueue(), adminApi.auditLogs()])
      .then(([m, u, r, a]) => {
        if (!mounted) {
          return;
        }
        setMetrics(m);
        setUsers(u);
        setReports(r);
        setAuditLogs(a);
        setError("");
      })
      .catch((err: unknown) => {
        if (!mounted) {
          return;
        }
        setError(err instanceof Error ? err.message : "Failed to load admin data");
      });
    return () => {
      mounted = false;
    };
  }, []);

  return (
    <main className="space-y-4">
      <section className="stratyx-panel p-5">
        <h1 className="text-2xl font-semibold">Admin Panel</h1>
        {error && <p className="mt-2 text-sm text-red-500">{error}</p>}
        {metrics && (
          <div className="mt-3 grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Users: {metrics.usersTotal}
            </article>
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Active Users: {metrics.usersActive}
            </article>
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Items: {metrics.itemsTotal}
            </article>
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Sessions: {metrics.sessionsActive}
            </article>
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Comments: {metrics.commentsTotal}
            </article>
            <article className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              Notifications: {metrics.notificationsTotal}
            </article>
          </div>
        )}
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">User Management</h2>
        <div className="mt-3 space-y-2">
          {users.map((user) => (
            <article key={user.id} className="rounded-md border border-[var(--border)] p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <p className="font-semibold">
                    {user.name} ({user.email})
                  </p>
                  <p className="text-xs opacity-65">
                    Role: {user.role} | Active: {user.isActive ? "Yes" : "No"}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={async () => {
                      await adminApi.setRole(user.id, user.role === "ADMIN" ? "USER" : "ADMIN");
                      await load();
                    }}
                    className="rounded-md border border-[var(--border)] px-2 py-1 text-xs"
                  >
                    Toggle Role
                  </button>
                  <button
                    onClick={async () => {
                      await adminApi.setActive(user.id, !user.isActive);
                      await load();
                    }}
                    className="rounded-md border border-[var(--border)] px-2 py-1 text-xs"
                  >
                    {user.isActive ? "Deactivate" : "Activate"}
                  </button>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Moderation Queue</h2>
        <div className="mt-3 space-y-2">
          {reports.length === 0 ? (
            <p className="text-sm opacity-70">No reports pending.</p>
          ) : (
            reports.map((report) => (
              <article key={report.id} className="rounded-md border border-[var(--border)] p-3">
                <p className="text-sm font-semibold">
                  {report.reason} - Status: {report.status}
                </p>
                <div className="mt-2 flex gap-2">
                  <button
                    onClick={async () => {
                      await adminApi.reviewModeration(report.id, "APPROVED");
                      await load();
                    }}
                    className="rounded-md border border-emerald-500 px-2 py-1 text-xs text-emerald-500"
                  >
                    Approve
                  </button>
                  <button
                    onClick={async () => {
                      await adminApi.reviewModeration(report.id, "REJECTED");
                      await load();
                    }}
                    className="rounded-md border border-red-500 px-2 py-1 text-xs text-red-500"
                  >
                    Reject
                  </button>
                </div>
              </article>
            ))
          )}
        </div>
      </section>

      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Audit Logs</h2>
        <div className="mt-3 space-y-2">
          {auditLogs.slice(0, 20).map((log) => (
            <article key={log.id} className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
              <p>{log.message}</p>
              <p className="mt-1 text-xs opacity-60">{new Date(log.createdAt).toLocaleString()}</p>
            </article>
          ))}
        </div>
      </section>
    </main>
  );
}
