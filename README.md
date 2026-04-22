# aisand

A TUI for managing Lima VMs for AI agent workloads.

## Install

```sh
go install github.com/japacx/aisand@latest
```

## Prerequisites

- macOS
- `limactl` — installed automatically on first run if missing

## Usage

```sh
aisand
```

## Key Features

- **Split-panel VM list** — browse and inspect your Lima VMs at a glance
- **Contextual actions** — start, stop, and shell into VMs with keyboard shortcuts
- **Real-time log streaming** — tail VM output directly inside the TUI
- **Embedded cloud-init provisioning** — ships with ready-to-use cloud-init templates for AI agent workloads
