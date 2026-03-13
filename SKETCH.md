# skillsmith

`skillsmith` は、既存のコマンドラインツールに対して、AI Agent 向けの agentskills 管理用サブコマンドを後付けしやすくするための、Go ライブラリである。

主な用途は次の通り。

- CLI ツールに `skills` サブコマンドを追加する
- 埋め込み済みの skill ファイル群をインストールできるようにする
- Agent 向け skill の一覧表示、展開、状態確認などを提供する
- 既存の CLI 実装を大きく壊さずに機能追加する


## 設計目標

`skillsmith` の目標は以下。

- 既存 CLI に後付けしやすくするために
    - 特定のflag / cliライブラリに依存しない
    - `cobra` などの大きな CLI フレームワークに依存しない
- Agent Skills の管理用途に十分な機能を持つ
- `embed.FS` を利用する想定


## ユースケース

### 1. 既存 CLI への `skills` サブコマンド追加

たとえば `mytool` に以下のようなコマンドを生やす。

```bash
mytool skills list
mytool skills install
mytool skills install --agent codex
mytool skills install --agent claude --scope user
mytool skills status
mytool skills uninstall
mytool skills reinstall
mytool skills udpate
```

### 2. 埋め込み skill ファイルの展開


`$repo/skills` にagentskill のファイル群を置いておくことがプラクティスになりつつある。これによって、`add-skills` コマンドでインストールできるようになる。

skillsmithはさらに、それをGoのCLIバイナリに埋め込むことで、cliツールをAIフレンドリーにする。

具体的には `skills/` ディレクトリを `embed.FS` で埋め込み、ユーザー環境へ展開することを想定している。


### 3. Agent ごとの既定配置先へのインストール
- Codex: `~/.codex/skills` または `.agents/skills`
- Claude / Copilot: `~/.claude/skills` または `.claude/skills`
  - Copilot も `.claude/skills` を読むため、`claude` agent で両方カバーする

### 4. ツール独自コマンドとの共存
`skillsmith` が担当するサブコマンドのみを処理し、それ以外は既存CLIが処理する。既存CLIがskillsmithに処理を委譲する形になる。

---

## 全体アーキテクチャ

`skillsmith` は大きく以下の層に分かれる。初期実装では 2〜4 を優先し、1 は後回しとする。

1. **サブコマンドアタッチ層**
    - 各種コマンドラインライブラリに対して、`skills` サブコマンドを追加するためのインターフェースと実装を提供する。
    - `cmd.RegisteSubcommand(skillsmith.NewSubcommand())` のような形で既存CLIに組み込むことを想定
        - Cobra, urfave/cli, flag など、複数のCLIライブラリに対応するためのアダプタを提供
        - アダプタは必要に応じてユーザーが実装することも可能
        - 依存ライブラリを増やしたくないので、インターフェースの実装に留めるか、インターフェース追加だけではだめで、CLIライブラリに依存が必要であれば、アダプタ自体は同じrepo内のサブディレクトリの別Go Modulesにわける
    - **※ 後回し。まずはコア機能を固める。**
2. **コマンド実行層**
   - `skills` コマンド以下の `list/install/status/uninstall` を処理する
3. **Skill 配布層**
   - `embed.FS` または任意の `fs.FS` から skill ディレクトリを列挙・コピーする
