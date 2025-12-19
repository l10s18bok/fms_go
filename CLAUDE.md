# FMS (Firewall Management System) - Claude 개발 가이드

## 프로젝트 개요

Go Fyne 패키지를 이용하여 구현하는 방화벽 관리 시스템(FMS) 데스크톱 애플리케이션입니다.
기존 웹 애플리케이션(index.html)의 기능을 Go Fyne으로 재구현합니다.

---

## 중요 지침

### 기존 코드 수정 금지

- **`smartfw_hs/` 폴더 내의 모든 파일 수정 금지**
- **`index.html` 파일 수정 금지**
- 기존 코드의 함수명, 변수명 등을 임의로 수정하면 안됨
- 기존 코드는 참조용으로만 사용

### 언어 및 커뮤니케이션

- **모든 응답, 주석, 문서, 커밋 메시지를 한글로 작성**
- 질문에 대답할 때 확실하지 않으면 추론으로 대답하지 말 것
- 모르면 코드를 찾아보고 답변해야 함

### 개발 환경

| 환경 | 플랫폼 | 용도 |
|------|--------|------|
| 개발 환경 | macOS (로컬) | 코드 작성 및 빌드 |
| 테스트 환경 | Linux 서버 (원격, x86_64) | 실제 테스트 및 배포 |

---

## 프로젝트 구조

```
smartfw_add_rules/
├── CLAUDE.md           # Claude 개발 가이드 (현재 파일)
├── FMS_SPEC.md         # FMS 기능 상세 명세서
├── index.html          # 원본 웹 애플리케이션 (참조용, 수정 금지)
├── smartfw_hs/         # 커널 모듈 소스 (참조용, 수정 금지)
│   └── Makefile        # 커널 모듈 빌드 스크립트
└── (Go Fyne 소스 파일들 - 신규 생성)
```

---

## 주요 기능

### 1. 템플릿 관리
- 템플릿 목록 표시 (라디오 버튼 선택)
- 템플릿 조회/저장/삭제

### 2. 장비(방화벽) 관리
- 장비 목록 테이블 (체크박스, 장비명, 서버상태, 배포상태, 버전)
- 장비 추가/저장/삭제

### 3. 서버 운영
- 서버 상태 확인 (SSH 연결 테스트)
- 배포 (템플릿을 원격 장비의 `/proc/smartfw`에 전달)
- 배포 결과 상세 표시

### 4. 데이터 Import/Export
- JSON 파일로 내보내기/가져오기
- 데이터 초기화

---

## 기술 스택

### 필수 패키지
- `fyne.io/fyne/v2` - GUI 프레임워크
- `golang.org/x/crypto/ssh` - SSH 연결
- `encoding/json` - 데이터 저장

### SSH 인증 방식
- **SSH 키 인증 (권장)** - 운영 환경
- **암호화된 비밀번호 인증** - 개발/테스트 환경

---

## 데이터 구조

### 템플릿
```go
type Template struct {
    Version  string `json:"version"`
    Contents string `json:"contents"`
}
```

### 장비
```go
type Firewall struct {
    Index        int    `json:"index"`
    DeviceName   string `json:"deviceName"`
    ServerStatus string `json:"serverStatus"`
    DeployStatus string `json:"deployStatus"`
    Version      string `json:"version"`
    AuthType     string `json:"authType"`     // "key" 또는 "password"
    SSHUser      string `json:"sshUser"`
    SSHKeyPath   string `json:"sshKeyPath"`
    SSHPassword  string `json:"sshPassword"`  // 암호화된 비밀번호
    SSHPort      int    `json:"sshPort"`
    RemoteDir    string `json:"remoteDir"`
}
```

---

## 템플릿 규칙 포맷

규칙은 `/proc/smartfw` 커널 모듈 인터페이스로 전달됩니다.

```
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}
```

### 예시
```
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||
```

---

## 참조 문서

- [FMS_SPEC.md](FMS_SPEC.md) - 상세 기능 명세 및 구현 예시 코드
- [DEV_PROC.md](DEV_PROC.md) - 개발 절차서, UI 레이아웃, 반응형 가이드, 커스텀 컴포넌트
- [index.html](index.html) - 원본 웹 애플리케이션 (참조용)
- [smartfw_hs/Makefile](smartfw_hs/Makefile) - 규칙 포맷 및 SSH 설정 참조

---

## 빌드 및 실행

### 빌드 파일명 규칙

| 플랫폼 | 파일명 | 비고 |
|--------|--------|------|
| Linux | `fms` | 운영 환경 |
| macOS | `fms_mac` | 개발 환경 |
| Android | `fms.apk` | 모바일 (Android SDK/NDK 필요) |

### macOS에서 빌드 (개발)
```bash
go mod init fms
go get fyne.io/fyne/v2
go get golang.org/x/crypto/ssh
go build -o fms_mac .
./fms_mac
```

### Linux용 크로스 컴파일
```bash
GOOS=linux GOARCH=amd64 go build -o fms .
```

### Android용 빌드
```bash
# fyne CLI 설치 (최초 1회)
go install fyne.io/fyne/v2/cmd/fyne@latest

# Android 패키징 (Android SDK/NDK 필요)
fyne package -os android -appid com.smartfw.fms -icon icon.png -name fms

# 디바이스에 설치 (ADB)
adb install fms.apk
```

### Linux 서버에서 실행 (테스트)
```bash
scp fms user@192.168.x.x:/path/to/
ssh user@192.168.x.x
./fms
```
