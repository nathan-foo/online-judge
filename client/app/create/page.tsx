"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";
import { apiFetch } from "@/lib/api";

const LANGUAGES = [
  "python", "c", "cpp", "java", "javascript", "go", "typescript", "kotlin", "rust", "csharp",
] as const;
type Lang = (typeof LANGUAGES)[number];

type Choice = { id: string; text: string };
type TestCase = { id: string; stdin: string; expected_stdout: string; is_example: boolean };

type McProblem = {
  kind: "multiple_choice";
  title: string;
  points: number;
  prompt: string;
  choices: Choice[];
  correctIds: string[];
  multipleCorrect: boolean;
};

type CodeProblem = {
  kind: "code";
  title: string;
  points: number;
  prompt: string;
  languages: Lang[];
  testCases: TestCase[];
  timeLimitMs: number;
  memoryLimitMb: number;
};

type Problem = McProblem | CodeProblem;

let seq = 0;
const nextId = (p: string) => `${p}${++seq}`;

function newMcProblem(): McProblem {
  const a = nextId("c");
  const b = nextId("c");
  return {
    kind: "multiple_choice",
    title: "",
    points: 1000,
    prompt: "",
    choices: [
      { id: a, text: "" },
      { id: b, text: "" },
    ],
    correctIds: [a],
    multipleCorrect: false,
  };
}

function newCodeProblem(): CodeProblem {
  return {
    kind: "code",
    title: "",
    points: 1000,
    prompt: "",
    languages: ["python"],
    testCases: [{ id: nextId("t"), stdin: "", expected_stdout: "", is_example: true }],
    timeLimitMs: 2000,
    memoryLimitMb: 256,
  };
}

