package domain

import "time"

// ProcessInfo 업데이트 프로그램 정보 엔티티
type ProcessInfo struct {
	ID                int    `json:"id"`                  // 고유 ID (Auto Increment)
	ProcessName       string `json:"process_name"`        // 프로그램 이름
	ProcessFilePath   string `json:"process_file_path"`   // 로컬 파일 경로
	ProcessUploadPath string `json:"process_upload_path"` // 서버 업로드 경로
	ProcessVersion    string `json:"process_version"`     // 프로그램 버전
	ProcessCreatedAt  string `json:"process_created_at"`  // 추가/수정 시간
}

// NewProcessInfo 새로운 프로그램 정보를 생성합니다.
func NewProcessInfo(name, version string) *ProcessInfo {
	return &ProcessInfo{
		ID:                -1, // 새 프로그램은 -1로 시작, 저장 시 ID 할당
		ProcessName:       name,
		ProcessVersion:    version,
		ProcessUploadPath: "/download/", // 기본 업로드 경로
		ProcessCreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}
}

// IsValid 프로그램 정보가 유효한지 검사합니다.
func (p *ProcessInfo) IsValid() bool {
	return p.ProcessName != "" && p.ProcessVersion != "" && p.ProcessFilePath != ""
}

// UpdateTimestamp 수정 시간을 현재 시간으로 업데이트합니다.
func (p *ProcessInfo) UpdateTimestamp() {
	p.ProcessCreatedAt = time.Now().Format("2006-01-02 15:04:05")
}

// GetDisplayName 표시용 이름을 반환합니다. (이름 + 버전)
func (p *ProcessInfo) GetDisplayName() string {
	return p.ProcessName + " " + p.ProcessVersion
}

// Clone 프로그램 정보의 복사본을 반환합니다.
func (p *ProcessInfo) Clone() *ProcessInfo {
	return &ProcessInfo{
		ID:                p.ID,
		ProcessName:       p.ProcessName,
		ProcessFilePath:   p.ProcessFilePath,
		ProcessUploadPath: p.ProcessUploadPath,
		ProcessVersion:    p.ProcessVersion,
		ProcessCreatedAt:  p.ProcessCreatedAt,
	}
}
