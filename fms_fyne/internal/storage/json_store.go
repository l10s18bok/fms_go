package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fms/internal/model"
)

// JSON 파일 기반 저장소입니다.
type JSONStore struct {
	configDir string
	mu        sync.RWMutex

	// 캐시된 데이터
	ruleFiles map[string]*model.RuleFile
	firewalls map[int]*model.Firewall
	history   map[int]*model.DeployHistory
	programs  map[int]*model.ProcessInfo // 신규: 패키지 정보

	// Auto increment 카운터
	nextFirewallID int
	nextHistoryID  int
	nextProgramID  int // 신규
}

// 파일명 상수
const (
	ruleFilesFile = "rule_files.json"
	firewallsFile = "firewalls.json"
	historyFile   = "history.json"
	configFile    = "config.json"
	programsFile  = "programs.json" // 신규
)

// 새로운 JSON 저장소를 생성.
func NewJSONStore(configDir string) (*JSONStore, error) {
	store := &JSONStore{
		configDir: configDir,
		ruleFiles: make(map[string]*model.RuleFile),
		firewalls: make(map[int]*model.Firewall),
		history:   make(map[int]*model.DeployHistory),
		programs:  make(map[int]*model.ProcessInfo), // 신규
	}

	// 설정 디렉토리 생성
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("설정 디렉토리 생성 실패: %v", err)
	}

	// 기존 데이터 로드
	if err := store.loadAll(); err != nil {
		return nil, fmt.Errorf("데이터 로드 실패: %v", err)
	}

	return store, nil
}

// 모든 데이터 파일을 로드합니다.
func (s *JSONStore) loadAll() error {
	if err := s.loadRuleFiles(); err != nil {
		return err
	}
	if err := s.loadFirewalls(); err != nil {
		return err
	}
	if err := s.loadHistory(); err != nil {
		return err
	}
	if err := s.loadPrograms(); err != nil {
		return err
	}
	return nil
}

// 규칙 파일 데이터를 로드합니다.
func (s *JSONStore) loadRuleFiles() error {
	path := filepath.Join(s.configDir, ruleFilesFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // 파일이 없으면 빈 상태로 시작
	}
	if err != nil {
		return err
	}

	var ruleFiles []*model.RuleFile
	if err := json.Unmarshal(data, &ruleFiles); err != nil {
		return err
	}

	for _, r := range ruleFiles {
		s.ruleFiles[r.Version] = r
	}
	return nil
}

