---
name: fyne-docs
description: |
  Fyne GUI 라이브러리 코드 작성 시 자동 활성화.
  fyne.io/fyne/v2 import, widget, container, dialog, canvas 관련 작업 감지 시 사용.
---

# Fyne Documentation Skill

## 자동 활성화 조건

다음 상황에서 이 Skill이 자동으로 활성화됩니다:

- `fyne.io/fyne/v2` 패키지 관련 코드 작성
- Fyne widget, container, dialog, canvas 구현
- Go GUI 개발 관련 질문
- DEV_PROC.md의 커스텀 컴포넌트 구현

## Context7 MCP 연동

Fyne 최신 문서 조회 시 Context7 사용:

- **Library ID**: `/fyne-io/docs.fyne.io`
- **조회 방법**: `mcp__context7__get-library-docs` 호출
- **topic 파라미터**: 필요한 주제 (widget, container, dialog, canvas, theme 등)

## 주요 Fyne 패키지

| 패키지 | 용도 |
|--------|------|
| `fyne.io/fyne/v2` | 코어 타입 (App, Window, Canvas, Size, Position) |
| `fyne.io/fyne/v2/widget` | UI 위젯 (Button, Entry, Label, Table, List) |
| `fyne.io/fyne/v2/container` | 레이아웃 (VBox, HBox, Border, Grid, Split, Tabs) |
| `fyne.io/fyne/v2/dialog` | 다이얼로그 (Confirm, Information, Error, Progress) |
| `fyne.io/fyne/v2/canvas` | 그래픽 (Rectangle, Circle, Text, Image) |
| `fyne.io/fyne/v2/theme` | 테마 및 아이콘 |
| `fyne.io/fyne/v2/data/binding` | 데이터 바인딩 |

## 프로젝트 참조

이 프로젝트의 Fyne 관련 문서:
- `DEV_PROC.md` - 커스텀 컴포넌트 구현 예시
- `FMS_SPEC.md` - UI 기능 명세

## 사용 예시

Fyne 코드 작성 시 자동으로:
1. Context7에서 해당 위젯/기능 문서 조회
2. 최신 API 사용법 확인
3. 프로젝트의 DEV_PROC.md 패턴과 일관성 유지
