# 🤖 Code Reviewer

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/code-reviewer)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/code-reviewer)](https://github.com/shouni/code-reviewer/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 概要 (About) - AI 連携のインプットからアウトプットまでを一気通貫で。

**Code Reviewer** は、Google Gemini API を核とした AI コードレビューのための **高機能ツールキット** です。

ソースコードの抽出（Input）、AI モデルへの最適化（Process）、そしてマルチプラットフォームへの配信（Output）まで。AI 連携に必要なすべてのコンポーネントを独立したパッケージとして提供します。CLI ツールの構築からエンタープライズな Web アプリケーションのバックエンドまで、一貫したレビューパイプラインを迅速に実装可能です。

---

## 🏗 アーキテクチャ設計 (Architecture)

本プロジェクトは、保守性とテスト容易性を最大化するため、**クリーンアーキテクチャ**の思想を取り入れ、外部依存（インフラ）とビジネスロジックを厳格に分離したレイヤー構造を採用しています。

### レイヤー構成の境界

* **Core (抽象定義・共通ドメイン):**
  `core/ports` を中心としたシステムの核。インターフェース定義（Ports）により、具体的な実装（Adapters）に依存しないビジネスルールを規定します。
* **Adapters (具象実装):**
  `gemini`, `slack`, `remoteio`, `md` など、外部サービスや特定の技術スタックに依存する実装。Core の Port を実装し、プラグインのように容易に差し替え可能です。
* **Cross-Cutting Concerns (基盤・ユーティリティ):**
  `armor`, `httpkit`, `utils` など、全レイヤーから利用されるセキュリティ、リトライ、通信、共通処理の基盤。

---

## 🔄 シーケンスフロー (Sequence Flow)

```mermaid
sequenceDiagram
  autonumber
  actor App as Application (Pipeline)

  box rgb(240, 248, 255) pkg/core & publisher (Library)
    participant Git as core/git
    participant Prompts as core/prompts (Builder)
    participant Pub as publisher (Impl)
    participant MD as ports/Render (Interface)
  end

  box rgb(255, 250, 240) internal/runner
    participant Review as runner/Review
    participant AppPub as runner/Publish (Wrapper)
  end

  box rgb(250, 250, 250) Adapters [Implementations]
    participant Gemini as gemini (Adapter)
    participant S3 as remoteio (Writer)
    participant Slack as slack (Adapter)
  end

  Note over App, Review: [1. INPUT & PROCESS Phase]

  App->>Review: 1.1 レビュー実行要求
  activate Review
  Review->>Git: 1.2 SSH認証 & 差分抽出
  Git-->>Review: 1.3 差分データ (diff)

%% Prompts Builder の登場
  Review->>Prompts: 1.4 Build(mode, data)
  activate Prompts
  Note right of Prompts: template.Execute で<br/>プロンプトを動的生成
  Prompts-->>Review: 1.5 構築済みプロンプト文字列
  deactivate Prompts

  Review->>Gemini: 1.6 AI解析要求
  Gemini-->>Review: 1.7 レビュー結果 (Markdown)
  Review-->>App: 1.8 ReviewData 返却
  deactivate Review

  Note over App, AppPub: [2. OUTPUT Phase (Unified Workflow)]

  App->>AppPub: 2.1 成果物公開要求
  activate AppPub

  AppPub->>Pub: 2.2 Publish(ctx, uri, data)
  activate Pub

  Note right of Pub: 2.3 convertMarkdownToHTML<br/>(メタ情報を付与)

  Pub->>MD: 2.4 HTML変換要求 (Run)
  activate MD
  MD-->>Pub: 2.5 HTMLデータ返却
  deactivate MD

  Pub->>S3: 2.6 Write(uri, html, contentType)
  activate S3
  S3-->>Pub: 2.7 書き込み完了
  deactivate S3

  Pub-->>AppPub: 2.8 公開処理完了
  deactivate Pub

  AppPub->>Slack: 2.9 レビュー完了通知 (公開URL含む)
  activate Slack
  Slack-->>AppPub: 2.10 通知完了
  deactivate Slack

  AppPub-->>App: 2.11 パイプライン完了
  deactivate AppPub

```

---

### プロジェクト構成図 (Directory Tree)

```text
code-reviewer/
├── core/                # 【最重要】ドメイン境界・抽象インターフェース
│   ├── ports/           # AI, Git, Publisher などの抽象定義 (Interface)
│   ├── ai/              # AI 連携の共通ロジック
│   ├── git/             # Git 操作のコア・認証・ローカル管理
│   ├── prompts/         # プロンプト構築エンジン
│   ├── publisher/       # 成果物出力のオーケストレーション
│   └── resource/        # リソースローダー
├── md/                  # Markdown 変換・HTML レンダリング基盤
│   ├── converter/       # Markdown 解析・変換
│   ├── renderer/        # HTML テンプレート・CSS 定義
│   └── runner/          # レンダリング実行パイプライン
├── gemini/              # Gemini API クライアント (File API 対応)
├── slack/               # Slack 通知基盤 (Block Kit 対応)
├── remoteio/            # クラウドストレージ (GCS/S3) 抽象化レイヤー
├── armor/               # 防御層 (SSRF 対策・リトライ戦略)
├── httpkit/             # 高機能 HTTP クライアント (Stream, Option 対応)
├── clibase/             # CLI 実行基盤 (Cobra 統合)
└── utils/               # 共通ユーティリティ (env, text, time, path)
```

---

### 主要パッケージの役割詳細

| パッケージ | 概要 |
| :--- | :--- |
| **`core/ports`** | システムが外部（AI, Git, 保存先）に求める契約を定義。DIP の要。 |
| **`core/git`** | SSH認証をサポートした、セキュアで柔軟な Git リポジトリ操作。|
| **`armor`** | `securenet` による通信先検証と `retry` による堅牢な実行制御。 |
| **`gemini`** | `File API` を含む Google Gemini 固有の通信・型定義を集約。 |
| **`md`** | テンプレートエンジンを用いた HTML レポート生成を担う独立した変換層。 |
| **`remoteio`** | プロトコル（gs://, s3://）を意識せずにファイルを読み書きする抽象化。 |

---

### 💡 設計のハイライト

* **DIP (依存性逆転の原則):** `core/ports` を介することで、ビジネスロジックを変更せずに AI モデルや通知先を自在にスワップ可能です。
* **Security by Design:** `armor/securenet` により、内部ネットワークへの意図しないアクセス (SSRF) を防ぎ、エンタープライズ品質の安全な通信を保証します。
* **High Reliability:** `httpkit` と `armor/retry` が連携し、不安定なネットワーク環境下でも AI 連携の完遂率を最大化します。
* **Cloud Native Ready:** `remoteio` により、ローカル、GCS、S3 などのストレージ環境を意識せず、シームレスに成果物をデプロイできます。

---

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---

