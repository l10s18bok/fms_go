// Package usecase는 비즈니스 로직을 구현합니다.
package usecase

import (
	"fmt"
	"path/filepath"
	"time"

	"fms/internal/infrastructure/ssh"
	"fms/internal/model"
	"fms/internal/utils"
)

// DefaultRemotePath 프로그램 업로드 기본 경로
const DefaultRemotePath = "/download/"

// UpdateUseCase 프로그램 업데이트 유스케이스
type UpdateUseCase struct {
	sshClientFactory  func() ssh.SSHClient
	sftpClientFactory func() ssh.SFTPClient
}

// NewUpdateUseCase 새로운 업데이트 유스케이스를 생성합니다.
func NewUpdateUseCase() *UpdateUseCase {
	return &UpdateUseCase{
		sshClientFactory:  ssh.NewSSHClient,
		sftpClientFactory: ssh.NewSFTPClient,
	}
}

// UpdateResult 업데이트 결과
type UpdateResult struct {
	Success bool
	Message string
	History *model.DeployHistory
}

// UpdateProgram 단일 장비에 프로그램을 업데이트합니다.
func (u *UpdateUseCase) UpdateProgram(
	device *model.Firewall,
	program *model.ProcessInfo,
) *UpdateResult {
	result := &UpdateResult{
		Success: false,
		History: model.NewProgramUpdateHistory(device.DeviceName, device.DeviceIP, program.ProcessName, program.ProcessVersion),
	}

	// SSH 인증 정보 확인
	if device.DeviceID == "" {
		result.Message = "SSH 사용자 ID가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	// 로컬 파일 확인
	if program.ProcessFilePath == "" {
		result.Message = "프로그램 파일 경로가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	// SSH 클라이언트 생성
	sshClient := u.sshClientFactory()
	defer sshClient.Close()

	// SSH 연결
	var err error
	if device.DevicePPK != "" {
		// 키 인증
		err = sshClient.ConnectWithKey(device.DeviceName, 22, device.DeviceID, device.DevicePPK)
	} else if device.DevicePW != "" {
		// 비밀번호 인증
		err = sshClient.Connect(device.DeviceName, 22, device.DeviceID, device.DevicePW)
	} else {
		result.Message = "SSH 인증 정보(비밀번호 또는 키)가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	if err != nil {
		result.Message = fmt.Sprintf("SSH 연결 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	// SFTP 클라이언트 생성 및 연결
	sftpClient := u.sftpClientFactory()
	defer sftpClient.Close()

	if err := sftpClient.Connect(sshClient); err != nil {
		result.Message = fmt.Sprintf("SFTP 연결 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	// 원격 경로 결정
	remotePath := program.ProcessUploadPath
	if remotePath == "" {
		remotePath = DefaultRemotePath
	}
	// 파일명 추가
	remoteFilePath := filepath.Join(remotePath, filepath.Base(program.ProcessFilePath))
	// Linux 경로로 변환
	remoteFilePath = filepath.ToSlash(remoteFilePath)

	// 파일 업로드
	if err := sftpClient.Upload(program.ProcessFilePath, remoteFilePath); err != nil {
		result.Message = fmt.Sprintf("파일 업로드 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		return result
	}

	// 장비의 프로그램 버전 업데이트
	if device.ProgramVersions == nil {
		device.ProgramVersions = make(map[string]string)
	}
	device.ProgramVersions[program.ProcessName] = program.ProcessVersion

	// 성공
	result.Success = true
	result.Message = fmt.Sprintf("프로그램 업데이트 완료: %s → %s", remoteFilePath, program.ProcessVersion)
	result.History.Status = model.DeployStatusSuccess
	result.History.Message = result.Message
	result.History.Timestamp = utils.Now()

	return result
}

// UpdateProgramBatch 여러 장비에 프로그램을 업데이트합니다.
func (u *UpdateUseCase) UpdateProgramBatch(
	devices []*model.Firewall,
	program *model.ProcessInfo,
	onProgress func(current, total int, deviceIP string, success bool),
) []*UpdateResult {
	results := make([]*UpdateResult, len(devices))
	total := len(devices)

	for i, device := range devices {
		result := u.UpdateProgram(device, program)
		results[i] = result

		if onProgress != nil {
			onProgress(i+1, total, device.DeviceName, result.Success)
		}

		// 연속 요청 시 약간의 딜레이
		if i < total-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return results
}
