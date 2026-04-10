"use client";

import { useState } from "react";
import { useAuth } from "@clerk/nextjs";

import { Button } from "@/components/ui/button";

type ServiceKey = "test" | "test2";

type ServiceState = {
  body: string | null;
  error: string | null;
  loading: boolean;
  status: number | null;
};

const SERVICES: Record<ServiceKey, { label: string; path: string }> = {
  test: {
    label: "Test Service 1",
    path: "/test",
  },
  test2: {
    label: "Test Service 2",
    path: "/test-2",
  },
};

const INITIAL_STATE: Record<ServiceKey, ServiceState> = {
  test: {
    body: null,
    error: null,
    loading: false,
    status: null,
  },
  test2: {
    body: null,
    error: null,
    loading: false,
    status: null,
  },
};

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
  const [results, setResults] =
    useState<Record<ServiceKey, ServiceState>>(INITIAL_STATE);

  async function callService(serviceKey: ServiceKey) {
    const service = SERVICES[serviceKey];

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

      const response = await fetch(`${getGatewayBaseUrl()}${service.path}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
        cache: "no-store",
      });

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
          {(Object.keys(SERVICES) as ServiceKey[]).map((serviceKey) => {
            const service = SERVICES[serviceKey];
            const result = results[serviceKey];

            return (
              <section
                key={serviceKey}
                className="rounded-lg border bg-card p-4 shadow-sm"
              >
                <div className="flex items-center justify-between gap-4">
                  <div>
                    <h2 className="font-medium">{service.label}</h2>
                    <p className="text-xs text-muted-foreground">
                      {getGatewayBaseUrl()}
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

                <div className="mt-4 rounded-md bg-muted p-3">
                  <p className="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    {result.status ? `Status ${result.status}` : "No response yet"}
                  </p>
                  <pre className="overflow-x-auto text-sm whitespace-pre-wrap break-words">
                    {result.error ?? result.body ?? "Click the button to test this service."}
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
