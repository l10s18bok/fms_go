### Windows 용 앱 개발 프레임워크 비교


### 핵심 비교표

> **측정 기준**: 동일한 FMS 앱을 각 프레임워크로 구현하여 측정 (fms_wails, fms_fyne, fms_flutter)
> - 앱 크기: Release 빌드 후 실행에 필요한 전체 파일 크기
> - 메모리: 앱 실행 후 5초 대기 시 메모리 측정

| 항목 | Flutter | Wails | Fyne |
|------|---------|-------|------|
| **차트 품질** | 기본 | 최고 (Grafana급) | 없음 |
| **모바일 지원** | iOS/Android | 미지원 | 실험적 |
| **앱 크기** | 26.6MB (EXE+DLL+data 포함) | 10.5MB (단일 EXE) | 23.7MB (단일 EXE) |
| **메모리 점유율** | 71MB | 28MB | 127MB |
| **Agent 통신** | HTTP, WebSocket | HTTP, WebSocket | HTTP, WebSocket |

---

### 1. 차트 비교

### Flutter - fl_chart
- **라이브러리**: [fl_chart](https://pub.dev/packages/fl_chart)
- **타입**: Line, Bar, Pie, Scatter (기본)
- **인터랙션**: 제한적 (줌/드래그 부족)
- **실시간 차트반영**: 성능 저하
- **평가**: 기본적 수준

### Wails - recharts
- **라이브러리**: [recharts](https://recharts.org/) (React 차트 라이브러리)
- **타입**: Line, Area, Bar, Pie, Radar 등
- **인터랙션**: 우수 (툴팁, 범례, 반응형)
- **실시간 차트반영**: 좋음
- **평가**: 웹 수준의 고품질 차트

### Fyne
- **라이브러리**: 없음 (직접 구현 필요)
- **평가**: 프로덕션 부적합

---

### 2. 모바일 지원

### Flutter
- **iOS/Android**: 완벽 지원
- **코드 재사용**: 높음

### Wails
- **모바일**: 미지원 (데스크톱 전용)

### Fyne
- **모바일**: 실험적 (프로덕션 비추천)
- **문제**: 성능 이슈, UI 어색함

---

### 3. 렌더링 엔진 비교

| 프레임워크 | 렌더링 방식 | 배포 형태 |
|------------|-------------|-----------|
| **Flutter** | Skia 엔진 (DLL 분리) | EXE + DLL + data 폴더 |
| **Wails** | OS WebView2 사용 | 단일 EXE |
| **Fyne** | OpenGL 자체 렌더링 | 단일 EXE |

- **Wails**: WebView2는 Windows에 기본 설치되어 있어 앱 크기가 작음
- **Fyne**: OpenGL 렌더링 엔진을 바이너리에 포함하여 메모리 사용량이 높음
- **Flutter**: Skia 엔진이 DLL로 분리되어 다중 파일 배포 필요


