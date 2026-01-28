# WebSocket 기반 장비 모니터링 전환 조사 문서

## 1. 개요

### 1.1 목적
장비 관리에서 주기적인 상태 체크 방식을 HTTP 폴링에서 WebSocket 실시간 방식으로 전환하여:
- 실시간 상태 모니터링 지원 (연결 상태 + 시스템 리소스)
- 네트워크 오버헤드 감소
- 서버 부하 감소

### 1.2 모니터링 데이터

| 데이터 | 설명 | 예시 |
|--------|------|------|
| 연결 상태 | 장비 연결 여부 | running, stop |
| CPU 사용률 | CPU 사용 퍼센트 | 22.7% |
| 메모리 사용률 | RAM 사용 퍼센트 | 98.6% |
| 디스크 사용률 | 스토리지 사용 퍼센트 | 50.9% |
| 네트워크 | 네트워크 트래픽 | 11.28 MB/s |

### 1.3 현재 방식 vs 목표 방식

| 항목 | 현재 (HTTP 폴링) | 목표 (WebSocket) |
|------|------------------|------------------|
| 연결 방식 | 주기적 HTTP GET 요청 | 상시 연결 유지 |
| 상태 업데이트 | 60초마다 전체 장비 체크 | 서버에서 실시간 푸시 |
| 데이터 종류 | 연결 상태만 | 연결 상태 + 시스템 리소스 |
| 네트워크 사용 | 요청마다 새 연결 | 단일 연결 재사용 |
| 지연 시간 | 최대 60초 | 즉시 |

---

## 2. 현재 구현 분석

### 2.1 핵심 파일 및 역할

| 파일 | 역할 | 주요 함수/구조체 |
|------|------|------------------|
| `internal/ui/device_tab.go` | 자동 상태 체크 UI 로직 | `startAutoStatusCheck()`, `performAutoStatusCheck()` |
| `internal/http/client.go` | HTTP 클라이언트 | `CheckHealthDirect()`, `CheckHealth()` |
| `internal/deploy/deployer.go` | 배포 및 상태 체크 | `HealthCheckBatch()`, `HealthCheck()` |
| `internal/model/config.go` | 설정 모델 | `Config`, `DefaultConfig()` |

### 2.2 현재 상태 체크 흐름

```
┌─────────────────────────────────────────────────────────────────┐
│                        device_tab.go                             │
│                                                                  │
│   ┌─────────────────┐                                           │
│   │ time.Ticker     │──60초 주기──▶ performAutoStatusCheck()   │
│   │ (autoCheckTicker)│                      │                   │
│   └─────────────────┘                       ▼                   │
│                              deployer.HealthCheckBatch()         │
│                                            │                    │
│                               ┌────────────┴────────────┐       │
│                               ▼                         ▼       │
│                        sync.WaitGroup (병렬 처리)               │
│                        각 장비에 HTTP GET 요청                   │
└─────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                        http/client.go                            │
│                                                                  │
│   CheckHealthDirect(fw *model.Firewall)                         │
│   GET http://{IP}:{port}/agent/respCheck                        │
│   응답 200 OK → "running", 그 외 → "stop"                        │
└─────────────────────────────────────────────────────────────────┘
```

### 2.3 주요 코드 분석

#### device_tab.go - 자동 상태 체크

```go
type DeviceTab struct {
    autoCheckEnabled  bool           // 자동 체크 활성화 여부
    autoCheckInterval int            // 체크 주기 (초)
    autoCheckTicker   *time.Ticker   // 타이머
    stopAutoCheck     chan struct{}  // 중지 신호 채널
}

func (d *DeviceTab) startAutoStatusCheck() {
    interval := time.Duration(d.autoCheckInterval) * time.Second
    d.autoCheckTicker = time.NewTicker(interval)
    d.stopAutoCheck = make(chan struct{})

    go func() {
        for {
            select {
            case <-d.autoCheckTicker.C:
                d.performAutoStatusCheck()
            case <-d.stopAutoCheck:
                return
            }
        }
    }()
}

func (d *DeviceTab) performAutoStatusCheck() {
    // deployer.HealthCheckBatch() 호출
    // 각 장비 상태 업데이트
    // fyne.Do()로 UI 갱신
}
```

#### http/client.go - HTTP 상태 체크

