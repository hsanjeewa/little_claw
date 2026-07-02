---
name: data-processor-api
description: Orchestrate HDP tasks and modules through the server API with robust lifecycle handling and observability.
---

# Herodotus AI Skill: Data Processor API (v1)

## Purpose

Use this skill to orchestrate HDP tasks/modules through the server API with robust lifecycle handling and observability.

## When to use

- Schedule tasks from compiled module inputs.
- Track status/state transitions.
- Retry failed tasks safely.
- Manage module upload/publish/version workflows.

## Source-of-truth

- https://docs.herodotus.cloud/data-processor-api/introduction
- https://docs.herodotus.cloud/data-processor-api/health-check
- https://docs.herodotus.cloud/data-processor-api/create-task
- https://docs.herodotus.cloud/data-processor-api/list-tasks
- https://docs.herodotus.cloud/data-processor-api/task-status
- https://docs.herodotus.cloud/data-processor-api/get-task-details
- https://docs.herodotus.cloud/data-processor-api/get-mmrs
- https://docs.herodotus.cloud/data-processor-api/get-task-output-preimage
- https://docs.herodotus.cloud/data-processor-api/retry-task
- https://docs.herodotus.cloud/data-processor-api/decommitment-only
- https://docs.herodotus.cloud/data-processor-api/get-task-state-transitions
- https://docs.herodotus.cloud/data-processor-api/upload-module
- https://docs.herodotus.cloud/data-processor-api/list-modules
- https://docs.herodotus.cloud/data-processor-api/get-module
- https://docs.herodotus.cloud/data-processor-api/update-module
- https://docs.herodotus.cloud/data-processor-api/publish-module
- https://docs.herodotus.cloud/data-processor-api/unpublish-module
- https://docs.herodotus.cloud/data-processor-api/get-module-versions
- OpenAPI spec: `openapi-hdp-server.json` (available in the docs repo)

## Architecture pattern

Treat this API as orchestration control plane:

`module registry -> task scheduler -> status tracker -> output consumer`

Use webhook fast path + polling fallback for reliability.

## Implementation workflow

1. Upload and publish module.
2. Create task with destination chain + params + optional webhook.
3. Track status until terminal state.
4. Fetch details/MMRs/output preimage on success.
5. Classify failure and apply retry policy.

## Reliability requirements

- Store task UUID, latest status, and full transition log.
- Use idempotent submission behavior.
- Implement bounded retries by failure type.
- Keep structured logs for state transitions and failures.

## Anti-hallucination guardrails

- Do not invent statuses/endpoints absent in docs/OpenAPI.
- Do not assume undocumented production URL.
- Apply auth requirements per endpoint definition.
- If docs and OpenAPI differ, prefer OpenAPI wire shape.

## Self-contained reference example

```ts
async function scheduleAndTrackTask(input: CreateTaskInput) {
  const taskId = await createTask(input);
  while (true) {
    const status = await getTaskStatus(taskId);
    if (isSuccessTerminal(status.status)) return await getTaskDetails(taskId);
    if (isFailureTerminal(status.status))
      throw new Error(status.errorMessage ?? "task failed");
    await sleep(nextBackoff());
  }
}
```

## Output checklist

- Task lifecycle observable
- Webhook + polling reconciliation implemented
- Retry policy explicit
- Output retrieval guarded by terminal success
