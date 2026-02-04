# 리소스 모니터링 기능 PRD

## 1. 개요

### 1.1 목적
장비 관리 테이블에서 실시간 리소스 정보(CPU, 메모리, 디스크, 네트워크)를 표시하고, 상세 정보를 별도 창으로 제공하는 기능을 구현한다.

### 1.2 배경
- DMS(Device Management System)에서 `/device-report` API를 통해 장비 리소스 정보를 수집하는 방식 참조
- 기존 fms_fyne 장비 테이블에 CPU, 메모리, 디스크, 네트워크 컬럼은 존재하나 데이터 미연동 상태

### 1.3 범위
- 장비 테이블 리소스 컬럼 데이터 연동
- 리소스 상세 모니터링 창 구현
- 설정의 `StatusCheckInterval` 주기로 자동 갱신 (기본 60초)

---

## 2. 기능 요구사항

### 2.1 장비 테이블 리소스 표시

#### 2.1.1 표시 컬럼
| 컬럼 | 데이터 소스 | 표시 형식 |
|------|-------------|-----------|
| CPU | `totalUsed.cpu` | `45.2%` |
| 메모리 | `totalUsed.memory` | `68.5%` |
| 디스크 | `totalUsed.disk` | `55.0%` |
| 네트워크 | `totalUsed.rxBps + txBps` | `1.2 MB/s` |

#### 2.1.2 데이터 수집
- 상태 체크 시 `/device-report` API 응답에서 `totalUsed` 파싱
- `Firewall` 모델의 `CPUUsage`, `MemoryUsage`, `DiskUsage`, `NetworkUsage` 필드에 저장

### 2.2 리소스 상세 모니터링 창

#### 2.2.1 창 열기 조건
- 장비 테이블의 CPU, 메모리, 디스크, 네트워크 컬럼 클릭 시 해당 장비의 리소스 상세 창 열림
- **하나의 장비당 하나의 창만 허용** (이미 열려있으면 해당 창 포커스)

#### 2.2.2 창 구성 (B 방식 - 전체 표시)
어떤 컬럼을 클릭해도 모든 상세 테이블을 한번에 표시:

```
┌─────────────────────────────────────────────────────────┐
│  리소스 모니터링 - 장비명 (IP)                           │
│  마지막 갱신: HH:MM:SS                       [갱신] [X] │
├─────────────────────────────────────────────────────────┤
│  ▼ CPU                                                  │
│  ┌────────┬────────┬────────┬────────┬────────┐        │
│  │ User   │ Kernel │ Idle   │ IOWait │ Avg1   │ ...    │
│  └────────┴────────┴────────┴────────┴────────┘        │
│                                                         │
│  ▼ 메모리                                               │
│  ┌────────┬────────┬────────┬────────┬────────┐        │
│  │ Total  │ Used   │ Free   │ Avail  │ Swap   │ ...    │
│  └────────┴────────┴────────┴────────┴────────┘        │
│                                                         │
│  ▼ 디스크                                               │
│  ┌────────────┬────────┬────────┬────────┬──────┐      │
│  │ 파티션     │ 마운트 │ 전체   │ 사용   │ 사용률│      │
│  └────────────┴────────┴────────┴────────┴──────┘      │
│                                                         │
│  ▼ 네트워크                                             │
│  ┌──────────┬─────────────┬─────────┬─────────┐        │
│  │ 인터페이스│ IP          │ RX      │ TX      │        │
│  └──────────┴─────────────┴─────────┴─────────┘        │
│                                                         │
│  ▼ 프로세스                                             │
│  ┌────────┬─────────┬───────┬───────┬──────┬──────┐    │
│  │ 이름   │ 버전    │ CPU%  │ MEM%  │ 상태 │ 시간 │    │
│  └────────┴─────────┴───────┴───────┴──────┴──────┘    │
└─────────────────────────────────────────────────────────┘
```

#### 2.2.3 테이블 상세

