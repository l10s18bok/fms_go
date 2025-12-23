package main

import (
	"log"
	"os"
	"path/filepath"

	"fms/internal/storage"
	"fms/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	// 실행 파일 경로 기준으로 설정 디렉토리 설정
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("실행 파일 경로를 찾을 수 없습니다: %v", err)
	}
	// 심볼릭 링크 해결 (실제 경로 획득)
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		log.Fatalf("실행 파일 경로를 해석할 수 없습니다: %v", err)
	}
	execDir := filepath.Dir(execPath)
	configDir := filepath.Join(execDir, "config")

	// 로그 파일 설정 (config/fms.log)
	// logFile, err := os.OpenFile(filepath.Join(configDir, "fms.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	// if err == nil {
	// 	// GUI 모드에서는 stdout이 없으므로 파일에만 출력
	// 	log.SetOutput(logFile)
	// 	defer logFile.Close()
	// }
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// log.Println("=== FMS 애플리케이션 시작 ===")

	// 저장소 초기화
	store, err := storage.NewJSONStore(configDir)
	if err != nil {
		log.Fatalf("저장소 초기화 실패: %v", err)
	}

	// Fyne 애플리케이션 생성
	a := app.New()

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
		w.Resize(fyne.NewSize(1200, 800))
		log.Println("데스크톱 환경 감지됨")
	}

	// 메인 UI 생성 및 설정
	mainUI := ui.NewMainUI(w, store)
	w.SetContent(mainUI.Content())

	// 윈도우 표시 및 실행
	w.ShowAndRun()
}
