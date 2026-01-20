package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"fms/internal/model"
)

// FileStore data 디렉토리의 파일들을 관리합니다.
type FileStore struct {
	dataDir string
	mu      sync.RWMutex
}

// NewFileStore 새로운 FileStore를 생성합니다.
func NewFileStore(baseDir string) (*FileStore, error) {
	dataDir := filepath.Join(baseDir, "data")

	// data 디렉토리 생성
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("data 디렉토리 생성 실패: %v", err)
	}

	return &FileStore{dataDir: dataDir}, nil
}

// GetAllFiles 모든 방화벽 파일 목록을 반환합니다.
func (s *FileStore) GetAllFiles() ([]*model.FirewallFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, err
	}

	var files []*model.FirewallFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := &model.FirewallFile{
			FileName:   entry.Name(),
			FilePath:   filepath.Join(s.dataDir, entry.Name()),
			CreatedAt:  info.ModTime(), // Windows에서는 생성일 접근이 복잡하므로 수정일 사용
			ModifiedAt: info.ModTime(),
			Version:    model.ExtractVersion(entry.Name()),
		}
		files = append(files, file)
	}

	// 수정일 내림차순 정렬 (최신순)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt.After(files[j].ModifiedAt)
	})

	return files, nil
}

// GetFile 특정 파일을 읽어서 반환합니다.
func (s *FileStore) GetFile(fileName string) (*model.FirewallFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filePath := filepath.Join(s.dataDir, fileName)

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일을 찾을 수 없습니다: %s", fileName)
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &model.FirewallFile{
		FileName:   fileName,
		FilePath:   filePath,
		CreatedAt:  info.ModTime(),
		ModifiedAt: info.ModTime(),
		Version:    model.ExtractVersion(fileName),
		Contents:   string(contents),
	}, nil
}

// SaveFile 파일을 저장합니다.
func (s *FileStore) SaveFile(file *model.FirewallFile) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.dataDir, file.FileName)
	return os.WriteFile(filePath, []byte(file.Contents), 0644)
}

// DeleteFile 파일을 삭제합니다.
func (s *FileStore) DeleteFile(fileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.dataDir, fileName)
	return os.Remove(filePath)
}

// DeleteFiles 여러 파일을 삭제합니다.
func (s *FileStore) DeleteFiles(fileNames []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error
	for _, fileName := range fileNames {
		filePath := filepath.Join(s.dataDir, fileName)
		if err := os.Remove(filePath); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// FileExists 파일 존재 여부를 확인합니다.
func (s *FileStore) FileExists(fileName string) bool {
	filePath := filepath.Join(s.dataDir, fileName)
	_, err := os.Stat(filePath)
	return err == nil
}

// GetDataDir data 디렉토리 경로를 반환합니다.
func (s *FileStore) GetDataDir() string {
	return s.dataDir
}

// CopyFileToData 외부 파일을 data 디렉토리로 복사합니다.
func (s *FileStore) CopyFileToData(srcPath string, destFileName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("원본 파일을 열 수 없습니다: %v", err)
	}
	defer srcFile.Close()

	destPath := filepath.Join(s.dataDir, destFileName)
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("대상 파일을 생성할 수 없습니다: %v", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("파일 복사 실패: %v", err)
	}

	return nil
}

// GetFileNames 모든 파일명 목록을 반환합니다. (배포 시 드롭다운용)
func (s *FileStore) GetFileNames() ([]string, error) {
	files, err := s.GetAllFiles()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.FileName
	}
	return names, nil
}

// ClearAllFiles 모든 파일을 삭제합니다. (Reset용)
func (s *FileStore) ClearAllFiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filePath := filepath.Join(s.dataDir, entry.Name())
		os.Remove(filePath)
	}

	return nil
}

// RenameFile 파일 이름을 변경합니다.
func (s *FileStore) RenameFile(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldPath := filepath.Join(s.dataDir, oldName)
	newPath := filepath.Join(s.dataDir, newName)

	return os.Rename(oldPath, newPath)
}
