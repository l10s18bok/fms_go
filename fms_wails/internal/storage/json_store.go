package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"fms_wails/internal/model"
)

// JSONStore는 JSON 파일 기반 저장소입니다.
type JSONStore struct {
	configDir string
	mu        sync.RWMutex

	// 캐시된 데이터
	templates map[string]*model.Template
	firewalls map[int]*model.Firewall
	history   map[int]*model.DeployHistory

	// Auto increment 카운터
	nextFirewallID int
	nextHistoryID  int
}

// 파일명 상수
const (
	templatesFile = "templates.json"
	firewallsFile = "firewalls.json"
	historyFile   = "history.json"
	configFile    = "config.json"
)

// NewJSONStore는 새로운 JSON 저장소를 생성합니다.
func NewJSONStore(configDir string) (*JSONStore, error) {
	store := &JSONStore{
		configDir:      configDir,
		templates:      make(map[string]*model.Template),
		firewalls:      make(map[int]*model.Firewall),
		history:        make(map[int]*model.DeployHistory),
		nextFirewallID: 1,
		nextHistoryID:  1,
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

// loadAll은 모든 데이터 파일을 로드합니다.
func (s *JSONStore) loadAll() error {
	if err := s.loadTemplates(); err != nil {
		return err
	}
	if err := s.loadFirewalls(); err != nil {
		return err
	}
	if err := s.loadHistory(); err != nil {
		return err
	}
	return nil
}

// loadTemplates는 템플릿 데이터를 로드합니다.
func (s *JSONStore) loadTemplates() error {
	path := filepath.Join(s.configDir, templatesFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var templates []*model.Template
	if err := json.Unmarshal(data, &templates); err != nil {
		return err
	}

	for _, t := range templates {
		s.templates[t.Version] = t
	}
	return nil
}

// loadFirewalls는 장비 데이터를 로드합니다.
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

// loadHistory는 배포 이력 데이터를 로드합니다.
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

// saveTemplates는 템플릿 데이터를 저장합니다.
func (s *JSONStore) saveTemplates() error {
	templates := make([]*model.Template, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t)
	}

	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.configDir, templatesFile)
	return os.WriteFile(path, data, 0644)
}

// saveFirewalls는 장비 데이터를 저장합니다.
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

// saveHistory는 배포 이력 데이터를 저장합니다.
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

// ===== Template 메서드 =====

// GetAllTemplates는 모든 템플릿을 반환합니다.
func (s *JSONStore) GetAllTemplates() ([]*model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	templates := make([]*model.Template, 0, len(s.templates))
	for _, t := range s.templates {
		templates = append(templates, t.Clone())
	}
	return templates, nil
}

// GetTemplate는 특정 버전의 템플릿을 반환합니다.
func (s *JSONStore) GetTemplate(version string) (*model.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.templates[version]
	if !ok {
		return nil, fmt.Errorf("템플릿을 찾을 수 없습니다: %s", version)
	}
	return t.Clone(), nil
}

// SaveTemplate는 템플릿을 저장합니다.
func (s *JSONStore) SaveTemplate(template *model.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.templates[template.Version] = template.Clone()
	return s.saveTemplates()
}

// DeleteTemplate는 템플릿을 삭제합니다.
func (s *JSONStore) DeleteTemplate(version string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.templates[version]; !ok {
		return fmt.Errorf("템플릿을 찾을 수 없습니다: %s", version)
	}

	delete(s.templates, version)
	return s.saveTemplates()
}

// ===== Firewall 메서드 =====

// GetAllFirewalls는 모든 장비를 반환합니다.
func (s *JSONStore) GetAllFirewalls() ([]*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	firewalls := make([]*model.Firewall, 0, len(s.firewalls))
	for _, f := range s.firewalls {
		firewalls = append(firewalls, f.Clone())
	}
	return firewalls, nil
}

// GetFirewall는 특정 인덱스의 장비를 반환합니다.
func (s *JSONStore) GetFirewall(index int) (*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, ok := s.firewalls[index]
	if !ok {
		return nil, fmt.Errorf("장비를 찾을 수 없습니다: %d", index)
	}
	return f.Clone(), nil
}

// SaveFirewall는 장비를 저장합니다.
func (s *JSONStore) SaveFirewall(firewall *model.Firewall) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 새 장비인 경우 ID 할당
	if firewall.Index == 0 {
		firewall.Index = s.nextFirewallID
		s.nextFirewallID++
	}

	s.firewalls[firewall.Index] = firewall.Clone()
	return s.saveFirewalls()
}

// DeleteFirewall는 장비를 삭제합니다.
func (s *JSONStore) DeleteFirewall(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.firewalls[index]; !ok {
		return fmt.Errorf("장비를 찾을 수 없습니다: %d", index)
	}

	delete(s.firewalls, index)
	return s.saveFirewalls()
}

// ===== History 메서드 =====

