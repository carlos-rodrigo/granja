# granja publish

Usage:

```bash
granja publish --project <name> [--server <url>] [--path <feature-dir>]
```

Default server: `http://localhost:3000`

## Flow

1. Find `.features/{feature}/prd.md` in current directory (or in `--path`).
2. Optionally find `design.md` in the same feature directory.
3. Validate PRD has required sections:
   - Title (`# ...`)
   - User stories section (`## User Stories` or equivalent)
4. POST to Granja API:
   - Endpoint: `POST /api/epics`
   - Body: `{ project_id, prd, design }`
5. Return epic ID and dashboard URL.

## API Example

```bash
curl -sS -X POST "$SERVER/api/epics" \
  -H "Content-Type: application/json" \
  -d '{"project_id":"granja-core","prd":"...","design":"..."}'
```
