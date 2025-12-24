package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fms_wails/internal/deploy"
	"fms_wails/internal/model"
	"fms_wails/internal/storage"
	"fms_wails/internal/version"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	store    *storage.JSONStore
	deployer *deploy.Deployer
	config   *model.Config
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
	a.deployer = deploy.NewDeployer(a.config)

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
	a.deployer.UpdateConfig(&config)
	return a.store.SaveConfig(&config)
}

// ===== 템플릿 API =====

// GetAllTemplates는 모든 템플릿을 반환합니다.
func (a *App) GetAllTemplates() []*model.Template {
	if a.store == nil {
		return []*model.Template{}
	}
	templates, _ := a.store.GetAllTemplates()
	return templates
}

// GetTemplate는 특정 버전의 템플릿을 반환합니다.
func (a *App) GetTemplate(version string) *model.Template {
	if a.store == nil {
		return nil
	}
	template, err := a.store.GetTemplate(version)
	if err != nil {
		return nil
	}
	return template
}

// SaveTemplate는 템플릿을 저장합니다.
func (a *App) SaveTemplate(version, contents string) error {
	if a.store == nil {
		return nil
	}
	template := model.NewTemplate(version, contents)
	if !template.IsValid() {
		return fmt.Errorf("유효하지 않은 템플릿입니다. 버전과 내용을 확인해주세요.")
	}
	return a.store.SaveTemplate(template)
}

// DeleteTemplate는 템플릿을 삭제합니다.
func (a *App) DeleteTemplate(version string) error {
	if a.store == nil {
		return nil
	}
	return a.store.DeleteTemplate(version)
}

// ===== 장비 API =====

// GetAllFirewalls는 모든 장비를 반환합니다.
func (a *App) GetAllFirewalls() []*model.Firewall {
	if a.store == nil {
		return []*model.Firewall{}
	}
	firewalls, _ := a.store.GetAllFirewalls()
	return firewalls
}

// GetFirewall는 특정 장비를 반환합니다.
func (a *App) GetFirewall(index int) *model.Firewall {
	if a.store == nil {
		return nil
	}
	firewall, err := a.store.GetFirewall(index)
	if err != nil {
		return nil
	}
	return firewall
}

// SaveFirewall는 장비를 저장합니다.
func (a *App) SaveFirewall(firewallJSON string) error {
	if a.store == nil {
		return nil
	}
	var firewall model.Firewall
	if err := json.Unmarshal([]byte(firewallJSON), &firewall); err != nil {
		return err
	}
	return a.store.SaveFirewall(&firewall)
}

// DeleteFirewall는 장비를 삭제합니다.
func (a *App) DeleteFirewall(index int) error {
	if a.store == nil {
		return nil
	}
	return a.store.DeleteFirewall(index)
}

// CheckServerStatus는 서버 상태를 확인합니다.
func (a *App) CheckServerStatus(index int) string {
	if a.store == nil || a.deployer == nil {
		return model.ServerStatusStop
	}

	firewall, err := a.store.GetFirewall(index)
	if err != nil {
		return model.ServerStatusStop
	}

	a.deployer.HealthCheck(firewall)

	// 상태 업데이트
	a.store.SaveFirewall(firewall)

	return firewall.ServerStatus
}

// CheckAllServerStatus는 모든 장비의 상태를 확인합니다.
func (a *App) CheckAllServerStatus() {
	if a.store == nil || a.deployer == nil {
		return
	}

	firewalls, _ := a.store.GetAllFirewalls()
	if len(firewalls) == 0 {
		return
	}

	// Agent 모드면 배치 호출, 아니면 개별 호출
	if a.config.IsAgentMode() {
		a.deployer.HealthCheckBatch(firewalls)
	} else {
		a.deployer.HealthCheckMultiple(firewalls, nil)
	}

	// 상태 저장
	for _, fw := range firewalls {
		a.store.SaveFirewall(fw)
	}
}

// CheckSelectedServerStatus는 선택된 장비들의 상태를 병렬로 확인합니다.
func (a *App) CheckSelectedServerStatus(indexes []int) {
	if a.store == nil || a.deployer == nil {
		return
	}

	// 선택된 장비들을 가져옴
	var selectedFirewalls []*model.Firewall
	for _, idx := range indexes {
		fw, err := a.store.GetFirewall(idx)
		if err == nil && fw != nil {
			selectedFirewalls = append(selectedFirewalls, fw)
		}
	}

	if len(selectedFirewalls) == 0 {
		return
	}

	// Agent 모드면 배치 호출, 아니면 병렬 개별 호출
	if a.config.IsAgentMode() {
		a.deployer.HealthCheckBatch(selectedFirewalls)
	} else {
		a.deployer.HealthCheckMultiple(selectedFirewalls, nil)
	}

	// 상태 저장
	for _, fw := range selectedFirewalls {
		a.store.SaveFirewall(fw)
	}
}

// ===== 배포 API =====

// Deploy는 템플릿을 장비에 배포합니다.
func (a *App) Deploy(firewallIndex int, templateVersion string) (*model.DeployHistory, error) {
	if a.store == nil || a.deployer == nil {
		return nil, nil
	}

	firewall, err := a.store.GetFirewall(firewallIndex)
	if err != nil {
		return nil, err
	}

	template, err := a.store.GetTemplate(templateVersion)
	if err != nil {
		return nil, err
	}

	result := a.deployer.Deploy(firewall, template)

	// 이력 저장
	a.store.SaveHistory(result.History)

	// 장비 상태 업데이트
	a.store.SaveFirewall(firewall)

	return result.History, nil
}

// ===== 이력 API =====

// GetAllHistory는 모든 배포 이력을 반환합니다.
func (a *App) GetAllHistory() []*model.DeployHistory {
	if a.store == nil {
		return []*model.DeployHistory{}
	}
	history, _ := a.store.GetAllHistory()
	return history
}

// DeleteHistory는 배포 이력을 삭제합니다.
func (a *App) DeleteHistory(id int) error {
	if a.store == nil {
		return nil
	}
	return a.store.DeleteHistory(id)
}

// SaveHistory는 배포 이력을 저장합니다. (Import용)
func (a *App) SaveHistory(historyJSON string) error {
	if a.store == nil {
		return nil
	}
	var history model.DeployHistory
	if err := json.Unmarshal([]byte(historyJSON), &history); err != nil {
		return err
	}
	return a.store.SaveHistory(&history)
}

// ===== Export/Import API =====

// ExportData는 모든 데이터를 JSON으로 내보냅니다.
func (a *App) ExportData() string {
	if a.store == nil {
		return "{}"
	}
	data, _ := a.store.ExportAll()
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonBytes)
}

// ImportData는 JSON 데이터를 가져옵니다.
func (a *App) ImportData(jsonData string) error {
	if a.store == nil {
		return nil
	}
	var data storage.ExportData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}
	return a.store.ImportAll(&data)
}

// ===== Reset API =====

// ResetAll은 모든 데이터를 초기화합니다.
func (a *App) ResetAll() error {
	if a.store == nil {
		return nil
	}
	return a.store.ClearAll()
}

// GetConfigDir은 설정 디렉토리 경로를 반환합니다.
func (a *App) GetConfigDir() string {
	if a.store == nil {
		return ""
	}
	return a.store.GetConfigDir()
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
