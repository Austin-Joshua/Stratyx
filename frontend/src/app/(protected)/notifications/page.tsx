"use client";

import { useEffect, useState } from "react";

import { notificationsApi, type Notification } from "@/services/api";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);

  async function load() {
    setLoading(true);
    try {
      const data = await notificationsApi.list();
      setNotifications(data);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <main className="stratyx-panel p-5">
      <h1 className="text-2xl font-semibold">Notifications</h1>
      {loading ? (
        <p className="mt-3 text-sm opacity-70">Loading...</p>
      ) : notifications.length === 0 ? (
        <p className="mt-3 text-sm opacity-70">No notifications yet.</p>
      ) : (
        <div className="mt-3 space-y-2">
          {notifications.map((n) => (
            <article key={n.id} className="rounded-md border border-[var(--border)] p-3">
              <div className="flex items-center justify-between">
                <p className="font-semibold">{n.title}</p>
                <button
                  onClick={async () => {
                    await notificationsApi.markRead(n.id, !n.read);
                    await load();
                  }}
                  className="text-xs text-[var(--primary)]"
                >
                  Mark as {n.read ? "unread" : "read"}
                </button>
              </div>
              <p className="mt-1 text-sm opacity-80">{n.body}</p>
            </article>
          ))}
        </div>
      )}
    </main>
  );
}
