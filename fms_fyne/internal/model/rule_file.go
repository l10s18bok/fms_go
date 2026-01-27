// Package model은 FMS 애플리케이션의 데이터 모델을 정의합니다.
package model

import "strings"

// 방화벽 규칙 파일을 나타냅니다.
type RuleFile struct {
	Version  string `json:"version"`  // 규칙 파일 버전명 (Primary Key)
	Contents string `json:"contents"` // 방화벽 규칙 내용 (줄 단위)
}

// 새로운 규칙 파일을 생성합니다.
func NewRuleFile(version, contents string) *RuleFile {
	return &RuleFile{
		Version:  version,
		Contents: contents,
	}
}

// 규칙 파일이 유효한지 검사합니다.
func (r *RuleFile) IsValid() bool {
	return r.Version != "" && strings.TrimSpace(r.Contents) != ""
}

// 규칙 파일의 복사본을 반환합니다.
func (r *RuleFile) Clone() *RuleFile {
	return &RuleFile{
		Version:  r.Version,
		Contents: r.Contents,
	}
}
