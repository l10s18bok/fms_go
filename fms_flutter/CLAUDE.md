# FMS Flutter - 개발 가이드

## 프로젝트 개요

Flutter로 구현한 FMS(Firewall Management System) Windows 데스크톱 애플리케이션입니다.
fms_wails (Go Wails + React)의 UI/UX를 완전 모방하여 동일한 기능과 디자인을 제공합니다.

---

## 중요 지침

### 기존 코드 참조

- **`fms_wails/` 폴더**: Wails 버전 (원본 UI/UX 참조)
- **루트 디렉토리 `internal/`**: Fyne 버전 백엔드 로직 참조

### 언어 및 커뮤니케이션

- **모든 응답, 주석, 문서, 커밋 메시지를 한글로 작성**
- 질문에 대답할 때 확실하지 않으면 추론으로 대답하지 말 것

---

## 프로젝트 구조

```
fms_flutter/
├── CLAUDE.md                    # 개발 가이드 (현재 파일)
├── pubspec.yaml                 # Flutter 프로젝트 설정
├── lib/
│   ├── main.dart               # 앱 진입점 및 메인 화면
│   │
│   ├── models/                 # 데이터 모델
│   │   ├── models.dart         # 모델 export
│   │   ├── template.dart       # 템플릿 모델
│   │   ├── firewall.dart       # 장비 모델
│   │   ├── deploy_result.dart  # 배포 결과 모델
│   │   ├── history.dart        # 배포 이력 모델
│   │   └── config.dart         # 앱 설정 모델
│   │
│   ├── services/               # 비즈니스 로직
│   │   ├── services.dart       # 서비스 export
│   │   ├── storage_service.dart # JSON 파일 저장소
│   │   └── deploy_service.dart  # HTTP 배포 서비스
│   │
│   ├── screens/                # 화면 (탭)
│   │   ├── screens.dart        # 화면 export
│   │   ├── template_tab.dart   # 템플릿 관리 탭
│   │   ├── device_tab.dart     # 장비 관리 탭
│   │   └── history_tab.dart    # 배포 이력 탭
│   │
│   ├── widgets/                # 공통 위젯
│   │   ├── widgets.dart        # 위젯 export
│   │   ├── fms_card.dart       # 카드 위젯
│   │   ├── fms_button.dart     # 버튼 위젯
│   │   ├── fms_table.dart      # 테이블 위젯
│   │   ├── fms_dialog.dart     # 다이얼로그 위젯
│   │   ├── status_badge.dart   # 상태 뱃지 위젯
│   │   └── empty_state.dart    # 빈 상태 위젯
│   │
│   └── theme/                  # 테마
│       └── app_theme.dart      # 다크 테마 (Wails CSS 매핑)
│
├── build/                      # 빌드 출력
│   └── windows/x64/runner/Release/
│       └── fms_flutter.exe     # Windows 실행 파일
│
└── windows/                    # Windows 플랫폼 설정
```

---

## 주요 기능

### 1. 템플릿 관리
- 템플릿 목록 (클릭 선택)
- 템플릿 조회/저장/삭제
- 버전 및 규칙 내용 편집

### 2. 장비(방화벽) 관리
- 장비 목록 테이블 (체크박스, 장비명, 서버상태, 배포상태, 버전)
- 장비 추가/편집/삭제
- 서버 상태 확인 (선택된 장비)
- 템플릿 배포

### 3. 배포 이력
- 배포 이력 목록 (최신순)
- 상세 정보 (규칙별 성공/실패)
- 이력 삭제/전체 삭제

### 4. 설정
- Connection Mode (Agent/Direct)
- Timeout 설정
- 설정 저장 경로 표시

### 5. Import/Export
- 현재 탭 데이터 JSON 내보내기/가져오기
- 전체 초기화

### 6. 차트 데모
- 하단 버전 클릭 시 월별 통계 차트 표시

---

## 테마 색상 (Wails CSS 매핑)

```dart
// 주요 색상
primaryColor: #E94560    // 강조색 (핑크)
backgroundColor: #1A1A2E // 배경
surfaceColor: #16213E    // 카드/표면
borderColor: #0F3460     // 테두리

// 상태 색상
successColor: #27AE60    // 성공 (녹색)
dangerColor: #E74C3C     // 위험 (빨강)
warningColor: #F1C40F    // 경고 (노랑)
infoColor: #3498DB       // 정보 (파랑)

// 텍스트 색상
textPrimary: #EEEEEE     // 기본 텍스트
textSecondary: #AAAAAA   // 보조 텍스트
textMuted: #666666       // 비활성 텍스트
```

---

## 데이터 저장 경로

Flutter 앱은 `Documents/config_flutter/` 폴더에 데이터를 저장합니다:
- `templates.json` - 템플릿 데이터
- `firewalls.json` - 장비 데이터
- `history.json` - 배포 이력
- `config.json` - 앱 설정

---

## 빌드 명령어

### 개발 실행
```bash
cd fms_flutter
flutter run -d windows
```

### 릴리즈 빌드
```bash
cd fms_flutter
flutter build windows --release
```

빌드 결과: `build/windows/x64/runner/Release/fms_flutter.exe`

---

## 의존성 패키지

```yaml
dependencies:
  http: ^1.2.0          # HTTP 통신
  path_provider: ^2.1.2  # 파일 경로
  path: ^1.9.0          # 경로 유틸리티
  file_picker: ^8.0.0   # 파일 선택 다이얼로그
  fl_chart: ^0.69.0     # 차트 라이브러리
  intl: ^0.20.0         # 날짜 포맷팅
```

---

## Wails 버전과의 비교

| 기능 | Wails (React) | Flutter |
|------|---------------|---------|
| 메뉴바 | CSS dropdown | PopupMenuButton |
| 탭 | CSS tab-btn | TabController |
| 테이블 | HTML table | Table widget |
| 모달 | CSS modal | showDialog |
| 뱃지 | CSS badge | StatusBadge widget |
| 차트 | recharts | fl_chart |
| 파일 다이얼로그 | Wails runtime | file_picker |

---

## 참조 문서

- [fms_wails/frontend/src/](../fms_wails/frontend/src/) - Wails React 소스
- [fms_wails/frontend/src/App.css](../fms_wails/frontend/src/App.css) - 원본 CSS
- [FMS_SPEC.md](../FMS_SPEC.md) - FMS 기능 명세서
