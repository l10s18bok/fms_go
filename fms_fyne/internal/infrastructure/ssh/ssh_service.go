package ssh

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// DefaultSSHClient SSH 클라이언트 구현
type DefaultSSHClient struct {
	client  *ssh.Client
	config  *SSHConfig
	timeout time.Duration
}

// NewSSHClient 새로운 SSH 클라이언트를 생성합니다.
func NewSSHClient() SSHClient {
	return &DefaultSSHClient{
		timeout: 30 * time.Second,
	}
}

// NewSSHClientWithTimeout 타임아웃을 지정하여 새로운 SSH 클라이언트를 생성합니다.
func NewSSHClientWithTimeout(timeout time.Duration) SSHClient {
	return &DefaultSSHClient{
		timeout: timeout,
	}
}

// Connect 비밀번호를 사용하여 SSH 연결을 수립합니다.
func (c *DefaultSSHClient) Connect(host string, port int, user, password string) error {
	if port == 0 {
		port = 22
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 프로덕션에서는 적절한 호스트 키 검증 필요
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH 연결 실패: %w", err)
	}

	c.client = client
	c.config = &SSHConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	}

	return nil
}

// ConnectWithKey PPK/PEM 키 파일을 사용하여 SSH 연결을 수립합니다.
func (c *DefaultSSHClient) ConnectWithKey(host string, port int, user, keyPath string) error {
	if port == 0 {
		port = 22
	}

	// 키 파일 읽기
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("키 파일 읽기 실패: %w", err)
	}

	// PPK 형식 감지 및 변환
	if isPPKFormat(keyData) {
		keyData, err = convertPPKToOpenSSH(keyData)
		if err != nil {
			return fmt.Errorf("PPK 변환 실패: %w", err)
		}
	}

	// 개인 키 파싱
	signer, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return fmt.Errorf("개인 키 파싱 실패: %w", err)
	}

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("SSH 연결 실패: %w", err)
	}

	c.client = client
	c.config = &SSHConfig{
		Host:    host,
		Port:    port,
		User:    user,
		KeyPath: keyPath,
	}

	return nil
}

// Close SSH 연결을 닫습니다.
func (c *DefaultSSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// RunCommand 원격 명령을 실행하고 결과를 반환합니다.
func (c *DefaultSSHClient) RunCommand(cmd string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("SSH 연결이 없습니다")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("세션 생성 실패: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(cmd)
	if err != nil {
		// stderr가 있으면 함께 반환
		if stderr.Len() > 0 {
			return "", fmt.Errorf("명령 실행 실패: %s", stderr.String())
		}
		return "", fmt.Errorf("명령 실행 실패: %w", err)
	}

	return stdout.String(), nil
}

// IsConnected 연결 상태를 확인합니다.
func (c *DefaultSSHClient) IsConnected() bool {
	return c.client != nil
}

// GetUnderlyingClient 내부 ssh.Client를 반환합니다 (SFTP 연결용).
func (c *DefaultSSHClient) GetUnderlyingClient() *ssh.Client {
	return c.client
}

// isPPKFormat PPK 형식인지 확인합니다.
func isPPKFormat(data []byte) bool {
	return strings.HasPrefix(string(data), "PuTTY-User-Key-File")
}

// convertPPKToOpenSSH PPK 형식을 OpenSSH 형식으로 변환합니다.
// 참고: 실제 PPK 변환은 복잡하므로 기본 구현만 제공합니다.
// 프로덕션에서는 외부 라이브러리 사용을 권장합니다.
func convertPPKToOpenSSH(ppkData []byte) ([]byte, error) {
	// PPK 형식은 복잡하므로 OpenSSH 형식(.pem)을 사용하도록 안내
	return nil, fmt.Errorf("PPK 형식은 지원되지 않습니다. OpenSSH 형식(.pem)으로 변환하여 사용해주세요. " +
		"PuTTYgen에서 Conversions > Export OpenSSH key로 변환할 수 있습니다")
}
