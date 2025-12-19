# FMS (Firewall Management System) 기능 분석

## 개요
Go Fyne 패키지를 이용하여 구현할 FMS 웹 애플리케이션의 기능 분석 문서입니다.

---

## 주요 기능 목록

### 1. 템플릿 관리

| 기능 | 설명 |
|------|------|
| 템플릿 목록 표시 | 좌측 패널에 저장된 템플릿 버전 목록 (라디오 버튼 선택) |
| 템플릿 조회 | 선택한 템플릿의 내용을 textarea에 표시 |
| 템플릿 저장 | 버전명과 함께 새 템플릿 저장 또는 기존 템플릿 업데이트 |
| 템플릿 삭제 | 선택한 템플릿 삭제 |

---

### 2. 장비(방화벽) 관리

| 기능 | 설명 |
|------|------|
| 장비 목록 테이블 | 체크박스, 장비명, 서버상태, 배포상태, 버전 컬럼 |
| 장비 추가 | 새로운 장비 행 추가 (장비명 입력 가능) |
| 장비 저장 | 장비 목록을 저장소에 저장 |
| 장비 삭제 | 체크된 장비 삭제 |
| 전체 선택/해제 | 헤더의 체크박스로 전체 선택 토글 |

---

### 3. 서버 운영

| 기능 | 설명 |
|------|------|
| 서버 상태 확인 | 선택된 장비들의 상태 체크 (정상/정지) |
| 배포 | 선택한 템플릿을 체크된 장비에 배포 |
| 배포 결과 상세 | 배포 상태 클릭 시 상세 결과 모달 표시 |

---

### 4. 데이터 Import/Export

| 기능 | 설명 |
|------|------|
| Export | 템플릿/방화벽 데이터를 JSON 파일로 다운로드 |
| Import | JSON 파일에서 데이터 불러오기 |
| Reset | 선택한 데이터(템플릿/방화벽) 초기화 |

---

### 5. UI 컴포넌트

| 컴포넌트 | 용도 |
|----------|------|
| 확인 모달 | 작업 전 사용자 확인 (확인/취소) |
| 배포 결과 모달 | 배포 상세 결과 테이블 표시 |
| Toast 알림 | 성공/에러 메시지 표시 |
| 로딩 스피너 | 서버 상태/배포 진행 중 표시 |

---

### 6. 데이터 저장소

원본 웹 애플리케이션은 **IndexedDB**를 사용하며, Fyne 구현 시 파일 기반 저장소(JSON/SQLite)로 대체 예정입니다.

- `templateList`: 템플릿 저장 (keyPath: version)
- `firewallList`: 장비 정보 저장 (keyPath: index, autoIncrement)

---

## 화면 구성 (레이아웃)

```
┌─────────────────────────────────────────────────────────┐
│  FMS - Firewall Management System                       │
├───────────────┬─────────────────────────────────────────┤
│ 템플릿 목록    │  [템플릿 저장] [Reset][Export][Import]  │
│ ○ v1.0       │  [Delete]                               │
│ ○ v1.1       │  ┌─────────────────────────────────────┐ │
│ ● v2.0       │  │  req|INSERT|...|INPUT|FLUSH|...     │ │
│              │  │  req|INSERT|...|INPUT|ACCEPT|TCP|...│ │
│              │  │  req|INSERT|...|INPUT|DROP|UDP|...  │ │
│              │  └─────────────────────────────────────┘ │
├───────────────┴─────────────────────────────────────────┤
│ [서버상태확인] [배포]              [저장] [추가] [삭제]   │
├─────────────────────────────────────────────────────────┤
│ ☑ │ 장비명          │ 서버상태 │ 배포상태 │ 버전       │
│ ☐ │ 192.168.1.1    │ 정상     │ 성공     │ v2.0      │
│ ☐ │ 192.168.1.2    │ 정지     │ 실패     │ -         │
└─────────────────────────────────────────────────────────┘
```

---

## 상태 코드

| 코드 | 표시 텍스트 |
|------|------------|
| running | 정상 |
| stop | 정지 |
| success | 성공 |
| error | 확인요망 |
| fail | 실패 |

---

## 기술 스택

### 원본 (Web)
- Vue.js 2
- Bootstrap Vue
- Axios (HTTP Client)
- IndexedDB (로컬 저장소)

### 구현 예정 (Desktop)
- Go
- Fyne (GUI Framework)
- JSON/SQLite (로컬 저장소)
- golang.org/x/crypto/ssh (원격 장비 연결)

---

## 템플릿 규칙 포맷

템플릿에는 방화벽 규칙이 줄 단위로 나열됩니다. 규칙은 `/proc/smartfw` 커널 모듈 인터페이스로 전달됩니다.