```go
func (c *Client) CheckHealthDirect(fw *model.Firewall) (bool, error) {
    port := fw.HealthCheckPort
    if port == 0 {
        port = c.config.HealthCheckPort
    }
    if port == 0 {
        port = model.DefaultAPIPort
    }

    path := fw.HealthCheckPath
    if path == "" {
        path = c.config.HealthCheckPath
    }
    if path == "" {
        path = model.DefaultHealthCheckPath
    }

    url := fmt.Sprintf("http://%s:%d%s", fw.DeviceIP, port, path)

    resp, err := c.httpClient.Get(url)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()

    return resp.StatusCode == http.StatusOK, nil
}
```

#### deployer.go - 배치 상태 체크 (병렬 처리)

```go
func (d *Deployer) HealthCheckBatch(firewalls []*model.Firewall) error {
    var wg sync.WaitGroup

    for _, fw := range firewalls {
        wg.Add(1)
        go func(f *model.Firewall) {
            defer wg.Done()
            d.HealthCheck(f)
        }(fw)
    }

    wg.Wait()
    return nil
}
```

### 2.4 현재 설정 구조

```go
type Config struct {
    TimeoutSeconds      int    `json:"timeoutSeconds"`      // HTTP 타임아웃 (5~120초)
    AutoStatusCheck     bool   `json:"autoStatusCheck"`     // 자동 체크 활성화
    StatusCheckInterval int    `json:"statusCheckInterval"` // 체크 주기 (10~300초)
    HealthCheckPath     string `json:"healthCheckPath"`     // 기본: /agent/respCheck
    HealthCheckPort     int    `json:"healthCheckPort"`     // 기본: 8080
}
```

---

## 3. WebSocket 라이브러리 선택

### 3.1 후보 라이브러리 비교

| 항목 | gorilla/websocket | nhooyr/websocket |
|------|-------------------|------------------|
| GitHub Stars | 22k+ | 4k+ |
| 성숙도 | 매우 성숙 (2013~) | 비교적 신규 (2019~) |
| 유지보수 | 현재 유지보수 모드 | 활발히 개발 중 |
| API 스타일 | 저수준, 세밀한 제어 | 고수준, context 기반 |
| 자동 재연결 | 직접 구현 필요 | 직접 구현 필요 |
| Ping/Pong | 내장 지원 | 내장 지원 |
| 압축 | 지원 | 지원 |
| 문서/예제 | 풍부함 | 보통 |

### 3.2 선택: gorilla/websocket

**선택 이유**:
1. Go 생태계에서 사실상 표준으로 사용됨
2. 문서와 예제가 풍부하여 구현 용이
3. 저수준 API로 연결 관리를 세밀하게 제어 가능
4. 안정성이 검증됨 (10년 이상 운영)

**설치 방법**:
```bash
cd fms_fyne
go get github.com/gorilla/websocket@v1.5.3
```

---

## 4. WebSocket 메시지 프로토콜 설계

### 4.1 기본 메시지 구조

```go
// 모든 메시지의 공통 프레임
type WSMessage struct {
    Type      string          `json:"type"`      // 메시지 유형
    Timestamp string          `json:"timestamp"` // ISO 8601 형식
    Data      json.RawMessage `json:"data"`      // 페이로드
}
```

### 4.2 데이터 구조체

```go
// 시스템 리소스 정보
type SystemResources struct {
    CPU     float64 `json:"cpu"`      // CPU 사용률 (%)
    Memory  float64 `json:"memory"`   // 메모리 사용률 (%)
    Disk    float64 `json:"disk"`     // 디스크 사용률 (%)
    Network float64 `json:"network"`  // 네트워크 트래픽 (MB/s)
}

// 장비 상태 (연결 상태 + 시스템 리소스)
type DeviceStatus struct {
    DeviceIP   string           `json:"device_ip"`   // 장비 IP
    Status     string           `json:"status"`      // "running", "stop"
    Resources  *SystemResources `json:"resources"`   // 시스템 리소스 (running일 때만)
    CheckedAt  string           `json:"checked_at"`  // 체크 시간
}
```

### 4.3 메시지 유형

| Type | 방향 | 용도 |
|------|------|------|
| `subscribe` | 클라이언트 → 서버 | 장비 모니터링 구독 |
| `unsubscribe` | 클라이언트 → 서버 | 구독 해제 |
| `status_update` | 서버 → 클라이언트 | 연결 상태 변경 알림 |
| `resources_update` | 서버 → 클라이언트 | 시스템 리소스 업데이트 |
| `ping` | 양방향 | 연결 유지 확인 |
| `pong` | 양방향 | Ping 응답 |
| `error` | 서버 → 클라이언트 | 오류 알림 |

### 4.4 메시지 예시

#### 클라이언트 → 서버: 장비 구독

