package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"fms/internal/model"

	"github.com/fsnotify/fsnotify"
)

// FileStore rule_files 디렉토리의 파일들을 관리합니다.
type FileStore struct {
	dataDir   string
	mu        sync.RWMutex
	watcher   *fsnotify.Watcher
	onChange  func() // 파일 변경 시 콜백
	stopWatch chan struct{}
}

// NewFileStore 새로운 FileStore를 생성합니다.
func NewFileStore(configDir string) (*FileStore, error) {
	dataDir := filepath.Join(configDir, "rule_files")

	// rule_files 디렉토리 생성
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("rule_files 디렉토리 생성 실패: %v", err)
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

// StartWatch rule_files 디렉토리의 파일 변경을 감시합니다.
func (s *FileStore) StartWatch(onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("파일 감시 생성 실패: %w", err)
	}

	if err := watcher.Add(s.dataDir); err != nil {
		watcher.Close()
		return fmt.Errorf("디렉토리 감시 등록 실패: %w", err)
	}

	s.watcher = watcher
	s.onChange = onChange
	s.stopWatch = make(chan struct{})

	go func() {
		var debounce *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Write|fsnotify.Rename) != 0 {
					// 디바운스: 연속 이벤트를 300ms 내 하나로 처리
					if debounce != nil {
						debounce.Stop()
					}
					debounce = time.AfterFunc(300*time.Millisecond, func() {
						if s.onChange != nil {
							s.onChange()
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("파일 감시 오류: %v", err)
			case <-s.stopWatch:
				return
			}
		}
	}()

	return nil
}

// StopWatch 파일 감시를 중지합니다.
func (s *FileStore) StopWatch() {
	if s.stopWatch != nil {
		close(s.stopWatch)
	}
	if s.watcher != nil {
		s.watcher.Close()
	}
}
