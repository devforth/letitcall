- Keep the portal UI in a mock black-and-white style without additional color styling; use alignment and spacing, and implement elements such as buttons, inputs, checkboxes, dropdowns, and calendars as reusable components.
- Every backend API must have tests in `api/tests`.
- Every backend setting must be configurable through a strictly structured uppercase environment variable in the `SECTION__SETTING__PROVIDER` form, such as `LOGIN__OAUTH__GOOGLE__CLIENT_ID` or `MAILING__SENDING__MAILGUN__API_KEY`.
- Do not add configuration settings unless the user explicitly requests them; implement behavior directly when no operator choice is needed.
- After changing Go API code, automatically restart the VS Code task `Dev: Go API (41784)` before testing the running application so it uses the newly compiled backend; do not wait for the user to restart it.
- Every instant in REST API contracts and LevelDB must use UTC with second precision. Convert UTC instants to local time only in the frontend or in delivery modules such as email rendering. Weekly availability is a wall-clock recurrence rule, not an instant: store it with its IANA timezone and convert generated booking instants to UTC.
- Every backend API error response must be a JSON object with the single `error` string field. Svelte callers must use the shared `callApi` function so API errors are shown through the global notification stack.
- For portal icons, use Iconify's official Svelte integration and look for an existing Tabler icon first. Prefer offline icon data so the portal's same-origin Content Security Policy is preserved.



## Important rules:

- Never add additional guards "just in case"; every piece of code should target the exact case we are solving now. Do not think ahead; ask the user in chat whether they want additional handling, and start with minimal changes.

- Always trust our own contracts. In this repo, we control both the API backend and frontend, so never treat the backend as unpredictable. Always follow a single contract, and do not handle edge cases that the API or frontend contract says will not happen.

- Always prefer less code to more code. Never copy-paste; instead, extract reusable functions. Never inflate code, each change should be minimal and targeted to the exact case we are solving now.

- Prefer smaller files organized by purpose within the folder tree. A file name should make its purpose clear.

- Never run compile/transpile/check commands in a way that emits generated files into repo root or source folders. If temporary output is needed, use `/tmp` or a gitignored scratch dir such as `bo/.tmp/`.

## Bug solving

Instead of making a guess and adding several guards to suspicious places in code, first add logs that fully reveal the bug and ask the user to provide the log, or ask the user to reproduce it if you are connected to logs.
