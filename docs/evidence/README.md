# Evidence Index

This directory contains the project-facing index for verification artifacts. Full phase evidence is retained under `openspec/changes/ux-securities-hkex-trading-terminal/results/`.

| Evidence | Location | Required contents |
|---|---|---|
| UI screenshots | `screenshots/` and the OpenSpec change's `results/screenshots/` | Chrome/Computer Use captures at 1440x900, 1280x800 and 390x844, with states and findings recorded |
| Verification reports | `reports/` and the OpenSpec change's `results/` | Commands, exit status, scope, failures, skipped checks and residual risks |
| Database reports | OpenSpec change's `results/db-change-reports/` | Migration id, isolated target, validation SQL, rollback result, failure/rework and risk |
| Docker logs | OpenSpec change's `results/` | Compose ps, health, migration logs, actual host ports and service URLs |

Unrun checks must be marked `待执行`; a planned artifact or a successful unrelated check is not a pass.