```json
{
    "type": "subscribe",
    "timestamp": "2026-01-26T10:00:00Z",
    "data": {
        "device_ips": ["192.168.1.10", "192.168.1.11", "192.168.1.12"]
    }
}
```

#### 클라이언트 → 서버: 구독 해제

```json
{
    "type": "unsubscribe",
    "timestamp": "2026-01-26T10:05:00Z",
    "data": {
        "device_ips": ["192.168.1.10"]
    }
}
```

#### 서버 → 클라이언트: 연결 상태 업데이트

```json
{
    "type": "status_update",
    "timestamp": "2026-01-26T10:00:05Z",
    "data": {
        "device_ip": "192.168.1.10",
        "status": "running",
        "checked_at": "2026-01-26T10:00:05Z"
    }
}
```

#### 서버 → 클라이언트: 시스템 리소스 업데이트 (개별)

```json
{
    "type": "resources_update",
    "timestamp": "2026-01-26T10:00:05Z",
    "data": {
        "device_ip": "192.168.1.10",
        "status": "running",
        "resources": {
            "cpu": 22.7,
            "memory": 98.6,
            "disk": 50.9,
            "network": 11.28
        },
        "checked_at": "2026-01-26T10:00:05Z"
    }
}
```

#### 서버 → 클라이언트: 시스템 리소스 업데이트 (일괄)

```json
{
    "type": "resources_update",
    "timestamp": "2026-01-26T10:00:05Z",
    "data": {
        "devices": [
            {
                "device_ip": "192.168.1.10",
                "status": "running",
                "resources": {
                    "cpu": 22.7,
                    "memory": 98.6,
                    "disk": 50.9,
                    "network": 11.28
                },
                "checked_at": "2026-01-26T10:00:05Z"
            },
            {
                "device_ip": "192.168.1.11",
                "status": "running",
                "resources": {
                    "cpu": 45.2,
                    "memory": 67.3,
                    "disk": 30.1,
                    "network": 5.67
                },
                "checked_at": "2026-01-26T10:00:05Z"
            },
            {
                "device_ip": "192.168.1.12",
                "status": "stop",
                "resources": null,
                "checked_at": "2026-01-26T10:00:05Z"
            }
        ]
    }
}
```

#### 서버 → 클라이언트: 오류

```json
{
    "type": "error",
    "timestamp": "2026-01-26T10:00:05Z",
    "data": {
        "code": "DEVICE_NOT_FOUND",
        "message": "장비를 찾을 수 없습니다: 192.168.1.99"
    }
}
```

---

## 5. 클라이언트 구현 설계

### 5.1 파일 구조

```
fms_fyne/internal/
├── websocket/                 # 신규 패키지
│   ├── message.go            # 메시지 타입 정의
│   ├── client.go             # WebSocket 클라이언트 핵심 로직
│   └── reconnect.go          # 재연결 및 하트비트 로직
├── http/
│   └── client.go             # 기존 HTTP 클라이언트 (유지)
├── model/
│   └── config.go             # 설정 필드 추가
└── ui/
    └── device_tab.go         # WebSocket 통합
```

### 5.2 WebSocket 클라이언트 구조

```go
// internal/websocket/client.go

package websocket

type ClientConfig struct {
    ServerURL         string        // ws://agent-server:8080/ws/status
    ReconnectInterval time.Duration // 재연결 간격 (기본 5초)
    MaxReconnectTries int           // 최대 재연결 시도 (0=무한)
    PingInterval      time.Duration // Ping 주기 (기본 60초, 게이트웨이 타임아웃 2분 미만)
    WriteTimeout      time.Duration // 쓰기 타임아웃 (기본 10초)
    ReadTimeout       time.Duration // 읽기 타임아웃 (기본 60초)
}

type Client struct {
    config     *ClientConfig
    conn       *websocket.Conn
    mu         sync.RWMutex

    // 상태
    connected  bool
    ctx        context.Context
    cancel     context.CancelFunc

    // 이벤트 채널
    statusChan    chan *StatusUpdate    // 연결 상태 업데이트 수신
    resourcesChan chan *ResourcesUpdate // 시스템 리소스 업데이트 수신
    errorChan     chan error            // 에러 수신

    // 구독 중인 장비 IP 목록
    subscribed map[string]bool

    // Pong 응답 추적 (타임아웃 리셋용)
    lastPong   time.Time           // 마지막 Pong 수신 시간
}

// 이벤트 타입
type StatusUpdate struct {
    DeviceIP  string `json:"device_ip"`
    Status    string `json:"status"`     // "running", "stop"
    CheckedAt string `json:"checked_at"`
}

type ResourcesUpdate struct {
    DeviceIP  string           `json:"device_ip"`
    Status    string           `json:"status"`
    Resources *SystemResources `json:"resources"`
    CheckedAt string           `json:"checked_at"`
}

// 주요 메서드
func NewClient(config *ClientConfig) *Client
func (c *Client) Connect() error
func (c *Client) Close()
func (c *Client) Subscribe(deviceIPs []string) error
func (c *Client) Unsubscribe(deviceIPs []string) error
func (c *Client) StatusUpdates() <-chan *StatusUpdate
func (c *Client) ResourcesUpdates() <-chan *ResourcesUpdate  // 신규
func (c *Client) Errors() <-chan error
func (c *Client) IsConnected() bool
```