### 규칙 포맷

```
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}
```

| 필드 | 설명 | 예시 값 |
|------|------|---------|
| 1 | 요청 타입 | `req` |
| 2 | 명령어 | `INSERT` |
| 3 | 규칙 ID | `3813792919` |
| 4 | 체인 | `INPUT` |
| 5 | 액션 | `FLUSH`, `LOG_ON`, `LOG_OFF`, `ACCEPT`, `DROP` 등 |
| 6 | 프로토콜 | `ANY`, `TCP`, `UDP` 등 |
| 7 | 출발지 | `ANY` 또는 IP/CIDR |
| 8 | 목적지 | `ANY` 또는 IP/CIDR |
| 9-11 | 추가 옵션 | 포트 등 |

### 규칙 예시

```
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|LOG_ON|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||
```

---

## 시스템 아키텍처

### 현재 (Web 방식)

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌─────────────┐
│ 브라우저  │ ──► │ 백엔드 서버   │ ──► │ 원격 장비     │ ──► │/proc/smartfw│
│index.html│ HTTP│/agent/deploy │     │ (에이전트)    │     │ 커널 모듈    │
└──────────┘     └──────────────┘     └──────────────┘     └─────────────┘
```

### 변경 후 (Go Fyne 방식)

**로컬 장비:**
```
┌──────────────┐                      ┌─────────────┐
│  Fyne 앱     │ ──────────────────► │/proc/smartfw│
│              │   직접 파일 쓰기      │ 커널 모듈    │
└──────────────┘                      └─────────────┘
```

**원격 장비:**
```
┌──────────────┐     ┌──────────────┐     ┌─────────────┐
│  Fyne 앱     │ ──► │ 원격 장비     │ ──► │/proc/smartfw│
│              │ SSH │              │     │ 커널 모듈    │
└──────────────┘     └──────────────┘     └─────────────┘
```

---

## 아키텍처 비교

| 항목 | Web (현재) | Go Fyne (변경 후) |
|------|-----------|------------------|
| 백엔드 서버 | 필요 | **불필요** |
| 로컬 배포 | 서버 경유 | **직접 쓰기** |
| 원격 배포 | 에이전트 필요 | **SSH 직접 연결** |
| 의존성 | 브라우저 + 서버 + 에이전트 | **Fyne 앱만** |
| 보안 | HTTP 통신 | **SSH 암호화** |

---

## Go Fyne 구현 예시

### 로컬 장비: 직접 `/proc/smartfw` 쓰기

```go
func writeRule(rule string) error {
    f, err := os.OpenFile("/proc/smartfw", os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()

    _, err = f.WriteString(rule)
    return err
}
```

### 원격 장비: SSH 인증 방식

#### 방법 1: SSH 키 인증 (권장)

비밀번호 없이 키 파일로 자동 인증하는 방식입니다.

```go
import (
    "os"
    "fmt"
    "golang.org/x/crypto/ssh"
)

// SSH 키 인증 설정
func getSSHConfigWithKey(user, keyPath string) (*ssh.ClientConfig, error) {
    key, err := os.ReadFile(keyPath)  // 예: ~/.ssh/id_rsa
    if err != nil {
        return nil, fmt.Errorf("키 파일 읽기 실패: %v", err)
    }

    signer, err := ssh.ParsePrivateKey(key)
    if err != nil {
        return nil, fmt.Errorf("키 파싱 실패: %v", err)
    }

    return &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.PublicKeys(signer),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }, nil
}

