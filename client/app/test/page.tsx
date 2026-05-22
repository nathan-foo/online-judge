"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";

import { Button } from "@/components/ui/button";

type Method = "GET" | "POST" | "PATCH" | "DELETE";

type ServiceConfig = {
  label: string;
  path: string;
  method: Method;
  defaultBody?: string;
};

type ServiceState = {
  body: string | null;
  error: string | null;
  loading: boolean;
  status: number | null;
};

const USER_PATCH_BODY = `{
  "username": "new_username"
}`;

const PROBLEM_CREATE_BODY = `{
  "title": "Sample MC question",
  "payload": {
    "type": "multiple_choice",
    "prompt": "What is 2 + 2?",
    "choices": [
      { "id": "a", "text": "3" },
      { "id": "b", "text": "4" },
      { "id": "c", "text": "5" }
    ],
    "correct_choice_ids": ["b"]
  }
}`;

const PROBLEM_PATCH_BODY = `{
  "title": "Renamed problem"
}`;

const QUIZ_CREATE_BODY = `{
  "title": "Sample quiz",
  "description": "A quiz",
  "is_public": false,
  "problems": []
}`;

const QUIZ_PATCH_BODY = `{
  "title": "Renamed quiz"
}`;

const SERVICES: Record<string, ServiceConfig> = {
  test: {
    label: "Test Service 1",
    path: "/test",
    method: "GET",
  },
  userGet: {
    label: "User Service - GET /me",
    path: "/users/me",
    method: "GET",
  },
  userPatch: {
    label: "User Service - PATCH /me",
    path: "/users/me",
    method: "PATCH",
    defaultBody: USER_PATCH_BODY,
  },
  problemCreate: {
    label: "Problem - Create",
    path: "/content/problems/",
    method: "POST",
    defaultBody: PROBLEM_CREATE_BODY,
  },
  problemList: {
    label: "Problem - List",
    path: "/content/problems/",
    method: "GET",
  },
  problemGet: {
    label: "Problem - Get",
    path: "/content/problems/:id",
    method: "GET",
  },
  problemPatch: {
    label: "Problem - Patch",
    path: "/content/problems/:id",
    method: "PATCH",
    defaultBody: PROBLEM_PATCH_BODY,
  },
  problemDelete: {
    label: "Problem - Delete",
    path: "/content/problems/:id",
    method: "DELETE",
  },
  quizCreate: {
    label: "Quiz - Create",
    path: "/content/quizzes/",
    method: "POST",
    defaultBody: QUIZ_CREATE_BODY,
  },
  quizList: {
    label: "Quiz - List",
    path: "/content/quizzes/",
    method: "GET",
  },
  quizListPublic: {
    label: "Quiz - List public",
    path: "/content/quizzes/public",
    method: "GET",
  },
  quizGet: {
    label: "Quiz - Get",
    path: "/content/quizzes/:id",
    method: "GET",
  },
  quizPatch: {
    label: "Quiz - Patch",
    path: "/content/quizzes/:id",
    method: "PATCH",
    defaultBody: QUIZ_PATCH_BODY,
  },
  quizDelete: {
    label: "Quiz - Delete",
    path: "/content/quizzes/:id",
    method: "DELETE",
  },
  quizPublish: {
    label: "Quiz - Publish",
    path: "/content/quizzes/:id/publish",
    method: "POST",
  },
};

type ServiceKey = keyof typeof SERVICES;

const SERVICE_KEYS = Object.keys(SERVICES) as ServiceKey[];

function buildInitial<T>(value: () => T): Record<ServiceKey, T> {
  return Object.fromEntries(SERVICE_KEYS.map((key) => [key, value()])) as Record<
    ServiceKey,
    T
  >;
}

function getGatewayBaseUrl() {
  const configuredUrl = process.env.NEXT_PUBLIC_API_GATEWAY_URL;

  if (configuredUrl) {
    return configuredUrl.replace(/\/$/, "");
  }

  if (typeof window === "undefined") {
    return "http://localhost:8080";
  }

  return `http://${window.location.hostname}:8080`;
}

function formatResponseBody(body: string, contentType: string | null) {
  if (contentType?.includes("application/json")) {
    try {
      return JSON.stringify(JSON.parse(body), null, 2);
    } catch {
      return body;
    }
  }

  return body;
}

