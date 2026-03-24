"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function OAuthCallbackClient({
  accessToken,
  refreshToken,
}: {
  accessToken: string;
  refreshToken: string;
}) {
  const router = useRouter();

  useEffect(() => {
    if (!accessToken || !refreshToken) {
      router.replace("/login?error=oauth_callback_failed");
      return;
    }
    localStorage.setItem("stratyx_access_token", accessToken);
    localStorage.setItem("stratyx_refresh_token", refreshToken);
    router.replace("/dashboard");
  }, [accessToken, refreshToken, router]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-sm opacity-70">Completing sign in...</p>
    </main>
  );
}
