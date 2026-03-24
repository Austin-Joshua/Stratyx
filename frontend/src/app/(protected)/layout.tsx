import ProtectedLayout from "@/layouts/protected-layout";

export default function ProtectedAppLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <ProtectedLayout>{children}</ProtectedLayout>;
}
