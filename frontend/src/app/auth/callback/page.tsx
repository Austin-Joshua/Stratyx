import { Suspense } from "react";

import OAuthCallbackClient from "./page_client";

export default function OAuthCallbackPage({
  searchParams,
}: {
  searchParams: { accessToken?: string; refreshToken?: string };
}) {
  return (
    <Suspense
      fallback={
        <main className="flex min-h-screen items-center justify-center">
          <p className="text-sm opacity-70">Completing sign in...</p>
        </main>
      }
    >
      <OAuthCallbackClient
        accessToken={searchParams.accessToken ?? ""}
        refreshToken={searchParams.refreshToken ?? ""}
      />
    </Suspense>
  );
}