4. **インストール先解決層**
   - `--dir` / `--agent` / `--scope` から配置先を決定する


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
        s := &skillsmith.Skillsmith{
            FS:      skillsFS,
            Version: version,
            Name:    "mytool",
        }
        return s.Execute(ctx, args[1:])
    }
    // ... 既存コマンド処理
    return nil
}
```


## skills コマンド


### サポート対象

- `skills list`
- `skills install`
- `skills status`
- `skills uninstall`
- `skills reinstall`
- `skills update`
- `skills show <name>` は任意

### コマンドイメージ

```bash
mytool skills list
mytool skills install
mytool skills install --dir ~/.codex/skills
mytool skills install --agent codex --scope user
mytool skills install --agent claude --scope repo
mytool skills status
mytool skills uninstall --agent codex --scope user
```


### 解決ルール

優先順位は以下。

1. `--dir`
2. `--agent` + `--scope`
3. ライブラリ既定値

ライブラリ既定値は `claude` + `user` とする。つまり、引数なしで `skills install` を実行した場合は、`~/.claude/skills` にインストールされることになる。


### 想定される `Agent`

- `codex`
- `claude` (Claude Code と GitHub Copilot の両方をカバー)

### 想定される `Scope`

- `user`
- `repo`


## インストール先解決

### 例

- `--agent codex --scope user` → `~/.codex/skills`
- `--agent codex --scope repo` → `.agents/skills`
- `--agent claude --scope user` → `~/.claude/skills`
- `--agent claude --scope repo` → `.claude/skills`

### 任意ディレクトリ

```bash
mytool skills install --dir /tmp/skills
```

この場合は `Agent` や `Scope` の解決は行わない。


## ファイルコピー方針

`skillsmith` は `fs.FS` からディレクトリ構造を維持したままファイルを展開する。

### 要件

- ディレクトリを再帰的にコピーできる
- 既存ファイルの扱いを制御できる
- install 対象の一覧取得ができる


## インストール時のチェック

- skills ディレクトリ以下のskillsを標準では全部インストールする
    - 個別インストールは将来の追加機能とする
- 既に既存のインストール先に同名のスキルがある場合は警告をだす
    - 複数インストールのケースも考慮して、エラー終了とはしない
    - reinstall や update を用いてくださいと案内する
- インストール時に、各スキルディレクトリ内に `.skillsmith.json` を生成してメタ情報を記録する
    - `--force` 等で上書きは可能にする
- `.skillsmith.json` が存在しないスキルは skillsmith 管理外と見なす


## メタ情報管理（`.skillsmith.json`）

### 方針

- SKILL.md を直接改変しない。メタ情報は各スキルディレクトリ内に `.skillsmith.json` として配置する
- この JSON ファイルの有無で skillsmith が管理するスキルかどうかを判別する

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

- `.skillsmith.json` が存在する → skillsmith 管理下。update / uninstall / reinstall の対象
- `.skillsmith.json` が存在しない → ユーザー作成スキルの可能性。install / reinstall では上書きしない（`--force` 時を除く）

## update
- インストール時に付加したバージョン情報をもとに、更新が必要なスキルを検出する
- `skills update` コマンドで、更新が必要なスキルのみを再インストールする
- バージョン情報が付与されていないスキルは、skillsmithによってインストールされたものではない可能性があるため、更新の対象外とする
  - 具体的には `.skillsmith.json` の有無で判定する

## reinstall
- すべてのスキルを再インストールする
- 既存のスキルがあっても、バージョンに関係なく上書きする
- skillsmithがインストールしたものではないスキル（`.skillsmith.json` が存在しない）は上書きしない（`--force` 時を除く）


## status API

`status` は次を確認する。

- install 先が存在するか
- skill が配置されているか
- 同梱 skill 一覧との差分
- 必要ならバージョンファイルの比較
  - `.skillsmith.json` のバージョンと現在の CLI バージョンを比較


## uninstall API

### 方針

- `skillsmith` が install した skill だけを削除する（`.skillsmith.json` が存在するスキルのみ）
- install 済み一覧に基づいて対象を限定する
- ディレクトリ全削除はデフォルトでは行わない


## 埋め込みスキルの前提

`skillsmith` 自体が期待するskillsのフォーマットはagentskillsの仕様に従うものとする。

```text
skills/
  mytool-cli/
    SKILL.md
    README.md
    examples/
      basic.md
```

アプリ側は以下のように埋め込む。

```go
//go:embed skills/**
var embeddedSkills embed.FS
```

### `cmd/skillsmith`

`cmd/skillsmith` はデモ用バイナリであり、ライブラリの動作確認に用いる。

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
