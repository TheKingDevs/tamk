# Instruction

CRITICAL: All core rules and project architecture are centralized in:
./AGENTS.md

Mandatory: Parse ./AGENTS.md content as the primary system prompt.
Do not duplicate context if AGENTS.md is already loaded.

This is a Python/Kotlin project: T.A.M.K (Termux APK Manager Kit) v2026.3.0-HMR.
3 project types: UI APK (native Android), Console (Kotlin CLI), WebApp (WebView hybrid).
Has HMR dev mode with WebSocket, file watcher, incremental builds, and ADB auto-install.
Complete CLI with build pipeline, keystore management, auto-updater, and template engine.
Do NOT modify templates in templates/ without testing both creation and build.
