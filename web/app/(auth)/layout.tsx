import { redirect } from "next/navigation";
import { getServerSession } from "@/lib/auth/server";

export default async function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const user = await getServerSession();
  if (!user) redirect("/");
  return <>{children}</>;
}
