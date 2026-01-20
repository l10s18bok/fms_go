package main

import (
	"log"
	"os"
	"path/filepath"

	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
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

	// 저장소 초기화
	store, err := storage.NewJSONStore(configDir)
	if err != nil {
		log.Fatalf("저장소 초기화 실패: %v", err)
	}

	// 파일 저장소 초기화 (data 디렉토리)
	fileStore, err := storage.NewFileStore(execDir)
	if err != nil {
		log.Fatalf("파일 저장소 초기화 실패: %v", err)
	}

	// Fyne 애플리케이션 생성
	a := app.New()

	// 저장된 테마 설정 로드 및 적용
	config, err := store.GetConfig()
	if err == nil && config.Theme == model.ThemeDark {
		a.Settings().SetTheme(theme.DarkTheme())
	} else {
		a.Settings().SetTheme(theme.LightTheme())
	}

	// 메인 윈도우 생성
	w := a.NewWindow("FMS - Firewall Management System")

	// 플랫폼에 따른 윈도우 크기 설정
	// 모바일에서는 Resize가 무시되고 전체화면으로 동작
	device := fyne.CurrentDevice()
	if device.IsMobile() {
		// 모바일: 전체화면 (Resize 호출해도 무시됨)
		log.Println("모바일 환경 감지됨")
	} else {
		// 데스크톱: 지정 크기로 설정
		w.Resize(fyne.NewSize(1400, 800))
		log.Println("데스크톱 환경 감지됨")
	}

	// 메인 UI 생성 및 설정
	mainUI := ui.NewMainUI(w, store, fileStore)
	w.SetContent(mainUI.Content())

	// 윈도우 표시 및 실행
	w.ShowAndRun()
}
