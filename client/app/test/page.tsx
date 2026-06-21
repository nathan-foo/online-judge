"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";
import { Button } from "@/components/ui/button";

// Browser -> gateway. The gateway strips the route prefix and proxies to the
// quiz/attempt services, validating the Clerk bearer token. Override via env
// when the gateway is not on localhost:8080.
const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

const LANGUAGES = ["python", "c", "cpp", "java", "javascript", "go", "typescript", "kotlin", "rust", "csharp"] as const;
type Lang = (typeof LANGUAGES)[number];

// Realistic "double the number" solutions. The code-eval-service runs these for
// real in a sandboxed exec pod, so a correct solution grades as fully passing.
const STARTER: Record<Lang, string> = {
  python: "n = int(input())\nprint(n * 2)\n",
  c: '#include <stdio.h>\nint main() { int n; scanf("%d", &n); printf("%d", n * 2); }\n',
  cpp: "#include <iostream>\nint main() { int n; std::cin >> n; std::cout << n * 2; }\n",
  java:
    "import java.util.*;\npublic class Main {\n  public static void main(String[] a) {\n    System.out.print(new Scanner(System.in).nextInt() * 2);\n  }\n}\n",
  javascript:
    "const n = parseInt(require('fs').readFileSync(0, 'utf8'));\nconsole.log(n * 2);\n",
  go: 'package main\nimport "fmt"\nfunc main() { var n int; fmt.Scan(&n); fmt.Print(n * 2) }\n',
  typescript:
    "const n = parseInt(require('fs').readFileSync(0, 'utf8'));\nconsole.log(n * 2);\n",
  kotlin: "fun main() { print(readLine()!!.trim().toInt() * 2) }\n",
  rust: "use std::io::*;\nfn main() {\n  let mut s = String::new();\n  stdin().read_line(&mut s).unwrap();\n  print!(\"{}\", s.trim().parse::<i64>().unwrap() * 2);\n}\n",
  csharp:
    "using System;\nclass Program {\n  static void Main() {\n    Console.Write(int.Parse(Console.ReadLine()) * 2);\n  }\n}\n",
};

type LogLevel = "info" | "request" | "response" | "success" | "warn" | "error";

interface LogEntry {
  id: number;
  time: string;
  level: LogLevel;
  msg: string;
  data?: unknown;
}

interface Quiz {
  id: string;
  version: number;
}

interface AttemptProblem {
  id: string;
  type: string;
}

interface AttemptAnswer {
  problem_id: string;
  problem_type: string;
  is_correct: boolean | null;
  points_awarded: number | null;
  eval_status: string | null;
  eval_result: unknown;
}

interface Attempt {
  id: string;
  status: string;
  score: number | null;
  max_score: number;
  quiz: { problems: AttemptProblem[] };
  answers: AttemptAnswer[];
}

