# 🤖 Code Reviewer

[![Language](https://img.shields.io/badge/Language-Go-blue)](https://golang.org/)
[![Go Version](https://img.shields.io/github/go-mod/go-version/shouni/code-reviewer)](https://golang.org/)
[![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/shouni/code-reviewer)](https://github.com/shouni/code-reviewer/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 概要 (About) — Go 1.24 世代の安定版、AI レビュー・ツールキット

**Code Reviewer** は、コアエンジン **[`shouni/gemini-reviewer-core`](https://github.com/shouni/gemini-reviewer-core)** をベースに、Go 1.24 世代のプロジェクト向けに構成を簡略化・再構築したサブ・プロジェクトです。

---

## 🖇 パッケージツリー (Package Tree)

```text
code-reviewer/
├── core/                # ドメイン境界・インターフェース (Ports)
├── md/                  # Markdown 変換・HTML レンダリング
├── gemini/              # Gemini API クライアント (File API 対応)
├── slack/               # Slack 通知基盤 (Block Kit 対応)
├── remoteio/            # クラウドストレージ (GCS/S3) 抽象化
├── armor/               # 防御層 (SSRF 対策・リトライ戦略)
├── httpkit/             # 高機能 HTTP クライアント (Stream, Option 対応)
├── clibase/             # CLI 実行基盤 (Cobra 統合)
└── utils/               # 共通ユーティリティ (env, text, time, path)
```

---

### 📜 ライセンス (License)

このプロジェクトは [MIT License](https://opensource.org/licenses/MIT) の下で公開されています。

---
