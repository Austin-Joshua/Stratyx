"use client";

import { useTheme } from "@/context/themeContext";

export default function ThemeToggle() {
  const { theme, toggleTheme } = useTheme();
  const isDark = theme === "dark";

  return (
    <button
      onClick={toggleTheme}
      aria-label="Toggle theme"
      className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-2 text-sm transition hover:opacity-80"
    >
      {isDark ? "☀️" : "🌙"}
    </button>
  );
}
