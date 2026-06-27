"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { apiFetch } from "@/lib/api";

interface Choice { id: string; text: string }
interface TestCase { id: string; stdin: string; expected_stdout: string; is_example: boolean }

interface ProblemPayload {
  type: "multiple_choice" | "code";
  prompt: string;
  choices?: Choice[];
  multiple_correct?: boolean;
  languages?: string[];
  starter_code?: Record<string, string>;
  test_cases?: TestCase[];
}

interface Problem {
  id: string;
  type: "multiple_choice" | "code";
  title: string;
  points: number;
  payload: ProblemPayload;
}

interface AttemptAnswer {
  problem_id: string;
  problem_type: string;
  is_correct: boolean | null;
  points_awarded: number | null;
  eval_status: string | null;
}

interface Attempt {
  id: string;
  status: "in_progress" | "grading" | "graded";
  score: number | null;
  max_score: number;
  quiz: { title: string; description: string | null; problems: Problem[] };
  answers: AttemptAnswer[];
}

type LocalAnswer =
  | { type: "multiple_choice"; selected: string[] }
  | { type: "code"; language: string; source: string };

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

function AttemptRunner() {
  const params = useSearchParams();
  const quizId = params.get("quiz");
  const { getToken } = useAuth();

  const [attempt, setAttempt] = useState<Attempt | null>(null);
  const [answers, setAnswers] = useState<Record<string, LocalAnswer>>({});
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const initAnswers = (problems: Problem[]) => {
    const init: Record<string, LocalAnswer> = {};
    for (const p of problems) {
      if (p.type === "multiple_choice") {
        init[p.id] = { type: "multiple_choice", selected: [] };
      } else {
        const lang = p.payload.languages?.[0] ?? "python";
        init[p.id] = { type: "code", language: lang, source: p.payload.starter_code?.[lang] ?? "" };
      }
    }
    return init;
  };

  const start = useCallback(async () => {
    if (!quizId) {
      setError("No quiz specified. Open this page from your dashboard.");
      setLoading(false);
      return;
    }
    try {
      const token = await getToken();
      const a = await apiFetch<Attempt>(token, "POST", "/attempts/", { quiz_id: quizId });
      setAttempt(a);
      setAnswers(initAnswers(a.quiz.problems));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start attempt");
    } finally {
      setLoading(false);
    }
  }, [quizId, getToken]);

  useEffect(() => {
    start();
  }, [start]);

  async function submit() {
    if (!attempt) return;
    setSubmitting(true);
    setError(null);
    try {
      const token = await getToken();

      for (const p of attempt.quiz.problems) {
        const a = answers[p.id];
        if (a.type === "multiple_choice") {
          if (a.selected.length === 0) continue;
          await apiFetch(token, "PUT", `/attempts/${attempt.id}/answers/${p.id}`, {
            type: "multiple_choice",
            selected_choice_ids: a.selected,
          });
        } else {
          if (!a.source.trim()) continue;
          await apiFetch(token, "PUT", `/attempts/${attempt.id}/answers/${p.id}`, {
            type: "code",
            language: a.language,
            source_code: a.source,
          });
        }
      }

      let current = await apiFetch<Attempt>(token, "POST", `/attempts/${attempt.id}/submit`);

      for (let i = 0; i < 6 && current.status !== "graded"; i++) {
        await sleep(1500);
        current = await apiFetch<Attempt>(token, "GET", `/attempts/${attempt.id}`);
      }
      setAttempt(current);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to submit");
    } finally {
      setSubmitting(false);
    }
  }

  if (loading) return <Centered>Starting attempt…</Centered>;
  if (error && !attempt)
    return (
      <Centered>
        <p className="text-destructive">{error}</p>
      </Centered>
    );
  if (!attempt) return null;

  const done = attempt.status === "graded";
  const grading = attempt.status === "grading";
  const resultById = Object.fromEntries(attempt.answers.map((a) => [a.problem_id, a]));

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10">
      <h1 className="text-2xl font-semibold">{attempt.quiz.title}</h1>
      {attempt.quiz.description && (
        <p className="mt-1 text-sm text-muted-foreground">{attempt.quiz.description}</p>
      )}

      {done && (
        <p className="mt-4 rounded-lg border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm font-medium text-emerald-500">
          Graded — score {attempt.score}/{attempt.max_score}
        </p>
      )}
      {grading && (
        <p className="mt-4 rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-sm text-amber-500">
          Still grading code answers. Refresh shortly if results are incomplete.
        </p>
      )}

      <div className="mt-6 grid gap-4">
        {attempt.quiz.problems.map((p, i) => {
          const result = resultById[p.id];
          return (
            <section key={p.id} className="grid gap-3 rounded-lg border border-border p-4">
              <div className="flex items-center justify-between">
                <h2 className="font-medium">
                  {i + 1}. {p.title}{" "}
                  <span className="text-xs text-muted-foreground">({p.points} pts)</span>
                </h2>
                {result && (
                  <span
                    className={
                      result.is_correct
                        ? "text-xs font-semibold text-emerald-500"
                        : "text-xs font-semibold text-destructive"
                    }
                  >
                    {result.is_correct ? "Correct" : "Incorrect"} · {result.points_awarded ?? 0} pts
                  </span>
                )}
              </div>
              <p className="whitespace-pre-wrap text-sm">{p.payload.prompt}</p>

              {p.type === "multiple_choice" ? (
                <McQuestion
                  problem={p}
                  answer={answers[p.id] as Extract<LocalAnswer, { type: "multiple_choice" }>}
                  disabled={done || grading || submitting}
                  onChange={(selected) =>
                    setAnswers((s) => ({ ...s, [p.id]: { type: "multiple_choice", selected } }))
                  }
                />
              ) : (
                <CodeQuestion
                  problem={p}
                  answer={answers[p.id] as Extract<LocalAnswer, { type: "code" }>}
                  disabled={done || grading || submitting}
                  onChange={(next) =>
                    setAnswers((s) => ({ ...s, [p.id]: { ...(s[p.id] as Extract<LocalAnswer, { type: "code" }>), ...next } }))
                  }
                />
              )}
            </section>
          );
        })}
      </div>

      {error && (
        <p className="mt-4 rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </p>
      )}

      {!done && (
        <div className="mt-6">
          <Button onClick={submit} disabled={submitting || grading}>
            {submitting ? "Submitting…" : "Submit attempt"}
          </Button>
        </div>
      )}
    </main>
  );
}

