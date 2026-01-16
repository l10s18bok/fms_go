package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fms/internal/themes"
)

// SearchBox 검색 입력 필드 + 찾기 버튼 공통 컴포넌트
type SearchBox struct {
	Entry     *widget.Entry
	Button    fyne.CanvasObject
	Container *fyne.Container
	onSearch  func(text string)
}

// SearchBoxConfig 검색 컴포넌트 설정
type SearchBoxConfig struct {
	Placeholder string  // 입력 필드 플레이스홀더
	Width       float32 // 입력 필드 너비 (기본값: 200)
	OnSearch    func(text string)
}

// NewSearchBox 새 검색 컴포넌트 생성
// "검색:" 라벨 + 입력 필드 + 돋보기 아이콘 찾기 버튼
func NewSearchBox(config SearchBoxConfig) *SearchBox {
	if config.Width <= 0 {
		config.Width = 200 // 기본값
	}

	s := &SearchBox{
		onSearch: config.OnSearch,
	}

	// 검색 입력 필드
	s.Entry = widget.NewEntry()
	if config.Placeholder != "" {
		s.Entry.SetPlaceHolder(config.Placeholder)
	}

	// 검색 실행 함수 (검색 후 필드 초기화)
	doSearch := func() {
		if s.onSearch != nil {
			s.onSearch(s.Entry.Text)
		}
		s.Entry.SetText("") // 검색 후 필드 초기화
	}

	// Enter 키 입력 시 검색 실행
	s.Entry.OnSubmitted = func(text string) {
		doSearch()
	}

	// 찾기 버튼 (돋보기 아이콘, NewCustomButton 사용)
	s.Button = NewCustomButton("", theme.SearchIcon(), nil, themes.Colors["darkgray"], func() {
		doSearch()
	})

	// 컨테이너 구성: "검색:" + 입력필드(고정크기) + 찾기버튼
	s.Container = container.NewHBox(
		widget.NewLabel("검색:"),
		container.NewGridWrap(fyne.NewSize(config.Width, 36), s.Entry),
		s.Button,
	)

	return s
}

// Content 컨테이너 반환
func (s *SearchBox) Content() *fyne.Container {
	return s.Container
}

// GetText 현재 입력된 텍스트 반환
func (s *SearchBox) GetText() string {
	return s.Entry.Text
}

// SetText 텍스트 설정
func (s *SearchBox) SetText(text string) {
	s.Entry.SetText(text)
}

// Clear 입력 필드 초기화
func (s *SearchBox) Clear() {
	s.Entry.SetText("")
}