**CPU 테이블**
| 컬럼 | 데이터 소스 |
|------|-------------|
| User | `cpuall.user` |
| Kernel | `cpuall.kernel` |
| Idle | `cpuall.idle` |
| IOWait | `cpuall.iowait` |
| Nice | `cpuall.ni` |
| HW IRQ | `cpuall.hwIrq` |
| SW IRQ | `cpuall.swIrq` |
| Avg 1분 | `cpuall.avg1` |
| Avg 5분 | `cpuall.avg5` |
| Avg 15분 | `cpuall.avg15` |

**메모리 테이블**
| 컬럼 | 데이터 소스 |
|------|-------------|
| Total | `memory.totalDisplay` |
| Used | `memory.usedDisplay` |
| Free | `memory.freeDisplay` |
| Available | `memory.availableDisplay` |
| Buff/Cache | `memory.buffCacheDisplay` |
| Swap Total | `memory.swapTotalDisplay` |
| Swap Used | `memory.swapUsedDisplay` |
| Swap Free | `memory.swapFreeDisplay` |

**디스크 테이블 (배열)**
| 컬럼 | 데이터 소스 |
|------|-------------|
| 파티션 | `disk[].filesystem` |
| 마운트 | `disk[].mountedOn` |
| 전체 | `disk[].sizeDisplay` |
| 사용 | `disk[].usedDisplay` |
| 여유 | `disk[].freeDisplay` |
| 사용률 | `disk[].usePercent` |

**네트워크 테이블 (배열)**
| 컬럼 | 데이터 소스 |
|------|-------------|
| 인터페이스 | `interface[].ifname` |
| IP | `interface[].ip4` |
| RX 속도 | `interface[].rxBpsDisplay` |
| TX 속도 | `interface[].txBpsDisplay` |
| RX 누적 | `interface[].rxBytesDisplay` |
| TX 누적 | `interface[].txBytesDisplay` |
| 에러 | `rxErrors + txErrors + rxDropped + txDropped` |

**프로세스 테이블 (배열)**
| 컬럼 | 데이터 소스 |
|------|-------------|
| 이름 | `process[].name` |
| 버전 | `process[].version` |
| CPU % | `process[].cpuPercent` |
| MEM % | `process[].memPercent` |
| 메모리 | `process[].rssDisplay` |
| 상태 | `process[].status` (한글 변환) |
| 실행시간 | `process[].runTime` |

> **참고**: 프로세스는 `name`과 `version`이 모두 있는 항목만 표시 (DMS 방식)

### 2.3 데이터 저장 방식

#### 2.3.1 A 방식 - 저장 안함 (채택)
- 상태 체크 시: `totalUsed` → Firewall 모델의 4개 필드에 저장 (테이블 표시용)
- 상세 데이터(cpuall, memory, disk, interface, process): DB 저장 안함
- 창 열 때: `/device-report` API 호출하여 상세 데이터 조회 (메모리에만 유지)
- 창 갱신 시: 다시 API 호출

#### 2.3.2 선택 이유
- fms_fyne은 **방화벽 관리**가 주 목적 (리소스 모니터링은 부가 기능)
- 이력 관리/TOP 5 등은 DMS 영역
- 구현 복잡도 낮음, DB 용량 부담 없음

---

### 2.4 자동 갱신

#### 2.4.1 갱신 주기
- 설정의 `StatusCheckInterval` 값 사용 (기본 60초, 최소 10초, 최대 300초)
- 창이 열려있는 동안 주기적으로 `/device-report` API 호출

#### 2.4.2 갱신 표시
- 창 상단에 "마지막 갱신: HH:MM:SS" 표시
- [갱신] 버튼으로 수동 갱신 가능

#### 2.4.3 연결 실패 처리
- 장비 연결 실패 시 "연결 안됨" 표시
- 기존 데이터 유지 + 갱신 시간 업데이트 안함

---

### 2.5 UI/UX 개선사항 (구현 완료)

#### 2.5.1 동적 컬럼 너비
- 창 크기 변경 시 테이블 컬럼 너비가 자동으로 조절됨
- 각 테이블별 컬럼 비율 유지

