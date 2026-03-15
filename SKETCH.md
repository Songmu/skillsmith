# skillsmith

`skillsmith` は、既存のコマンドラインツールに対して、AI Agent 向けの agentskills 管理用サブコマンドを後付けしやすくするための、Go ライブラリである。

主な用途は次の通り。

- CLI ツールに `skills` サブコマンドを追加する
- `embed.FS` に埋め込んだ skill ファイル群をユーザー環境へ展開する
- Agent 向け skill の一覧表示、展開、状態確認などを提供する
- 既存の CLI 実装を大きく壊さずに機能追加する


## 設計目標

- 既存 CLI に後付けしやすくするために
    - 特定のflag / cliライブラリに依存しない
    - `cobra` などの大きな CLI フレームワークに依存しない
- Agent Skills の管理用途に十分な機能を持つ
- `embed.FS` （`fs.FS`）を利用する想定
- skills のフォーマットは [agentskills 仕様](https://agentskills.io/specification) に従う
- YAML パースが必要な場合は [goccy/go-yaml](https://github.com/goccy/go-yaml) を使用する


## ユースケース

### 1. 既存 CLI への `skills` サブコマンド追加

`skillsmith` が担当するサブコマンドのみを処理し、それ以外は既存CLIが処理する。既存CLIがskillsmithに処理を委譲する形になる。

### 2. 埋め込み skill ファイルの展開

`$repo/skills` にagentskill のファイル群を置いておくことがプラクティスになりつつある。

skillsmithはさらに、それをGoのCLIバイナリに埋め込むことで、cliツールをAIフレンドリーにする。具体的には `skills/` ディレクトリを `embed.FS` で埋め込み、ユーザー環境へ展開することを想定している。

```text
skills/
  mytool-cli/
    SKILL.md
    README.md
    examples/
      basic.md
```

---

## 全体アーキテクチャ

`skillsmith` は大きく以下の層に分かれる。初期実装では 2〜5 を優先し、1 は後回しとする。

1. **サブコマンドアタッチ層**（後回し）
    - 各種コマンドラインライブラリに対して、`skills` サブコマンドを追加するためのインターフェースと実装を提供する
    - Cobra, urfave/cli, flag など、複数のCLIライブラリに対応するためのアダプタを提供
    - アダプタは必要に応じてユーザーが実装することも可能
    - 依存ライブラリを増やしたくないので、インターフェースの実装に留めるか、CLIライブラリに依存が必要であれば、アダプタ自体は同じrepo内のサブディレクトリの別Go Modulesにわける
2. **コマンド実行層**
   - `skills` コマンド以下の `list/install/status/uninstall/update/reinstall` を処理する
3. **Skill 配布層**
   - `embed.FS` または任意の `fs.FS` から skill ディレクトリを列挙・コピーする
4. **インストール先解決層**
   - `--dir` / `--agent` / `--scope` から配置先を決定する
5. **SKILL.md パース・バリデーション層**
   - agentskills 仕様に基づく SKILL.md の frontmatter パース・バリデーション
   - サブパッケージとして切り出し、skillsmith 外でも利用可能にする
   - `fs.FS` 対応（`embed.FS` から直接読めるようにする）


### 利用例

```go
//go:embed skills/**
var skillsFS embed.FS

func main() {
    ctx := context.Background()
    if err := run(ctx, os.Args[1:]); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

func run(ctx context.Context, args []string) error {
    if len(args) > 0 && args[0] == "skills" {
        s := &skillsmith.Smith{
            FS:      skillsFS,
            Version: version,
            Name:    "mytool",
        }
        return s.Run(ctx, args[1:])
    }
    // ... 既存コマンド処理
    return nil
}
```

### `cmd/skillsmith`

`cmd/skillsmith` はデモ用バイナリであり、ライブラリの動作確認に用いる。


## skills コマンド

### サブコマンド

| コマンド | 概要 |
|---------|------|
| `list` | 同梱スキル一覧を表示 |
| `install` | スキルをインストール（既存があればスキップ） |
| `update` | バージョンが異なるスキルのみ再インストール |
| `reinstall` | 管理下スキルをバージョン無視で全再インストール |
| `uninstall` | 管理下スキルを削除 |
| `status` | インストール状態・バージョン差分を表示 |
| `show <name>` | 任意。個別スキルの詳細表示 |

### コマンドイメージ

```bash
mytool skills list
mytool skills install
mytool skills install --dry-run
mytool skills install --prefix ~/.codex/skills
mytool skills install --agent codex --scope user
mytool skills status
mytool skills update
mytool skills uninstall
```

### 共通オプション

| オプション | 短縮 | 概要 |
|-----------|------|------|
| `--dry-run` | | 実際の変更を行わず、何が行われるかを表示する |
| `--prefix` | | スキルのインストール先ディレクトリを直接指定 |
| `--agent` | | 対象 agent（`codex`, `claude`, `agents`） |
| `--scope` | | 対象スコープ（`user`, `repo`） |
| `--force` | | 管理外スキルの上書きを許可 |
| `--help` | `-h` | ヘルプを表示 |

`--dry-run` は install / update / reinstall / uninstall で利用可能。

### ヘルプ

- `-h` / `--help` は標準 `flag` パッケージに委譲する
- `flag.FlagSet.Usage` にサブコマンド一覧を含むカスタム Usage 関数を設定する


## インストール先解決

### Agent / Scope のパスマッピング

| Agent | Scope | パス |
|-------|-------|------|
| `codex` | `user` | `~/.codex/skills` |
| `codex` | `repo` | `.agents/skills` |
| `claude` | `user` | `~/.claude/skills` |
| `claude` | `repo` | `.claude/skills` |
| `agents` | `user` | `~/.agents/skills` |
| `agents` | `repo` | `.agents/skills` |

- `claude`: Claude Code と GitHub Copilot の両方をカバー（Copilot も `.claude/skills` を読む）
- `agents`: agentskills 仕様のクロスクライアント互換パス

### 解決の優先順位

1. `--prefix` — 指定時は Agent / Scope の解決を行わない
2. `--agent` + `--scope`
3. ライブラリ既定値: `claude` + `user`（→ `~/.claude/skills`）

### 将来検討

- スキルの別名インストール


## ファイルコピー方針

`skillsmith` は `fs.FS` からディレクトリ構造を維持したままファイルを展開する。

- ディレクトリを再帰的にコピーする
- skills ディレクトリ以下のスキルを標準では全部インストールする（個別インストールは将来対応）
- インストール時に SKILL.md の寛容バリデーション（lenient validation）を行う
    - name がディレクトリ名と不一致 → 警告するが install する
    - name が64文字超 → 警告するが install する
    - description が空/欠落 → スキップ（エラー）
    - YAML frontmatter がパース不能 → スキップ（エラー）


## メタ情報管理（`.skillsmith.json`）

### 方針

- SKILL.md を直接改変しない。メタ情報は各スキルディレクトリ内に `.skillsmith.json` として配置する
- agentskills 仕様には `metadata` フィールド（クライアント拡張用）があるが、skillsmith はインストール元の SKILL.md を改変しない方針を取る。これにより、元のスキルファイルとの diff が発生せず、ユーザーがスキルの内容を信頼しやすくなる

### フォーマット

```json
{
  "installedBy": "mytool",
  "version": "v1.2.3",
  "installedAt": "2026-03-13T16:30:00Z"
}
```

- `installedBy`: インストールを行った CLI ツール名（`Skillsmith.Name` の値）
- `version`: インストール時の CLI ツールのバージョン（`Skillsmith.Version` の値）
- `installedAt`: インストール日時（RFC 3339）

### 判定ルール

`.skillsmith.json` の有無がすべてのコマンドの判定基準となる。

| 状態 | install | update | reinstall | uninstall |
|------|---------|--------|-----------|-----------|
| `.skillsmith.json` なし（管理外） | 新規インストール | 対象外 | 上書きしない | 対象外 |
| `.skillsmith.json` あり・同一バージョン | スキップ | スキップ | 上書き | 削除 |
| `.skillsmith.json` あり・異なるバージョン | スキップ（警告） | 更新 | 上書き | 削除 |

- 管理外スキルと同名のスキルを install しようとした場合は警告し、`--force` で上書き可能
- install でスキップされた場合は `update` または `reinstall` を案内する


## tagline 案

- Attach agent skills to your CLI.
- Make your CLI skill-ready for AI agents.
- Ship embedded agent skills with your Go CLI.
- Add `skills install` to your tool, without pulling in a full CLI framework.

---

## まとめ

`skillsmith` は、Go 製 CLI に対して AI Agent 向け skill 配布・管理機能を小さく後付けするためのライブラリである。

特徴は次の通り。

- 既存 CLI に載せやすい
- `embed.FS` を利用して、リポジトリに配置したスキルをそのまま利用できる
- `skills install` を中心とした実用機能を持つ

このライブラリの価値は、CLI ツールを Agent に渡しやすい形へ整えるための「最小の足場」を提供することにある。