### 5.3 재연결 로직

```go
// 지수 백오프 재연결
func (c *Client) reconnectLoop() {
    backoff := c.config.ReconnectInterval  // 초기값: 5초
    maxBackoff := 60 * time.Second
    attempts := 0

    for {
        select {
        case <-c.ctx.Done():
            return
        default:
            if c.IsConnected() {
                time.Sleep(time.Second)
                continue
            }

            log.Printf("[WebSocket] 재연결 시도 #%d", attempts+1)

            if err := c.connect(); err != nil {
                log.Printf("[WebSocket] 연결 실패: %v", err)
                attempts++

                // 지수 백오프: 5s → 10s → 20s → 40s → 60s (최대)
                time.Sleep(backoff)
                backoff = min(backoff*2, maxBackoff)
                continue
            }

            // 연결 성공 - 백오프 초기화
            log.Printf("[WebSocket] 재연결 성공")
            backoff = c.config.ReconnectInterval
            attempts = 0

            // 이전 구독 복원
            c.resubscribeAll()
        }
    }
}
```

### 5.4 하트비트 (Ping/Pong)

> **중요**: 중간 게이트웨이(라우터/방화벽)가 idle 연결을 끊는 것을 방지하기 위해
> Ping/Pong을 사용합니다. 일반적으로 게이트웨이 타임아웃이 2분(120초)이므로,
> PingInterval은 60초로 설정하여 연결을 유지합니다.

#### 5.4.1 Pong 핸들러 설정 (연결 시)

```go
// connect() 함수 내에서 Pong 핸들러 설정
func (c *Client) connect() error {
    // ... 연결 로직 ...

    // Pong 수신 시 타임아웃 리셋 (핵심!)
    c.conn.SetPongHandler(func(appData string) error {
        c.mu.Lock()
        defer c.mu.Unlock()

        c.lastPong = time.Now()
        // 읽기 타임아웃 리셋 → 게이트웨이 타임아웃 방지
        c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
        log.Printf("[WebSocket] Pong 수신 - 타임아웃 리셋")
        return nil
    })

    // 초기 lastPong 설정
    c.lastPong = time.Now()

    return nil
}
```

#### 5.4.2 Ping 전송 루프

```go
func (c *Client) pingLoop() {
    ticker := time.NewTicker(c.config.PingInterval)  // 60초
    defer ticker.Stop()

    for {
        select {
        case <-c.ctx.Done():
            return
        case <-ticker.C:
            c.mu.Lock()
            if c.conn != nil {
                deadline := time.Now().Add(c.config.WriteTimeout)
                if err := c.conn.WriteControl(
                    websocket.PingMessage,
                    []byte{},
                    deadline,
                ); err != nil {
                    log.Printf("[WebSocket] Ping 실패: %v", err)
                    c.connected = false
                }

                // Pong 응답 확인: 마지막 Pong이 PingInterval*2 보다 오래됐으면 연결 끊김 판단
                if time.Since(c.lastPong) > c.config.PingInterval*2 {
                    log.Printf("[WebSocket] Pong 응답 없음 (마지막: %v) - 연결 끊김 판단",
                        c.lastPong.Format("15:04:05"))
                    c.connected = false
                }
            }
            c.mu.Unlock()
        }
    }
}
```

#### 5.4.3 Ping/Pong 동작 흐름

