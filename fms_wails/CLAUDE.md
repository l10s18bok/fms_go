# FMS Wails - Claude 개발 가이드

## 중요 지침

### 언어 및 커뮤니케이션
- **모든 응답, 주석, 문서, 커밋 메시지를 한글로 작성**
- 확실하지 않으면 추론으로 대답하지 말고 코드를 찾아볼 것

### 개발 환경
| 환경 | 플랫폼 | 용도 |
|------|--------|------|
| 개발 환경 | macOS (로컬) | 코드 작성 및 빌드 |
| 테스트 환경 | Linux 서버 (원격, x86_64) | 실제 테스트 및 배포 |

---

## 빠른 참조

### 빌드 명령어
```bash
cd fms_wails

# 개발 모드
wails dev

# 프로덕션 빌드
wails build

# Linux용
wails build -platform linux/amd64
```

### 주요 파일 위치
| 파일 | 설명 |
|------|------|
| `app.go` | Wails 백엔드 API |
| `internal/version/version.go` | 앱 버전 상수 |
| `internal/model/` | 데이터 모델 (Config, Template, Firewall, DeployHistory) |
| `internal/storage/json_store.go` | JSON 파일 저장소 |
| `frontend/src/App.tsx` | 메인 앱 컴포넌트 |
| `frontend/src/components/` | 탭 컴포넌트 (Template, Device, History) |

### 설정 파일 위치
앱 실행 파일과 같은 위치의 `config/` 디렉토리:
- `config.json` - 앱 설정
- `templates.json` - 템플릿 데이터
- `firewalls.json` - 장비 데이터
- `history.json` - 배포 이력

---

## 상세 문서

프로젝트 구조, 데이터 모델, API 목록 등 상세 내용은 [README.md](README.md)를 참조하세요.
