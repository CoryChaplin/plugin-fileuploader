# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

File sharing plugin for [Kiwi IRC](https://kiwiirc.com). Two components: a Go backend server implementing the TUS (resumable upload) protocol, and a Vue.js frontend plugin that integrates into Kiwi IRC's UI via Uppy.

## Build & Development Commands

### Backend (Go) — run from repo root
- `go run .` — run server from source (uses `fileuploader.config.toml` by default, override with `-config path`)
- `go build` — build production binary
- Server supports SIGHUP for config reload without restart

### Frontend (JS/Vue) — run from `fileuploader-kiwiirc-plugin/`
- `yarn install` — install dependencies
- `yarn build` — production build (outputs `dist/main.js`)
- `yarn watch` — rebuild on file changes
- `yarn dev` — webpack dev server with hot-reload (port 9001)
- `yarn lint` — ESLint

### No test suite exists for either component.

## Architecture

### Backend (Go)

Entry point: `main.go` → `server.NewRunContext()` → `runCtx.Run()`

Key packages:
- **server/** — Gin HTTP server, TUS protocol handlers (tusd), JWT authentication, CORS, graceful shutdown (SIGINT/SIGTERM)
- **shardedfilestore/** — File storage with configurable shard layers (subdirectory distribution to avoid filesystem limits)
- **db/** — Database layer (sqlx) supporting SQLite3 and MySQL, with sql-migrate migrations
- **expirer/** — Background goroutine that garbage-collects expired uploads based on configurable TTL (different for authenticated vs anonymous)
- **config/** — TOML config parsing
- **logging/** — Structured logging via zerolog with multiple output targets (file, UDP, Unix socket, stderr/stdout)
- **events/** — TUS upload lifecycle event broadcasting

The server can run standalone or as a WebIRCGateway plugin (see `webircgateway-plugin/`).

### Frontend (`fileuploader-kiwiirc-plugin/`)

Vue 2.7 plugin using Uppy for file upload UI (dashboard, drag-drop, webcam, audio, image editor).

- **src/plugin.js** — Kiwi IRC plugin entry point, registers UI hooks
- **src/uppy-setup.js** — Uppy instance configuration
- **src/components/** — Vue components (e.g., SidebarFileList)
- **src/handlers/** — Event handlers for paste, drag-drop, upload completion
- **src/utils/** — Helpers for URL encoding, file metadata decoding
- **src/token-manager.js** — JWT token management for authenticated uploads

## Configuration

TOML format: `fileuploader.config.toml` (see `fileuploader.config.example.toml` for documented options).

Key sections: `[Server]` (listen address, base path, CORS, reverse proxy), `[Storage]` (upload path, shard layers, max size), `[Database]` (sqlite3/mysql), `[Expiration]` (TTL, JWT secrets per issuer), `[PreFinishCommands]` (post-upload commands like metadata stripping), `[[Loggers]]` (multiple outputs).

## Code Style

- Tabs for indentation (see `.editorconfig`)
- Go: standard Go conventions
- JS: ESLint with `eslint-config-standard` + Vue plugin
