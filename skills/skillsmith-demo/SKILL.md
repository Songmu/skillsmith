---
name: skillsmith-demo
description: Learn how to use skillsmith to distribute agent skills with your CLI
license: MIT
compatibility:
  - claude
  - codex
  - agents
allowed_tools:
  - Bash
  - Read
---

# skillsmith-demo

`skillsmith` is a Go library that makes it easy to attach AI agent skill
distribution to any existing CLI tool.

## Overview

skillsmith lets you embed `agentskills`-compatible skill files into your Go
binary and expose an `skills` subcommand that installs, updates, and manages
those skills in the user's agent environment.

## Usage

```bash
# List skills bundled with this tool
mytool skills list

# Install skills into the default agent directory (~/.claude/skills)
mytool skills install

# Install skills for a specific agent
mytool skills install --agent codex --scope user

# Install skills to a custom directory
mytool skills install --prefix /path/to/skills

# Preview what would be installed without making changes
mytool skills install --dry-run

# Check which skills are installed and whether updates are available
mytool skills status

# Update skills whose version has changed
mytool skills update

# Reinstall all managed skills
mytool skills reinstall

# Remove all managed skills
mytool skills uninstall
```

## Notes

- Skills are installed alongside a `.skillsmith.json` metadata file that
  records the installing tool, version, and timestamp.
- Unmanaged skills (placed manually without `.skillsmith.json`) are never
  overwritten unless `--force` is provided.