```
┌─────────────────────────────────────────────────────────────────┐
│  Ping/Pong + 타임아웃 리셋 흐름                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   pingLoop: 60초마다 Ping 전송                                   │
│        │                                                        │
│        ▼                                                        │
│   서버/프록시에서 Pong 응답                                       │
│        │                                                        │
│        ▼                                                        │
│   SetPongHandler 호출됨                                          │
│        │                                                        │
│        ├─▶ lastPong = time.Now()                                │
│        └─▶ SetReadDeadline 리셋  ← 핵심!                         │
│                                                                 │
│   → 게이트웨이도 "활성 연결"로 인식                                │
│   → idle 타임아웃 발생 안 함                                      │
│                                                                 │
│   ※ Pong 응답이 PingInterval*2 (120초) 동안 없으면               │
│     → connected = false → 재연결 시도                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## 6. 설정 변경 사항

### 6.1 Config 모델 확장

```go
// internal/model/config.go

// 상태 체크 모드 상수
const (
    StatusCheckModeWebSocket = "websocket"
    StatusCheckModeHTTP      = "http"
)

// 기본값
const (
    DefaultWebSocketURL          = "ws://localhost:8080/ws/status"
    DefaultWebSocketReconnect    = 5   // 5초
    DefaultWebSocketPingInterval = 60  // 60초 (게이트웨이 타임아웃 2분 미만으로 설정)
)

type Config struct {
    // 기존 필드 유지...
    TimeoutSeconds      int    `json:"timeoutSeconds"`
    AutoStatusCheck     bool   `json:"autoStatusCheck"`
    StatusCheckInterval int    `json:"statusCheckInterval"`

    // WebSocket 관련 신규 필드
    StatusCheckMode       string `json:"statusCheckMode"`       // "websocket" 또는 "http"
    WebSocketURL          string `json:"webSocketURL"`          // ws://agent:8080/ws/status
    WebSocketReconnect    int    `json:"webSocketReconnect"`    // 재연결 간격 (초)
    WebSocketPingInterval int    `json:"webSocketPingInterval"` // Ping 주기 (초)

    // 기존 API 경로 (HTTP Direct 모드용) 유지...
}
```

---

## 7. UI 통합 설계

### 7.1 리소스 모니터링 UI 구조

```
┌─────────────────────────────────────────────────────────────────────┐
│  장비 모니터링 화면                                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐   │
│  │ 🖥️ CPU      │ │ 💾 메모리    │ │ 💿 디스크    │ │ 🌐 네트워크  │   │
│  │   22.7%    │ │   98.6%    │ │   50.9%    │ │  11.28 MB/s │   │
│  │  CPU 사용률 │ │ 메모리 사용률│ │ 디스크 사용률│ │   네트워크   │   │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘   │
│                                                                     │
│  장비 목록:                                                          │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ IP            │ 상태    │ CPU   │ Memory │ Disk  │ Network   │ │
│  ├───────────────────────────────────────────────────────────────┤ │
│  │ 192.168.1.10  │ 🟢 정상 │ 22.7% │ 98.6%  │ 50.9% │ 11.28MB/s │ │
│  │ 192.168.1.11  │ 🟢 정상 │ 45.2% │ 67.3%  │ 30.1% │ 5.67MB/s  │ │
│  │ 192.168.1.12  │ 🔴 중지 │  -    │   -    │   -   │    -      │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 7.2 device_tab.go 수정

