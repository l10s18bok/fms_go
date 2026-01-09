package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"fms_wails/internal/deploy"
	"fms_wails/internal/model"
	"fms_wails/internal/service"
	"fms_wails/internal/storage"
	"fms_wails/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct - Wails 프레임워크와 서비스 계층을 연결하는 파사드
type App struct {
	ctx context.Context

	// 서비스 계층
	templateSvc *service.TemplateService
	firewallSvc *service.FirewallService
	deploySvc   *service.DeployService
	parserSvc   *service.ParserService

	// 설정
	store  *storage.JSONStore
	config *model.Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 실행 파일 경로 기준으로 설정 디렉토리 설정
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("실행 파일 경로를 찾을 수 없습니다: %v", err)
		return
	}

	// 심볼릭 링크 해결 시도, 실패하면 원본 경로 사용
	resolvedPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Printf("심볼릭 링크 해결 실패, 원본 경로 사용: %v", err)
		resolvedPath = execPath
	}

	execDir := filepath.Dir(resolvedPath)
	configDir := filepath.Join(execDir, "config2")

	store, err := storage.NewJSONStore(configDir)
	if err != nil {
		log.Printf("저장소 초기화 실패: %v", err)
		return
	}
	a.store = store

	// 설정 로드
	config, err := a.store.GetConfig()
	if err != nil {
		config = model.DefaultConfig()
	}
	a.config = config

	// Deployer 초기화
	deployer := deploy.NewDeployer(a.config)

	// 서비스 계층 초기화
	a.templateSvc = service.NewTemplateService(store)
	a.firewallSvc = service.NewFirewallService(store, deployer, config)
	a.deploySvc = service.NewDeployService(store, store, store, store, deployer)
	a.parserSvc = service.NewParserService()

	log.Printf("저장소 초기화 완료: %s", configDir)
}

// ===== 설정 API =====

// GetConfig는 현재 설정을 반환합니다.
func (a *App) GetConfig() *model.Config {
	if a.config == nil {
		return model.DefaultConfig()
	}
	return a.config
}

// SaveConfig는 설정을 저장합니다.
func (a *App) SaveConfig(configJSON string) error {
	if a.store == nil {
		return nil
	}
	var config model.Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return err
	}
	a.config = &config
	a.firewallSvc.UpdateConfig(&config)
	return a.store.SaveConfig(&config)
}

// ===== 템플릿 API =====

// GetAllTemplates는 모든 템플릿을 반환합니다.
func (a *App) GetAllTemplates() []*model.Template {
	return a.templateSvc.GetAll()
}

// GetTemplate는 특정 버전의 템플릿을 반환합니다.
func (a *App) GetTemplate(version string) *model.Template {
	return a.templateSvc.Get(version)
}

// SaveTemplate는 템플릿을 저장합니다.
func (a *App) SaveTemplate(version, contents string) error {
	return a.templateSvc.Save(version, contents)
}

// DeleteTemplate는 템플릿을 삭제합니다.
func (a *App) DeleteTemplate(version string) error {
	return a.templateSvc.Delete(version)
}

// DeleteAllTemplates는 모든 템플릿을 삭제합니다.
func (a *App) DeleteAllTemplates() error {
	return a.templateSvc.DeleteAll()
}

// ===== 장비 API =====

// GetAllFirewalls는 모든 장비를 반환합니다.
func (a *App) GetAllFirewalls() []*model.Firewall {
	return a.firewallSvc.GetAll()
}

// GetFirewall는 특정 장비를 반환합니다.
func (a *App) GetFirewall(index int) *model.Firewall {
	return a.firewallSvc.Get(index)
}

// SaveFirewall는 장비를 저장합니다.
func (a *App) SaveFirewall(firewallJSON string) error {
	return a.firewallSvc.Save(firewallJSON)
}

// DeleteFirewall는 장비를 삭제합니다.
func (a *App) DeleteFirewall(index int) error {
	return a.firewallSvc.Delete(index)
}

// DeleteAllFirewalls는 모든 장비를 삭제합니다.
func (a *App) DeleteAllFirewalls() error {
	return a.firewallSvc.DeleteAll()
}

// CheckServerStatus는 서버 상태를 확인합니다.
func (a *App) CheckServerStatus(index int) string {
	return a.firewallSvc.CheckServerStatus(index)
}

// CheckAllServerStatus는 모든 장비의 상태를 확인합니다.
func (a *App) CheckAllServerStatus() {
	a.firewallSvc.CheckAllServerStatus()
}

// CheckSelectedServerStatus는 선택된 장비들의 상태를 병렬로 확인합니다.
func (a *App) CheckSelectedServerStatus(indexes []int) {
	a.firewallSvc.CheckSelectedServerStatus(indexes)
}

// ===== 배포 API =====

// Deploy는 템플릿을 장비에 배포합니다.
func (a *App) Deploy(firewallIndex int, templateVersion string) (*model.DeployHistory, error) {
	return a.deploySvc.Deploy(firewallIndex, templateVersion)
}

// ===== 이력 API =====

// GetAllHistory는 모든 배포 이력을 반환합니다.
func (a *App) GetAllHistory() []*model.DeployHistory {
	return a.deploySvc.GetAllHistory()
}

// DeleteHistory는 배포 이력을 삭제합니다.
func (a *App) DeleteHistory(id int) error {
	return a.deploySvc.DeleteHistory(id)
}

// DeleteAllHistory는 모든 배포 이력을 삭제합니다.
func (a *App) DeleteAllHistory() error {
	return a.deploySvc.DeleteAllHistory()
}