export default function TestPage() {
  const { getToken } = useAuth();
  const [results, setResults] = useState<Record<ServiceKey, ServiceState>>(() =>
    buildInitial(() => ({
      body: null,
      error: null,
      loading: false,
      status: null,
    }))
  );
  const [bodies, setBodies] = useState<Record<ServiceKey, string>>(() =>
    Object.fromEntries(
      SERVICE_KEYS.map((key) => [key, SERVICES[key].defaultBody ?? ""])
    ) as Record<ServiceKey, string>
  );
  const [ids, setIds] = useState<Record<ServiceKey, string>>(() =>
    buildInitial(() => "")
  );

  async function callService(serviceKey: ServiceKey) {
    const service = SERVICES[serviceKey];
    const id = ids[serviceKey].trim();
    const needsId = service.path.includes(":id");

    if (needsId && !id) {
      setResults((current) => ({
        ...current,
        [serviceKey]: {
          body: null,
          error: "Enter an id.",
          loading: false,
          status: null,
        },
      }));
      return;
    }

    setResults((current) => ({
      ...current,
      [serviceKey]: {
        body: null,
        error: null,
        loading: true,
        status: null,
      },
    }));

    try {
      const token = await getToken();

      if (!token) {
        throw new Error("Missing Clerk session token.");
      }

      const headers: Record<string, string> = {
        Authorization: `Bearer ${token}`,
      };

      const init: RequestInit = {
        method: service.method,
        headers,
        cache: "no-store",
      };

      if (service.defaultBody !== undefined) {
        headers["Content-Type"] = "application/json";
        init.body = bodies[serviceKey];
      }

      const path = needsId
        ? service.path.replace(":id", encodeURIComponent(id))
        : service.path;

      const response = await fetch(`${getGatewayBaseUrl()}${path}`, init);

      const body = await response.text();
      const formattedBody = formatResponseBody(
        body,
        response.headers.get("content-type")
      );

      setResults((current) => ({
        ...current,
        [serviceKey]: {
          body: formattedBody,
          error: response.ok ? null : `Request failed with status ${response.status}.`,
          loading: false,
          status: response.status,
        },
      }));
    } catch (error) {
      setResults((current) => ({
        ...current,
        [serviceKey]: {
          body: null,
          error:
            error instanceof Error ? error.message : "Something went wrong.",
          loading: false,
          status: null,
        },
      }));
    }
  }

  return (
    <main className="flex-1 px-4 pt-24 pb-10">
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-6">
        <div className="space-y-2">
          <h1 className="text-2xl font-semibold">Gateway Test</h1>
          <p className="text-sm text-muted-foreground">
            Use the buttons below to call each Docker service through the API
            gateway.
          </p>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {SERVICE_KEYS.map((serviceKey) => {
            const service = SERVICES[serviceKey];
            const result = results[serviceKey];
            const needsId = service.path.includes(":id");
            const hasBody = service.defaultBody !== undefined;

            return (
              <section
                key={serviceKey}
                className="rounded-lg border bg-card p-4 shadow-sm"
              >
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <h2 className="font-medium">{service.label}</h2>
                    <p className="text-xs text-muted-foreground">
                      {service.method} {getGatewayBaseUrl()}
                      {service.path}
                    </p>
                  </div>
                  <Button
                    onClick={() => void callService(serviceKey)}
                    disabled={result.loading}
                  >
                    {result.loading ? "Loading..." : "Call Service"}
                  </Button>
                </div>

                {needsId && (
                  <input
                    value={ids[serviceKey]}
                    onChange={(event) =>
                      setIds((current) => ({
                        ...current,
                        [serviceKey]: event.target.value,
                      }))
                    }
                    placeholder="id"
                    spellCheck={false}
                    className="mt-4 w-full rounded-md border bg-background p-2 font-mono text-xs"
                  />
                )}

                {hasBody && (
                  <textarea
                    value={bodies[serviceKey]}
                    onChange={(event) =>
                      setBodies((current) => ({
                        ...current,
                        [serviceKey]: event.target.value,
                      }))
                    }
                    spellCheck={false}
                    className="mt-4 w-full rounded-md border bg-background p-2 font-mono text-xs"
                    rows={service.method === "POST" ? 12 : 5}
                  />
                )}

                <div className="mt-4 rounded-md bg-muted p-3">
                  <p className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    {result.status ? `Status ${result.status}` : "No response yet"}
                  </p>
                  <pre className="overflow-x-auto text-sm whitespace-pre-wrap break-words">
                    {result.error
                      ? `${result.error}${result.body ? `\n\n${result.body}` : ""}`
                      : (result.body ?? "Click the button to test this service.")}
                  </pre>
                </div>
              </section>
            );
          })}
        </div>
      </div>
    </main>
  );
}
