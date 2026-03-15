skillsmith
=======

[![Test Status](https://github.com/Songmu/skillsmith/actions/workflows/test.yaml/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/skillsmith/branch/main/graph/badge.svg)][codecov]
[![MIT License](https://img.shields.io/github/license/Songmu/skillsmith)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/skillsmith)][PkgGoDev]

[actions]: https://github.com/Songmu/skillsmith/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/skillsmith
[license]: https://github.com/Songmu/skillsmith/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/skillsmith

Ship embedded [Agent Skills](https://agentskills.io/) with your Go CLI.

## What is skillsmith?

`skillsmith` is a Go library for distributing [Agent Skills](https://agentskills.io/specification) through your CLI tool. Agent Skills are an open format for giving AI agents new capabilities — portable instruction sets that work across Claude Code, GitHub Copilot, OpenAI Codex, and other compatible agents.

With skillsmith, you can embed skill files into your Go binary using `embed.FS` and expose a `skills` subcommand that lets users install, update, and manage those skills on their machine. This makes your CLI tool AI-agent-friendly without pulling in a full CLI framework.

## Synopsis

```go
//go:embed skills/**
var skillsFS embed.FS

func run(ctx context.Context, args []string) error {
    if len(args) > 0 && args[0] == "skills" {
        s, err := skillsmith.New("mytool", version, skillsFS)
        if err != nil {
            log.Fatal(err)
        }
        return s.Run(ctx, args[1:])
    }
    // ... existing command handling
    return nil
}
```

This gives your tool the following subcommands:

```bash
mytool skills list        # List embedded skills
mytool skills install     # Install skills to ~/.agents/skills
mytool skills update      # Update skills when version differs
mytool skills reinstall   # Reinstall all managed skills
mytool skills uninstall   # Remove managed skills
mytool skills status      # Show install status and version diff
```

## Features

- **Drop-in integration** — add a `skills` subcommand to your existing CLI with a few lines of code
- **No CLI framework dependency** — uses the standard `flag` package only
- **`embed.FS` support** — embed skill files in your binary; the `skills/` prefix is auto-detected and stripped
- **[agentskills](https://agentskills.io/specification) compliant** — follows the open Agent Skills specification for SKILL.md format and directory structure
- **Metadata tracking** — writes `.skillsmith.json` alongside each installed skill for version-aware update and uninstall
- **Semantic versioning** — uses `semver.Compare` for version comparison; prevents accidental downgrades
- **Lenient validation** — warns on name mismatches; only skips skills with missing descriptions or unparseable YAML
- **`agentskills` subpackage** — SKILL.md parsing and discovery can be used independently of the CLI layer

## Options

| Option | Description |
|--------|-------------|
| `--prefix` | Override the install directory (ignores `--scope`) |
| `--scope` | `user` (default: `~/.agents/skills`) or `repo` (auto-detects repository root) |
| `--dry-run` | Preview changes without applying |
| `--force` | Overwrite unmanaged skills or force downgrade |

## Installation

```console
% go get github.com/Songmu/skillsmith
```

## Author

[Songmu](https://github.com/Songmu)
