package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
)

// DefaultSFTPClient SFTP 클라이언트 구현
type DefaultSFTPClient struct {
	client *sftp.Client
}

// NewSFTPClient 새로운 SFTP 클라이언트를 생성합니다.
func NewSFTPClient() SFTPClient {
	return &DefaultSFTPClient{}
}

// Connect SSH 연결 위에 SFTP 세션을 수립합니다.
func (c *DefaultSFTPClient) Connect(sshClient SSHClient) error {
	// DefaultSSHClient에서 내부 클라이언트 가져오기
	defaultClient, ok := sshClient.(*DefaultSSHClient)
	if !ok {
		return fmt.Errorf("지원되지 않는 SSH 클라이언트 타입입니다")
	}

	underlyingClient := defaultClient.GetUnderlyingClient()
	if underlyingClient == nil {
		return fmt.Errorf("SSH 연결이 없습니다")
	}

	sftpClient, err := sftp.NewClient(underlyingClient)
	if err != nil {
		return fmt.Errorf("SFTP 클라이언트 생성 실패: %w", err)
	}

	c.client = sftpClient
	return nil
}

// Close SFTP 연결을 닫습니다.
func (c *DefaultSFTPClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Upload 로컬 파일을 원격 서버에 업로드합니다.
func (c *DefaultSFTPClient) Upload(localPath, remotePath string) error {
	if c.client == nil {
		return fmt.Errorf("SFTP 연결이 없습니다")
	}

	// 로컬 파일 열기
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("로컬 파일 열기 실패: %w", err)
	}
	defer localFile.Close()

	// 파일 정보 가져오기
	localInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("로컬 파일 정보 가져오기 실패: %w", err)
	}

	// 원격 디렉토리 생성 (Linux 경로이므로 path.Dir 사용)
	remoteDir := path.Dir(remotePath)
	if err := c.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("원격 디렉토리 생성 실패: %w", err)
	}

	// 원격 파일 생성
	remoteFile, err := c.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("원격 파일 생성 실패: %w", err)
	}
	defer remoteFile.Close()

	// 파일 복사
	written, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("파일 업로드 실패: %w", err)
	}

	if written != localInfo.Size() {
		return fmt.Errorf("파일 크기 불일치: 예상 %d, 실제 %d", localInfo.Size(), written)
	}

	return nil
}

// UploadFromReader Reader에서 읽어서 원격 서버에 업로드합니다.
func (c *DefaultSFTPClient) UploadFromReader(reader io.Reader, remotePath string, size int64) error {
	if c.client == nil {
		return fmt.Errorf("SFTP 연결이 없습니다")
	}

	// 원격 디렉토리 생성 (Linux 경로이므로 path.Dir 사용)
	remoteDir := path.Dir(remotePath)
	if err := c.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("원격 디렉토리 생성 실패: %w", err)
	}

	// 원격 파일 생성
	remoteFile, err := c.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("원격 파일 생성 실패: %w", err)
	}
	defer remoteFile.Close()

	// 파일 복사
	written, err := io.Copy(remoteFile, reader)
	if err != nil {
		return fmt.Errorf("파일 업로드 실패: %w", err)
	}

	if size > 0 && written != size {
		return fmt.Errorf("파일 크기 불일치: 예상 %d, 실제 %d", size, written)
	}

	return nil
}

// Download 원격 파일을 로컬로 다운로드합니다.
func (c *DefaultSFTPClient) Download(remotePath, localPath string) error {
	if c.client == nil {
		return fmt.Errorf("SFTP 연결이 없습니다")
	}

	// 원격 파일 열기
	remoteFile, err := c.client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("원격 파일 열기 실패: %w", err)
	}
	defer remoteFile.Close()

	// 로컬 디렉토리 생성
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("로컬 디렉토리 생성 실패: %w", err)
	}

	// 로컬 파일 생성
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("로컬 파일 생성 실패: %w", err)
	}
	defer localFile.Close()

	// 파일 복사
	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("파일 다운로드 실패: %w", err)
	}

	return nil
}

// Stat 원격 파일 정보를 가져옵니다.
func (c *DefaultSFTPClient) Stat(remotePath string) (FileInfo, error) {
	if c.client == nil {
		return nil, fmt.Errorf("SFTP 연결이 없습니다")
	}

	info, err := c.client.Stat(remotePath)
	if err != nil {
		return nil, fmt.Errorf("파일 정보 가져오기 실패: %w", err)
	}

	return &sftpFileInfo{info: info}, nil
}

// MkdirAll 원격 디렉토리를 생성합니다 (상위 디렉토리 포함).
func (c *DefaultSFTPClient) MkdirAll(remotePath string) error {
	if c.client == nil {
		return fmt.Errorf("SFTP 연결이 없습니다")
	}

	// 디렉토리가 이미 존재하는지 확인
	if _, err := c.client.Stat(remotePath); err == nil {
		return nil
	}

	return c.client.MkdirAll(remotePath)
}

// Remove 원격 파일을 삭제합니다.
func (c *DefaultSFTPClient) Remove(remotePath string) error {
	if c.client == nil {
		return fmt.Errorf("SFTP 연결이 없습니다")
	}

	return c.client.Remove(remotePath)
}

// sftpFileInfo FileInfo 인터페이스 구현
type sftpFileInfo struct {
	info os.FileInfo
}

func (f *sftpFileInfo) Name() string { return f.info.Name() }
func (f *sftpFileInfo) Size() int64  { return f.info.Size() }
func (f *sftpFileInfo) IsDir() bool  { return f.info.IsDir() }
