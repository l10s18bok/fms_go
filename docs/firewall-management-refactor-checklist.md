# 방화벽 관리 탭 UI 재구성 체크리스트

## Phase 1: 데이터 레이어

### 1.1 데이터 모델 생성
- [x] `model/firewall_file.go` 생성
  - [x] FirewallFile 구조체 정의
  - [x] NewFirewallFile() 생성자
  - [x] ExtractVersion() 버전 추출 함수
  - [x] GetDisplayCreatedAt() 표시용 생성일
  - [x] GetDisplayModifiedAt() 표시용 수정일
  - [x] Clone() 복사 메서드

### 1.2 파일 저장소 생성
- [x] `storage/file_store.go` 생성
  - [x] FileStore 구조체 정의
  - [x] NewFileStore() 생성자 (data 디렉토리 생성)
  - [x] GetAllFiles() 모든 파일 목록 조회
  - [x] GetFile() 특정 파일 읽기
  - [x] SaveFile() 파일 저장
  - [x] DeleteFile() 파일 삭제
  - [x] DeleteFiles() 여러 파일 삭제
  - [x] FileExists() 파일 존재 확인
  - [x] GetDataDir() data 디렉토리 경로 반환

### 1.3 초기화 코드 수정
- [x] `main.go` 수정
  - [x] data 디렉토리 경로 설정
  - [x] FileStore 초기화
  - [x] MainUI에 FileStore 전달

---

## Phase 2: 초기화면 UI

### 2.1 FirewallTab 구조체
- [x] `ui/firewall_tab.go` 생성
  - [x] FirewallTab 구조체 정의
  - [x] window, fileStore, mainUI 필드
  - [x] searchBox, fileTable 컴포넌트 필드
  - [x] files, filteredFiles 데이터 필드

### 2.2 UI 생성
- [x] NewFirewallTab() 생성자
- [x] createUI() UI 생성 메서드
  - [x] SearchBox 생성 (component.NewSearchBox)
  - [x] 삭제 버튼 (component.NewCustomButton, 빨간색)
  - [x] 파일추가/수정 버튼 (component.NewCustomButton, 파란색)
  - [x] PagedTable 설정 (5컬럼)

### 2.3 테이블 설정
- [x] PagedTable 컬럼 정의
  - [x] 선택 (50px, 체크박스)
  - [x] 파일이름 (200px)
  - [x] 만든 날짜 (180px)
  - [x] 수정한 날짜 (180px)
  - [x] 버전 (100px)
- [x] CellFactory 구현
- [x] OnRowDoubleClick 콜백 설정

### 2.4 데이터 로드
- [x] loadFiles() 파일 목록 로드
- [x] filterFiles() 검색 필터링
- [x] RefreshFiles() 외부 호출용 새로고침

### 2.5 CRUD 기능
- [x] onDeleteSelected() 선택 삭제
- [x] onAddOrEdit() 파일추가/수정 버튼 핸들러

---

## Phase 3: 파일추가/수정 다이얼로그

### 3.1 다이얼로그 UI
- [x] showFileDialog() 메서드 구현
  - [x] 제목: "파일 추가"
  - [x] 선택 드롭다운 (파일생성/파일찾기)
  - [x] 파일명 입력 필드
  - [x] 찾아보기 버튼 (조건부 표시)
  - [x] 취소/완료 버튼

### 3.2 동작 로직
- [x] 파일생성 모드
  - [x] 파일명 입력 활성화
  - [x] 찾아보기 버튼 숨김
  - [x] 완료 시 빈 파일 생성
- [x] 파일찾기 모드
  - [x] 파일명 입력 읽기 전용
  - [x] 찾아보기 버튼 표시
  - [x] 파일 선택 다이얼로그 (.txt 필터)
  - [x] 완료 시 파일 복사

### 3.3 유효성 검사
- [x] 파일명 빈값 검사
- [x] 중복 파일명 검사
- [x] .txt 확장자 자동 추가

---

## Phase 4: 편집 탭 UI

### 4.1 FirewallEditTab 구조체
- [x] `ui/firewall_edit_tab.go` 생성
  - [x] FirewallEditTab 구조체 정의
  - [x] window, fileStore, mainUI, file 필드
  - [x] subTabs, textEditor, ruleBuilder, natBuilder 컴포넌트

