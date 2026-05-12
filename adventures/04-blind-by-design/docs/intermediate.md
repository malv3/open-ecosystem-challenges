# 🟡 Intermediate: Outcome by cohort

Populate all three OpenFeature evaluation-context layers on a Spring Boot service and register a custom `Hook`:

- **Transaction context** (request-scoped) — populated by a Spring `HandlerInterceptor` that reads `?species=`, and clears on `afterCompletion` so values don't leak across pooled threads.
- **Global context** (process-scoped) — set once at startup from the `COUNTRY` environment variable.
- **Invocation context** (call-site) — passed as a third argument to `client.getStringDetails(...)`, carrying the per-evaluation `dose` attribute.
- **Audit `Hook`** — fires after every flag evaluation, writes an `[AUDIT]` log line with a fixed PII-safe attribute allowlist.

The broken-state lab already has the SDK and flagd provider wired in `Resolver.RPC` mode. The targeting in `flags.json` already carries three branches — `species == zyklop`, improper-`dose` for non-zyklops, `country == de` — but none of those attributes are in the eval context yet, so every request lands on the default variant. Your job is to make the targeting fire by wiring the three context layers and the audit hook.

## 🪐 The Backstory

The trial is widening. Subjects from outside the lab's local population are getting the wrong reading on their chart, and the lab director has just walked into the lab holding a stack of complaint forms. She wants the audit log to tell her, after the fact, exactly which `vision_state` the lab recorded for which subject — and she wants the lab to read the chart properly before it records any more bad readings.

The protocol is the same for every subject; the lab is not varying the trial. What differs is the **observed outcome**, because subjects don't all start from the same place — some have a biology that responds enhancedly to the same serum, some absorb less or more than the protocol's standard dose, and the trial is registered in different jurisdictions with different baselines.

Your shift: teach the lab to read each subject's species off the request, attach the trial's **country of registration** (set on the JVM via the `COUNTRY` environment variable) to the global context, pass the **dose** as invocation context at the moment of the flag evaluation, and register an audit hook that records every dose with its variant and reason.

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│  Spring Boot lab  (this challenge)                                   │
│                                                                      │
│  HTTP ──► SpeciesInterceptor ──► Trial ──► OpenFeature client        │
│           (transaction ctx:     (invocation ctx:   (global ctx:      │
│            species ← ?species=)  dose ← computed   country ←         │
│                                  at call site,     $COUNTRY env)     │
│                                  overridable                         │
│                                  with ?dose=)                        │
│                                                            │         │
│                                                            ▼         │
│                                                       AuditHook      │
│                                                       (audit log)    │
│                                                            │         │
│                                                            ▼         │
│                                                       FlagdProvider  │
│                                                       (Resolver.RPC) │
└────────────────────────────────────────────────────────────┬─────────┘
                                                             │  gRPC :8013
                                                             ▼
                                          ┌─────────────────────────────┐
                                          │   flagd  (sibling container) │
                                          │   reads + watches flags.json │
                                          └─────────────────────────────┘
