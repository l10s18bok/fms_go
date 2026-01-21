// Package usecase는 비즈니스 로직을 구현합니다.
package usecase

import (
	"fmt"
	"log"
	"net"
	"path/filepath"
	"time"

	"fms/internal/http"
	"fms/internal/infrastructure/ssh"
	"fms/internal/model"
	"fms/internal/utils"
)

// UpdateUseCase 패키지 업데이트 유스케이스
type UpdateUseCase struct {
	sshClientFactory  func() ssh.SSHClient
	sftpClientFactory func() ssh.SFTPClient
	config            *model.Config
}

// NewUpdateUseCase 새로운 업데이트 유스케이스를 생성합니다.
func NewUpdateUseCase(config *model.Config) *UpdateUseCase {
	return &UpdateUseCase{
		sshClientFactory:  ssh.NewSSHClient,
		sftpClientFactory: ssh.NewSFTPClient,
		config:            config,
	}
}

// UpdateResult 업데이트 결과
type UpdateResult struct {
	Success bool
	Message string
	History *model.DeployHistory
}

// UpdateProgram 단일 장비에 패키지을 업데이트합니다.
func (u *UpdateUseCase) UpdateProgram(
	device *model.Firewall,
	program *model.ProcessInfo,
) *UpdateResult {
	deviceIP := device.DeviceIP
	if deviceIP == "" {
		deviceIP = device.DeviceName
	}
	log.Printf("[UpdateProgram] 시작 - 장비: %s, 패키지: %s %s", deviceIP, program.ProcessName, program.ProcessVersion)

	result := &UpdateResult{
		Success: false,
		History: model.NewProgramUpdateHistory(device.DeviceName, device.DeviceIP, program.ProcessName, program.ProcessVersion),
	}

	// SSH 인증 정보 확인
	if device.DeviceID == "" {
		result.Message = "SSH 사용자 ID가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}

	// 로컬 파일 확인
	if program.ProcessFilePath == "" {
		result.Message = "패키지 파일 경로가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}
	log.Printf("[UpdateProgram] 로컬 파일: %s", program.ProcessFilePath)

	// SSH 클라이언트 생성
	sshClient := u.sshClientFactory()
	defer sshClient.Close()

	// SSH 연결 (DeviceIP 사용, 없으면 DeviceName 사용)
	// DeviceIP에 포트가 포함된 경우 (예: 192.168.1.1:8090) IP만 추출
	sshHost := device.DeviceIP
	if sshHost == "" {
		sshHost = device.DeviceName
	}
	// 포트 분리 (SSH는 항상 22번 포트 사용)
	if host, _, err := net.SplitHostPort(sshHost); err == nil {
		sshHost = host
	}

	var err error
	if device.DevicePPK != "" {
		log.Printf("[UpdateProgram] SSH 키 인증 시도 - %s@%s:22", device.DeviceID, sshHost)
		err = sshClient.ConnectWithKey(sshHost, 22, device.DeviceID, device.DevicePPK)
	} else if device.DevicePW != "" {
		log.Printf("[UpdateProgram] SSH 비밀번호 인증 시도 - %s@%s:22", device.DeviceID, sshHost)
		err = sshClient.Connect(sshHost, 22, device.DeviceID, device.DevicePW)
	} else {
		result.Message = "SSH 인증 정보(비밀번호 또는 키)가 설정되지 않았습니다"
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}

	if err != nil {
		result.Message = fmt.Sprintf("SSH 연결 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}
	log.Printf("[UpdateProgram] SSH 연결 성공 - %s", sshHost)

	// SFTP 클라이언트 생성 및 연결
	sftpClient := u.sftpClientFactory()
	defer sftpClient.Close()

	log.Printf("[UpdateProgram] SFTP 연결 시도 - %s", sshHost)
	if err := sftpClient.Connect(sshClient); err != nil {
		result.Message = fmt.Sprintf("SFTP 연결 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}
	log.Printf("[UpdateProgram] SFTP 연결 성공 - %s", sshHost)

	// 원격 경로 결정
	remotePath := program.ProcessUploadPath
	if remotePath == "" {
		remotePath = model.DefaultRemotePath
	}
	// 파일명 추가
	remoteFilePath := filepath.Join(remotePath, filepath.Base(program.ProcessFilePath))
	// Linux 경로로 변환
	remoteFilePath = filepath.ToSlash(remoteFilePath)

	// 파일 업로드
	log.Printf("[UpdateProgram] 파일 업로드 시도 - %s -> %s:%s", program.ProcessFilePath, sshHost, remoteFilePath)
	if err := sftpClient.Upload(program.ProcessFilePath, remoteFilePath); err != nil {
		result.Message = fmt.Sprintf("파일 업로드 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", sshHost, result.Message)
		return result
	}
	log.Printf("[UpdateProgram] 파일 업로드 성공 - %s:%s", sshHost, remoteFilePath)

	// 장비에 업데이트 요청 API 호출
	httpClient := http.NewClient(u.config)
	updateReq := &http.ProgramUpdateRequest{
		FilePath:       remoteFilePath,
		ProcessName:    program.ProcessName,
		ProcessVersion: program.ProcessVersion,
	}

	log.Printf("[UpdateProgram] 업데이트 요청 시도 - %s", deviceIP)
	if err := httpClient.ProgramUpdateDirect(deviceIP, updateReq); err != nil {
		result.Message = fmt.Sprintf("업데이트 요청 실패: %v", err)
		result.History.Status = model.DeployStatusFail
		result.History.Message = result.Message
		log.Printf("[UpdateProgram] 실패 - %s: %s", deviceIP, result.Message)
		return result
	}
	log.Printf("[UpdateProgram] 업데이트 요청 성공 - %s", deviceIP)

	// 장비 정보 조회로 업데이트 확인
	log.Printf("[UpdateProgram] 장비 정보 조회 시도 - %s", deviceIP)
	report, err := httpClient.GetDeviceReportDirect(deviceIP)
	if err != nil {
		// 장비 정보 조회 실패해도 업데이트는 성공으로 처리 (경고만 출력)
		log.Printf("[UpdateProgram] 경고 - %s: 장비 정보 조회 실패: %v", deviceIP, err)
	} else {
		// 장비에서 반환한 버전 확인
		for _, proc := range report.Processes {
			if proc.Name == program.ProcessName {
				log.Printf("[UpdateProgram] 장비 버전 확인 - %s: %s=%s", deviceIP, proc.Name, proc.Version)
				if proc.Version != program.ProcessVersion {
					log.Printf("[UpdateProgram] 경고 - %s: 버전 불일치 (요청: %s, 실제: %s)", deviceIP, program.ProcessVersion, proc.Version)
				}
				break
			}
		}
	}

	// 장비의 패키지 버전 업데이트
	if device.ProgramVersions == nil {
		device.ProgramVersions = make(map[string]string)
	}
	device.ProgramVersions[program.ProcessName] = program.ProcessVersion

	// 성공
	result.Success = true
	result.Message = fmt.Sprintf("패키지 업데이트 완료: %s → %s", remoteFilePath, program.ProcessVersion)
	result.History.Status = model.DeployStatusSuccess
	result.History.Message = result.Message
	result.History.Timestamp = utils.Now()

	log.Printf("[UpdateProgram] 완료 - %s: %s", deviceIP, result.Message)
	return result
}

// UpdateProgramBatch 여러 장비에 패키지을 업데이트합니다.
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