```go
type DeviceTab struct {
    // 기존 필드 유지...
    autoCheckTicker   *time.Ticker
    stopAutoCheck     chan struct{}

    // WebSocket 관련 신규 필드
    wsClient  *websocket.Client
    wsEnabled bool

    // 리소스 모니터링 UI 위젯
    cpuLabel     *widget.Label
    memoryLabel  *widget.Label
    diskLabel    *widget.Label
    networkLabel *widget.Label
}

// 자동 상태 체크 시작 (모드에 따라 분기)
func (d *DeviceTab) startAutoStatusCheck() {
    config, _ := d.store.GetConfig()

    if config.StatusCheckMode == model.StatusCheckModeWebSocket {
        d.startWebSocketStatusCheck()
    } else {
        d.startHTTPPollingStatusCheck()  // 기존 로직
    }
}

// WebSocket 상태 체크 시작
func (d *DeviceTab) startWebSocketStatusCheck() {
    config, _ := d.store.GetConfig()

    wsConfig := &websocket.ClientConfig{
        ServerURL:         config.WebSocketURL,
        ReconnectInterval: time.Duration(config.WebSocketReconnect) * time.Second,
        PingInterval:      time.Duration(config.WebSocketPingInterval) * time.Second,
    }

    d.wsClient = websocket.NewClient(wsConfig)
    d.wsEnabled = true

    // 연결 (실패해도 재연결 루프가 자동으로 시작됨)
    if err := d.wsClient.Connect(); err != nil {
        log.Printf("[ERROR] WebSocket 초기 연결 실패: %v", err)
    }

    // 모든 장비 구독
    ips := d.getAllDeviceIPs()
    d.wsClient.Subscribe(ips)

    // 상태 및 리소스 업데이트 수신 고루틴
    go d.handleWebSocketUpdates()
}

// WebSocket 업데이트 처리 (상태 + 리소스)
func (d *DeviceTab) handleWebSocketUpdates() {
    for {
        select {
        // 연결 상태 업데이트
        case update := <-d.wsClient.StatusUpdates():
            fyne.Do(func() {
                d.updateDeviceStatus(update.DeviceIP, update.Status, update.CheckedAt)
            })

        // 시스템 리소스 업데이트
        case resources := <-d.wsClient.ResourcesUpdates():
            fyne.Do(func() {
                d.updateDeviceResources(resources)
            })

        case err := <-d.wsClient.Errors():
            log.Printf("[WebSocket] 오류: %v", err)

        case <-d.stopAutoCheck:
            return
        }
    }
}

// 장비 리소스 UI 업데이트
func (d *DeviceTab) updateDeviceResources(update *websocket.ResourcesUpdate) {
    if update.Resources == nil {
        return  // 연결 안 된 장비는 리소스 없음
    }

    // 선택된 장비의 리소스를 상단 카드에 표시
    if d.isSelectedDevice(update.DeviceIP) {
        d.cpuLabel.SetText(fmt.Sprintf("%.1f%%", update.Resources.CPU))
        d.memoryLabel.SetText(fmt.Sprintf("%.1f%%", update.Resources.Memory))
        d.diskLabel.SetText(fmt.Sprintf("%.1f%%", update.Resources.Disk))
        d.networkLabel.SetText(fmt.Sprintf("%.2f MB/s", update.Resources.Network))

        // 색상 업데이트 (임계값 초과 시 경고색)
        d.updateResourceColors(update.Resources)
    }

    // 테이블의 해당 장비 행 업데이트
    d.updateDeviceTableRow(update)
}

// 리소스 색상 업데이트 (임계값 기반)
func (d *DeviceTab) updateResourceColors(res *websocket.SystemResources) {
    // CPU: 80% 이상 빨강, 60% 이상 노랑
    // Memory: 90% 이상 빨강, 70% 이상 노랑
    // Disk: 90% 이상 빨강, 70% 이상 노랑
}
```

### 7.3 리소스 카드 위젯 생성

```go
// 리소스 모니터링 카드 생성
func (d *DeviceTab) createResourceCards() *fyne.Container {
    // CPU 카드
    cpuIcon := widget.NewIcon(theme.ComputerIcon())
    d.cpuLabel = widget.NewLabel("-")
    d.cpuLabel.TextStyle = fyne.TextStyle{Bold: true}
    cpuCard := container.NewVBox(
        cpuIcon,
        d.cpuLabel,
        widget.NewLabel("CPU 사용률"),
    )

    // 메모리 카드
    memIcon := widget.NewIcon(theme.StorageIcon())
    d.memoryLabel = widget.NewLabel("-")
    d.memoryLabel.TextStyle = fyne.TextStyle{Bold: true}
    memCard := container.NewVBox(
        memIcon,
        d.memoryLabel,
        widget.NewLabel("메모리 사용률"),
    )

    // 디스크 카드
    diskIcon := widget.NewIcon(theme.FolderIcon())
    d.diskLabel = widget.NewLabel("-")
    d.diskLabel.TextStyle = fyne.TextStyle{Bold: true}
    diskCard := container.NewVBox(
        diskIcon,
        d.diskLabel,
        widget.NewLabel("디스크 사용률"),
    )

    // 네트워크 카드
    netIcon := widget.NewIcon(theme.MailSendIcon())
    d.networkLabel = widget.NewLabel("-")
    d.networkLabel.TextStyle = fyne.TextStyle{Bold: true}
    netCard := container.NewVBox(
        netIcon,
        d.networkLabel,
        widget.NewLabel("네트워크"),
    )

    return container.NewHBox(
        cpuCard, memCard, diskCard, netCard,
    )
}
```

---

## 8. 에러 처리 전략

### 8.1 연결 실패 처리 흐름도

> **참고**: HTTP 폴링 폴백은 사용하지 않습니다. WebSocket 연결 실패 시 주기적으로 토스트 메시지를 표시하고, 사용자가 장비체크를 중지할 때까지 재연결을 시도합니다.
>
> **중요**: 사용자가 자동장비체크를 중지하면 WebSocket 연결과 토스트 알림 모두 즉시 중지됩니다.

