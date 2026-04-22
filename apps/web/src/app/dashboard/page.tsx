"use client";

import { useEffect, useState } from "react";
import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";

type MeResponse = {
  user: {
    id: string;
    full_name: string;
    email: string;
  };
};

type AccountsResponse = {
  accounts: Array<{
    id: string;
    account_name: string;
    current_balance: string;
    currency: string;
  }>;
};

type GoalsResponse = {
  goals: Array<{
    id: string;
    title: string;
    target_amount: string;
    current_amount: string;
    progress_percent: number;
    status: string;
  }>;
};

type SafeToSaveResponse = {
  safe_to_save: {
    engine: string;
    recommended_amount: string;
    available_balance: string;
    avg_monthly_income: string;
    avg_monthly_expense: string;
    reason: string;
  };
};

export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [me, setMe] = useState<MeResponse["user"] | null>(null);
  const [accounts, setAccounts] = useState<AccountsResponse["accounts"]>([]);
  const [goals, setGoals] = useState<GoalsResponse["goals"]>([]);
  const [safeToSave, setSafeToSave] =
    useState<SafeToSaveResponse["safe_to_save"] | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    async function loadDashboard() {
      try {
        const [meData, accountsData, goalsData, safeToSaveData] =
          await Promise.all([
            apiFetch<MeResponse>("/auth/me"),
            apiFetch<AccountsResponse>("/accounts"),
            apiFetch<GoalsResponse>("/goals"),
            apiFetch<SafeToSaveResponse>("/insights/safe-to-save"),
          ]);

        setMe(meData.user);
        setAccounts(accountsData.accounts);
        setGoals(goalsData.goals);
        setSafeToSave(safeToSaveData.safe_to_save);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load dashboard");
      } finally {
        setLoading(false);
      }
    }

    loadDashboard();
  }, []);

  if (loading) {
    return (
      <AppShell>
        <div>Loading dashboard...</div>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">
            Welcome{me ? `, ${me.full_name}` : ""}
          </h1>
          <p className="text-sm text-gray-500">
            Your savings intelligence overview
          </p>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {error}
          </div>
        ) : null}

        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Accounts</p>
            <h2 className="mt-2 text-3xl font-bold">{accounts.length}</h2>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Goals</p>
            <h2 className="mt-2 text-3xl font-bold">{goals.length}</h2>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Recommended Save Now</p>
            <h2 className="mt-2 text-3xl font-bold">
              ₦{safeToSave?.recommended_amount ?? "0.00"}
            </h2>
            <p className="mt-2 text-xs text-gray-500">
              Engine: {safeToSave?.engine ?? "unknown"}
            </p>
          </div>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <h3 className="text-xl font-semibold">Accounts</h3>
            <div className="mt-4 space-y-3">
              {accounts.map((account) => (
                <div
                  key={account.id}
                  className="rounded-xl border p-4 flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium">{account.account_name}</p>
                    <p className="text-sm text-gray-500">{account.currency}</p>
                  </div>
                  <p className="font-semibold">₦{account.current_balance}</p>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <h3 className="text-xl font-semibold">Goals</h3>
            <div className="mt-4 space-y-3">
              {goals.map((goal) => (
                <div key={goal.id} className="rounded-xl border p-4">
                  <div className="flex items-center justify-between">
                    <p className="font-medium">{goal.title}</p>
                    <p className="text-sm text-gray-500">{goal.status}</p>
                  </div>
                  <p className="mt-2 text-sm text-gray-500">
                    ₦{goal.current_amount} / ₦{goal.target_amount}
                  </p>
                  <div className="mt-3 h-3 w-full rounded-full bg-gray-100">
                    <div
                      className="h-3 rounded-full bg-black"
                      style={{ width: `${goal.progress_percent}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="rounded-2xl border bg-white p-6">
          <h3 className="text-xl font-semibold">Savings Insight</h3>
          <p className="mt-4 text-sm text-gray-700">
            {safeToSave?.reason ?? "No recommendation available yet."}
          </p>
        </div>
      </div>
    </AppShell>
  );
}