// SSH 키로 원격 배포
func deployWithKey(ip, user, keyPath, rule string) error {
    config, err := getSSHConfigWithKey(user, keyPath)
    if err != nil {
        return err
    }

    client, err := ssh.Dial("tcp", ip+":22", config)
    if err != nil {
        return err
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()

    cmd := fmt.Sprintf("echo '%s' > /proc/smartfw", rule)
    return session.Run(cmd)
}
```

**사전 설정:**
```bash
# 1. SSH 키 생성 (클라이언트)
ssh-keygen -t rsa -b 4096 -f ~/.ssh/fms_key

# 2. 공개키를 원격 서버에 복사
ssh-copy-id -i ~/.ssh/fms_key.pub user@192.168.1.1
```

#### 방법 2: 암호화된 비밀번호 인증

비밀번호를 AES 암호화하여 저장하고, 연결 시 자동 복호화하는 방식입니다.

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "io"
    "golang.org/x/crypto/ssh"
)

// AES 암호화 (저장 시)
func encrypt(plaintext, key string) (string, error) {
    block, err := aes.NewCipher([]byte(key)) // key는 16/24/32 바이트
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }

    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AES 복호화 (연결 시)
func decrypt(ciphertext, key string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(ciphertext)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return "", fmt.Errorf("암호문이 너무 짧음")
    }

    nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
    if err != nil {
        return "", err
    }

    return string(plaintext), nil
}

// 암호화된 비밀번호로 원격 배포
func deployWithPassword(ip, user, encryptedPassword, encryptionKey, rule string) error {
    // 저장된 암호화 비밀번호 → 복호화
    password, err := decrypt(encryptedPassword, encryptionKey)
    if err != nil {
        return fmt.Errorf("비밀번호 복호화 실패: %v", err)
    }

    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    client, err := ssh.Dial("tcp", ip+":22", config)
    if err != nil {
        return err
    }
    defer client.Close()

    session, err := client.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()

    cmd := fmt.Sprintf("echo '%s' > /proc/smartfw", rule)
    return session.Run(cmd)
}
```

#### 통합 SSH 클라이언트

두 가지 인증 방식을 모두 지원하는 통합 구현입니다.

```go
// 인증 방식에 따른 SSH 설정
func getSSHConfig(fw Firewall, encryptionKey string) (*ssh.ClientConfig, error) {
    var authMethod ssh.AuthMethod

    if fw.AuthType == "key" {
        // SSH 키 인증
        key, err := os.ReadFile(fw.SSHKeyPath)
        if err != nil {
            return nil, err
        }
        signer, err := ssh.ParsePrivateKey(key)
        if err != nil {
            return nil, err
        }
        authMethod = ssh.PublicKeys(signer)
    } else {
        // 비밀번호 인증 (자동 복호화)
        password, err := decrypt(fw.SSHPassword, encryptionKey)
        if err != nil {
            return nil, err
        }
        authMethod = ssh.Password(password)
    }

    return &ssh.ClientConfig{
        User:            fw.SSHUser,
        Auth:            []ssh.AuthMethod{authMethod},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }, nil
}
```

#### 인증 방식 비교

| 항목 | SSH 키 인증 | 암호화 비밀번호 |
|------|-----------|---------------|
| 보안성 | 높음 | 중간 |
| 설정 복잡도 | 초기 설정 필요 | 간단 |
| 저장 데이터 | 키 파일 경로 | 암호화된 비밀번호 |
| 권장 상황 | 운영 환경 | 개발/테스트 환경 |
```

---

## 데이터 구조 (Go Struct)

### 템플릿 구조체

```go
type Template struct {
    Version  string `json:"version"`  // 템플릿 버전명 (Primary Key)
    Contents string `json:"contents"` // 방화벽 규칙 내용 (줄 단위)
}
```

### 장비(방화벽) 구조체

```go
type Firewall struct {
    Index        int    `json:"index"`        // 고유 ID (Auto Increment)
    DeviceName   string `json:"deviceName"`   // 장비 IP 주소
    ServerStatus string `json:"serverStatus"` // 서버 상태 (running/stop)
    DeployStatus string `json:"deployStatus"` // 배포 상태 (success/fail/error)
    Version      string `json:"version"`      // 배포된 템플릿 버전
    // SSH 접속 정보
    AuthType     string `json:"authType"`     // 인증 방식: "key" 또는 "password"
    SSHUser      string `json:"sshUser"`      // SSH 계정 (예: root, hyung500)
    SSHKeyPath   string `json:"sshKeyPath"`   // SSH 키 파일 경로 (AuthType=key)
    SSHPassword  string `json:"sshPassword"`  // 암호화된 비밀번호 (AuthType=password)
    SSHPort      int    `json:"sshPort"`      // SSH 포트 (기본값: 22)
    RemoteDir    string `json:"remoteDir"`    // 원격 배포 경로 (예: /root/dev2)
}
```

### 배포 결과 구조체

```go
type DeployResult struct {
    IP     string       `json:"ip"`     // 장비 IP
    Status string       `json:"status"` // 배포 상태
    Info   []RuleResult `json:"info"`   // 규칙별 결과
}

type RuleResult struct {
    Rule   string `json:"rule"`   // 규칙 내용
    Text   string `json:"text"`   // 규칙 설명
    Status string `json:"status"` // 결과 (ok/error/unfind/validation)
    Reason string `json:"reason"` // 실패 사유
}
```

---

## 보안 참고사항

현재 Makefile에 SSH 비밀번호가 평문으로 노출되어 있습니다. Go Fyne 구현 시 권장사항:

| 항목 | 현재 | 권장 |
|------|------|------|
| 인증 방식 | 비밀번호 평문 | **SSH 키 인증** |
| 비밀번호 저장 | 평문 | **암호화 저장** |
| 설정 파일 | 소스코드 내 하드코딩 | **별도 설정 파일 또는 환경변수**