```
┌─────────────────────────────────────────────────────────────┐
│                    연결 실패 처리 흐름도                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   장비체크 시작                                              │
│    │                                                         │
│    ▼                                                         │
│   WebSocket 연결 시도                                        │
│    │                                                         │
│    ├─성공──▶ WebSocket으로 상태 수신 (실시간 푸시)           │
│    │         │                                               │
│    │         └─연결 끊김 감지──▶ 재연결 시도                  │
│    │                            │                            │
│    │                            └───────────────┐            │
│    │                                            │            │
│    └─실패──▶ 재연결 시도 (지수 백오프)          │            │
│               │                                 │            │
│               ├─성공──▶ WebSocket 복구 ─────────┘            │
│               │                                              │
│               └─실패──▶ 토스트 메시지 표시                   │
│                         "WebSocket 연결 실패"                │
│                         │                                    │
│                         ▼                                    │
│                  ┌──────────────────┐                        │
│                  │ 재연결 대기      │                        │
│                  │ (지수 백오프)    │                        │
│                  └────────┬─────────┘                        │
│                           │                                  │
│                           ▼                                  │
│                  ┌──────────────────┐                        │
│                  │ 사용자가         │                        │
│                  │ 장비체크 중지?   │                        │
│                  └────────┬─────────┘                        │
│                           │                                  │
│                    ├─Yes──▶ 종료                             │
│                    │                                         │
│                    └─No───▶ 재연결 시도 (반복)               │
│                             + 주기적 토스트 알림             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 8.2 연결 실패 알림 구현

```go
type DeviceTab struct {
    // ...
    wsClient              *websocket.Client
    wsConnected           bool
    lastFailureNotifyTime time.Time
}

// 연결 실패 시 주기적 토스트 알림
// 알림 주기: config.StatusCheckInterval (장비 자동 체크 주기) 사용
func (d *DeviceTab) handleConnectionFailure(err error) {
    now := time.Now()

    config, _ := d.store.GetConfig()
    notifyInterval := time.Duration(config.GetStatusCheckInterval()) * time.Second

    // 마지막 알림으로부터 StatusCheckInterval 경과 시에만 토스트 표시
    if now.Sub(d.lastFailureNotifyTime) >= notifyInterval {
        d.lastFailureNotifyTime = now

        fyne.Do(func() {
            component.ShowErrorToast(
                d.window,
                fmt.Sprintf("WebSocket 연결 실패: %v", err),
            )
        })
    }
}

// 재연결 루프 (지수 백오프)
func (c *Client) reconnectLoop() {
    backoff := c.config.ReconnectInterval  // 초기값: 5초
    maxBackoff := 60 * time.Second

    for {
        select {
        case <-c.ctx.Done():
            // 사용자가 장비체크 중지
            log.Printf("[WebSocket] 장비체크 중지됨")
            return
        default:
            if c.IsConnected() {
                backoff = c.config.ReconnectInterval  // 연결 중이면 백오프 초기화
                time.Sleep(time.Second)
                continue
            }

            log.Printf("[WebSocket] 재연결 시도 (대기: %v)", backoff)

            if err := c.connect(); err != nil {
                // 연결 실패 - 에러 채널로 전파 (토스트 알림용)
                select {
                case c.errorChan <- err:
                default:
                }

                // 지수 백오프: 5s → 10s → 20s → 40s → 60s (최대)
                time.Sleep(backoff)
                backoff = min(backoff*2, maxBackoff)
                continue
            }

            // 연결 성공
            log.Printf("[WebSocket] 연결 성공")
            backoff = c.config.ReconnectInterval

            // 이전 구독 복원
            c.resubscribeAll()
        }
    }
}

// 장비체크 중지 시 WebSocket 정리
func (d *DeviceTab) stopAutoStatusCheck() {
    if d.wsClient != nil {
        d.wsClient.Close()  // context 취소 → reconnectLoop 종료
        d.wsClient = nil
    }
    d.wsConnected = false

    // 기존 타이머 정리 (사용하지 않지만 안전을 위해)
    if d.autoCheckTicker != nil {
        d.autoCheckTicker.Stop()
    }
    if d.stopAutoCheck != nil {
        close(d.stopAutoCheck)
    }
}
```

---

## 9. Agent 서버 측 요구사항

### 9.1 WebSocket 엔드포인트

Agent 서버에서 WebSocket 엔드포인트 구현 필요:

- **URL 형식**: `ws://{host}:{port}/ws/status`
- **포트**: 설정 가능 (Config.WebSocketURL에서 지정)
- **프로토콜**: JSON 메시지

