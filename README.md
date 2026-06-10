# Online Judge

## Overview

This project is a work in progress.

## Repository Structure

```
.
├── client
│   ├── app
│   │   ├── explore
│   │   ├── pricing
│   │   ├── problems
│   │   ├── sign-in
│   │   │   └── [[...sign-in]]
│   │   └── sign-up
│   │       └── [[...sign-up]]
│   ├── components
│   │   └── ui
│   ├── lib
│   └── public
├── gateway
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
└── services
    └── test-service
        ├── app
        └── tests
    └── attempt-service
```

## Benchmark Testing

The following benchmark was measured against a local Docker setup using [hey](https://github.com/rakyll/hey), with a sustained load of 1,000 concurrent connections over 30 seconds. The benchmark targeted the http://localhost:8080/test endpoint with rate limiting disabled.

```
hey -z 30s -c 1000 http://localhost:8080/test
```

```
Summary:
  Total:	30.0448 secs
  Slowest:	0.8186 secs
  Fastest:	0.0005 secs
  Average:	0.0550 secs
  Requests/sec:	18094.9502
  
  Total data:	14639589 bytes
  Size/request:	27 bytes

Response time histogram:
  0.001 [1]	|
  0.082 [469194]	|■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■■
  0.164 [69467]	|■■■■■■
  0.246 [2338]	|
  0.328 [678]	|
  0.410 [301]	|
  0.491 [137]	|
  0.573 [57]	|
  0.655 [21]	|
  0.737 [6]	|
  0.819 [7]	|


Latency distribution:
  10%% in 0.0249 secs
  25%% in 0.0356 secs
  50%% in 0.0495 secs
  75%% in 0.0688 secs
  90%% in 0.0890 secs
  95%% in 0.1048 secs
  99%% in 0.1493 secs

Details (average, fastest, slowest):
  DNS+dialup:	0.0000 secs, 0.0000 secs, 0.1257 secs
  DNS-lookup:	0.0001 secs, 0.0000 secs, 0.0643 secs
  req write:	0.0000 secs, 0.0000 secs, 0.0378 secs
  resp wait:	0.0544 secs, 0.0005 secs, 0.4648 secs
  resp read:	0.0000 secs, 0.0000 secs, 0.0374 secs

Status code distribution:
  [200]	542207 responses
```
