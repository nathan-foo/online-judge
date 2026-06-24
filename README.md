# Online Judge

## Overview

This repository contains a production-grade code evaluation platform allowing users to run and submit code against test cases and receive a scored result.

## Repository Structure

```
.
├── client
│   ├── app
│   ├── components
│   ├── lib
│   └── public
├── gateway
│   ├── bench
│   ├── cmd
│   │   └── api-gateway
│   └── internal
│       ├── auth
│       ├── config
│       ├── health
│       ├── logger
│       ├── proxy
│       ├── ratelimit
│       ├── redis
│       └── router
├── k8s
│   ├── base
│   │   ├── attempt-service
│   │   ├── client
│   │   ├── code-eval-service
│   │   ├── gateway
│   │   ├── postgres
│   │   ├── quiz-service
│   │   ├── rabbitmq
│   │   ├── redis
│   │   └── user-service
│   └── overlays
│       ├── local
│       └── prod
├── scripts
└── services
    ├── attempt-service
    │   ├── app
    │   └── tests
    ├── code-eval-service
    │   ├── build
    │   ├── cmd
    │   │   ├── eval-service
    │   │   └── exec-agent
    │   └── internal
    │       ├── broker
    │       ├── config
    │       ├── executor
    │       ├── health
    │       ├── judge
    │       ├── logger
    │       ├── pool
    │       └── router
    ├── quiz-service
    │   ├── app
    │   └── tests
    └── user-service
        ├── app
        └── tests
```

## Technology Stack

| Service | Technologies |
|---|---|
| Client | React.js / Next.js / TypeScript |
| API Gateway | Go, Redis |
| Code Evaluation Service | Go / RabbitMQ |
| User / Quiz / Attempt Service | Python / FastAPI / PostgreSQL / RabbitMQ |
| Infrastructure | Docker, Kubernetes |

## Methodology

### Architecture

The platform is built as a set of independent microservices communicating through HTTP and async messaging. All requests enter and are validated through a custom API Gateway, centralizing request logging, CORS, security headers, JWT authentication, global and route-specific rate limiting, and reverse proxying to appropriate services. This enforces access control and removes burden from the other microservices, allowing them to focus on domain logic.

The services are as follows:

- **User Service:** handles identity and account management
- **Quiz Service:** handles problem/quiz/test case creation for reusable problems
- **Attempt Service:** handles attempt submissions and grading
- **Code Eval Service:** handles untrusted code execution


### Code Submission and Execution

On submission, the attempt service publishes a request to RabbitMQ. The code evaluation service consumes this request, runs the submission against test cases, and publishes a result, which is then applied idempotently once all code answers have been resolved.

The evaluation service executes untrusted user code on network-isolated, single-use Kubernetes pods with strict memory/cpu limits, restricted user permissions, and a sandboxed gVisor runtime. To migitate the pod cold start latency, a per-language warm pool is used with concurrent refilling using goroutines.

## Benchmark Testing

The following benchmark was measured against a local Docker setup using [hey](https://github.com/rakyll/hey) and tests the overhead latency of the custom API gateway.

```
hey -z 30s -c 100 http://localhost:8080/
```

### Direct Path

```
Summary:
  Total:	30.0036 secs
  Slowest:	0.1072 secs
  Fastest:	0.0003 secs
  Average:	0.0030 secs
  Requests/sec:	38379.2928
  
  Total data:	2303036 bytes
  Size/request:	2 bytes

Response time histogram:
  0.000 [1]	|
  0.011 [998040]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.022 [1594]	|
  0.032 [99]	|
  0.043 [70]	|
  0.054 [12]	|
  0.064 [64]	|
  0.075 [20]	|
  0.086 [0]	|
  0.096 [2]	|
  0.107 [98]	|


Latency distribution:
  10%% in 0.0015 secs
  25%% in 0.0017 secs
  50%% in 0.0021 secs
  75%% in 0.0032 secs
  90%% in 0.0041 secs
  95%% in 0.0046 secs
  99%% in 0.0064 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0000 secs, 0.0959 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0390 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0013 secs
  resp wait:	0.0030 secs, 0.0003 secs, 0.0661 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0055 secs

Status code distribution:
  [200]	1000000 responses
```

### Gateway Path

```
Summary:
  Total:	30.0093 secs
  Slowest:	0.1670 secs
  Fastest:	0.0003 secs
  Average:	0.0061 secs
  Requests/sec:	16392.1954
  
  Total data:	983838 bytes
  Size/request:	2 bytes

Response time histogram:
  0.000 [1]	|
  0.017 [487402]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.034 [3628]	|
  0.050 [320]	|
  0.067 [184]	|
  0.084 [48]	|
  0.100 [115]	|
  0.117 [78]	|
  0.134 [50]	|
  0.150 [65]	|
  0.167 [28]	|


Latency distribution:
  10%% in 0.0030 secs
  25%% in 0.0039 secs
  50%% in 0.0054 secs
  75%% in 0.0074 secs
  90%% in 0.0097 secs
  95%% in 0.0115 secs
  99%% in 0.0167 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0000 secs, 0.0968 secs
  DNS-lookup:	0.0000 secs, 0.0000 secs, 0.0243 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0013 secs
  resp wait:	0.0061 secs, 0.0003 secs, 0.1669 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0069 secs

Status code distribution:
  [200]	491919 responses
```

### Results

Results were calculated by recording the delta between the direct API call and the gateway chain.

| target  | rps    | avg ms | p50 ms | p95 ms | p99 ms |
|---------|--------|--------|--------|--------|--------|
| direct  | 38,379 | 3.0    | 2.1    | 4.6    | 6.4    |
| gateway | 16,392 | 6.1    | 5.4    | 11.5   | 16.7   |

| metric | overhead |
|--------|----------|
| avg    | +3.1 ms  |
| p50    | +3.3 ms  |
| p95    | +6.9 ms  |
| p99    | +10.3 ms |
