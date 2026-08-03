import { useMemo } from "react";
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { api } from "../api/client";
import { useFetch } from "../lib/useFetch";
import { useRun } from "../state/RunContext";
import { useScreenShortcuts } from "../state/ShortcutContext";
import { ErrorNote } from "../components/ErrorNote";
import { formatInt, formatPct, formatUsd } from "../lib/format";

export function LedgerPage() {
  const { refreshTick } = useRun();
  const { data, loading, error, reload } = useFetch(() => api.ledger(), [refreshTick]);

  useScreenShortcuts([{ keys: "r", label: "reload" }], (e) => {
    if (e.key === "r") reload();
  });

  const byDay = useMemo(() => {
    const acc = new Map<string, number>();
    for (const row of data?.rows ?? []) {
      acc.set(row.day, (acc.get(row.day) ?? 0) + row.cost_usd);
    }
    return [...acc.entries()]
      .sort((a, b) => (a[0] < b[0] ? -1 : 1))
      .map(([day, cost]) => ({ day, cost: Number(cost.toFixed(4)) }));
  }, [data]);

  return (
    <div className="mx-auto max-w-5xl p-6">
      <h1 className="mb-4 text-lg font-semibold text-ink-100">Ledger</h1>
      {error && <ErrorNote error={error} onRetry={reload} />}
      {loading && !data && <div className="text-ink-500">Loading…</div>}

      {data && (
        <>
          <div className="mb-6 flex flex-wrap gap-4">
            <Stat
              label={data.unpriced_models?.length ? "Total cost (at least)" : "Total cost"}
              value={formatUsd(data.total_cost_usd)}
            />
            <Stat label="Total calls" value={formatInt(data.total_calls)} />
            <Stat label="Days tracked" value={formatInt(byDay.length)} />
          </div>

          {/* Spend the total does not include. A model absent from the pricing
              table used to be billed at zero, so a grounded run reported its web
              search as free — a wrong number that looked authoritative, which is
              worse than an absent one. */}
          {data.unpriced_models?.length ? (
            <div className="mb-6 rounded-lg border border-amber-500/40 bg-amber-500/5 p-3 text-[13px] text-amber-200/90">
              <strong className="font-semibold">
                {formatInt(data.unpriced_tokens ?? 0)} tokens are not in the total.
              </strong>{" "}
              No price is known for {data.unpriced_models.join(", ")}, so their spend is
              missing rather than zero — the figure above is a floor. Add them to{" "}
              <code className="font-mono text-[12px]">modelPricing</code> in{" "}
              <code className="font-mono text-[12px]">internal/studio/ledger.go</code> to
              close the gap.
            </div>
          ) : null}

          {byDay.length > 0 && (
            <div className="mb-8 h-64 rounded-lg border border-ink-800 bg-ink-900 p-3">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={byDay} margin={{ top: 8, right: 8, bottom: 0, left: 0 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#26262c" />
                  <XAxis dataKey="day" stroke="#71717c" fontSize={11} />
                  <YAxis stroke="#71717c" fontSize={11} tickFormatter={(v) => formatUsd(v as number)} />
                  <Tooltip
                    formatter={(v) => formatUsd(v as number)}
                    contentStyle={{
                      background: "#131316",
                      border: "1px solid #33333b",
                      borderRadius: 6,
                      fontSize: 12,
                    }}
                    labelStyle={{ color: "#c6c6cf" }}
                  />
                  <Bar dataKey="cost" fill="#38bdf8" radius={[3, 3, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}

          {data.quotas.length > 0 && (
            <div className="mb-8">
              <h2 className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-ink-500">
                Quotas
              </h2>
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {data.quotas.map((q) => {
                  const ratio = q.per_day > 0 ? q.day_used / q.per_day : 0;
                  return (
                    <div
                      key={q.provider}
                      className="rounded-lg border border-ink-800 bg-ink-900 p-3"
                    >
                      <div className="mb-1 flex items-center justify-between">
                        <span className="font-medium text-ink-200">{q.provider}</span>
                        <span className="text-[11px] text-ink-500">
                          {formatInt(q.day_used)}/{formatInt(q.per_day)} today
                        </span>
                      </div>
                      <div className="h-1.5 overflow-hidden rounded-full bg-ink-800">
                        <div
                          className={`h-full ${ratio >= 0.9 ? "bg-red-500" : ratio >= 0.7 ? "bg-amber-500" : "bg-emerald-500"}`}
                          style={{ width: `${Math.min(100, ratio * 100)}%` }}
                        />
                      </div>
                      <div className="mt-1 text-[11px] text-ink-500">
                        {formatPct(ratio)} of daily · {formatInt(q.per_minute)}/min limit
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          <h2 className="mb-2 text-[11px] font-semibold uppercase tracking-wide text-ink-500">
            By day · provider · model
          </h2>
          <div className="overflow-x-auto rounded-lg border border-ink-800">
            <table className="w-full text-left">
              <thead className="bg-ink-900 text-[11px] uppercase tracking-wide text-ink-500">
                <tr>
                  <th className="px-3 py-2 font-medium">Day</th>
                  <th className="px-3 py-2 font-medium">Provider</th>
                  <th className="px-3 py-2 font-medium">Model</th>
                  <th className="px-3 py-2 text-right font-medium">Calls</th>
                  <th className="px-3 py-2 text-right font-medium">Prompt</th>
                  <th className="px-3 py-2 text-right font-medium">Completion</th>
                  <th className="px-3 py-2 text-right font-medium">Cost</th>
                </tr>
              </thead>
              <tbody>
                {data.rows.map((r, i) => (
                  <tr key={`${r.day}-${r.provider}-${r.model}-${i}`} className="border-t border-ink-850">
                    <td className="px-3 py-1.5">{r.day}</td>
                    <td className="px-3 py-1.5 text-ink-300">{r.provider}</td>
                    <td className="px-3 py-1.5 font-mono text-[12px] text-ink-300">{r.model}</td>
                    <td className="px-3 py-1.5 text-right">{formatInt(r.calls)}</td>
                    <td className="px-3 py-1.5 text-right text-ink-400">{formatInt(r.prompt_tokens)}</td>
                    <td className="px-3 py-1.5 text-right text-ink-400">
                      {formatInt(r.completion_tokens)}
                    </td>
                    {/* Never $0.00 for a model we cannot price — the tokens were
                        really spent and somebody really paid for them. */}
                    <td className="px-3 py-1.5 text-right">
                      {r.priced ? (
                        formatUsd(r.cost_usd)
                      ) : (
                        <span className="text-amber-300/80" title="No price known for this model">
                          unpriced
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
                {data.rows.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-3 py-4 text-center text-ink-500">
                      No usage recorded yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-ink-800 bg-ink-900 px-4 py-3">
      <div className="text-[11px] uppercase tracking-wide text-ink-500">{label}</div>
      <div className="mt-0.5 text-xl font-semibold text-ink-100">{value}</div>
    </div>
  );
}