export default function CreateQuizPage() {
  const router = useRouter();
  const { getToken } = useAuth();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [isPublic, setIsPublic] = useState(false);
  const [problems, setProblems] = useState<Problem[]>([newMcProblem()]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const patch = (i: number, next: Partial<Problem>) =>
    setProblems((ps) => ps.map((p, j) => (j === i ? ({ ...p, ...next } as Problem) : p)));

  function buildPayload(p: Problem) {
    if (p.kind === "multiple_choice") {
      return {
        type: "multiple_choice",
        prompt: p.prompt,
        choices: p.choices,
        correct_choice_ids: p.correctIds,
        multiple_correct: p.multipleCorrect,
      };
    }
    return {
      type: "code",
      prompt: p.prompt,
      languages: p.languages,
      test_cases: p.testCases,
      time_limit_ms: p.timeLimitMs,
      memory_limit_mb: p.memoryLimitMb,
    };
  }

  async function submit() {
    setSaving(true);
    setError(null);
    try {
      const token = await getToken();
      const body = {
        title,
        description: description || null,
        is_public: isPublic,
        problems: problems.map((p, i) => ({
          title: p.title,
          position: i + 1,
          points: p.points,
          payload: buildPayload(p),
        })),
      };
      await apiFetch(token, "POST", "/quizzes/", body);
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create quiz");
      setSaving(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-10">
      <h1 className="text-2xl font-semibold">Create quiz</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Add problems, then create. Publish it from your dashboard to make it attemptable.
      </p>

      <section className="mt-6 grid gap-3 rounded-lg border border-border p-4">
        <label className="grid gap-1 text-sm font-medium">
          Title
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="h-8 rounded-lg border border-border bg-background px-2 text-sm font-normal"
            placeholder="My quiz"
          />
        </label>
        <label className="grid gap-1 text-sm font-medium">
          Description
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            className="rounded-lg border border-border bg-background p-2 text-sm font-normal"
            placeholder="Optional"
          />
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={isPublic} onChange={(e) => setIsPublic(e.target.checked)} />
          Public (listed once published)
        </label>
      </section>

      <div className="mt-6 flex items-center justify-between">
        <h2 className="text-lg font-medium">Problems</h2>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => setProblems((ps) => [...ps, newMcProblem()])}>
            + Multiple choice
          </Button>
          <Button variant="outline" size="sm" onClick={() => setProblems((ps) => [...ps, newCodeProblem()])}>
            + Code
          </Button>
        </div>
      </div>

      <div className="mt-3 grid gap-4">
        {problems.map((p, i) => (
          <section key={i} className="grid gap-3 rounded-lg border border-border p-4">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase text-muted-foreground">
                #{i + 1} · {p.kind === "code" ? "Code" : "Multiple choice"}
              </span>
              <Button
                variant="destructive"
                size="xs"
                onClick={() => setProblems((ps) => ps.filter((_, j) => j !== i))}
                disabled={problems.length === 1}
              >
                Remove
              </Button>
            </div>

            <div className="flex gap-2">
              <input
                value={p.title}
                onChange={(e) => patch(i, { title: e.target.value })}
                placeholder="Problem title"
                className="h-8 flex-1 rounded-lg border border-border bg-background px-2 text-sm"
              />
              <input
                type="number"
                value={p.points}
                onChange={(e) => patch(i, { points: Number(e.target.value) })}
                className="h-8 w-24 rounded-lg border border-border bg-background px-2 text-sm"
                title="Points"
              />
            </div>

            <textarea
              value={p.prompt}
              onChange={(e) => patch(i, { prompt: e.target.value })}
              rows={2}
              placeholder="Prompt"
              className="rounded-lg border border-border bg-background p-2 text-sm"
            />

            {p.kind === "multiple_choice" ? (
              <McEditor problem={p} onChange={(next) => patch(i, next)} />
            ) : (
              <CodeEditor problem={p} onChange={(next) => patch(i, next)} />
            )}
          </section>
        ))}
      </div>

      {error && (
        <p className="mt-4 rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </p>
      )}

      <div className="mt-6 flex gap-2">
        <Button onClick={submit} disabled={saving || !title || problems.length === 0}>
          {saving ? "Creating…" : "Create quiz"}
        </Button>
        <Button variant="outline" onClick={() => router.push("/dashboard")} disabled={saving}>
          Cancel
        </Button>
      </div>
    </main>
  );
}

function McEditor({ problem, onChange }: { problem: McProblem; onChange: (p: Partial<McProblem>) => void }) {
  const { choices, correctIds, multipleCorrect } = problem;

  const toggleCorrect = (id: string) => {
    if (multipleCorrect) {
      onChange({
        correctIds: correctIds.includes(id)
          ? correctIds.filter((c) => c !== id)
          : [...correctIds, id],
      });
    } else {
      onChange({ correctIds: [id] });
    }
  };

  return (
    <div className="grid gap-2">
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={multipleCorrect}
          onChange={(e) => onChange({ multipleCorrect: e.target.checked, correctIds: correctIds.slice(0, 1) })}
        />
        Allow multiple correct answers
      </label>
      {choices.map((c) => (
        <div key={c.id} className="flex items-center gap-2">
          <input
            type={multipleCorrect ? "checkbox" : "radio"}
            checked={correctIds.includes(c.id)}
            onChange={() => toggleCorrect(c.id)}
            title="Correct?"
          />
          <input
            value={c.text}
            onChange={(e) =>
              onChange({ choices: choices.map((x) => (x.id === c.id ? { ...x, text: e.target.value } : x)) })
            }
            placeholder="Choice text"
            className="h-8 flex-1 rounded-lg border border-border bg-background px-2 text-sm"
          />
          <Button
            variant="ghost"
            size="xs"
            disabled={choices.length <= 2}
            onClick={() =>
              onChange({
                choices: choices.filter((x) => x.id !== c.id),
                correctIds: correctIds.filter((x) => x !== c.id),
              })
            }
          >
            ✕
          </Button>
        </div>
      ))}
      <Button
        variant="outline"
        size="xs"
        className="justify-self-start"
        disabled={choices.length >= 10}
        onClick={() => onChange({ choices: [...choices, { id: nextId("c"), text: "" }] })}
      >
        + Choice
      </Button>
    </div>
  );
}

function CodeEditor({ problem, onChange }: { problem: CodeProblem; onChange: (p: Partial<CodeProblem>) => void }) {
  const { languages, testCases, timeLimitMs, memoryLimitMb } = problem;

  const toggleLang = (l: Lang) =>
    onChange({
      languages: languages.includes(l) ? languages.filter((x) => x !== l) : [...languages, l],
    });

  return (
    <div className="grid gap-3">
      <div>
        <p className="mb-1 text-sm font-medium">Languages</p>
        <div className="flex flex-wrap gap-2">
          {LANGUAGES.map((l) => (
            <label key={l} className="flex items-center gap-1 text-xs">
              <input type="checkbox" checked={languages.includes(l)} onChange={() => toggleLang(l)} />
              {l}
            </label>
          ))}
        </div>
      </div>

      <div className="flex gap-2">
        <label className="grid gap-1 text-xs font-medium">
          Time limit (ms)
          <input
            type="number"
            value={timeLimitMs}
            onChange={(e) => onChange({ timeLimitMs: Number(e.target.value) })}
            className="h-8 w-28 rounded-lg border border-border bg-background px-2 text-sm font-normal"
          />
        </label>
        <label className="grid gap-1 text-xs font-medium">
          Memory limit (MB)
          <input
            type="number"
            value={memoryLimitMb}
            onChange={(e) => onChange({ memoryLimitMb: Number(e.target.value) })}
            className="h-8 w-28 rounded-lg border border-border bg-background px-2 text-sm font-normal"
          />
        </label>
      </div>

      <div className="grid gap-2">
        <p className="text-sm font-medium">Test cases</p>
        {testCases.map((t) => (
          <div key={t.id} className="grid gap-2 rounded-lg border border-border/60 p-2">
            <div className="grid grid-cols-2 gap-2">
              <textarea
                value={t.stdin}
                onChange={(e) =>
                  onChange({ testCases: testCases.map((x) => (x.id === t.id ? { ...x, stdin: e.target.value } : x)) })
                }
                rows={2}
                placeholder="stdin"
                className="rounded-lg border border-border bg-background p-2 font-mono text-xs"
              />
              <textarea
                value={t.expected_stdout}
                onChange={(e) =>
                  onChange({
                    testCases: testCases.map((x) =>
                      x.id === t.id ? { ...x, expected_stdout: e.target.value } : x,
                    ),
                  })
                }
                rows={2}
                placeholder="expected stdout"
                className="rounded-lg border border-border bg-background p-2 font-mono text-xs"
              />
            </div>
            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-xs">
                <input
                  type="checkbox"
                  checked={t.is_example}
                  onChange={(e) =>
                    onChange({
                      testCases: testCases.map((x) =>
                        x.id === t.id ? { ...x, is_example: e.target.checked } : x,
                      ),
                    })
                  }
                />
                Example (shown to taker)
              </label>
              <Button
                variant="ghost"
                size="xs"
                disabled={testCases.length <= 1}
                onClick={() => onChange({ testCases: testCases.filter((x) => x.id !== t.id) })}
              >
                ✕
              </Button>
            </div>
          </div>
        ))}
        <Button
          variant="outline"
          size="xs"
          className="justify-self-start"
          disabled={testCases.length >= 50}
          onClick={() =>
            onChange({
              testCases: [...testCases, { id: nextId("t"), stdin: "", expected_stdout: "", is_example: false }],
            })
          }
        >
          + Test case
        </Button>
      </div>
    </div>
  );
}
