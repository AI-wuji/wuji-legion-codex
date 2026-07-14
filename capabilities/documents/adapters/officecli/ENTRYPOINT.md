# OfficeCLI stateless adapter

Use only for an explicit `.docx` or `.xlsx` structural export or one-shot HTML rendering task. Call `invoke-officecli.ps1` by absolute path with one allowed operation:

- `ViewHtml`: `officecli view <file> html -o <output>`
- `DumpJson`: `officecli dump <file> --json`, captured as JSON output

The adapter is read-only with respect to the source file, resolves `%LOCALAPPDATA%\OfficeCLI\officecli.exe` without PATH lookup, and forces `OFFICECLI_NO_AUTO_RESIDENT=1`. It rejects PPT/PPTX, output overwrite, resident commands, MCP, and installation commands. Do not use it to modify an existing file unless the user separately gives explicit write authority.

Use the existing document and presentation Skills for all other Office work.
