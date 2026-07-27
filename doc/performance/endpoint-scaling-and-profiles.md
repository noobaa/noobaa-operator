# MCG Endpoint Scaling and Performance Profiles

[NooBaa Operator](../../README.md) /

## Introduction

This document records the endpoint performance measurements behind the MCG performance profile
values, why the `small-objects` profile sets an endpoint CPU request of 1 core, and why the endpoint
HPA carries a `behavior` block. It is intended for anyone changing
`pkg/system/performance_profiles.go`, the endpoint HPA, or reasoning about how MCG endpoints scale.

The central finding is that **a NooBaa endpoint pod is a single Node.js process, and its throughput
is bounded by one event loop — roughly 1 CPU core on small objects.** Several profile values had been
chosen on the assumption that an endpoint pod could use its full CPU limit. It cannot, and that
assumption silently disabled CPU-based autoscaling for small-object workloads.

## Background: what limits an endpoint

A NooBaa endpoint container runs one Node.js process. Node executes JavaScript on a single **event
loop** thread, with a small pool of helper threads (libuv, default 4) for filesystem and crypto work,
plus V8 worker threads. For the S3 data path:

- **Small objects are event-loop bound.** Per-request overhead (TLS, SigV4 signing, hashing, RPC to
  core, object metadata) runs as JavaScript on the main thread. Per-thread sampling inside a loaded
  endpoint consistently showed the main thread at **0.92–0.99 cores** with helper threads
  near-idle — i.e. one core is the ceiling, and the container's higher CPU limit is unreachable.
- **Large objects engage the helper threads.** At 1 MiB, V8 worker and libuv threads do real payload
  work, and a single process reaches roughly **2.5 cores**.
