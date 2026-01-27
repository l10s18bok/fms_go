// Package storage는 데이터 저장소를 관리합니다.
package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// MigrationResult 마이그레이션 결과
type MigrationResult struct {
	Success      bool
	TotalCount   int
	MigratedCount int
	ErrorCount   int
	Errors       []string
}

// MigrateHistoryToSQLite JSON 배포 이력을 SQLite로 마이그레이션합니다.
func MigrateHistoryToSQLite(jsonStore *JSONStore, sqliteStore *SQLiteStore) (*MigrationResult, error) {
	result := &MigrationResult{
		Success: false,
		Errors:  []string{},
	}

	// JSON에서 모든 이력 조회
	histories, err := jsonStore.GetAllHistory()
	if err != nil {
		return result, fmt.Errorf("JSON 이력 조회 실패: %w", err)
	}

	result.TotalCount = len(histories)

	if result.TotalCount == 0 {
		result.Success = true
		log.Println("[마이그레이션] 마이그레이션할 이력이 없습니다.")
		return result, nil
	}

	log.Printf("[마이그레이션] %d개의 이력을 마이그레이션합니다...\n", result.TotalCount)

	// 각 이력을 SQLite에 저장
	for _, h := range histories {
		// ID를 초기화하여 새로운 ID가 할당되도록 함
		originalID := h.ID
		h.ID = 0

		if err := sqliteStore.SaveHistory(h); err != nil {
			errMsg := fmt.Sprintf("이력 #%d 마이그레이션 실패: %v", originalID, err)
			result.Errors = append(result.Errors, errMsg)
			result.ErrorCount++
			log.Println("[마이그레이션] " + errMsg)
			continue
		}

		result.MigratedCount++
	}

	result.Success = result.ErrorCount == 0
	log.Printf("[마이그레이션] 완료: 총 %d개 중 %d개 성공, %d개 실패\n",
		result.TotalCount, result.MigratedCount, result.ErrorCount)

	return result, nil
}

// NeedsMigration 마이그레이션이 필요한지 확인합니다.
// history.json 파일이 존재하고 데이터가 있으며, SQLite DB가 비어있는 경우 true 반환
func NeedsMigration(configDir string, jsonStore *JSONStore, sqliteStore *SQLiteStore) bool {
	// history.json 파일 존재 여부 확인
	historyPath := filepath.Join(configDir, "history.json")
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return false // JSON 파일이 없으면 마이그레이션 불필요
	}

	// JSON에 데이터가 있는지 확인
	jsonCount := jsonStore.CountHistory()
	if jsonCount == 0 {
		return false // JSON에 데이터가 없으면 마이그레이션 불필요
	}

	// SQLite에 데이터가 있는지 확인
	sqliteCount := sqliteStore.CountHistory()
	if sqliteCount > 0 {
		return false // SQLite에 이미 데이터가 있으면 마이그레이션 불필요
	}

	return true // 마이그레이션 필요
}

// BackupHistoryJSON history.json 파일을 백업합니다.
func BackupHistoryJSON(configDir string) error {
	srcPath := filepath.Join(configDir, "history.json")
	dstPath := filepath.Join(configDir, "history.json.backup")

	// 파일 존재 여부 확인
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil // 파일이 없으면 백업할 것 없음
	}

	// 파일 복사
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("history.json 읽기 실패: %w", err)
	}

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return fmt.Errorf("백업 파일 쓰기 실패: %w", err)
	}

	log.Printf("[마이그레이션] history.json 백업 완료: %s\n", dstPath)
	return nil
}
