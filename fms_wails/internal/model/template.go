package model

// Template은 방화벽 규칙 템플릿입니다.
type Template struct {
	Version  string `json:"version"`
	Contents string `json:"contents"`
}

// Clone은 템플릿의 복사본을 반환합니다.
func (t *Template) Clone() *Template {
	return &Template{
		Version:  t.Version,
		Contents: t.Contents,
	}
}
