"use client";

import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";
import useSWR from "swr";
import {
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Tooltip,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
} from "recharts";

type MeResponse = {
  user: {
    id: string;
    full_name: string;
    email: string;
  };
};

type Account = {
  id: string;
  account_name: string;
  current_balance: string;
  currency: string;
};

type AccountsResponse = {
  accounts: Account[];
};

type Goal = {
  id: string;
  title: string;
  target_amount: string;
  current_amount: string;
  progress_percent: number;
  status: string;
};

type GoalsResponse = {
  goals: Goal[];
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

type SafeToSave = {
  engine: string;
  model_name: string;
  recommended_amount: string;
  available_balance: string;
  avg_monthly_income: string;
  avg_monthly_expense: string;
  reason: string;
};

type SafeToSaveResponse = {
  safe_to_save: SafeToSave;
};

const fetchMe = () => apiFetch<MeResponse>("/auth/me");
const fetchAccounts = () => apiFetch<AccountsResponse>("/accounts");
const fetchGoals = () => apiFetch<GoalsResponse>("/goals");
const fetchTransactions = () => apiFetch<TransactionsResponse>("/transactions");
const fetchInsight = () => apiFetch<SafeToSaveResponse>("/insights/safe-to-save");

const COLORS = ["#111827", "#374151", "#6B7280", "#9CA3AF", "#D1D5DB"];

function toNumber(value: string | undefined) {
  return Number.parseFloat(value || "0") || 0;
}

export default function DashboardPage() {
  const { data: meData, error: meError } = useSWR("/auth/me", fetchMe);
  const { data: accountsData, error: accountsError } = useSWR("/accounts", fetchAccounts);
  const { data: goalsData, error: goalsError } = useSWR("/goals", fetchGoals);
  const { data: transactionsData, error: transactionsError } = useSWR(
    "/transactions",
    fetchTransactions
  );
  const { data: insightData, error: insightError, mutate: refreshInsight } = useSWR(
    "/insights/safe-to-save",
    fetchInsight
  );

  const error = meError || accountsError || goalsError || transactionsError || insightError;

  const me = meData?.user;
  const accounts = accountsData?.accounts ?? [];
  const goals = goalsData?.goals ?? [];
  const transactions = transactionsData?.transactions ?? [];
  const safeToSave = insightData?.safe_to_save;

  const totalBalance = accounts.reduce(
    (sum, account) => sum + toNumber(account.current_balance),
    0
  );

  const activeGoals = goals.filter((goal) => goal.status === "active");
  const completedGoals = goals.filter((goal) => goal.status === "completed");

  const recentTransactions = [...transactions].slice(0, 8);

  const spendingByCategoryMap = new Map<string, number>();
  transactions
    .filter((txn) => txn.direction === "debit")
    .forEach((txn) => {
      const category = txn.category || "uncategorized";
      const amount = toNumber(txn.amount);
      spendingByCategoryMap.set(
        category,
        (spendingByCategoryMap.get(category) || 0) + amount
      );
    });

  const spendingByCategory = Array.from(spendingByCategoryMap.entries())
    .map(([name, value]) => ({ name, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 5);

  const dailyFlowMap = new Map<string, number>();
  transactions.slice(0, 20).forEach((txn) => {
    const date = new Date(txn.txn_date).toLocaleDateString("en-GB", {
      day: "2-digit",
      month: "short",
    });
    const amount = toNumber(txn.amount);
    const signedAmount = txn.direction === "credit" ? amount : -amount;

    dailyFlowMap.set(date, (dailyFlowMap.get(date) || 0) + signedAmount);
  });

  const dailyFlow = Array.from(dailyFlowMap.entries()).map(([date, value]) => ({
    date,
    value,
  }));

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">
            Welcome{me ? `, ${me.full_name}` : ""}
          </h1>
          <p className="text-sm text-gray-500">
            Your financial intelligence dashboard
          </p>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {error instanceof Error ? error.message : "Failed to load dashboard"}
          </div>
        ) : null}

        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Total Balance</p>
            <h2 className="mt-2 text-3xl font-bold">₦{totalBalance.toFixed(2)}</h2>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Accounts</p>
            <h2 className="mt-2 text-3xl font-bold">{accounts.length}</h2>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <p className="text-sm text-gray-500">Active Goals</p>
            <h2 className="mt-2 text-3xl font-bold">{activeGoals.length}</h2>
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

        <div className="grid gap-6 xl:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold">Spending by Category</h3>
            </div>

            <div className="mt-6 h-80">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={spendingByCategory}
                    dataKey="value"
                    nameKey="name"
                    outerRadius={110}
                    innerRadius={55}
                    paddingAngle={3}
                  >
                    {spendingByCategory.map((entry, index) => (
                      <Cell key={entry.name} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>

            <div className="mt-4 space-y-2">
              {spendingByCategory.map((item) => (
                <div
                  key={item.name}
                  className="flex items-center justify-between text-sm"
                >
                  <span>{item.name}</span>
                  <span className="font-medium">₦{item.value.toFixed(2)}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <h3 className="text-xl font-semibold">Recent Cash Flow</h3>

            <div className="mt-6 h-80">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={dailyFlow}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="value" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>

        <div className="grid gap-6 xl:grid-cols-2">
          <div className="rounded-2xl border bg-white p-6">
            <h3 className="text-xl font-semibold">Goals Progress</h3>

            <div className="mt-4 space-y-4">
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

                  <p className="mt-2 text-xs text-gray-500">
                    {goal.progress_percent.toFixed(1)}% completed
                  </p>
                </div>
              ))}

              {!goals.length ? (
                <p className="text-sm text-gray-500">No goals yet.</p>
              ) : null}
            </div>
          </div>

          <div className="rounded-2xl border bg-white p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-semibold">Savings Intelligence</h3>
              <button
                onClick={() => refreshInsight()}
                className="rounded-xl border px-4 py-2 text-sm hover:bg-gray-100"
              >
                Refresh
              </button>
            </div>

            <div className="mt-4 space-y-4">
              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Available Balance</p>
                <h4 className="mt-2 text-2xl font-bold">
                  ₦{safeToSave?.available_balance ?? "0.00"}
                </h4>
              </div>

              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Average Monthly Income</p>
                <h4 className="mt-2 text-2xl font-bold">
                  ₦{safeToSave?.avg_monthly_income ?? "0.00"}
                </h4>
              </div>

              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Average Monthly Expense</p>
                <h4 className="mt-2 text-2xl font-bold">
                  ₦{safeToSave?.avg_monthly_expense ?? "0.00"}
                </h4>
              </div>

              <div className="rounded-xl border p-4">
                <p className="text-sm text-gray-500">Recommendation Reason</p>
                <p className="mt-2 text-sm leading-7 text-gray-700">
                  {safeToSave?.reason ?? "No recommendation available yet."}
                </p>
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-2xl border bg-white p-6">
          <h3 className="text-xl font-semibold">Recent Transactions</h3>

          <div className="mt-4 space-y-3">
            {recentTransactions.map((txn) => (
              <div
                key={txn.id}
                className="rounded-xl border p-4 flex items-center justify-between"
              >
                <div>
                  <p className="font-medium">{txn.narration || "No narration"}</p>
                  <p className="text-sm text-gray-500">
                    {txn.category || "uncategorized"} •{" "}
                    {new Date(txn.txn_date).toLocaleDateString()}
                  </p>
                </div>

                <p
                  className={`font-semibold ${
                    txn.direction === "credit" ? "text-green-600" : "text-red-600"
                  }`}
                >
                  {txn.direction === "credit" ? "+" : "-"}₦{txn.amount}
                </p>
              </div>
            ))}

            {!recentTransactions.length ? (
              <p className="text-sm text-gray-500">No transactions yet.</p>
            ) : null}
          </div>

          <div className="mt-6 grid gap-4 sm:grid-cols-2">
            <div className="rounded-xl border p-4">
              <p className="text-sm text-gray-500">Completed Goals</p>
              <h4 className="mt-2 text-2xl font-bold">{completedGoals.length}</h4>
            </div>

            <div className="rounded-xl border p-4">
              <p className="text-sm text-gray-500">Total Transactions Loaded</p>
              <h4 className="mt-2 text-2xl font-bold">{transactions.length}</h4>
            </div>
          </div>
        </div>
      </div>
    </AppShell>
  );
}