// 장비 데이터를 로드합니다.
func (s *JSONStore) loadFirewalls() error {
	path := filepath.Join(s.configDir, firewallsFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var firewalls []*model.Firewall
	if err := json.Unmarshal(data, &firewalls); err != nil {
		return err
	}

	maxID := 0
	for _, f := range firewalls {
		s.firewalls[f.Index] = f
		if f.Index > maxID {
			maxID = f.Index
		}
	}
	s.nextFirewallID = maxID + 1
	return nil
}

// 배포 이력 데이터를 로드합니다.
func (s *JSONStore) loadHistory() error {
	path := filepath.Join(s.configDir, historyFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var history []*model.DeployHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return err
	}

	maxID := 0
	for _, h := range history {
		s.history[h.ID] = h
		if h.ID > maxID {
			maxID = h.ID
		}
	}
	s.nextHistoryID = maxID + 1
	return nil
}

// 패키지 데이터를 로드합니다.
func (s *JSONStore) loadPrograms() error {
	path := filepath.Join(s.configDir, programsFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // 파일이 없으면 빈 상태로 시작
	}
	if err != nil {
		return err
	}

	var programs []*model.ProcessInfo
	if err := json.Unmarshal(data, &programs); err != nil {
		return err
	}

	maxID := 0
	for _, p := range programs {
		s.programs[p.ID] = p
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	s.nextProgramID = maxID + 1
	return nil
}

// 규칙 파일 데이터를 저장합니다.
func (s *JSONStore) saveRuleFiles() error {
	ruleFiles := make([]*model.RuleFile, 0, len(s.ruleFiles))
	for _, r := range s.ruleFiles {
		ruleFiles = append(ruleFiles, r)
	}

	data, err := json.MarshalIndent(ruleFiles, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, ruleFilesFile)
	return os.WriteFile(path, data, 0644)
}

// 장비 데이터를 저장합니다.
func (s *JSONStore) saveFirewalls() error {
	firewalls := make([]*model.Firewall, 0, len(s.firewalls))
	for _, f := range s.firewalls {
		firewalls = append(firewalls, f)
	}

	data, err := json.MarshalIndent(firewalls, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, firewallsFile)
	return os.WriteFile(path, data, 0644)
}

// 배포 이력 데이터를 저장합니다.
func (s *JSONStore) saveHistory() error {
	history := make([]*model.DeployHistory, 0, len(s.history))
	for _, h := range s.history {
		history = append(history, h)
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, historyFile)
	return os.WriteFile(path, data, 0644)
}

// 패키지 데이터를 저장합니다.
func (s *JSONStore) savePrograms() error {
	programs := make([]*model.ProcessInfo, 0, len(s.programs))
	for _, p := range s.programs {
		programs = append(programs, p)
	}

	data, err := json.MarshalIndent(programs, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, programsFile)
	return os.WriteFile(path, data, 0644)
}

// ===== RuleFile 메서드 =====

// 모든 규칙 파일을 반환합니다.
func (s *JSONStore) GetAllRuleFiles() ([]*model.RuleFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ruleFiles := make([]*model.RuleFile, 0, len(s.ruleFiles))
	for _, r := range s.ruleFiles {
		ruleFiles = append(ruleFiles, r.Clone())
	}
	return ruleFiles, nil
}

// 특정 버전의 규칙 파일을 반환합니다.
func (s *JSONStore) GetRuleFile(version string) (*model.RuleFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.ruleFiles[version]
	if !ok {
		return nil, fmt.Errorf("규칙 파일을 찾을 수 없습니다: %s", version)
	}
	return r.Clone(), nil
}

// 규칙 파일을 저장합니다.
func (s *JSONStore) SaveRuleFile(ruleFile *model.RuleFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ruleFiles[ruleFile.Version] = ruleFile.Clone()
	return s.saveRuleFiles()
}

// 규칙 파일을 삭제합니다.
func (s *JSONStore) DeleteRuleFile(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.ruleFiles[version]; !ok {
		return fmt.Errorf("규칙 파일을 찾을 수 없습니다: %s", version)
	}

	delete(s.ruleFiles, version)
	return s.saveRuleFiles()
}

// 모든 규칙 파일을 삭제합니다.
func (s *JSONStore) ClearRuleFiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ruleFiles = make(map[string]*model.RuleFile)
	return s.saveRuleFiles()
}

// ===== Firewall 메서드 =====

// 모든 장비를 반환합니다.
func (s *JSONStore) GetAllFirewalls() ([]*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	firewalls := make([]*model.Firewall, 0, len(s.firewalls))
	for _, f := range s.firewalls {
		firewalls = append(firewalls, f.Clone())
	}
	return firewalls, nil
}

// 특정 인덱스의 장비를 반환합니다.
func (s *JSONStore) GetFirewall(index int) (*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.firewalls[index]
	if !ok {
		return nil, fmt.Errorf("장비를 찾을 수 없습니다: %d", index)
	}
	return f.Clone(), nil
}

// 장비를 저장합니다.
func (s *JSONStore) SaveFirewall(firewall *model.Firewall) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 기존 장비인지 확인 (Index로 판단)
	_, exists := s.firewalls[firewall.Index]

	// 새 장비인 경우에만 ID 할당 (Index가 -1인 경우)
	if firewall.Index < 0 {
		firewall.Index = s.nextFirewallID
		s.nextFirewallID++
	}

	// 기존 장비가 아니고 Index >= 0인 경우, nextFirewallID 업데이트
	if !exists && firewall.Index >= s.nextFirewallID {
		s.nextFirewallID = firewall.Index + 1
	}

	s.firewalls[firewall.Index] = firewall.Clone()
	return s.saveFirewalls()
}

