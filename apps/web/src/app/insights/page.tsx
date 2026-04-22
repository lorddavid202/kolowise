"use client";

import AppShell from "@/components/app-shell";
import { apiFetch } from "@/lib/api";
import useSWR from "swr";

type SafeToSave = {
  engine: string;
  model_name: string;
  recommended_amount: string;
  available_balance: string;
  avg_monthly_income: string;
  avg_monthly_expense: string;
  emergency_buffer: string;
  monthly_surplus: string;
  active_goal_need: string;
  reason: string;
};

type SafeToSaveResponse = {
  safe_to_save: SafeToSave;
};

const fetcher = () => apiFetch<SafeToSaveResponse>("/insights/safe-to-save");

export default function InsightsPage() {
  const { data, error, isLoading, mutate } = useSWR(
    "/insights/safe-to-save",
    fetcher
  );

  const safeToSave = data?.safe_to_save;

  return (
    <AppShell>
      <div className="space-y-8">
        <div>
          <h1 className="text-3xl font-bold">Insights</h1>
          <p className="text-sm text-gray-500">
            Savings intelligence and explanation panel
          </p>
        </div>

        {error ? (
          <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">
            {error instanceof Error ? error.message : "Failed to load insights"}
          </div>
        ) : null}

        {isLoading ? (
          <div className="rounded-2xl border bg-white p-6">
            Loading insights...
          </div>
        ) : (
          <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-3">
              <div className="rounded-2xl border bg-white p-6">
                <p className="text-sm text-gray-500">Recommended Save Now</p>
                <h2 className="mt-2 text-3xl font-bold">
                  ₦{safeToSave?.recommended_amount}
                </h2>
              </div>

              <div className="rounded-2xl border bg-white p-6">
                <p className="text-sm text-gray-500">Engine</p>
                <h2 className="mt-2 text-2xl font-bold">
                  {safeToSave?.engine}
                </h2>
                <p className="mt-2 text-xs text-gray-500">
                  {safeToSave?.model_name}
                </p>
              </div>

              <div className="rounded-2xl border bg-white p-6">
                <p className="text-sm text-gray-500">Monthly Surplus</p>
                <h2 className="mt-2 text-3xl font-bold">
                  ₦{safeToSave?.monthly_surplus}
                </h2>
              </div>
            </div>

            <div className="grid gap-6 lg:grid-cols-2">
              <div className="rounded-2xl border bg-white p-6">
                <h3 className="text-xl font-semibold">Financial Inputs</h3>

                <div className="mt-4 space-y-3 text-sm">
                  <div className="flex justify-between">
                    <span>Available Balance</span>
                    <span className="font-medium">
                      ₦{safeToSave?.available_balance}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span>Average Monthly Income</span>
                    <span className="font-medium">
                      ₦{safeToSave?.avg_monthly_income}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span>Average Monthly Expense</span>
                    <span className="font-medium">
                      ₦{safeToSave?.avg_monthly_expense}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span>Emergency Buffer</span>
                    <span className="font-medium">
                      ₦{safeToSave?.emergency_buffer}
                    </span>
                  </div>

                  <div className="flex justify-between">
                    <span>Active Goal Need</span>
                    <span className="font-medium">
                      ₦{safeToSave?.active_goal_need}
                    </span>
                  </div>
                </div>
              </div>

              <div className="rounded-2xl border bg-white p-6">
                <h3 className="text-xl font-semibold">How this recommendation was made</h3>
                <div className="mt-4 space-y-3 text-sm leading-7 text-gray-700">
                  <p>
                    <strong>Step 1:</strong> The system checks your available balance.
                  </p>
                  <p>
                    <strong>Step 2:</strong> It estimates your average monthly income and expenses.
                  </p>
                  <p>
                    <strong>Step 3:</strong> It keeps an emergency buffer aside first.
                  </p>
                  <p>
                    <strong>Step 4:</strong> It compares your surplus against your active goal need.
                  </p>
                  <p>
                    <strong>Step 5:</strong> The ML model proposes a value, then the rule-based guardrail caps it for safety.
                  </p>
                </div>
              </div>
            </div>

            <div className="rounded-2xl border bg-white p-6">
              <h3 className="text-xl font-semibold">Recommendation Reason</h3>
              <p className="mt-4 text-sm leading-7 text-gray-700">
                {safeToSave?.reason}
              </p>
            </div>

            <button
              onClick={() => mutate()}
              className="rounded-xl border px-4 py-3 font-medium hover:bg-gray-100"
            >
              Refresh Insight
            </button>
          </div>
        )}
      </div>
    </AppShell>
  );
}