```

The lab and a flagd sidecar run as siblings in the devcontainer's compose stack. The OpenFeature client uses `Resolver.RPC` to reach `flagd:8013`; flagd is the one watching `flags.json` and serving evaluations from it. The targeting rules live entirely inside `flags.json`; your job is to make sure the attributes the rules reference (`species`, `country`, `dose`) are populated on every evaluation.

## 🎯 Objective

By the end of this level, the lab hits each of these observable outcomes:

- **Targeting by species fires.** `curl /?species=zyklop` returns `"enhanced"` regardless of dose or country.
- **Targeting by country fires.** With `COUNTRY=de`, `curl /?dose=standard` returns `"sharp"`; with `COUNTRY=at` the same call falls through to the default — Austria isn't a country branch in `flags.json`.
- **Targeting by dose fires, and species takes precedence over it.** `curl /?dose=underdose` returns `"clouded"`; `curl /?species=zyklop&dose=underdose` still returns `"enhanced"`.
- **Every evaluation produces an `[AUDIT]` log line** naming the flag, the resolved variant, the reason, and the attributes that drove the outcome (`species`, `country`, `dose`).
- **The response is never `"untreated"`.** That fallback only fires when the SDK can't reach flagd at all — if you see it, the provider isn't registered.

> 📋 **Run with `tee app.log`.** The verifier greps `[AUDIT]` lines from `app.log` next to `pom.xml`. The `./run-germany.sh` / `./run-austria.sh` scripts handle this for you; if you run `./mvnw spring-boot:run` directly, pipe through `| tee app.log` or the verifier has nothing to grep.

## 🧠 What You'll Learn

- How OpenFeature's **transaction-context propagation** works in a thread-per-request server, and why a `ThreadLocalTransactionContextPropagator` is the right primitive for Servlet-based apps
- The difference between **request-scoped context** (the subject's species) and **global evaluation context** (the trial's country) — and when each is the right tool
- How **hooks** let you attach cross-cutting behaviour — audit logging today, OpenTelemetry tracing tomorrow — without modifying every flag evaluation call site

## 🧰 Toolbox

Your Codespace comes pre-configured with the following tools:

- [Java 21](https://adoptium.net/) toolchain (Temurin)
- The Spring Boot Maven Wrapper (`./mvnw`) — no global Maven install required
- `curl` and `jq` for poking at the lab
- `tail -f` for watching the application log live

The flagd sibling that the Beginner level introduced is still running here — the broken-state `OpenFeatureConfig` already targets it via `Resolver.RPC` (`flagd:8013` from the workspace, `localhost:8013` from your host).

## ⏰ Deadline

Tuesday, 26 May 2026 at 23:59 CET

> ℹ️ You can still complete the challenge after this date, but points will only be awarded for submissions before the
> deadline.

## 💬 Join the discussion

Share your solutions and questions in
the [challenge thread](https://community.open-ecosystem.com/t/outcome-by-cohort-adventure-04-intermediate/1485)
in the Open Ecosystem Community.

## ✅ How to Play

### 1. Start Your Challenge

> 📖 **First time?** Check out the [Getting Started Guide](../../start-a-challenge) for detailed instructions on forking, starting a Codespace, and waiting for infrastructure setup.

Quick start:

- Fork the repo
- Create a Codespace
- Select "Adventure 04 | 🟡 Intermediate (Outcome by cohort)"
- Wait ~2-3 minutes for the Java toolchain to install (`Cmd/Ctrl + Shift + P` → `View Creation Log` to view progress)

When the post-create finishes you'll have Java 21, the Maven wrapper, and the broken-state lab ready in `adventures/04-blind-by-design/intermediate/`.

### 2. Start the Lab

The lab is a terminal-only level — no port is forwarded to your host, you `curl` it from inside the Codespace. Boot it once so it's actually serving on `localhost:8080`. Either click **Run** on `Laboratory` in the Spring Boot Dashboard panel (or press **F5** with `Laboratory.java` open), or, from the terminal:

```bash
./mvnw spring-boot:run
```

In another terminal, confirm the broken-state symptom:

```bash
curl 'http://localhost:8080/?species=zyklop'
# => {"value":"blurry", ...}    ← wrong cohort, no targeting fired
```

That `"blurry"` is the starting point you want: even when the request shouts `species=zyklop`, the lab has nothing in its evaluation context, so flagd's targeting can't fire and every subject drops to the default variant.

Stop the app (`Ctrl+C`) and start fixing.

### 3. Inspect the Starting Point

The lab already has the OpenFeature SDK and the flagd contrib provider on the classpath, and the `FlagdProvider` is wired in `Resolver.RPC` mode against the flagd sibling. The `flags.json` shipping with this level is the targeting-rich version — all three branches (open `intermediate/flags.json` and you'll see this verbatim):

```json
"targeting": {
  "if": [
    { "===": [{"var": "species"}, "zyklop"] },                  "enhanced",
    { "in":  [{"var": "dose"},    ["underdose", "overdose"]] }, "clouded",
    { "===": [{"var": "country"}, "de"] },                      "sharp"
  ]
}
```

The catch: nothing in the application populates `species`, `country`, or `dose` yet. Every request lands with an empty evaluation context, so none of the branches fire and every subject walks out with `"blurry"` (the default variant) — exactly the symptom you just reproduced in step 2.

### 4. Implement the Objective

You need four pieces.

#### 4a. A `SpeciesInterceptor`

Create a Spring [`HandlerInterceptor`](https://docs.spring.io/spring-framework/docs/current/javadoc-api/org/springframework/web/servlet/HandlerInterceptor.html) — a per-request hook with `preHandle` running before your controller and `afterCompletion` running after the response — that:

- In `preHandle`, reads `?species=` from the request and puts it on the **transaction context** for the duration of the request.
- In `afterCompletion`, clears the transaction context. Servlet threads are pooled — if you don't clear, the previous request's species leaks into whichever request lands on the thread next.
- In a static initialiser, registers a `ThreadLocalTransactionContextPropagator` once on the OpenFeature API. Without it the SDK has no way to carry per-request context across the call into the controller, and the transaction context silently stays empty.

#### 4b. Wire the interceptor + global context + hook in `OpenFeatureConfig`

Update `OpenFeatureConfig` to:

- Register your `SpeciesInterceptor` with Spring (`WebMvcConfigurer.addInterceptors`).
- Read `COUNTRY` from the environment and set it as the **global** evaluation context — merged into every flag evaluation regardless of request.
- Register your `AuditHook` (you'll write that next) globally on the OpenFeature API.

The three context layers — *global* (this `country`), *transaction* (the `species` you set in 4a), and *invocation* (the `dose` your `Trial` controller will pass at each call site) — merge before flagd evaluates the rules. Precedence on conflict is **invocation > transaction > global**.

#### 4c. Pass the dose as invocation context from `Trial`

Update `Trial` so each evaluation carries a `dose` on the **invocation context** — the third argument to `client.getStringDetails(flag, fallback, ctx)`, evaluated per call and not stored anywhere afterwards. Two real-world details to think through:

- The dose should be `"standard"` most of the time but occasionally `"underdose"` or `"overdose"` — that's the lab tech mis-measuring, and it's what makes the improper-dosing branch in `flags.json` fire at all.
- Make it overridable via a `?dose=` query parameter so you can verify each branch by hand without waiting for the random pick to happen.

The flag rule that depends on this is the second branch in `flags.json`:

```json
{ "in": [{"var": "dose"}, ["underdose", "overdose"]] }
```

If your invocation context doesn't carry `dose`, that rule sees `null` and the branch never fires — every non-zyklop request lands on either the country branch or the default.

#### 4d. An `AuditHook`

A [`Hook`](https://openfeature.dev/docs/reference/concepts/hooks) is OpenFeature's interceptor for flag evaluations: `before` / `after` / `error` / `finallyAfter` fire around every `client.getXxxDetails(...)`, and `HookContext.getCtx()` exposes the **merged** context — that's what makes an audit trail useful instead of a "got here" log line. Create one that, on `after(...)`, writes an `[AUDIT]` log line naming the flag, the resolved variant, the reason, and the attributes that drove the outcome. Two design decisions worth thinking about:

- When the resolved variant is `clouded`, log at **`WARN`** so the safety officer can grep for it; otherwise `INFO`. Also implement `error(...)` — failed evaluations shouldn't disappear silently.
- Use a **fixed allowlist** of attribute keys (`List.of("species", "country", "dose")`) rather than iterating the whole eval context — audit logs outlive app logs and a discipline of "log only what you decided to log" pays off the moment something sensitive lands on the context.

The order matters less than you'd think — Spring will pick up `OpenFeatureConfig` as a `@Configuration` class on boot, the `@PostConstruct` will run once, and from then on every evaluation the `Trial` performs will see both contexts and trigger your hook.

### 5. Re-run the Lab with a Cohort

```bash
./run-germany.sh   # COUNTRY=de — exercises the country-targeting branch  (or `make lab-germany`)
```

`./run-austria.sh` (`COUNTRY=at`) ships alongside it for the no-targeting case. Three named launch configs in `.vscode/launch.json` (Germany / Austria / No country) give you one-click cohort switching from the **Run and Debug** view.

### 6. Verify Each Cohort by Hand

In another terminal — exercise all three context layers and the precedence between them:

```bash
# Transaction context — species wins, regardless of country / dose
curl -s 'http://localhost:8080/?species=zyklop' | jq .value
# => "enhanced"

