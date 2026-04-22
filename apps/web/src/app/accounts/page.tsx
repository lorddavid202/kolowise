"use client";

import { useState } from "react";
import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";
import useSWR from "swr";

type Account = {
  id: string;
  account_name: string;
  institution_name: string;
  account_type: string;
  current_balance: string;
  currency: string;
};

type AccountsResponse = {
  accounts: Account[];
};

const fetchAccounts = () => apiFetch<AccountsResponse>("/accounts");

export default function AccountsPage() {
  const {
    data,
    error: fetchError,
    isLoading,
    mutate,
  } = useSWR("/accounts", fetchAccounts);

  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const [form, setForm] = useState({
    account_name: "",
    institution_name: "",
    account_type: "checking",
    opening_balance: "",
    currency: "NGN",
  });

  const accounts = data?.accounts ?? [];

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");

    try {
      await apiFetch("/accounts", {
        method: "POST",
        body: JSON.stringify(form),
      });

      setForm({
        account_name: "",
        institution_name: "",
        account_type: "checking",
        opening_balance: "",
        currency: "NGN",
      });

      await mutate();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create account");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Accounts</h1>
          <p className="text-sm text-gray-500">
            Create and manage your financial accounts
          </p>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {error}
          </div>
        ) : null}

        {fetchError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {fetchError instanceof Error
              ? fetchError.message
              : "Failed to load accounts"}
          </div>
        ) : null}

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Create Account</h2>

            <form onSubmit={handleSubmit} className="mt-4 space-y-4">
              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Account name"
                value={form.account_name}
                onChange={(e) =>
                  setForm({ ...form, account_name: e.target.value })
                }
              />

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Institution name"
                value={form.institution_name}
                onChange={(e) =>
                  setForm({ ...form, institution_name: e.target.value })
                }
              />

              <select
                className="w-full rounded-xl border px-4 py-3"
                value={form.account_type}
                onChange={(e) =>
                  setForm({ ...form, account_type: e.target.value })
                }
              >
                <option value="checking">Checking</option>
                <option value="savings">Savings</option>
                <option value="wallet">Wallet</option>
              </select>

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Opening balance"
                value={form.opening_balance}
                onChange={(e) =>
                  setForm({ ...form, opening_balance: e.target.value })
                }
              />

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Currency"
                value={form.currency}
                onChange={(e) => setForm({ ...form, currency: e.target.value })}
              />

              <button
                type="submit"
                disabled={submitting}
                className="w-full rounded-xl bg-black px-4 py-3 text-white disabled:opacity-60"
              >
                {submitting ? "Creating..." : "Create Account"}
              </button>
            </form>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Your Accounts</h2>

            {isLoading ? (
              <p className="mt-4 text-gray-500">Loading accounts...</p>
            ) : (
              <div className="mt-4 space-y-3">
                {accounts.map((account) => (
                  <div
                    key={account.id}
                    className="rounded-xl border p-4 flex items-center justify-between"
                  >
                    <div>
                      <p className="font-medium">{account.account_name}</p>
                      <p className="text-sm text-gray-500">
                        {account.institution_name || "No institution"} •{" "}
                        {account.account_type}
                      </p>
                    </div>
                    <p className="font-semibold">
                      {account.currency} {account.current_balance}
                    </p>
                  </div>
                ))}

                {!accounts.length ? (
                  <p className="text-sm text-gray-500">No accounts yet.</p>
                ) : null}
              </div>
            )}
          </div>
        </div>
      </div>
    </AppShell>
  );
}