// 장비를 삭제합니다.
func (s *JSONStore) DeleteFirewall(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.firewalls[index]; !ok {
		return fmt.Errorf("장비를 찾을 수 없습니다: %d", index)
	}

	delete(s.firewalls, index)
	return s.saveFirewalls()
}

// 모든 장비를 삭제합니다.
func (s *JSONStore) ClearFirewalls() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.firewalls = make(map[int]*model.Firewall)
	s.nextFirewallID = 1
	return s.saveFirewalls()
}

// ===== DeployHistory 메서드 =====

// 모든 배포 이력을 반환합니다.
func (s *JSONStore) GetAllHistory() ([]*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]*model.DeployHistory, 0, len(s.history))
	for _, h := range s.history {
		// 복사본 생성
		hCopy := *h
		hCopy.Results = make([]model.RuleResult, len(h.Results))
		copy(hCopy.Results, h.Results)
		history = append(history, &hCopy)
	}
	return history, nil
}

// 특정 ID의 배포 이력을 반환합니다.
func (s *JSONStore) GetHistory(id int) (*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	h, ok := s.history[id]
	if !ok {
		return nil, fmt.Errorf("배포 이력을 찾을 수 없습니다: %d", id)
	}

	// 복사본 생성
	hCopy := *h
	hCopy.Results = make([]model.RuleResult, len(h.Results))
	copy(hCopy.Results, h.Results)
	return &hCopy, nil
}

// 배포 이력을 저장합니다.
func (s *JSONStore) SaveHistory(history *model.DeployHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 새 이력인 경우 ID 할당
	if history.ID == 0 {
		history.ID = s.nextHistoryID
		s.nextHistoryID++
	}

	// 복사본 저장
	hCopy := *history
	hCopy.Results = make([]model.RuleResult, len(history.Results))
	copy(hCopy.Results, history.Results)
	s.history[history.ID] = &hCopy

	return s.saveHistory()
}

// 배포 이력을 삭제합니다.
func (s *JSONStore) DeleteHistory(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.history[id]; !ok {
		return fmt.Errorf("배포 이력을 찾을 수 없습니다: %d", id)
	}

	delete(s.history, id)
	return s.saveHistory()
}

// 모든 배포 이력을 삭제합니다.
func (s *JSONStore) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = make(map[int]*model.DeployHistory)
	s.nextHistoryID = 1
	return s.saveHistory()
}

// ===== Export/Import 메서드 =====

// 모든 데이터를 반환합니다.
func (s *JSONStore) ExportAll() (*ExportData, error) {
	ruleFiles, err := s.GetAllRuleFiles()
	if err != nil {
		return nil, err
	}

	firewalls, err := s.GetAllFirewalls()
	if err != nil {
		return nil, err
	}

	history, err := s.GetAllHistory()
	if err != nil {
		return nil, err
	}

	return &ExportData{
		RuleFiles: ruleFiles,
		Firewalls: firewalls,
		History:   history,
	}, nil
}

// 데이터를 가져옵니다.
func (s *JSONStore) ImportAll(data *ExportData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 규칙 파일 가져오기
	for _, r := range data.RuleFiles {
		s.ruleFiles[r.Version] = r.Clone()
	}

	// 장비 가져오기
	maxFirewallID := s.nextFirewallID
	for _, f := range data.Firewalls {
		if f.Index == 0 {
			f.Index = maxFirewallID
			maxFirewallID++
		}
		s.firewalls[f.Index] = f.Clone()
		if f.Index >= s.nextFirewallID {
			s.nextFirewallID = f.Index + 1
		}
	}

	// 이력 가져오기
	maxHistoryID := s.nextHistoryID
	for _, h := range data.History {
		if h.ID == 0 {
			h.ID = maxHistoryID
			maxHistoryID++
		}
		hCopy := *h
		hCopy.Results = make([]model.RuleResult, len(h.Results))
		copy(hCopy.Results, h.Results)
		s.history[h.ID] = &hCopy
		if h.ID >= s.nextHistoryID {
			s.nextHistoryID = h.ID + 1
		}
	}

	// 모든 데이터 저장
	if err := s.saveRuleFiles(); err != nil {
		return err
	}
	if err := s.saveFirewalls(); err != nil {
		return err
	}
	if err := s.saveHistory(); err != nil {
		return err
	}

	return nil
}