// GetAllHistory는 모든 배포 이력을 반환합니다.
func (s *JSONStore) GetAllHistory() ([]*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]*model.DeployHistory, 0, len(s.history))
	for _, h := range s.history {
		hCopy := *h
		hCopy.Results = make([]model.RuleResult, len(h.Results))
		copy(hCopy.Results, h.Results)
		history = append(history, &hCopy)
	}
	return history, nil
}

// SaveHistory는 배포 이력을 저장합니다.
func (s *JSONStore) SaveHistory(history *model.DeployHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if history.ID == 0 {
		history.ID = s.nextHistoryID
		s.nextHistoryID++
	}

	hCopy := *history
	hCopy.Results = make([]model.RuleResult, len(history.Results))
	copy(hCopy.Results, history.Results)
	s.history[history.ID] = &hCopy

	return s.saveHistory()
}

// DeleteHistory는 배포 이력을 삭제합니다.
func (s *JSONStore) DeleteHistory(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.history[id]; !ok {
		return fmt.Errorf("배포 이력을 찾을 수 없습니다: %d", id)
	}

	delete(s.history, id)
	return s.saveHistory()
}

// ===== Export/Import =====

// ExportAll은 모든 데이터를 반환합니다.
func (s *JSONStore) ExportAll() (*ExportData, error) {
	templates, _ := s.GetAllTemplates()
	firewalls, _ := s.GetAllFirewalls()
	history, _ := s.GetAllHistory()

	return &ExportData{
		Templates: templates,
		Firewalls: firewalls,
		History:   history,
	}, nil
}

// ImportAll은 데이터를 가져옵니다.
func (s *JSONStore) ImportAll(data *ExportData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range data.Templates {
		s.templates[t.Version] = t.Clone()
	}

	for _, f := range data.Firewalls {
		if f.Index == 0 {
			f.Index = s.nextFirewallID
			s.nextFirewallID++
		}
		s.firewalls[f.Index] = f.Clone()
		if f.Index >= s.nextFirewallID {
			s.nextFirewallID = f.Index + 1
		}
	}

	for _, h := range data.History {
		if h.ID == 0 {
			h.ID = s.nextHistoryID
			s.nextHistoryID++
		}
		hCopy := *h
		hCopy.Results = make([]model.RuleResult, len(h.Results))
		copy(hCopy.Results, h.Results)
		s.history[h.ID] = &hCopy
		if h.ID >= s.nextHistoryID {
			s.nextHistoryID = h.ID + 1
		}
	}

	if err := s.saveTemplates(); err != nil {
		return err
	}
	if err := s.saveFirewalls(); err != nil {
		return err
	}
	return s.saveHistory()
}

// GetConfigDir는 설정 디렉토리 경로를 반환합니다.
func (s *JSONStore) GetConfigDir() string {
	return s.configDir
}

// ===== Config 메서드 =====

// GetConfig는 설정을 로드합니다.
func (s *JSONStore) GetConfig() (*model.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.configDir, configFile)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return model.DefaultConfig(), nil
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

// SaveConfig는 설정을 저장합니다.
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

// ===== Clear 메서드 =====

// ClearTemplates는 모든 템플릿을 삭제합니다.
func (s *JSONStore) ClearTemplates() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.templates = make(map[string]*model.Template)
	return s.saveTemplates()
}

// ClearFirewalls는 모든 장비를 삭제합니다.
func (s *JSONStore) ClearFirewalls() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.firewalls = make(map[int]*model.Firewall)
	s.nextFirewallID = 1
	return s.saveFirewalls()
}

// ClearHistory는 모든 배포 이력을 삭제합니다.
func (s *JSONStore) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history = make(map[int]*model.DeployHistory)
	s.nextHistoryID = 1
	return s.saveHistory()
}

// ClearAll은 모든 데이터를 삭제합니다.
func (s *JSONStore) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 메모리 캐시 초기화
	s.templates = make(map[string]*model.Template)
	s.firewalls = make(map[int]*model.Firewall)
	s.history = make(map[int]*model.DeployHistory)
	s.nextFirewallID = 1
	s.nextHistoryID = 1

	// 파일 저장
	if err := s.saveTemplates(); err != nil {
		return err
	}
	if err := s.saveFirewalls(); err != nil {
		return err
	}
	return s.saveHistory()
}

// ReloadAll은 파일에서 모든 데이터를 다시 로드합니다.
func (s *JSONStore) ReloadAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 기존 캐시 초기화
	s.templates = make(map[string]*model.Template)
	s.firewalls = make(map[int]*model.Firewall)
	s.history = make(map[int]*model.DeployHistory)
	s.nextFirewallID = 1
	s.nextHistoryID = 1

	// 파일에서 다시 로드
	if err := s.loadTemplates(); err != nil {
		return err
	}
	if err := s.loadFirewalls(); err != nil {
		return err
	}
	return s.loadHistory()
}
