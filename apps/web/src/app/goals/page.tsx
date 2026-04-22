"use client";

import { useState } from "react";
import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";
import useSWR from "swr";

type Account = {
  id: string;
  account_name: string;
};

type Goal = {
  id: string;
  title: string;
  target_amount: string;
  current_amount: string;
  remaining_amount: string;
  progress_percent: number;
  target_date: string;
  status: string;
};

type Contribution = {
  id: string;
  goal_id: string;
  account_id: string;
  amount: string;
  note: string;
  created_at: string;
};

type AccountsResponse = {
  accounts: Account[];
};

type GoalsResponse = {
  goals: Goal[];
};

type ContributionsResponse = {
  contributions: Contribution[];
};

const fetchAccounts = () => apiFetch<AccountsResponse>("/accounts");
const fetchGoals = () => apiFetch<GoalsResponse>("/goals");

export default function GoalsPage() {
  const {
    data: accountsData,
    error: accountsError,
    isLoading: accountsLoading,
    mutate: mutateAccounts,
  } = useSWR("/accounts", fetchAccounts);

  const {
    data: goalsData,
    error: goalsError,
    isLoading: goalsLoading,
    mutate: mutateGoals,
  } = useSWR("/goals", fetchGoals);

  const [creating, setCreating] = useState(false);
  const [error, setError] = useState("");
  const [historyByGoal, setHistoryByGoal] = useState<Record<string, Contribution[]>>({});
  const [loadingHistoryByGoal, setLoadingHistoryByGoal] = useState<Record<string, boolean>>({});

  const [goalForm, setGoalForm] = useState({
    title: "",
    target_amount: "",
    target_date: "",
    priority: 1,
  });

  const [contributionState, setContributionState] = useState<Record<string, string>>({});
  const [selectedAccountByGoal, setSelectedAccountByGoal] = useState<Record<string, string>>({});

  const accounts = accountsData?.accounts ?? [];
  const goals = goalsData?.goals ?? [];
  const loading = accountsLoading || goalsLoading;
  const fetchError = accountsError || goalsError;

  async function refreshAll() {
    await Promise.all([mutateAccounts(), mutateGoals()]);
  }

  async function loadHistory(goalId: string) {
    try {
      setLoadingHistoryByGoal((prev) => ({ ...prev, [goalId]: true }));
      const data = await apiFetch<ContributionsResponse>(`/goals/${goalId}/contributions`);
      setHistoryByGoal((prev) => ({ ...prev, [goalId]: data.contributions }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load history");
    } finally {
      setLoadingHistoryByGoal((prev) => ({ ...prev, [goalId]: false }));
    }
  }

  async function handleCreateGoal(e: React.FormEvent) {
    e.preventDefault();
    setCreating(true);
    setError("");

    try {
      await apiFetch("/goals", {
        method: "POST",
        body: JSON.stringify(goalForm),
      });

      setGoalForm({
        title: "",
        target_amount: "",
        target_date: "",
        priority: 1,
      });

      await refreshAll();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create goal");
    } finally {
      setCreating(false);
    }
  }

  async function contribute(goalId: string) {
    const amount = contributionState[goalId];
    const accountId =
      selectedAccountByGoal[goalId] || (accounts.length ? accounts[0].id : "");

    if (!amount || !accountId) {
      setError("Select an account and enter a contribution amount");
      return;
    }

    try {
      setError("");

      await apiFetch(`/goals/${goalId}/contribute`, {
        method: "POST",
        body: JSON.stringify({
          account_id: accountId,
          amount,
          note: "Contribution from dashboard",
        }),
      });

      setContributionState((prev) => ({ ...prev, [goalId]: "" }));
      await refreshAll();
      await loadHistory(goalId);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to contribute");
    }
  }

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Goals</h1>
          <p className="text-sm text-gray-500">
            Create and fund your savings goals
          </p>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {error}
          </div>
        ) : null}

        {fetchError ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {fetchError instanceof Error ? fetchError.message : "Failed to load goals page"}
          </div>
        ) : null}

        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Create Goal</h2>

            <form onSubmit={handleCreateGoal} className="mt-4 space-y-4">
              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Goal title"
                value={goalForm.title}
                onChange={(e) =>
                  setGoalForm({ ...goalForm, title: e.target.value })
                }
              />

              <input
                className="w-full rounded-xl border px-4 py-3"
                placeholder="Target amount"
                value={goalForm.target_amount}
                onChange={(e) =>
                  setGoalForm({ ...goalForm, target_amount: e.target.value })
                }
              />

              <input
                type="date"
                className="w-full rounded-xl border px-4 py-3"
                value={goalForm.target_date}
                onChange={(e) =>
                  setGoalForm({ ...goalForm, target_date: e.target.value })
                }
              />

              <select
                className="w-full rounded-xl border px-4 py-3"
                value={goalForm.priority}
                onChange={(e) =>
                  setGoalForm({
                    ...goalForm,
                    priority: Number(e.target.value),
                  })
                }
              >
                <option value={1}>Priority 1</option>
                <option value={2}>Priority 2</option>
                <option value={3}>Priority 3</option>
              </select>

              <button
                type="submit"
                disabled={creating}
                className="w-full rounded-xl bg-black px-4 py-3 text-white disabled:opacity-60"
              >
                {creating ? "Creating..." : "Create Goal"}
              </button>
            </form>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <h2 className="text-xl font-semibold">Goal Overview</h2>
            <div className="mt-4 grid gap-4 sm:grid-cols-2">
              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Total Goals</p>
                <p className="mt-2 text-3xl font-bold">{goals.length}</p>
              </div>

              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Completed</p>
                <p className="mt-2 text-3xl font-bold">
                  {goals.filter((g) => g.status === "completed").length}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border bg-white p-6">
          <h2 className="text-xl font-semibold">Your Goals</h2>

          {loading ? (
            <p className="mt-4 text-gray-500">Loading goals...</p>
          ) : (
            <div className="mt-4 space-y-4">
              {goals.map((goal) => (
                <div key={goal.id} className="rounded-xl border p-5">
                  <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                    <div>
                      <h3 className="text-lg font-semibold">{goal.title}</h3>
                      <p className="text-sm text-gray-500">
                        ₦{goal.current_amount} / ₦{goal.target_amount} • Remaining
                        ₦{goal.remaining_amount}
                      </p>
                      <p className="text-sm text-gray-500">
                        Status: {goal.status} • Target:{" "}
                        {new Date(goal.target_date).toLocaleDateString()}
                      </p>
                    </div>

                    <div className="w-full lg:w-96">
                      <div className="h-3 w-full rounded-full bg-gray-100">
                        <div
                          className="h-3 rounded-full bg-black"
                          style={{ width: `${goal.progress_percent}%` }}
                        />
                      </div>
                      <p className="mt-2 text-xs text-gray-500">
                        {goal.progress_percent.toFixed(1)}% complete
                      </p>
                    </div>
                  </div>

                  {goal.status !== "completed" ? (
                    <div className="mt-4 grid gap-3 lg:grid-cols-[1fr_1fr_160px_140px]">
                      <select
                        className="rounded-xl border px-4 py-3"
                        value={
                          selectedAccountByGoal[goal.id] ||
                          (accounts.length ? accounts[0].id : "")
                        }
                        onChange={(e) =>
                          setSelectedAccountByGoal((prev) => ({
                            ...prev,
                            [goal.id]: e.target.value,
                          }))
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
                        className="rounded-xl border px-4 py-3"
                        placeholder="Contribution amount"
                        value={contributionState[goal.id] || ""}
                        onChange={(e) =>
                          setContributionState((prev) => ({
                            ...prev,
                            [goal.id]: e.target.value,
                          }))
                        }
                      />

                      <button
                        onClick={() => contribute(goal.id)}
                        className="rounded-xl bg-black px-4 py-3 text-white"
                      >
                        Contribute
                      </button>

                      <button
                        onClick={() => loadHistory(goal.id)}
                        className="rounded-xl border px-4 py-3"
                      >
                        {loadingHistoryByGoal[goal.id] ? "Loading..." : "History"}
                      </button>
                    </div>
                  ) : (
                    <div className="mt-4">
                      <button
                        onClick={() => loadHistory(goal.id)}
                        className="rounded-xl border px-4 py-3"
                      >
                        {loadingHistoryByGoal[goal.id] ? "Loading..." : "History"}
                      </button>
                    </div>
                  )}

                  {historyByGoal[goal.id]?.length ? (
                    <div className="mt-4 rounded-xl border bg-gray-50 p-4">
                      <h4 className="font-medium">Contribution History</h4>
                      <div className="mt-3 space-y-2">
                        {historyByGoal[goal.id].map((item) => (
                          <div
                            key={item.id}
                            className="flex items-center justify-between rounded-lg border bg-white p-3"
                          >
                            <div>
                              <p className="font-medium">₦{item.amount}</p>
                              <p className="text-xs text-gray-500">
                                {item.note || "No note"}
                              </p>
                            </div>
                            <p className="text-xs text-gray-500">
                              {new Date(item.created_at).toLocaleString()}
                            </p>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </div>
              ))}

              {!goals.length ? (
                <p className="text-sm text-gray-500">No goals yet.</p>
              ) : null}
            </div>
          )}
        </div>
      </div>
    </AppShell>
  );
}