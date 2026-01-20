package model

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FirewallFile 방화벽 규칙 파일 정보
type FirewallFile struct {
	FileName   string    // 파일명 (예: rules-v1_0_1.txt)
	FilePath   string    // 전체 파일 경로
	CreatedAt  time.Time // 만든 날짜 (파일 시스템 메타데이터)
	ModifiedAt time.Time // 수정한 날짜 (파일 시스템 메타데이터)
	Version    string    // 버전 (파일명에서 추출: rules-v1_0_1.txt -> v1.0.1)
	Contents   string    // 파일 내용
}

// NewFirewallFile 새로운 방화벽 파일 정보를 생성합니다.
func NewFirewallFile(fileName string) *FirewallFile {
	now := time.Now()
	return &FirewallFile{
		FileName:   fileName,
		CreatedAt:  now,
		ModifiedAt: now,
		Version:    ExtractVersion(fileName),
		Contents:   "",
	}
}

// ExtractVersion 파일명에서 버전을 추출합니다.
// 예: rules-v1_0_1.txt -> v1.0.1
// 예: firewall-v2_3.txt -> v2.3
// 예: test-v1.txt -> v1
// 형식이 맞지 않으면 "-" 반환
func ExtractVersion(fileName string) string {
	// 확장자 제거
	name := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	// "-" 구분자로 분리
	parts := strings.Split(name, "-")
	if len(parts) < 2 {
		return "-"
	}

	// 마지막 부분이 버전 패턴인지 확인 (v로 시작)
	versionPart := parts[len(parts)-1]
	if !strings.HasPrefix(versionPart, "v") {
		return "-"
	}

	// v1_0_1 -> v1.0.1 변환
	version := strings.ReplaceAll(versionPart, "_", ".")

	// 버전 형식 검증 (v숫자 또는 v숫자.숫자... 패턴)
	versionPattern := regexp.MustCompile(`^v\d+(\.\d+)*$`)
	if !versionPattern.MatchString(version) {
		return "-"
	}

	return version
}

// GetDisplayCreatedAt 표시용 생성일 반환
func (f *FirewallFile) GetDisplayCreatedAt() string {
	return f.CreatedAt.Format("2006/01/02 15:04:05")
}

// GetDisplayModifiedAt 표시용 수정일 반환
func (f *FirewallFile) GetDisplayModifiedAt() string {
	return f.ModifiedAt.Format("2006/01/02 15:04:05")
}

// Clone 복사본을 반환합니다.
func (f *FirewallFile) Clone() *FirewallFile {
	return &FirewallFile{
		FileName:   f.FileName,
		FilePath:   f.FilePath,
		CreatedAt:  f.CreatedAt,
		ModifiedAt: f.ModifiedAt,
		Version:    f.Version,
		Contents:   f.Contents,
	}
}

// IsValid 파일 정보가 유효한지 확인합니다.
func (f *FirewallFile) IsValid() bool {
	return f.FileName != ""
}