const LEVEL_STYLES: Record<LogLevel, string> = {
  info: "text-muted-foreground",
  request: "text-blue-500",
  response: "text-emerald-500",
  success: "text-emerald-400 font-semibold",
  warn: "text-amber-500",
  error: "text-destructive font-semibold",
};

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export default function TestFlowPage() {
  const { getToken } = useAuth();
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [running, setRunning] = useState(false);
  const [language, setLanguage] = useState<Lang>("python");
  const [source, setSource] = useState<string>(STARTER.python);
  const [result, setResult] = useState<Attempt | null>(null);

  let logSeq = 0;
  const log = (level: LogLevel, msg: string, data?: unknown) => {
    const entry: LogEntry = {
      id: ++logSeq + Date.now(),
      time: new Date().toLocaleTimeString(),
      level,
      msg,
      data,
    };
    setLogs((prev) => [...prev, entry]);
    // Mirror to the devtools console too.
    const fn = level === "error" ? console.error : level === "warn" ? console.warn : console.log;
    fn(`[test-flow] ${msg}`, data ?? "");
  };

  const onLanguageChange = (next: Lang) => {
    setLanguage(next);
    setSource(STARTER[next]);
  };

  async function api<T>(
    token: string,
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    log("request", `${method} ${path}`, body);
    const res = await fetch(`${API_BASE}${path}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    const text = await res.text();
    let parsed: unknown = text || null;
    try {
      parsed = text ? JSON.parse(text) : null;
    } catch {
      // leave parsed as raw text
    }

    if (!res.ok) {
      log("error", `${method} ${path} -> ${res.status}`, parsed);
      throw new Error(`${method} ${path} failed (${res.status})`);
    }
    log("response", `${method} ${path} -> ${res.status}`, parsed);
    return parsed as T;
  }

  async function runFlow() {
    setRunning(true);
    setResult(null);
    setLogs([]);

    try {
      log("info", "Fetching Clerk session token…");
      const token = await getToken();
      if (!token) {
        log("error", "No auth token — are you signed in?");
        return;
      }
      log("info", `Using gateway at ${API_BASE}`);

      // 1. Create a mock quiz with one MCQ + one code problem.
      const quizBody = {
        title: "Mock Quiz (test flow)",
        description: "Auto-generated by /test",
        is_public: false,
        problems: [
          {
            title: "Arithmetic",
            position: 1,
            points: 1000,
            payload: {
              type: "multiple_choice",
              prompt: "What is 2 + 2?",
              choices: [
                { id: "a", text: "3" },
                { id: "b", text: "4" },
                { id: "c", text: "5" },
              ],
              correct_choice_ids: ["b"],
              multiple_correct: false,
            },
          },
          {
            title: "Double the number",
            position: 2,
            points: 1000,
            payload: {
              type: "code",
              prompt: "Read an integer n from stdin and print n * 2.",
              languages: [...LANGUAGES],
              test_cases: [
                { id: "t1", stdin: "2\n", expected_stdout: "4", is_example: true },
                { id: "t2", stdin: "5\n", expected_stdout: "10", is_example: false },
              ],
              time_limit_ms: 2000,
              memory_limit_mb: 256,
            },
          },
        ],
      };
      log("info", "Step 1 — create quiz");
      const quiz = await api<Quiz>(token, "POST", "/quizzes/", quizBody);

      // 2. Publish it — an attempt can only snapshot a published quiz.
      log("info", "Step 2 — publish quiz");
      await api(token, "POST", `/quizzes/${quiz.id}/publish`, { version: quiz.version });

      // 3. Start an attempt (returns the taker view with server-assigned problem ids).
      log("info", "Step 3 — start attempt");
      const attempt = await api<Attempt>(token, "POST", "/attempts/", { quiz_id: quiz.id });

      const mcq = attempt.quiz.problems.find((p) => p.type === "multiple_choice");
      const code = attempt.quiz.problems.find((p) => p.type === "code");
      if (!mcq || !code) {
        log("error", "Could not find both problems in attempt snapshot", attempt.quiz.problems);
        return;
      }
      log("info", `Mapped problems — mcq=${mcq.id} code=${code.id}`);

      // 4. Answer the MCQ (correct choice).
      log("info", "Step 4 — answer MCQ");
      await api(token, "PUT", `/attempts/${attempt.id}/answers/${mcq.id}`, {
        type: "multiple_choice",
        selected_choice_ids: ["b"],
      });

      // 5. Answer the code question in the selected language.
      log("info", `Step 5 — answer code (${language})`);
      await api(token, "PUT", `/attempts/${attempt.id}/answers/${code.id}`, {
        type: "code",
        language,
        source_code: source,
      });

      // 6. Submit — MCQ grades synchronously; the code answer is queued to the
      //    code-eval-service, so the attempt enters "grading".
      log("info", "Step 6 — submit attempt");
      let current = await api<Attempt>(token, "POST", `/attempts/${attempt.id}/submit`);

      // 7. Poll until graded (kept under the 10 req/min route limit).
      log("info", "Step 7 — poll until graded");
      for (let i = 0; i < 5 && current.status !== "graded"; i++) {
        await sleep(1500);
        current = await api<Attempt>(token, "GET", `/attempts/${attempt.id}`);
        log("info", `Poll ${i + 1} — status=${current.status}`);
      }

      setResult(current);
      if (current.status === "graded") {
        log("success", `Graded — score ${current.score}/${current.max_score}`, current.answers);
      } else {
        log(
          "warn",
          `Still "${current.status}" after polling. Is the code-eval-service running and consuming code_eval.requests?`,
        );
      }
    } catch (err) {
      log("error", "Flow aborted", err instanceof Error ? err.message : err);
    } finally {
      setRunning(false);
    }
  }

  return (
    <main className="mx-auto w-full max-w-4xl px-6 py-10">
      <h1 className="text-2xl font-semibold">End-to-end flow test</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Creates a mock quiz (1 MCQ + 1 code), starts an attempt, answers both, submits, and
        polls for the graded result. Every request and response is logged below.
      </p>

      <section className="mt-6 grid gap-4 rounded-lg border border-border p-4">
        <div className="flex flex-wrap items-center gap-3">
          <label className="text-sm font-medium" htmlFor="lang">
            Code language
          </label>
          <select
            id="lang"
            value={language}
            disabled={running}
            onChange={(e) => onLanguageChange(e.target.value as Lang)}
            className="h-8 rounded-lg border border-border bg-background px-2 text-sm"
          >
            {LANGUAGES.map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
        </div>
        <textarea
          value={source}
          disabled={running}
          onChange={(e) => setSource(e.target.value)}
          spellCheck={false}
          rows={6}
          className="w-full rounded-lg border border-border bg-background p-3 font-mono text-xs"
        />
        <div className="flex gap-2">
          <Button onClick={runFlow} disabled={running}>
            {running ? "Running…" : "Run full flow"}
          </Button>
          <Button variant="outline" onClick={() => setLogs([])} disabled={running}>
            Clear logs
          </Button>
        </div>
      </section>

      {result && (
        <section className="mt-6 rounded-lg border border-border p-4">
          <h2 className="text-lg font-medium">Result</h2>
          <p className="mt-1 text-sm">
            Status <span className="font-mono">{result.status}</span> — score{" "}
            <span className="font-mono">
              {result.score ?? "—"}/{result.max_score}
            </span>
          </p>
          <table className="mt-3 w-full text-left text-sm">
            <thead className="text-muted-foreground">
              <tr>
                <th className="py-1 pr-4 font-medium">Type</th>
                <th className="py-1 pr-4 font-medium">Correct</th>
                <th className="py-1 pr-4 font-medium">Points</th>
                <th className="py-1 pr-4 font-medium">Eval status</th>
              </tr>
            </thead>
            <tbody className="font-mono">
              {result.answers.map((a) => (
                <tr key={a.problem_id} className="border-t border-border">
                  <td className="py-1 pr-4">{a.problem_type}</td>
                  <td className="py-1 pr-4">{String(a.is_correct)}</td>
                  <td className="py-1 pr-4">{a.points_awarded ?? "—"}</td>
                  <td className="py-1 pr-4">{a.eval_status ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </section>
      )}

      <section className="mt-6">
        <h2 className="text-lg font-medium">Logs</h2>
        <div className="mt-2 max-h-[28rem] overflow-auto rounded-lg border border-border bg-muted/30 p-3 font-mono text-xs">
          {logs.length === 0 ? (
            <p className="text-muted-foreground">No logs yet — run the flow.</p>
          ) : (
            logs.map((e) => (
              <div key={e.id} className="border-b border-border/40 py-1 last:border-0">
                <span className="text-muted-foreground">{e.time} </span>
                <span className={LEVEL_STYLES[e.level]}>
                  [{e.level}] {e.msg}
                </span>
                {e.data !== undefined && e.data !== null && (
                  <pre className="mt-1 overflow-auto whitespace-pre-wrap text-muted-foreground">
                    {typeof e.data === "string" ? e.data : JSON.stringify(e.data, null, 2)}
                  </pre>
                )}
              </div>
            ))
          )}
        </div>
      </section>
    </main>
  );
}
