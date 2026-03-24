"use client";

import { useEffect, useState } from "react";
import Image from "next/image";

import { useAuth } from "@/hooks/useAuth";
import { authApi, type Session } from "@/services/api";

export default function ProfilePage() {
  const { user, refreshMe } = useAuth();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [name, setName] = useState("");
  const [avatarUploading, setAvatarUploading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let mounted = true;
    authApi.sessions().then((data) => {
      if (mounted) {
        setSessions(data);
      }
    });
    return () => {
      mounted = false;
    };
  }, []);

  useEffect(() => {
    setName(user?.name ?? "");
  }, [user?.name]);

  const saveProfile = async () => {
    setSaving(true);
    setMessage("");
    try {
      await authApi.updateProfile({
        name,
        avatarUrl: user?.avatarUrl ?? "",
        privacySettings: user
          ? { showEmail: false, showProfile: true }
          : { showEmail: false, showProfile: true },
        notificationPrefs: user
          ? { emailMentions: true, inAppMentions: true, productNews: true }
          : { emailMentions: true, inAppMentions: true, productNews: true },
      });
      await refreshMe();
      setMessage("Profile updated.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Failed to update profile.");
    } finally {
      setSaving(false);
    }
  };

  const uploadAvatar = async (file?: File) => {
    if (!file) return;
    setAvatarUploading(true);
    setMessage("");
    try {
      await authApi.uploadAvatar(file);
      await refreshMe();
      setMessage("Avatar uploaded.");
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Failed to upload avatar.");
    } finally {
      setAvatarUploading(false);
    }
  };

  return (
    <main className="space-y-4">
      <section className="stratyx-panel p-5">
        <h1 className="text-2xl font-semibold">Profile</h1>
        <div className="mt-4 grid gap-3 text-sm md:grid-cols-2">
          <div className="space-y-3">
            <p>
              <span className="opacity-70">Email:</span> {user?.email}
            </p>
            <p>
              <span className="opacity-70">Role:</span> {user?.role}
            </p>
            <p>
              <span className="opacity-70">User ID:</span> {user?.id}
            </p>
            <label className="block">
              <span className="mb-1 block opacity-70">Display name</span>
              <input
                className="w-full rounded-md border border-[var(--border)] bg-[var(--surface)] px-3 py-2"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </label>
            <button
              onClick={saveProfile}
              disabled={saving}
              className="rounded-md bg-[var(--primary)] px-4 py-2 text-white disabled:opacity-60"
            >
              {saving ? "Saving..." : "Save Profile"}
            </button>
          </div>
          <div className="space-y-3">
            <p className="opacity-70">Avatar</p>
            {user?.avatarUrl ? (
              <Image src={user.avatarUrl} alt="avatar" width={80} height={80} className="h-20 w-20 rounded-full object-cover" />
            ) : (
              <div className="flex h-20 w-20 items-center justify-center rounded-full bg-[var(--surface-muted)] text-xs">
                No avatar
              </div>
            )}
            <input
              type="file"
              accept="image/*"
              onChange={(e) => uploadAvatar(e.target.files?.[0])}
              className="block text-sm"
            />
            {avatarUploading && <p className="text-xs opacity-70">Uploading avatar...</p>}
          </div>
        </div>
        {message ? <p className="mt-3 text-sm opacity-80">{message}</p> : null}
        <div className="mt-4 grid gap-3 text-sm">
          <p>
            <span className="opacity-70">Name:</span> {user?.name}
          </p>
        </div>
      </section>
      <section className="stratyx-panel p-5">
        <h2 className="text-lg font-semibold">Active Sessions</h2>
        {sessions.length === 0 ? (
          <p className="mt-2 text-sm opacity-70">No active sessions found.</p>
        ) : (
          <div className="mt-3 space-y-2">
            {sessions.map((session) => (
              <article key={session.id} className="rounded-md bg-[var(--surface-muted)] p-3 text-sm">
                <p>{session.deviceName || "Unknown device"}</p>
                <p className="text-xs opacity-70">
                  {session.ipAddress} • Last seen {new Date(session.lastSeenAt).toLocaleString()}
                </p>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
