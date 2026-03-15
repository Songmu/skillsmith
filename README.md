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

Ship embedded agent skills with your Go CLI.

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

## Description

`skillsmith` is a Go library that makes it easy to attach
[agentskills](https://agentskills.io/specification)-compatible skill
distribution to any existing CLI tool.

Embed a `skills/` directory into your binary using `embed.FS`, then expose a
`skills` subcommand backed by `skillsmith.Smith`. Users can then install,
update, and manage your tool's agent skills with commands like:

```bash
mytool skills list
mytool skills install
mytool skills install --dry-run
mytool skills install --scope user
mytool skills install --prefix ~/.agents/skills
mytool skills status
mytool skills update
mytool skills reinstall
mytool skills uninstall
```

By default, skills are installed into `~/.agents/skills`. You can override this
location with `--prefix`, control the target scope with `--scope`, and use
`--dry-run` or `--force` to preview or force installations.

### Key features

- **No CLI framework dependency** — uses the standard `flag` package only.
- **`embed.FS` / `fs.FS` throughout** — easy to test and embed.
- **Lenient validation** — warns about name mismatches; only skips skills with
  missing descriptions or unparseable YAML.
- **Metadata tracking** — writes `.skillsmith.json` alongside each installed
  skill so that `update` and `reinstall` work correctly.
- **`agentskill` subpackage** — SKILL.md parsing and discovery can be used
  independently of the CLI layer.


## Author

[Songmu](https://github.com/Songmu)
