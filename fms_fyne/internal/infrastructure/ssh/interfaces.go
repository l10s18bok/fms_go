// Package ssh는 SSH 및 SFTP 연결을 위한 인프라를 제공합니다.
package ssh

import "io"

// SSHClient SSH 클라이언트 인터페이스
type SSHClient interface {
	// Connect SSH 연결을 수립합니다.
	Connect(host string, port int, user, password string) error

	// ConnectWithKey PPK/PEM 키 파일을 사용하여 SSH 연결을 수립합니다.
	ConnectWithKey(host string, port int, user, keyPath string) error

	// Close SSH 연결을 닫습니다.
	Close() error

	// RunCommand 원격 명령을 실행하고 결과를 반환합니다.
	RunCommand(cmd string) (string, error)

	// IsConnected 연결 상태를 확인합니다.
	IsConnected() bool
}

// SFTPClient SFTP 클라이언트 인터페이스
type SFTPClient interface {
	// Connect SFTP 연결을 수립합니다 (SSH 연결 위에)
	Connect(sshClient SSHClient) error

	// Close SFTP 연결을 닫습니다.
	Close() error

	// Upload 로컬 파일을 원격 서버에 업로드합니다.
	Upload(localPath, remotePath string) error

	// UploadFromReader Reader에서 읽어서 원격 서버에 업로드합니다.
	UploadFromReader(reader io.Reader, remotePath string, size int64) error

	// Download 원격 파일을 로컬로 다운로드합니다.
	Download(remotePath, localPath string) error

	// Stat 원격 파일 정보를 가져옵니다.
	Stat(remotePath string) (FileInfo, error)

	// MkdirAll 원격 디렉토리를 생성합니다 (상위 디렉토리 포함).
	MkdirAll(remotePath string) error

	// Remove 원격 파일을 삭제합니다.
	Remove(remotePath string) error
}

// FileInfo 파일 정보 인터페이스
type FileInfo interface {
	Name() string
	Size() int64
	IsDir() bool
}

// SSHConfig SSH 연결 설정
type SSHConfig struct {
	Host     string // 호스트 주소
	Port     int    // SSH 포트 (기본값: 22)
	User     string // 사용자 ID
	Password string // 비밀번호 (비밀번호 인증 시)
	KeyPath  string // PPK/PEM 키 파일 경로 (키 인증 시)
}

// AuthType 인증 유형
type AuthType string

const (
	AuthTypePassword AuthType = "password" // 비밀번호 인증
	AuthTypeKey      AuthType = "key"      // 키 인증
)

// GetAuthType 설정에서 인증 유형을 결정합니다.
func (c *SSHConfig) GetAuthType() AuthType {
	if c.KeyPath != "" {
		return AuthTypeKey
	}
	return AuthTypePassword
}
