# FMS (Firewall Management System) - Claude 개발 가이드

## 프로젝트 개요

Go Fyne 패키지를 이용하여 구현하는 방화벽 관리 시스템(FMS) 데스크톱 애플리케이션입니다.
기존 웹 애플리케이션(index.html)의 기능 플러스 Go Fyne으로 완전히 재구현합니다.

---

## 중요 지침

### 언어 및 커뮤니케이션

- **모든 응답, 주석, 문서, 커밋 메시지를 한글로 작성**
- 질문에 대답할 때 확실하지 않으면 추론으로 대답하지 말 것
- 모르면 코드를 찾아보고 답변해야 함

### 개발 환경

| 환경 | 플랫폼 | 용도 |
|------|--------|------|
| 개발 환경 | windows (로컬) | 코드 작성 및 빌드 |

---

## 프로젝트 구조

```
fms_go/
├── CLAUDE.md               # Claude 개발 가이드 (현재 파일)
├── flutter_wails_fyne.md   # 프레임워크 비교 문서
├── index.html              # 원본 웹 애플리케이션 (참조용, 수정 금지)
├── docs/                   # 개발 문서
│   ├── rule-builder-prd.md       # 규칙 빌더 PRD 문서
│   └── rule-builder-checklist.md # 규칙 빌더 구현 체크리스트
├── fms_fyne/               # Go Fyne 프로젝트
├── fms_flutter/            # Flutter 프로젝트
└── fms_wails/              # Wails 프로젝트
```

---

## 기술 스택

### 공통 패키지
- `net/http` - HTTP 통신 (Direct 모드)
- `encoding/json` - 데이터 저장

### fms_fyne 전용
- `fyne.io/fyne/v2` - GUI 프레임워크

### fms_wails 전용
- `github.com/wailsapp/wails/v2` - 데스크톱 앱 프레임워크
- React + TypeScript - 프론트엔드

---

## 통신 아키텍처

### Direct 모드
```
FMS 클라이언트 → 방화벽 장비 (직접 연결)
```

- 상태 확인: `GET http://{장비IP}:{장비PORT}/device-report`
- 배포: `POST http://{장비IP}:{장비PORT}/agent/req-deploy`

---

## 템플릿 규칙 포맷

규칙은 Agent 명령어 형식으로 장비에 전달됩니다.

### 일반 방화벽 규칙
```
./agent -f -I -c {CHAIN} -a {ACTION} -p {PROTOCOL} -s {SRC} -d {DST} --dport {PORT}
```

### NAT 규칙
```
./agent -f -I -a NAT -p {PROTOCOL}?{NAT_TYPE} -s {SRC} --dest {DEST_IP} --dport {PORT}
```

### 예시
```
./agent -f -I -c INPUT -a ACCEPT -p tcp -s 192.168.1.0/24 --dport 80
./agent -f -I -a NAT -p tcp?DNAT --dest 192.168.30.180 --dport 6080 --to-port 8080
./agent -f -I -a NAT -p tcp?SNAT -s 192.168.1.0/24 --dest 203.0.113.1
```

---

## 참조 문서

- [index.html](index.html) - 원본 웹 애플리케이션 (참조용)
- [docs/](docs/) - 개발 문서 (PRD, 체크리스트 등)

---

## 빌드 및 실행

### 빌드 파일명 규칙

| 프로젝트 | 파일명 | 비고 |
|----------|--------|------|
| fms_fyne | `fms_fyne.exe` | Go Fyne GUI |
| fms_wails | `fms_wails.exe` | Wails + React |

### fms_fyne 빌드 (Windows)

**배포용 빌드 (콘솔 창 없음):**
```bash
cd fms_fyne
go mod download
go mod tidy
go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .
```

**디버깅용 빌드 (콘솔 창 표시):**
```bash
cd fms_fyne
go build -ldflags "-H windows -s -w" -o fms_fyne.exe .
```
- 콘솔 창이 함께 열려 `log.Println()` 출력 확인 가능
- 개발 중 디버깅에 유용

### fms_wails 빌드 (Windows)

**CMD / Git Bash / PowerShell:**
```bash
cd fms_wails
wails build
# 또는 개발 모드
wails dev
```
빌드 결과: `build/bin/fms_wails.exe`

### 빌드 옵션 설명

`-ldflags "-H windowsgui -s -w"` 옵션의 의미:

| 옵션 | 설명 |
|------|------|
| `-H windowsgui` | Windows GUI 애플리케이션으로 빌드 (콘솔 창 없음) |
| `-s` | 심볼 테이블 제거 (파일 크기 감소) |
| `-w` | DWARF 디버그 정보 제거 (파일 크기 감소) |

**중요**: 이 옵션 없이 빌드하면 exe 파일이 실행되지 않을 수 있음