# Global context — country=de from the env. Pin ?dose=standard so the
# random dose pick can't trip the improper-dose branch.
curl -s 'http://localhost:8080/?dose=standard' | jq .value
# => "sharp"     (when running ./run-germany.sh — COUNTRY=de)
# => "blurry"    (when running ./run-austria.sh — COUNTRY=at: no targeting branch fires, default applies)

# Invocation context — improper dose for a non-zyklop subject
curl -s 'http://localhost:8080/?dose=underdose' | jq .value
# => "clouded"

# Precedence — species-zyklop is evaluated before improper-dose in flags.json
curl -s 'http://localhost:8080/?species=zyklop&dose=underdose' | jq .value
# => "enhanced"
```

Tail the log to see the audit trail:

```bash
grep '\[AUDIT\]' app.log | head
```

You should see one `[AUDIT] flag=vision_state variant=… reason=… species=… country=… dose=…` line per `curl` call. `clouded` outcomes log at `WARN` with the "improper dosing or off-protocol cohort, follow-up required" suffix.

### 7. Verify Your Solution

Once you think you've solved the challenge, run the verification script:

```bash
./verify.sh
```

**If the verification fails:**

The script will tell you which checks failed. Fix the issues and run it again.

**If the verification passes:**

1. The script will check if your changes are committed and pushed.
2. Follow the on-screen instructions to commit your changes if needed.
3. Once everything is ready, the script will generate a **Certificate of Completion**.
4. **Copy this certificate** and paste it into the [challenge thread](https://community.open-ecosystem.com/c/open-ecosystem-challenges/) to claim your victory! 🏆