#### 2.5.2 Sticky 헤더
- 테이블 스크롤 시 헤더 행이 상단에 고정되어 항상 표시
- Fyne 2.4+ `StickyRowCount = 1` 사용

#### 2.5.3 셀 선택 비활성화
- 테이블 셀 클릭 시 선택 하이라이트 표시 안함
- `OnSelected` 콜백에서 즉시 `UnselectAll()` 호출

#### 2.5.4 창 수명주기 관리
- 메인 창 닫기 시 열려있는 모든 리소스 창 자동 닫기
- `main.go`에서 `SetOnClosed` → `MainUI.Cleanup()` 호출

---

## 3. API 연동

### 3.1 Device Report API

**요청**
```
GET {scheme}://{ip}:{port}{path}
예: http://192.168.1.100:8080/device-report
```

**응답 JSON 구조**
```json
{
  "totalUsed": {
    "cpu": 45.2,
    "memory": 68.5,
    "disk": 55.0,
    "rxBps": 1258291,
    "txBps": 839527,
    "rxBpsDisplay": "1.2 MB/s",
    "txBpsDisplay": "0.8 MB/s"
  },
  "cpuall": {
    "user": 12.3,
    "kernel": 5.2,
    "idle": 80.5,
    "iowait": 1.5,
    "ni": 0,
    "hwIrq": 0.2,
    "swIrq": 0.3,
    "avg1": 0.5,
    "avg5": 0.8,
    "avg15": 1.2
  },
  "memory": {
    "total": 17179869184,
    "used": 11701322547,
    "free": 5478546637,
    "totalDisplay": "16.0 GB",
    "usedDisplay": "10.9 GB",
    "freeDisplay": "5.1 GB",
    "usedPercent": 68.1,
    "swapTotal": 0,
    "swapUsed": 0
  },
  "disk": [
    {
      "filesystem": "/dev/sda1",
      "mountedOn": "/",
      "size": 107374182400,
      "used": 59055800320,
      "sizeDisplay": "100 GB",
      "usedDisplay": "55 GB",
      "freeDisplay": "45 GB",
      "usePercent": 55.0
    }
  ],
  "interface": [
    {
      "ifname": "eth0",
      "ip4": "192.168.1.100",
      "rxBps": 1258291,
      "txBps": 839527,
      "rxBpsDisplay": "1.2 MB/s",
      "txBpsDisplay": "0.8 MB/s"
    }
  ],
  "process": [
    {
      "name": "nginx",
      "version": "1.18.0",
      "cpuPercent": 1.2,
      "memPercent": 0.5,
      "rssDisplay": "50 MB",
      "status": ["S"],
      "runTime": "10:23:45"
    }
  ]
}
```

---

## 4. 구현 계획

### 4.1 모델 추가 (model/)

1. `DeviceReport` 구조체 추가 (API 응답 파싱용)
2. `ReportTotalUsed`, `ReportCPUAll`, `ReportMemory`, `ReportDisk`, `ReportInterface`, `ReportProcess` 구조체 추가
3. `ProcessStatusMap` - 프로세스 상태 코드 한글 매핑

### 4.2 HTTP 클라이언트 수정 (http/client.go)

1. `CheckHealthDirect` 함수에서 응답 JSON 파싱
2. `totalUsed` 데이터를 `Firewall` 모델에 저장
3. 상세 데이터 반환을 위한 새 함수 추가: `GetDeviceReport(fw *Firewall) (*DeviceReport, error)`

### 4.3 UI 수정 (ui/device_tab.go)

1. 리소스 컬럼 클릭 이벤트 추가
2. 리소스 모니터링 창 관리 (`ResourceWindowManager`)
3. 창 열기/포커스 로직

### 4.4 리소스 모니터링 창 (ui/resource_window.go) - 신규