**URL 예시**:
```
ws://agent-server:8080/ws/status    # 기본 예시
ws://192.168.1.100:9000/ws/status   # 다른 포트 사용
ws://localhost:3000/ws/status       # 로컬 테스트
```

### 9.2 서버 측 구현 예시 (참고용)

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type StatusHub struct {
    clients    map[*websocket.Conn]map[string]bool  // conn → subscribed IPs
    mu         sync.RWMutex
}

func (hub *StatusHub) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket 업그레이드 실패: %v", err)
        return
    }
    defer conn.Close()

    hub.mu.Lock()
    hub.clients[conn] = make(map[string]bool)
    hub.mu.Unlock()

    defer func() {
        hub.mu.Lock()
        delete(hub.clients, conn)
        hub.mu.Unlock()
    }()

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }

        var wsMsg WSMessage
        if err := json.Unmarshal(msg, &wsMsg); err != nil {
            continue
        }

        switch wsMsg.Type {
        case "subscribe":
            hub.handleSubscribe(conn, wsMsg.Data)
        case "unsubscribe":
            hub.handleUnsubscribe(conn, wsMsg.Data)
        }
    }
}

// 장비 상태 변경 시 구독자에게 푸시
func (hub *StatusHub) broadcastStatusUpdate(deviceIP, status string) {
    hub.mu.RLock()
    defer hub.mu.RUnlock()

    update := WSMessage{
        Type:      "status_update",
        Timestamp: time.Now().Format(time.RFC3339),
        Data:      json.RawMessage(`{"device_ip":"` + deviceIP + `","status":"` + status + `"}`),
    }

    for conn, subscribed := range hub.clients {
        if subscribed[deviceIP] {
            conn.WriteJSON(update)
        }
    }
}
```

---

## 10. 구현 일정

### Phase 1: 기반 구축
- `internal/websocket/message.go` - 메시지 타입 정의 (SystemResources, DeviceStatus 등)
- `internal/websocket/client.go` - 기본 클라이언트 구조
- `internal/model/config.go` - 설정 필드 추가

### Phase 2: 핵심 기능
- WebSocket 연결/종료 구현
- Ping/Pong 하트비트 구현
- 메시지 송수신 구현 (status_update, resources_update)
- 재연결 로직 구현

### Phase 3: UI 통합
- `device_tab.go` 수정 - WebSocket 이벤트 처리
- 연결 상태 업데이트 UI 반영
- 리소스 모니터링 카드 UI 구현 (CPU, 메모리, 디스크, 네트워크)
- 장비 테이블에 리소스 컬럼 추가
- 연결 실패 시 토스트 알림 구현

### Phase 4: 테스트 및 마무리
- 빌드 테스트
- 통합 테스트
- 문서화

---

## 11. 검증 방법

### 11.1 빌드 테스트

```bash
cd fms_fyne
go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .
```

### 11.2 기능 테스트

| 테스트 항목 | 확인 사항 |
|-------------|-----------|
| WebSocket 연결 | 서버에 정상 연결되는지 |
| 구독 | subscribe 메시지 전송 및 응답 |
| 상태 수신 | status_update 메시지 수신 |
| 리소스 수신 | resources_update 메시지 수신 |
| 재연결 | 연결 끊김 후 자동 재연결 |
| 연결 실패 알림 | 연결 실패 시 토스트 메시지 표시 (StatusCheckInterval 주기) |
| 장비체크 중지 | 중지 시 WebSocket 연결 및 토스트 알림 모두 중지 |

### 11.3 UI 테스트

| 테스트 항목 | 확인 사항 |
|-------------|-----------|
| 장비 상태 표시 | 상태 색상 원 실시간 업데이트 |
| 상태 요약 | green/yellow/red 카운트 갱신 |
| CPU 카드 | CPU 사용률 실시간 표시 |
| 메모리 카드 | 메모리 사용률 실시간 표시 |
| 디스크 카드 | 디스크 사용률 실시간 표시 |
| 네트워크 카드 | 네트워크 트래픽 실시간 표시 |
| 임계값 경고 | 80%/90% 초과 시 색상 변경 |
| 장비 테이블 | 각 장비별 리소스 정보 표시 |
| 연결 끊김 장비 | 리소스 "-" 표시 |
| 토스트 알림 | 연결 실패 시 메시지 표시 |

---

## 12. 참고 자료

- [gorilla/websocket GitHub](https://github.com/gorilla/websocket)
- [gorilla/websocket 문서](https://pkg.go.dev/github.com/gorilla/websocket)
- [WebSocket RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455)