### 4.2 서브탭 구현
- [x] 목록보기 탭 (텍스트 편집기)
- [x] Firewall 탭 (RuleBuilder 재사용)
- [x] NAT 탭 (NATBuilder 재사용)
- [x] onSubTabChanged() 탭 전환 시 동기화

### 4.3 버튼
- [x] 추가 버튼 (+ 추가)
  - [x] 룰 빌더/NAT 탭에서만 표시
  - [x] onAddRule() 핸들러
- [x] 저장 버튼
  - [x] onSave() 저장 다이얼로그

### 4.4 저장 다이얼로그
- [x] 파일명 표시 (수정 가능)
- [x] 취소/저장 버튼
- [x] 저장 로직 (파일 쓰기)

### 4.5 탭 이름 처리
- [x] GetTabName() 메서드
- [x] 25자 초과 시 "..." 처리
- [x] 형식: "방화벽관리/{파일명}"

---

## Phase 5: 탭 관리 통합

### 5.1 MainUI 수정
- [x] `ui/app.go` 수정
  - [x] fileStore 필드 추가
  - [x] editTabs map 추가 (동적 탭 관리)
  - [x] FirewallTab 초기화 (TemplateTab 대체)

### 5.2 동적 탭 관리
- [x] OpenFirewallEditTab() 편집 탭 열기
  - [x] 이미 열려있으면 선택
  - [x] 새로 열면 탭 추가
  - [x] editTabs 맵에 등록
- [x] CloseFirewallEditTab() 편집 탭 닫기
  - [x] editTabs 맵에서 제거

### 5.3 탭 닫기 이벤트
- [x] OnClosed 핸들러 수정
  - [x] 편집 탭 닫기 시 editTabs 정리

---

## Phase 6: Import/Export 수정

### 6.1 Import 수정
- [x] 방화벽 관리 탭 Import 제거
  - [x] 방화벽 관리 탭에서 Import 시 안내 메시지 표시

### 6.2 Export 수정
- [x] 방화벽 관리 탭 Export 제거
  - [x] 방화벽 관리 탭에서 Export 시 안내 메시지 표시

### 6.3 Reset 수정
- [x] 방화벽 관리 Reset 시 data 디렉토리 파일 삭제

---

## Phase 7: 기존 코드 정리

### 7.1 삭제할 파일
- [x] `ui/template_tab.go` 삭제

### 7.2 수정할 코드
- [x] `storage/json_store.go` - 템플릿 관련 메서드 유지 (하위 호환성)

### 7.3 참조 업데이트
- [x] DeviceTab에서 TemplateTab 참조 수정
  - [x] templateTab → firewallTab 변경
  - [x] GetTemplateVersions() → GetFileNames() 변경
  - [x] GetTemplate() → GetFileContents() 변경

---

## Phase 8: 테스트

### 8.1 기능 테스트
- [ ] 파일 목록 표시 확인
- [ ] 검색 기능 확인
- [ ] 페이지네이션 확인
- [ ] 파일 추가 (파일생성) 확인
- [ ] 파일 추가 (파일찾기) 확인
- [ ] 파일 수정 확인
- [ ] 파일 삭제 확인
- [ ] 편집 탭 열기/닫기 확인
- [ ] 서브탭 전환 확인
- [ ] 파일 저장 확인

### 8.2 버전 추출 테스트
- [ ] `rules-v1_0_1.txt` → `v1.0.1`
- [ ] `firewall-v2_3.txt` → `v2.3`
- [ ] `test-v1.txt` → `v1`
- [ ] `invalid.txt` → `-`
- [ ] `no-version.txt` → `-`

### 8.3 탭 이름 테스트
- [ ] 짧은 이름: `방화벽관리/test.txt`
- [ ] 긴 이름: `방화벽관리/very-long-...`

### 8.4 빌드 테스트
- [x] `go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .`
- [ ] 실행 확인

---

## 완료 기준

- [x] 모든 Phase 완료 (Phase 1~7)
- [x] 빌드 성공
- [ ] 기능 테스트 통과 (수동 테스트 필요)
- [ ] 기존 기능 (장비 관리, 배포 이력, 패키지 관리) 영향 없음 확인