- `UV_THREADPOOL_SIZE` does not change either case (see [Results](#results)).
- `ENDPOINT_FORKS` makes the endpoint call `cluster.fork()`, giving N independent event loops in one
  pod. It is not set by the operator.

The consequence for autoscaling: the endpoint HPA scales on **CPU utilisation = per-pod CPU ÷ CPU
request**, with a target of 80%. If the request is larger than what a pod can physically consume, a
fully saturated pod never reaches the target and the HPA never scales.

## Test method

| | |
|---|---|
| Workload | `warp mixed` (GET / PUT / STAT / DELETE) via `noobaa bench warp` |
| Load | 3 client pods × 80 concurrency = **240 concurrent** |
| Duration | 5 min per run (shorter runs do not reach steady state) |
| Cluster | 3 workers, 15.5 cores allocatable each; DB 6 CPU / 16Gi × 2 instances |
| Profile | `small-objects` baseline; endpoint `req 2 / lim 4 CPU`, `2Gi / 4Gi` memory |
| Postgres | Pinned (`max_connections=383`, `work_mem=10782kB`) identically across every run |
| Replicas | Pinned `min = max` so the HPA could not change pod count mid-run |

Reported per-pod CPU is `max(rate(container_cpu_usage_seconds_total{container="endpoint"}[2m]))`
over the steady-state window. Two reporting rules matter for interpreting the tables:

- **`warp`'s blended `Total obj/s` is not sufficient on its own** — it is a sum across operation
  types dominated by cheap GET/STAT. Per-operation rates are always given alongside it.
- **CFS throttling is reported as steady-state mean and window peak separately.** They differ by an
  order of magnitude in some runs, and conflating them caused a heavily-throttled run to be read as
  clean during this study.

Each run was accepted only if: steady state was reached; warp clients were not CPU-saturated (peak
0.30 cores/client observed, far from saturation); DB CPU stayed well under its 6-core allowance;
Postgres parameters were unchanged; replica count held constant; and no endpoint restarted.

## Results

### Object-size sweep — 2 endpoint pods, stock `request: 2`

| Object size | Per-pod CPU | HPA utilisation | GET | PUT | STAT | DELETE | Total obj/s | Throughput |
|---|---|---|---|---|---|---|---|---|
| 4 KiB | 1.04 | 50–51% ✗ | 393.16 | 131.04 | 262.04 | 87.27 | 873.51 | 2.05 MiB/s |
| 32 KiB | 1.07 | 52–54% ✗ | 370.11 | 123.32 | 246.74 | 82.20 | 822.36 | 15.42 MiB/s |
| 46 KiB (mixed) | 1.02 | 53–55% ✗ | 366.79 | 122.19 | 244.50 | 81.44 | 814.92 | 22.14 MiB/s |
| 128 KiB | 1.36 | 63–70% ✗ | 306.48 | 102.12 | 204.30 | 68.04 | 680.94 | 51.08 MiB/s |
| 512 KiB | 2.28 | 103–107% ✓ | 184.30 | 61.42 | 122.89 | 40.94 | 409.55 | 122.86 MiB/s |
| 1 MiB | 2.57 | 119–124% ✓ | 125.09 | 41.63 | 83.36 | 27.81 | 277.64 | 166.58 MiB/s |
| 1.78 MiB (mixed) | 2.66 | 112–132% ✓ | 80.03 | 26.63 | 53.40 | 17.80 | 177.97 | 189.87 MiB/s |

**A saturated pod stays under the 80% target for every size at or below ~180 KiB.** Interpolating
between 128 KiB and 512 KiB, per-pod CPU crosses the 1.6-core trigger at approximately **183 KiB**.
Below that, the profile is blind: the endpoints are at their ceiling and the HPA reports ~50–70%.

The two "mixed" rows are `warp --obj.randsize`. Uncapped, it produced a **1.78 MiB** mean object —
larger than the fixed 1 MiB case — so an uncapped random-size run does *not* exercise small-object
behaviour. Capping with `--obj.size 256KiB` yielded a 46 KiB mean, which does. Mean size per data
operation was verified as `MiB/s ÷ (GET + PUT obj/s)`; the fixed-1 MiB run returns 1.00 MiB, which
validates the arithmetic.

### `UV_THREADPOOL_SIZE` has no effect

| Config (1 pod, 1 MiB) | Per-pod CPU | Total obj/s | Process threads |
|---|---|---|---|
| Default (4) | 2.18 | 149.83 | 15 |
| 8 | 2.24 | 149.80 | 19 |

CPU differs by 0.4% and throughput by 0.02%. The thread count rose as expected, so the pool did grow
— it simply does no additional work. **`UV_THREADPOOL_SIZE` is not a useful tuning knob for the S3
data path.**

### `ENDPOINT_FORKS` — event-loop scaling at 4 KiB

Measured with the CPU limit raised to 8 so that throttling did not distort the result:

| Event loops | Per-pod CPU | CPU per loop | Total obj/s | vs 1 loop | efficiency | obj/s per loop | DB CPU | GET latency |
|---|---|---|---|---|---|---|---|---|
| 1 | 0.91 | 0.91 | 493.07 | 1.00× | 100% | 493 | 0.42 | 489 ms |
| 2 | 1.75 | 0.88 | 909.10 | 1.84× | 92% | 455 | 0.80 | 271 ms |
| 4 | 3.63 | 0.91 | 1508.77 | 3.06× | 76% | 377 | 1.36 | 166 ms |

Two distinct behaviours:

- **CPU scales linearly.** CPU per loop is flat at 0.88–0.91 regardless of how many loops run. Each
  event loop draws its full ~0.9 core with no measurable interference. Adding endpoint capacity
  reliably adds CPU capacity 1:1.
- **Throughput scales sublinearly, and efficiency erodes** — 3.06× for 4× the loops, with
  obj/s-per-loop falling 493 → 455 → 377. Latency improves (489 → 166 ms) but less than 1/N, which
  points to a fixed per-request downstream cost (DB round trip, backing store, network) that adding
  endpoints does not reduce. There is no plateau or regression at 4 loops, and the DB was at 1.36 of
  6 cores, so this is diminishing returns rather than a hard ceiling.

> **Caveat:** offered load was fixed at 240 concurrency for every run, so these are "divide a fixed
> load across N loops" measurements. Some of the sublinearity may be offered-load starvation rather
> than a server-side limit. CPU stayed pegged at ~0.9 core/loop, which argues the loops were genuinely
> busy — but a run that scales offered load together with endpoint count was not performed. Treat the
> CPU linearity as solid and the throughput-efficiency curve as indicative.

### Forks versus pods

At equal event-loop count *and* equal node count, the two are equivalent:

| Config (2 event loops, 1 MiB) | Nodes | Total obj/s | Aggregate CPU | obj/s per core |
|---|---|---|---|---|
| 2 forks in 1 pod | 1 | 274.41 | 4.15 | 66.1 |
| 2 pods, 1 process each | **1 (pinned)** | 270.51 | 4.18 | 64.7 |
| 2 pods, 1 process each | 2 | 298.08 | 4.40 | 67.7 |

The 2-pods-on-2-nodes advantage (~10%) is a **node-count** effect — independent CPU, memory
bandwidth, network stack and page cache — not an architectural advantage of pods over forks. Any
comparison of forks against pods must hold node count fixed.

### Fork side effects

- **CPU limit becomes binding quickly.** With the stock 4-core limit, 2 forks at 1 MiB were throttled
  69% of periods and 4 forks 82%. The cost is real but modest: 4 forks at the stock limit delivered
  1442.71 obj/s versus 1508.77 uncapped (**−4.4%**), with ~5% higher GET latency.
- **Memory does not require a limit increase for small objects.** 4 forks at 4 KiB peaked at 3.47 GiB
  inside the **stock 4 GiB limit**, with no `OOMKilled` and no restarts, and memory was effectively
  identical at a 4 GiB and 8 GiB limit. An earlier 4.84 GiB observation came from a throttled run —
  throttling causes request queueing, which inflates in-flight buffers. **Untested at 1 MiB**, where a
  throttled run peaked at 4.47 GiB; do not assume the same result for large objects.
- **Postgres connections are over-provisioned by the current formula.** `max_connections` is derived
  as `63 + 80n`, i.e. 80 per endpoint *process*. Measured usage was **40–54 per process**; 4 forks
  peaked at 216 client backends against 383 available, with no connection errors.
- **`calculateMaxConnections` counts pods, not processes.** In `pkg/system/db_reconciler.go` it is fed
  `endpointMaxCount` (pods) while its own comment says each endpoint *process* uses up to 80
  connections. With forks enabled this under-provisions `max_connections` by the fork factor. **This
  must be fixed before any fork-based feature ships.**

### The fix: CPU request 1 for `small-objects`

| 4 KiB, single process per pod | Per-pod CPU | HPA utilisation | Scales at target 80%? |
|---|---|---|---|
| `request: 2` (previous) | 1.03 | 50–52% | **No** |
| `request: 1` | 1.03 | **100–102%** | **Yes** |

Identical workload and identical CPU consumption; only the denominator changed. The request now
matches the measured single-process ceiling, so a saturated pod reports ~100% and the HPA behaves as
intended.

**Why 1 and not a higher fractional value** (e.g. 1.2): the small-object ceiling is a hard ~1.0 core.
At `request: 1` the 80% target trips at 0.8 cores, leaving ~20% headroom to absorb the ~30–60 s pod
cold-start. At `request: 1.2` the target would trip at 0.96 of a 1.0-core ceiling — a fully saturated
pod would report only 85%, understating saturation and leaving almost no reaction headroom.

## Conclusions

### The operator should not set `ENDPOINT_FORKS`

Forks work: one pod with 4 forks delivered **1508.77 obj/s versus 493.07 for a single process — 3.06×
per pod** — with linear CPU scaling and no memory-limit increase needed at 4 KiB. For deployments
constrained by pod count or node count they are a legitimate capacity lever.

They are nevertheless the wrong mechanism for the autoscaling problem:

1. **Forks cannot participate in autoscaling.** Changing `ENDPOINT_FORKS` requires a pod restart, so
   it is a provisioning-time decision; the HPA still only scales pods.
2. **They are equivalent to pods**, at equal node count, so they add no throughput a replica would not.
3. **`calculateMaxConnections` must be corrected first**, since it under-provisions `max_connections`
   by the fork factor.
4. **`request: 2 → 1` achieves the same autoscaling repair** by changing one profile value, with none
   of the above.

### `mixed-workload` does not need an HPA target change

Taking the profile domains as **`small-objects` ≤ 1 MiB** and **`mixed-workload` ≥ 1 MiB**, every size
in `mixed-workload`'s range clears the trigger comfortably (1 MiB: 119–124%, 1.78 MiB: 112–132%). The
blind band below ~183 KiB lies entirely within `small-objects`, and is addressed there by the CPU
request change rather than by an HPA target change.

This distinction matters because `request` is the better lever on two counts. It is already a
per-profile field, whereas the 80% target is static and global — making it per-profile would require a
new struct field plus reconcile logic. And it tracks each profile's real ceiling (~1 core for small
objects, ~2.5 for large), whereas a global target change would apply to every profile including
`default`. Before this change, `mixed-workload` and `small-objects` shipped **byte-identical endpoint
specs**; the request value is the first thing that differentiates them at the endpoint.

## Applied changes

| Change | Location |
|---|---|
| `small-objects` endpoint CPU request `2 → 1` (limit stays 4; memory unchanged) | `pkg/system/performance_profiles.go` |
| Endpoint HPA `behavior` block — scale up 60 s window / ≤2 pods per 60 s; scale down 600 s window / 1 pod per 180 s | `deploy/internal/hpav2-autoscaling.yaml` |

The `behavior` block accompanies the request change. At `request: 1`, per-pod utilisation spans
102% (4 KiB) to 257% (1 MiB), so a stream mixing object sizes can push utilisation across the trigger
repeatedly; without damping, replicas may flap, and endpoint pod churn is expensive (V8 heap,
metadata cache and DB connection pool all re-warm). It is global rather than per-profile because
`mixed-workload` keeps `request: 2` and its utilisation is flat across its domain (112–133%), well
clear of the target — it needs less damping, not different damping.

**The `behavior` values are reasoned starting points, not measured.** Every run in this study pinned
`min = max` to hold replicas constant, so flapping was never exercised. They should be tuned against
an oscillating-load run before being treated as final.

Unchanged deliberately: the HPA target remains **80%**, and `ENDPOINT_FORKS` is **not** set.

## Open questions

| Question | Status |
|---|---|
| Endpoint scaling knee with offered load scaled alongside endpoint count | Not measured — offered load was fixed at 240 concurrency throughout |
| Whether `endpointMaxCount` should rise above 4 for `small-objects` | Undecided. `request: 1` halves the per-pod reservation, so the same footprint that held 4 pods now holds ~8, and DB CPU (1.36 of 6 cores at 4 loops) suggests room for ~12–16 loops. But throughput efficiency degrades to 76% at 4 loops, so extra replicas buy progressively less. Needs a target throughput and a scaling run. |
| 4 forks at 1 MiB within a 4 GiB memory limit | Untested; only 4 KiB was verified |
| `behavior` tuning under bursty mixed traffic | Not measured (see above) |
| `calculateMaxConnections` pods → processes | Confirmed defect; prerequisite for any fork-based feature |

Also observed but out of scope: `CONTAINER_CPU_REQUEST` is dead code in noobaa-core — the CPU request
has no runtime effect on the endpoint. CPU profiling of the 4 KiB path attributed 11.4% to crypto
(TLS/SigV4/hashing), 9.0% to `lodash` deep-clone, ~8% to RPC-to-core and 3.5% to GC, which suggests
per-request overhead in noobaa-core (SigV4 signing-key caching, removing the clone) as the route to
raising the single-loop ceiling itself.
