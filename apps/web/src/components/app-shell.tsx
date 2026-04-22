"use client";

import dynamic from "next/dynamic";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import Sidebar from "@/components/sidebar";
import { getToken } from "@/lib/auth";

function AppShellInner({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const token = getToken();

  useEffect(() => {
    if (!token) {
      router.replace("/login");
    }
  }, [token, router]);

  if (!token) {
    return <div className="p-8">Loading...</div>;
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar />
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}

export default dynamic(() => Promise.resolve(AppShellInner), {
  ssr: false,
});