// SaveHistory는 배포 이력을 저장합니다. (Import용)
func (a *App) SaveHistory(historyJSON string) error {
	return a.deploySvc.SaveHistory(historyJSON)
}

// ===== Export/Import API =====

// ExportData는 모든 데이터를 JSON으로 내보냅니다.
func (a *App) ExportData() string {
	return a.deploySvc.ExportData()
}

// ImportData는 JSON 데이터를 가져옵니다.
func (a *App) ImportData(jsonData string) error {
	return a.deploySvc.ImportData(jsonData)
}

// ===== Reset API =====

// ResetAll은 모든 데이터를 초기화합니다.
func (a *App) ResetAll() error {
	return a.deploySvc.ResetAll()
}

// GetConfigDir은 설정 디렉토리 경로를 반환합니다.
func (a *App) GetConfigDir() string {
	return a.deploySvc.GetConfigDir()
}

// ===== 네이티브 파일 다이얼로그 API =====

// OpenFileDialog는 파일 열기 다이얼로그를 표시합니다.
func (a *App) OpenFileDialog(title string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
}

// SaveFileDialog는 파일 저장 다이얼로그를 표시합니다.
func (a *App) SaveFileDialog(title, defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
}

// ReadFileContent는 파일 내용을 읽어 반환합니다.
func (a *App) ReadFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFileContent는 파일에 내용을 씁니다.
func (a *App) WriteFileContent(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ConfirmDialog는 확인 대화상자를 표시합니다.
func (a *App) ConfirmDialog(title, message string) (string, error) {
	return runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"확인", "취소"},
		DefaultButton: "취소",
	})
}

// AlertDialog는 알림 대화상자를 표시합니다.
func (a *App) AlertDialog(title, message string) error {
	_, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
	return err
}

// ===== 버전 API =====

// GetAppVersion은 앱 버전을 반환합니다.
func (a *App) GetAppVersion() string {
	return version.AppVersion
}

// GetAppName은 앱 이름을 반환합니다.
func (a *App) GetAppName() string {
	return version.AppName
}

// GetAppFullName은 앱 전체 이름을 반환합니다.
func (a *App) GetAppFullName() string {
	return version.AppFullName
}

// ===== 규칙 파서 API =====

// ParseRules는 텍스트를 규칙 배열로 파싱합니다.
func (a *App) ParseRules(text string) *service.ParseRulesResult {
	return a.parserSvc.ParseRules(text)
}

// RulesToText는 규칙 배열을 텍스트로 변환합니다.
func (a *App) RulesToText(rulesJSON string, commentsJSON string) string {
	return a.parserSvc.RulesToText(rulesJSON, commentsJSON)
}

// GetChainOptions는 Chain 옵션 목록을 반환합니다.
func (a *App) GetChainOptions() []string {
	return a.parserSvc.GetChainOptions()
}

// GetProtocolOptions는 Protocol 옵션 목록을 반환합니다.
func (a *App) GetProtocolOptions() []string {
	return a.parserSvc.GetProtocolOptions()
}

// GetActionOptions는 Action 옵션 목록을 반환합니다.
func (a *App) GetActionOptions() []string {
	return a.parserSvc.GetActionOptions()
}

// GetTCPFlagsPresets는 TCP Flags 프리셋 목록을 반환합니다.
func (a *App) GetTCPFlagsPresets() []model.TCPFlagsPreset {
	return a.parserSvc.GetTCPFlagsPresets()
}

// GetTCPFlagsList는 개별 TCP Flags 목록을 반환합니다.
func (a *App) GetTCPFlagsList() []string {
	return a.parserSvc.GetTCPFlagsList()
}

// GetICMPTypeOptions는 ICMP Type 옵션 목록을 반환합니다.
func (a *App) GetICMPTypeOptions() []string {
	return a.parserSvc.GetICMPTypeOptions()
}

// GetICMPCodeOptions는 ICMP Code 옵션 목록을 반환합니다.
func (a *App) GetICMPCodeOptions() []string {
	return a.parserSvc.GetICMPCodeOptions()
}

// NewFirewallRule은 기본값이 설정된 새 규칙을 생성합니다.
func (a *App) NewFirewallRule() *model.FirewallRule {
	return a.parserSvc.NewFirewallRule()
}

// ===== NAT 규칙 파서 API =====

// ParseNATRules는 텍스트를 NAT 규칙 배열로 파싱합니다.
func (a *App) ParseNATRules(text string) *service.ParseNATRulesResult {
	return a.parserSvc.ParseNATRules(text)
}

// NATRulesToText는 NAT 규칙 배열을 텍스트로 변환합니다.
func (a *App) NATRulesToText(rulesJSON string, commentsJSON string) string {
	return a.parserSvc.NATRulesToText(rulesJSON, commentsJSON)
}

// GetNATTypeOptions는 NAT 타입 옵션 목록을 반환합니다.
func (a *App) GetNATTypeOptions() []string {
	return a.parserSvc.GetNATTypeOptions()
}

// GetSNATTypeOptions는 SNAT 폼용 타입 옵션을 반환합니다.
func (a *App) GetSNATTypeOptions() []string {
	return a.parserSvc.GetSNATTypeOptions()
}

// NewNATRule은 기본값이 설정된 새 NAT 규칙을 생성합니다.
func (a *App) NewNATRule() *model.NATRule {
	return a.parserSvc.NewNATRule()
}

// NewDNATRule은 DNAT 규칙을 생성합니다.
func (a *App) NewDNATRule() *model.NATRule {
	return a.parserSvc.NewDNATRule()
}

// NewSNATRule은 SNAT 규칙을 생성합니다.
func (a *App) NewSNATRule() *model.NATRule {
	return a.parserSvc.NewSNATRule()
}