// 설정 디렉토리 경로를 반환합니다.
func (s *JSONStore) GetConfigDir() string {
	return s.configDir
}

// ===== Config 메서드 =====

// 설정을 로드합니다.
func (s *JSONStore) GetConfig() (*model.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.configDir, configFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// 설정 파일이 없으면 기본값으로 파일 생성
		config := model.DefaultConfig()
		configData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(path, configData, 0644)
		return config, nil
	}
	if err != nil {
		return nil, err
	}

	var config model.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// 설정을 저장합니다.
func (s *JSONStore) SaveConfig(config *model.Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, configFile)
	return os.WriteFile(path, data, 0644)
}

// ===== ProcessInfo 메서드 =====

// 모든 패키지을 반환합니다.
func (s *JSONStore) GetAllPrograms() ([]*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	programs := make([]*model.ProcessInfo, 0, len(s.programs))
	for _, p := range s.programs {
		programs = append(programs, p.Clone())
	}
	return programs, nil
}

// 특정 ID의 패키지을 반환합니다.
func (s *JSONStore) GetProgram(id int) (*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.programs[id]
	if !ok {
		return nil, fmt.Errorf("패키지을 찾을 수 없습니다: %d", id)
	}
	return p.Clone(), nil
}

// 이름으로 패키지을 검색합니다.
func (s *JSONStore) GetProgramByName(name string) (*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, p := range s.programs {
		if p.ProcessName == name {
			return p.Clone(), nil
		}
	}
	return nil, fmt.Errorf("패키지을 찾을 수 없습니다: %s", name)
}

// 패키지을 저장합니다.
func (s *JSONStore) SaveProgram(program *model.ProcessInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 기존 패키지인지 확인 (ID로 판단)
	_, exists := s.programs[program.ID]

	// 새 패키지인 경우에만 ID 할당 (ID가 -1인 경우)
	if program.ID < 0 {
		program.ID = s.nextProgramID
		s.nextProgramID++
	}

	// 기존 패키지이 아니고 ID >= nextProgramID인 경우, nextProgramID 업데이트
	if !exists && program.ID >= s.nextProgramID {
		s.nextProgramID = program.ID + 1
	}

	s.programs[program.ID] = program.Clone()
	return s.savePrograms()
}

// 패키지을 삭제합니다.
func (s *JSONStore) DeleteProgram(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.programs[id]; !ok {
		return fmt.Errorf("패키지을 찾을 수 없습니다: %d", id)
	}

	delete(s.programs, id)
	return s.savePrograms()
}

// 모든 패키지을 삭제합니다.
func (s *JSONStore) ClearPrograms() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.programs = make(map[int]*model.ProcessInfo)
	s.nextProgramID = 1
	return s.savePrograms()
}

// 패키지 수를 반환합니다.
func (s *JSONStore) CountPrograms() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.programs)
}

// IP로 장비를 조회합니다.
func (s *JSONStore) GetFirewallByIP(ip string) (*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, fw := range s.firewalls {
		if fw.DeviceIP == ip {
			return fw, nil
		}
	}
	return nil, fmt.Errorf("장비를 찾을 수 없습니다: %s", ip)
}

// 장비 수를 반환합니다.
func (s *JSONStore) CountFirewalls() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.firewalls)
}

// 이력 수를 반환합니다.
func (s *JSONStore) CountHistory() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.history)
}

// 규칙 파일 수를 반환합니다.
func (s *JSONStore) CountRuleFiles() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.ruleFiles)
}
