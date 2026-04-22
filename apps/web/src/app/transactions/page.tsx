"use client";

import { useState } from "react";
import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";
import { getToken } from "@/lib/auth";
import useSWR from "swr";

type Account = {
  id: string;
  account_name: string;
};

type AccountsResponse = {
  accounts: Account[];
};

type Transaction = {
  id: string;
  account_id: string;
  amount: string;
  direction: string;
  narration: string;
  merchant_name: string;
  category: string;
  txn_date: string;
  source: string;
};

type TransactionsResponse = {
  transactions: Transaction[];
};

const fetchAccounts = () => apiFetch<AccountsResponse>("/accounts");
const fetchTransactions = () => apiFetch<TransactionsResponse>("/transactions");

export default function TransactionsPage() {
  const {
    data: accountsData,
    error: accountsError,
    isLoading: accountsLoading,
    mutate: mutateAccounts,
  } = useSWR("/accounts", fetchAccounts);

  const {
    data: transactionsData,
    error: transactionsError,
    isLoading: transactionsLoading,
    mutate: mutateTransactions,
  } = useSWR("/transactions", fetchTransactions);

  const accounts = accountsData?.accounts ?? [];
  const transactions = transactionsData?.transactions ?? [];

  const [submitting, setSubmitting] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  const [manualForm, setManualForm] = useState({
    account_id: "",
    amount: "",
    direction: "debit",
    narration: "",
    merchant_name: "",
    category: "",
    txn_date: "",
  });

  const [csvAccountId, setCsvAccountId] = useState("");
  const [csvFile, setCsvFile] = useState<File | null>(null);

  const loading = accountsLoading || transactionsLoading;
  const fetchError = accountsError || transactionsError;

  async function refreshAll() {
    await Promise.all([mutateAccounts(), mutateTransactions()]);
  }

  async function handleManualSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError("");

    const accountId = manualForm.account_id || (accounts.length ? accounts[0].id : "");

    try {
      await apiFetch("/transactions/manual", {
        method: "POST",
        body: JSON.stringify({
          ...manualForm,
          account_id: accountId,
        }),
      });

      setManualForm((prev) => ({
        ...prev,
        account_id: accountId,
        amount: "",
        narration: "",
        merchant_name: "",
        category: "",
        txn_date: "",
      }));

      await refreshAll();
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to create transaction"
      );
    } finally {
      setSubmitting(false);
    }
  }

  async function handleCsvUpload(e: React.FormEvent) {
    e.preventDefault();

    const accountId = csvAccountId || (accounts.length ? accounts[0].id : "");

    if (!accountId || !csvFile) {
      setError("Please select an account and a CSV file");
      return;
    }

    setUploading(true);
    setError("");

    try {
      const formData = new FormData();
      formData.append("account_id", accountId);
      formData.append("file", csvFile);

      const token = getToken();
      const response = await fetch(
        `${process.env.NEXT_PUBLIC_API_BASE_URL}/transactions/upload-csv`,
        {
          method: "POST",
          headers: {
            Authorization: `Bearer ${token}`,
          },
          body: formData,
        }
      );

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "CSV upload failed");
      }

      setCsvFile(null);
      await refreshAll();
    } catch (err) {
      setError(err instanceof Error ? err.message : "CSV upload failed");
    } finally {
      setUploading(false);
    }
  }

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Transactions</h1>
          <p className="text-sm text-gray-500">
            Add manual transactions or upload CSV statements
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
              : "Failed to load transactions page"}
          </div>
        ) : null}

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Manual Transaction</h2>

            <form onSubmit={handleManualSubmit} className="mt-4 space-y-4">
              <select
                className="w-full rounded-xl border px-4 py-3"
                value={manualForm.account_id}
                onChange={(e) =>
                  setManualForm({ ...manualForm, account_id: e.target.value })
                }
              >
                <option value="">Select account</option>
                {accounts.map((account) => (
                  <option key={account.id} value={account.id}>
                    {account.account_name}
                  </option>
                ))}
              </select>

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Amount"
                value={manualForm.amount}
                onChange={(e) =>
                  setManualForm({ ...manualForm, amount: e.target.value })
                }
              />

              <select
                className="w-full rounded-xl border px-4 py-3"
                value={manualForm.direction}
                onChange={(e) =>
                  setManualForm({ ...manualForm, direction: e.target.value })
                }
              >
                <option value="debit">Debit</option>
                <option value="credit">Credit</option>
              </select>

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Narration"
                value={manualForm.narration}
                onChange={(e) =>
                  setManualForm({ ...manualForm, narration: e.target.value })
                }
              />

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Merchant name"
                value={manualForm.merchant_name}
                onChange={(e) =>
                  setManualForm({
                    ...manualForm,
                    merchant_name: e.target.value,
                  })
                }
              />

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Category (leave empty for ML auto-categorization)"
                value={manualForm.category}
                onChange={(e) =>
                  setManualForm({ ...manualForm, category: e.target.value })
                }
              />

              <input
                type="date"
                className="w-full rounded-xl border px-4 py-3"
                value={manualForm.txn_date}
                onChange={(e) =>
                  setManualForm({ ...manualForm, txn_date: e.target.value })
                }
              />

              <button
                type="submit"
                disabled={submitting}
                className="w-full rounded-xl bg-black px-4 py-3 text-white disabled:opacity-60"
              >
                {submitting ? "Saving..." : "Add Transaction"}
              </button>
            </form>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Upload CSV</h2>

            <form onSubmit={handleCsvUpload} className="mt-4 space-y-4">
              <select
                className="w-full rounded-xl border px-4 py-3"
                value={csvAccountId}
                onChange={(e) => setCsvAccountId(e.target.value)}
              >
                <option value="">Select account</option>
                {accounts.map((account) => (
                  <option key={account.id} value={account.id}>
                    {account.account_name}
                  </option>
                ))}
              </select>

              <input
                type="file"
                accept=".csv"
                className="w-full rounded-xl border px-4 py-3"
                onChange={(e) => setCsvFile(e.target.files?.[0] || null)}
              />

              <button
                type="submit"
                disabled={uploading}
                className="w-full rounded-xl border px-4 py-3 font-medium hover:bg-gray-100 disabled:opacity-60"
              >
                {uploading ? "Uploading..." : "Upload CSV"}
              </button>
            </form>

            <p className="mt-4 text-xs text-gray-500">
              CSV can omit category and the backend will try ML auto-classification.
            </p>
          </div>
        </div>

        <div className="rounded-2xl border bg-white p-6">
          <h2 className="text-xl font-semibold">Transaction History</h2>

          {loading ? (
            <p className="mt-4 text-gray-500">Loading transactions...</p>
          ) : (
            <div className="mt-4 space-y-3">
              {transactions.map((txn) => (
                <div
                  key={txn.id}
                  className="rounded-xl border p-4 flex items-center justify-between"
                >
                  <div>
                    <p className="font-medium">{txn.narration || "No narration"}</p>
                    <p className="text-sm text-gray-500">
                      {txn.category || "uncategorized"} • {txn.source} •{" "}
                      {new Date(txn.txn_date).toLocaleDateString()}
                    </p>
                  </div>

                  <div className="text-right">
                    <p
                      className={`font-semibold ${
                        txn.direction === "credit"
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {txn.direction === "credit" ? "+" : "-"}₦{txn.amount}
                    </p>
                    <p className="text-xs text-gray-500">{txn.direction}</p>
                  </div>
                </div>
              ))}

              {!transactions.length ? (
                <p className="text-sm text-gray-500">No transactions yet.</p>
              ) : null}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  );
}