function McQuestion({
  problem,
  answer,
  disabled,
  onChange,
}: {
  problem: Problem;
  answer: Extract<LocalAnswer, { type: "multiple_choice" }>;
  disabled: boolean;
  onChange: (selected: string[]) => void;
}) {
  const multiple = problem.payload.multiple_correct ?? false;
  const toggle = (id: string) => {
    if (multiple) {
      onChange(answer.selected.includes(id) ? answer.selected.filter((x) => x !== id) : [...answer.selected, id]);
    } else {
      onChange([id]);
    }
  };
  return (
    <div className="grid gap-1.5">
      {(problem.payload.choices ?? []).map((c) => (
        <label key={c.id} className="flex items-center gap-2 text-sm">
          <input
            type={multiple ? "checkbox" : "radio"}
            name={problem.id}
            checked={answer.selected.includes(c.id)}
            disabled={disabled}
            onChange={() => toggle(c.id)}
          />
          {c.text}
        </label>
      ))}
    </div>
  );
}

function CodeQuestion({
  problem,
  answer,
  disabled,
  onChange,
}: {
  problem: Problem;
  answer: Extract<LocalAnswer, { type: "code" }>;
  disabled: boolean;
  onChange: (next: Partial<Extract<LocalAnswer, { type: "code" }>>) => void;
}) {
  const languages = problem.payload.languages ?? [];
  const examples = problem.payload.test_cases ?? [];
  return (
    <div className="grid gap-2">
      <select
        value={answer.language}
        disabled={disabled}
        onChange={(e) => {
          const lang = e.target.value;
          onChange({ language: lang, source: problem.payload.starter_code?.[lang] ?? answer.source });
        }}
        className="h-8 w-40 rounded-lg border border-border bg-background px-2 text-sm"
      >
        {languages.map((l) => (
          <option key={l} value={l}>{l}</option>
        ))}
      </select>
      <textarea
        value={answer.source}
        disabled={disabled}
        onChange={(e) => onChange({ source: e.target.value })}
        spellCheck={false}
        rows={8}
        className="w-full rounded-lg border border-border bg-background p-3 font-mono text-xs"
        placeholder="Write your solution…"
      />
      {examples.length > 0 && (
        <details className="text-xs text-muted-foreground">
          <summary className="cursor-pointer">Example test cases ({examples.length})</summary>
          <div className="mt-2 grid gap-2">
            {examples.map((t) => (
              <div key={t.id} className="grid grid-cols-2 gap-2">
                <pre className="overflow-auto rounded border border-border/60 p-2">stdin:{"\n"}{t.stdin}</pre>
                <pre className="overflow-auto rounded border border-border/60 p-2">expected:{"\n"}{t.expected_stdout}</pre>
              </div>
            ))}
          </div>
        </details>
      )}
    </div>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-20 text-center text-sm text-muted-foreground">
      {children}
    </main>
  );
}

export default function AttemptPage() {
  return (
    <Suspense fallback={<Centered>Loading…</Centered>}>
      <AttemptRunner />
    </Suspense>
  );
}