1. 새 창 생성 (`fyne.App.NewWindow`)
2. 5개 테이블 구성 (CPU, 메모리, 디스크, 네트워크, 프로세스)
3. `StatusCheckInterval` 주기 자동 갱신 (ticker)
4. 창 닫기 시 리소스 정리
5. 동적 컬럼 너비 조절 (`resourceResizableContainer`)
6. Sticky 헤더 및 셀 선택 비활성화

### 4.5 메인 앱 수정 (main.go, app.go)

1. 메인 창 닫기 시 `Cleanup()` 호출
2. 모든 리소스 창 자동 닫기

---

## 5. 파일 구조

```
fms_fyne/
├── main.go                     # 수정: SetOnClosed에서 Cleanup 호출
├── internal/
│   ├── model/
│   │   ├── firewall.go         # 기존 (수정 없음)
│   │   └── device_report.go    # 신규: API 응답 모델
│   ├── http/
│   │   └── client.go           # 수정: GetDeviceReport 추가
│   └── ui/
│       ├── app.go              # 수정: Cleanup 메서드 추가
│       ├── device_tab.go       # 수정: 컬럼 클릭 이벤트, ResourceWindowManager
│       ├── component/table.go  # 수정: StickyRowCount 추가
│       └── resource_window.go  # 신규: 리소스 모니터링 창
```

---

## 6. 체크리스트

### 6.1 모델
- [x] `DeviceReport` 구조체 정의
- [x] `ReportTotalUsed` 구조체 정의
- [x] `ReportCPUAll` 구조체 정의
- [x] `ReportMemory` 구조체 정의
- [x] `ReportDisk` 구조체 정의
- [x] `ReportInterface` 구조체 정의
- [x] `ReportProcess` 구조체 정의
- [x] `ProcessStatusMap` 프로세스 상태 한글 매핑

### 6.2 HTTP 클라이언트
- [x] `GetDeviceReport` 함수 추가 (상세 데이터 반환)
- [ ] `CheckHealthDirect`에서 `totalUsed` 파싱 및 Firewall 저장

### 6.3 장비 테이블
- [x] CPU/메모리/디스크/네트워크 컬럼 클릭 이벤트 추가
- [x] `ResourceWindowManager` 구현 (장비별 창 맵)
- [x] 중복 창 방지 및 포커스 처리

### 6.4 리소스 모니터링 창
- [x] 창 생성 및 레이아웃
- [x] CPU 테이블
- [x] 메모리 테이블
- [x] 디스크 테이블
- [x] 네트워크 테이블
- [x] 프로세스 테이블
- [x] `StatusCheckInterval` 주기 자동 갱신 (ticker)
- [x] 수동 갱신 버튼
- [x] 마지막 갱신 시간 표시
- [x] 창 닫기 시 ticker 정리

### 6.5 UI/UX 개선
- [x] 동적 컬럼 너비 조절 (창 크기에 맞게)
- [x] Sticky 헤더 (스크롤 시 헤더 고정)
- [x] 셀 선택 비활성화 (UnselectAll)
- [x] 메인 창 닫기 시 모든 리소스 창 자동 닫기

### 6.6 커스텀 테이블 (PagedTable)
- [x] Sticky 헤더 적용 (`StickyRowCount = 1`)

---

## 7. 참고

### 7.1 DMS 참조 파일
- `C:\Users\igsgcns\Downloads\DMS\app.go` - `CallDeviceReport`, `SaveDeviceReportRecord`
- `C:\Users\igsgcns\Downloads\DMS\models\models.go` - `DeviceReport`, `ReportTotalUsed` 등
- `C:\Users\igsgcns\Downloads\DMS\frontend\js\monitoring.js` - UI 표시 로직

### 7.2 fms_fyne 기존 파일
- `internal/ui/device_tab.go` - 장비 테이블 (컬럼 7~10이 리소스)
- `internal/http/client.go` - `CheckHealthDirect` (현재 상태만 확인)
- `internal/model/firewall.go` - `CPUUsage`, `MemoryUsage`, `DiskUsage`, `NetworkUsage` 필드 존재
