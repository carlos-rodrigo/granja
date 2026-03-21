---
name: granja
description: "Publish PRDs and designs to Granja orchestration system. Triggers on: granja publish, /publish, publish to granja"
---

# Granja Skill

Publish a feature PRD (and optional design) from `.features/` to the Granja API.

## Command

### `publish`

Usage:

```bash
granja publish --project <name> [--server <url>] [--path <feature-dir>]
```

- `--project` (required): Project identifier/name to send as `project_id`
- `--server` (optional): Granja server base URL (default: `http://localhost:3000`)
- `--path` (optional): Feature folder path (example: `.features/user-auth`). If omitted, auto-discover from current directory.

## Examples

```bash
granja publish --project granja-core
granja publish --project granja-core --server http://localhost:3000
granja publish --project granja-core --path .features/user-auth
```

## Agent Procedure (step by step)

1. Parse command args (`project`, optional `server`, optional `path`).
2. Resolve feature directory:
   - If `--path` is provided, use it.
   - Otherwise, search current directory for `.features/*/prd.md`.
   - If multiple PRDs exist, ask user to choose feature path.
3. Read PRD from `{featureDir}/prd.md`.
4. Try to read optional design from `{featureDir}/design.md`.
5. Validate PRD contains required sections:
   - A title heading (for example `# ...`)
   - A user stories section (`## User Stories` or equivalent)
6. Build payload:

```json
{
  "project_id": "<project>",
  "prd": "<full prd markdown>",
  "design": "<full design markdown or null>"
}
```

7. Send request:

```bash
curl -sS -X POST "$SERVER/api/epics" \
  -H "Content-Type: application/json" \
  -d '<payload>'
```

8. Parse response and return:
   - `epic_id`
   - Dashboard URL (prefer `dashboard_url` from response; fallback to `$SERVER/epics/<epic_id>`)

## Expected Success Output

Provide a concise confirmation:

- Project
- Feature path
- Epic ID
- Dashboard URL

Example:

```text
Published to Granja
Project: granja-core
Feature: .features/user-auth
Epic ID: epic_12345
Dashboard: http://localhost:3000/epics/epic_12345
```

## Error Handling Guidance

- Missing `--project`: show usage and ask for project name.
- No `.features/*/prd.md` found: ask for `--path`.
- Multiple PRDs found: present options; do not guess.
- Missing `prd.md`: stop and report exact path checked.
- Invalid PRD format (missing title or user stories): report validation failure and missing section.
- API/network error: show HTTP status and server error body.
- Non-JSON response: show raw response snippet.
- Missing `epic_id` in response: mark publish as failed and print full response for debugging.

## Notes

- Default server is `http://localhost:3000`.
- `design.md` is optional; send `null` if absent.
- Never silently publish the wrong feature; require explicit user selection when ambiguous.
