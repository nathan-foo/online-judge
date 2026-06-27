"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { apiFetch } from "@/lib/api";

interface QuizSummary {
  id: string;
  title: string;
  description: string | null;
  is_published: boolean;
  is_public: boolean;
  problem_count: number;
  version: number;
  created_at: string;
  updated_at: string;
}

export default function DashboardPage() {
  const { getToken } = useAuth();
  const [quizzes, setQuizzes] = useState<QuizSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const token = await getToken();
      setQuizzes(await apiFetch<QuizSummary[]>(token, "GET", "/quizzes/"));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load quizzes");
    } finally {
      setLoading(false);
    }
  }, [getToken]);

  useEffect(() => {
    load();
  }, [load]);

  async function publish(q: QuizSummary) {
    setBusy(q.id);
    setError(null);
    try {
      const token = await getToken();
      await apiFetch(token, "POST", `/quizzes/${q.id}/publish`, { version: q.version });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to publish");
    } finally {
      setBusy(null);
    }
  }

  async function remove(q: QuizSummary) {
    if (!confirm(`Delete "${q.title}"?`)) return;
    setBusy(q.id);
    setError(null);
    try {
      const token = await getToken();
      await apiFetch(token, "DELETE", `/quizzes/${q.id}`);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    } finally {
      setBusy(null);
    }
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Your quizzes</h1>
        <Button asChild>
          <Link href="/create">New quiz</Link>
        </Button>
      </div>

      {error && (
        <p className="mt-4 rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </p>
      )}

      {loading ? (
        <p className="mt-6 text-sm text-muted-foreground">Loading…</p>
      ) : quizzes.length === 0 ? (
        <p className="mt-6 text-sm text-muted-foreground">
          No quizzes yet. <Link href="/create" className="underline">Create one</Link>.
        </p>
      ) : (
        <ul className="mt-6 grid gap-3">
          {quizzes.map((q) => (
            <li key={q.id} className="rounded-lg border border-border p-4">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="font-medium">{q.title}</h2>
                  {q.description && (
                    <p className="mt-0.5 text-sm text-muted-foreground">{q.description}</p>
                  )}
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
                    <span>{q.problem_count} problem{q.problem_count === 1 ? "" : "s"}</span>
                    <span>·</span>
                    <span className={q.is_published ? "text-emerald-500" : ""}>
                      {q.is_published ? "Published" : "Draft"}
                    </span>
                    {q.is_public && <><span>·</span><span>Public</span></>}
                  </div>
                </div>
                <div className="flex shrink-0 gap-2">
                  {q.is_published ? (
                    <Button asChild size="sm">
                      <Link href={`/attempt?quiz=${q.id}`}>Attempt</Link>
                    </Button>
                  ) : (
                    <Button size="sm" onClick={() => publish(q)} disabled={busy === q.id}>
                      {busy === q.id ? "…" : "Publish"}
                    </Button>
                  )}
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={() => remove(q)}
                    disabled={busy === q.id}
                  >
                    Delete
                  </Button>
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}
