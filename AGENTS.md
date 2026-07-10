- Keep the portal UI in a mock black-and-white style without additional color styling; use alignment and spacing, and implement elements such as buttons, inputs, checkboxes, dropdowns, and calendars as reusable components.
- Every backend API must have tests in `api/tests`.
- Every backend setting must be configurable through a strictly structured uppercase environment variable in the `SECTION__SETTING__PROVIDER` form, such as `LOGIN__OAUTH__GOOGLE__CLIENT_ID` or `MAILING__SENDING__MAILGUN__API_KEY`.



## Important rules:

- Never add additional guards "just in case"; every piece of code should target the exact case we are solving now. Do not think ahead; ask the user in chat whether they want additional handling, and start with minimal changes.

- Always trust our own contracts. In this repo, we control both the API backend and frontend, so never treat the backend as unpredictable. Always follow a single contract, and do not handle edge cases that the API or frontend contract says will not happen.

- Always prefer less code to more code. Never copy-paste; instead, extract reusable functions. Never inflate code, each change should be minimal and targeted to the exact case we are solving now.

- Prefer smaller files organized by purpose within the folder tree. A file name should make its purpose clear.

- Never run compile/transpile/check commands in a way that emits generated files into repo root or source folders. If temporary output is needed, use `/tmp` or a gitignored scratch dir such as `bo/.tmp/`.

## Bug solving

Instead of making a guess and adding several guards to suspicious places in code, first add logs that fully reveal the bug and ask the user to provide the log, or ask the user to reproduce it if you are connected to logs.