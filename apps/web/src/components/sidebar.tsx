"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { clearToken } from "@/lib/auth";

const links = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/accounts", label: "Accounts" },
  { href: "/transactions", label: "Transactions" },
  { href: "/goals", label: "Goals" },
  { href: "/insights", label: "Insights" },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();

  function logout() {
    clearToken();
    router.push("/login");
  }

  return (
    <aside className="w-64 min-h-screen border-r bg-white p-6">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">KoloWise</h1>
        <p className="text-sm text-gray-500">Savings Intelligence</p>
      </div>

      <nav className="space-y-2">
        {links.map((link) => {
          const active = pathname === link.href;
          return (
            <Link
              key={link.href}
              href={link.href}
              className={`block rounded-xl px-4 py-3 text-sm font-medium ${
                active
                  ? "bg-black text-white"
                  : "text-gray-700 hover:bg-gray-100"
              }`}
            >
              {link.label}
            </Link>
          );
        })}
      </nav>

      <button
        onClick={logout}
        className="mt-10 w-full rounded-xl border px-4 py-3 text-sm font-medium hover:bg-gray-100"
      >
        Logout
      </button>
    </aside>
  );
}