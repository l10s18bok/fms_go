package main

import (
	"log"
	"os"
	"path/filepath"

	"fms/internal/model"
	"fms/internal/repository"
	"fms/internal/storage"
	"fms/internal/translations"
	"fms/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	fynetooltip "github.com/dweymouth/fyne-tooltip"
)

func main() {
	// 실행 파일 경로 기준으로 설정 디렉토리 설정
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("실행 파일 경로를 찾을 수 없습니다: %v", err)
	}
	// 심볼릭 링크 해결 시도, 실패하면 원본 경로 사용
	resolvedPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Printf("심볼릭 링크 해결 실패, 원본 경로 사용: %v", err)
		resolvedPath = execPath
	}
	execDir := filepath.Dir(resolvedPath)
	configDir := filepath.Join(execDir, "config")

	// JSON 저장소 초기화 (기존 데이터용)
	jsonStore, err := storage.NewJSONStore(configDir)
	if err != nil {
		log.Fatalf("JSON 저장소 초기화 실패: %v", err)
	}

	// SQLite 저장소 초기화 (배포 이력용)
	sqliteStore, err := storage.NewSQLiteStore(configDir)
	if err != nil {
		log.Fatalf("SQLite 저장소 초기화 실패: %v", err)
	}
	defer sqliteStore.Close()

	// Repository 생성 (SQLite 사용)
	historyRepo := repository.NewSQLiteHistoryRepository(sqliteStore)
	firewallRepo := repository.NewSQLiteFirewallRepository(sqliteStore)
	programRepo := repository.NewSQLiteProgramRepository(sqliteStore)

	// 파일 저장소 초기화 (config/rule_files 디렉토리)
	fileStore, err := storage.NewFileStore(configDir)
	if err != nil {
		log.Fatalf("파일 저장소 초기화 실패: %v", err)
	}

	// 달력 한국어 번역 등록
	if err := lang.AddTranslationsForLocale(translations.KoJSON, lang.SystemLocale()); err != nil {
		log.Printf("한국어 번역 로드 실패: %v", err)
	}

	// Fyne 애플리케이션 생성
	a := app.New()

	// 저장된 테마 설정 로드 및 적용
	config, err := jsonStore.GetConfig()
	if err == nil && config.Theme == model.ThemeDark {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}

	// 메인 윈도우 생성
	w := a.NewWindow("FMS - Firewall Management System")
	w.Resize(fyne.NewSize(1400, 800))

	// 메인 UI 생성 및 설정
	mainUI := ui.NewMainUI(w, jsonStore, fileStore, historyRepo, firewallRepo, programRepo)
	w.SetContent(fynetooltip.AddWindowToolTipLayer(mainUI.Content(), w.Canvas()))

	// 윈도우 표시 및 실행
	w.ShowAndRun()
}
