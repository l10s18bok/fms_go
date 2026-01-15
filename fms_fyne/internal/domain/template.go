package domain

import "strings"

// Template 방화벽 규칙 템플릿 엔티티
type Template struct {
	Version  string `json:"version"`  // 템플릿 버전명 (Primary Key)
	Contents string `json:"contents"` // 방화벽 규칙 내용 (줄 단위)
}

// NewTemplate 새로운 템플릿을 생성합니다.
func NewTemplate(version, contents string) *Template {
	return &Template{
		Version:  version,
		Contents: contents,
	}
}

// IsValid 템플릿이 유효한지 검사합니다.
func (t *Template) IsValid() bool {
	return t.Version != "" && strings.TrimSpace(t.Contents) != ""
}

// Clone 템플릿의 복사본을 반환합니다.
func (t *Template) Clone() *Template {
	return &Template{
		Version:  t.Version,
		Contents: t.Contents,
	}
}
