# Xiaobai Image2 Direct Provider

This is a task-scoped provider client, not a local bridge or GUI. It sends one OpenAI-compatible image request, handles synchronous or task/poll responses, validates the returned image, and exits.

Set `XIAOBAI_API_KEY` in the process environment. Optionally set `XIAOBAI_BASE_URL`; the default is `https://new.xiaobaiapi.cc`. Credentials are never accepted from repository configuration, written to reports, or exposed through a resident HTTP endpoint.

```powershell
& ./invoke.ps1 -Prompt 'A precise product illustration' -OutputDir ./outputs/image/xiaobai
```

On provider failure, return control to the Wuji image scenario so it can use the declared default GPT image fallback.
