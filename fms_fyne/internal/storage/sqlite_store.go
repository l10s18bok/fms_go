// Package storage는 데이터 저장소를 관리합니다.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fms/internal/model"
	"fms/internal/utils"

	_ "modernc.org/sqlite"
)

const (
	sqliteDBFile      = "fms.db"
	sqliteTimeFormat  = "2006-01-02 15:04:05"
	currentDBVersion  = 1
)

// SQLiteStore SQLite 기반 데이터 저장소
type SQLiteStore struct {
	db        *sql.DB
	configDir string
	mu        sync.RWMutex
}

// NewSQLiteStore 새로운 SQLite 저장소를 생성합니다.
func NewSQLiteStore(configDir string) (*SQLiteStore, error) {
	// 디렉토리가 없으면 생성
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("설정 디렉토리 생성 실패: %w", err)
	}

	dbPath := filepath.Join(configDir, sqliteDBFile)

	// SQLite 연결 (파일이 없으면 자동 생성)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("SQLite 연결 실패: %w", err)
	}

	// 연결 테스트
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("SQLite 연결 확인 실패: %w", err)
	}

	// 외래 키 제약 조건 활성화
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("외래 키 활성화 실패: %w", err)
	}

	store := &SQLiteStore{
		db:        db,
		configDir: configDir,
	}

	// 스키마 초기화
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("스키마 초기화 실패: %w", err)
	}

	return store, nil
}

// Close 데이터베이스 연결을 종료합니다.
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB 데이터베이스 연결을 반환합니다. (Repository에서 사용)
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

// initSchema 데이터베이스 스키마를 초기화합니다.
func (s *SQLiteStore) initSchema() error {
	// 배포 이력 테이블
	createDeployHistoryTable := `
	CREATE TABLE IF NOT EXISTS deploy_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT,
		timestamp DATETIME,
		device_name TEXT,
		device_ip TEXT,
		template_version TEXT,
		program_name TEXT,
		program_version TEXT,
		message TEXT,
		status TEXT
	);`

	// 규칙 결과 테이블
	createRuleResultsTable := `
	CREATE TABLE IF NOT EXISTS rule_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		history_id INTEGER NOT NULL,
		rule TEXT,
		text TEXT,
		status TEXT,
		reason TEXT,
		FOREIGN KEY (history_id) REFERENCES deploy_history(id) ON DELETE CASCADE
	);`

	// 인덱스 생성
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_history_type ON deploy_history(type);",
		"CREATE INDEX IF NOT EXISTS idx_history_device_ip ON deploy_history(device_ip);",
		"CREATE INDEX IF NOT EXISTS idx_history_timestamp ON deploy_history(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_rule_results_history_id ON rule_results(history_id);",
	}

	// 테이블 생성
	if _, err := s.db.Exec(createDeployHistoryTable); err != nil {
		return fmt.Errorf("deploy_history 테이블 생성 실패: %w", err)
	}

	if _, err := s.db.Exec(createRuleResultsTable); err != nil {
		return fmt.Errorf("rule_results 테이블 생성 실패: %w", err)
	}

	// 인덱스 생성
	for _, idx := range createIndexes {
		if _, err := s.db.Exec(idx); err != nil {
			return fmt.Errorf("인덱스 생성 실패: %w", err)
		}
	}

	return nil
}

// ========== 배포 이력 관련 메서드 ==========

// GetAllHistory 모든 배포 이력을 조회합니다.
func (s *SQLiteStore) GetAllHistory() ([]*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, type, timestamp, device_name, device_ip,
		       template_version, program_name, program_version, message, status
		FROM deploy_history
		ORDER BY timestamp DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("이력 조회 실패: %w", err)
	}
	defer rows.Close()

	var histories []*model.DeployHistory
	for rows.Next() {
		h, err := s.scanHistory(rows)
		if err != nil {
			return nil, err
		}

		// 규칙 결과 조회
		results, err := s.getRuleResults(h.ID)
		if err != nil {
			return nil, err
		}
		h.Results = results

		histories = append(histories, h)
	}

	return histories, nil
}

// GetHistory ID로 배포 이력을 조회합니다.
func (s *SQLiteStore) GetHistory(id int) (*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, type, timestamp, device_name, device_ip,
		       template_version, program_name, program_version, message, status
		FROM deploy_history
		WHERE id = ?`

	row := s.db.QueryRow(query, id)
	h, err := s.scanHistoryRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("이력을 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("이력 조회 실패: %w", err)
	}

	// 규칙 결과 조회
	results, err := s.getRuleResults(h.ID)
	if err != nil {
		return nil, err
	}
	h.Results = results

	return h, nil
}

// GetHistoryByType 유형별 배포 이력을 조회합니다.
func (s *SQLiteStore) GetHistoryByType(historyType string) ([]*model.DeployHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, type, timestamp, device_name, device_ip,
		       template_version, program_name, program_version, message, status
		FROM deploy_history
		WHERE type = ?
		ORDER BY timestamp DESC`

	rows, err := s.db.Query(query, historyType)
	if err != nil {
		return nil, fmt.Errorf("이력 조회 실패: %w", err)
	}
	defer rows.Close()

	var histories []*model.DeployHistory
	for rows.Next() {
		h, err := s.scanHistory(rows)
		if err != nil {
			return nil, err
		}

		// 규칙 결과 조회
		results, err := s.getRuleResults(h.ID)
		if err != nil {
			return nil, err
		}
		h.Results = results

		histories = append(histories, h)
	}

	return histories, nil
}

// SaveHistory 배포 이력을 저장합니다.
func (s *SQLiteStore) SaveHistory(h *model.DeployHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	timestamp := h.Timestamp.Time().Format(sqliteTimeFormat)

	if h.ID == 0 {
		// 신규 삽입
		result, err := tx.Exec(`
			INSERT INTO deploy_history
			(type, timestamp, device_name, device_ip, template_version,
			 program_name, program_version, message, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			h.Type, timestamp, h.DeviceName, h.DeviceIP, h.TemplateVer,
			h.ProgramName, h.ProgramVer, h.Message, h.Status)
		if err != nil {
			return fmt.Errorf("이력 삽입 실패: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("ID 조회 실패: %w", err)
		}
		h.ID = int(id)
	} else {
		// 기존 업데이트
		_, err := tx.Exec(`
			UPDATE deploy_history SET
				type = ?, timestamp = ?, device_name = ?, device_ip = ?,
				template_version = ?, program_name = ?, program_version = ?,
				message = ?, status = ?
			WHERE id = ?`,
			h.Type, timestamp, h.DeviceName, h.DeviceIP, h.TemplateVer,
			h.ProgramName, h.ProgramVer, h.Message, h.Status, h.ID)
		if err != nil {
			return fmt.Errorf("이력 업데이트 실패: %w", err)
		}

		// 기존 규칙 결과 삭제
		if _, err := tx.Exec("DELETE FROM rule_results WHERE history_id = ?", h.ID); err != nil {
			return fmt.Errorf("규칙 결과 삭제 실패: %w", err)
		}
	}

	// 규칙 결과 삽입
	for _, r := range h.Results {
		_, err := tx.Exec(`
			INSERT INTO rule_results (history_id, rule, text, status, reason)
			VALUES (?, ?, ?, ?, ?)`,
			h.ID, r.Rule, r.Text, r.Status, r.Reason)
		if err != nil {
			return fmt.Errorf("규칙 결과 삽입 실패: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	return nil
}

// DeleteHistory 배포 이력을 삭제합니다.
func (s *SQLiteStore) DeleteHistory(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// CASCADE 설정으로 rule_results도 자동 삭제됨
	result, err := s.db.Exec("DELETE FROM deploy_history WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("이력 삭제 실패: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("삭제 결과 확인 실패: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("이력을 찾을 수 없습니다: %d", id)
	}

	return nil
}

// ClearHistory 모든 배포 이력을 삭제합니다.
func (s *SQLiteStore) ClearHistory() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// CASCADE 설정으로 rule_results도 자동 삭제됨
	if _, err := s.db.Exec("DELETE FROM deploy_history"); err != nil {
		return fmt.Errorf("이력 전체 삭제 실패: %w", err)
	}

	return nil
}

// CountHistory 배포 이력 수를 반환합니다.
func (s *SQLiteStore) CountHistory() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM deploy_history").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

// ========== 내부 헬퍼 메서드 ==========

// scanHistory rows에서 DeployHistory를 스캔합니다.
func (s *SQLiteStore) scanHistory(rows *sql.Rows) (*model.DeployHistory, error) {
	var h model.DeployHistory
	var timestamp string
	var deviceName, programName, programVer, message sql.NullString

	err := rows.Scan(
		&h.ID, &h.Type, &timestamp, &deviceName, &h.DeviceIP,
		&h.TemplateVer, &programName, &programVer, &message, &h.Status)
	if err != nil {
		return nil, fmt.Errorf("이력 스캔 실패: %w", err)
	}

	// 시간 파싱 (여러 포맷 지원)
	t, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, fmt.Errorf("시간 파싱 실패: %w", err)
	}
	h.Timestamp = utils.JSONTime(t)

	// NULL 처리
	if deviceName.Valid {
		h.DeviceName = deviceName.String
	}
	if programName.Valid {
		h.ProgramName = programName.String
	}
	if programVer.Valid {
		h.ProgramVer = programVer.String
	}
	if message.Valid {
		h.Message = message.String
	}

	return &h, nil
}

// scanHistoryRow QueryRow에서 DeployHistory를 스캔합니다.
func (s *SQLiteStore) scanHistoryRow(row *sql.Row) (*model.DeployHistory, error) {
	var h model.DeployHistory
	var timestamp string
	var deviceName, programName, programVer, message sql.NullString

	err := row.Scan(
		&h.ID, &h.Type, &timestamp, &deviceName, &h.DeviceIP,
		&h.TemplateVer, &programName, &programVer, &message, &h.Status)
	if err != nil {
		return nil, err
	}

	// 시간 파싱 (여러 포맷 지원)
	t, err := parseTimestamp(timestamp)
	if err != nil {
		return nil, fmt.Errorf("시간 파싱 실패: %w", err)
	}
	h.Timestamp = utils.JSONTime(t)

	// NULL 처리
	if deviceName.Valid {
		h.DeviceName = deviceName.String
	}
	if programName.Valid {
		h.ProgramName = programName.String
	}
	if programVer.Valid {
		h.ProgramVer = programVer.String
	}
	if message.Valid {
		h.Message = message.String
	}

	return &h, nil
}

// getRuleResults 특정 이력의 규칙 결과를 조회합니다.
func (s *SQLiteStore) getRuleResults(historyID int) ([]model.RuleResult, error) {
	rows, err := s.db.Query(`
		SELECT rule, text, status, reason
		FROM rule_results
		WHERE history_id = ?
		ORDER BY id`, historyID)
	if err != nil {
		return nil, fmt.Errorf("규칙 결과 조회 실패: %w", err)
	}
	defer rows.Close()

	var results []model.RuleResult
	for rows.Next() {
		var r model.RuleResult
		var text, reason sql.NullString

		if err := rows.Scan(&r.Rule, &text, &r.Status, &reason); err != nil {
			return nil, fmt.Errorf("규칙 결과 스캔 실패: %w", err)
		}

		if text.Valid {
			r.Text = text.String
		}
		if reason.Valid {
			r.Reason = reason.String
		}

		results = append(results, r)
	}

	return results, nil
}

// parseTimestamp 여러 시간 포맷을 지원하는 파싱 함수
func parseTimestamp(timestamp string) (time.Time, error) {
	// 1. SQLite 기본 포맷 (2006-01-02 15:04:05)
	if t, err := time.ParseInLocation(sqliteTimeFormat, timestamp, time.Local); err == nil {
		return t, nil
	}

	// 2. RFC3339 포맷 (2006-01-02T15:04:05Z)
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t, nil
	}

	// 3. RFC3339 with timezone (2006-01-02T15:04:05+09:00)
	if t, err := time.Parse(time.RFC3339Nano, timestamp); err == nil {
		return t, nil
	}

	// 4. 초 없는 포맷 (2006-01-02 15:04)
	if t, err := time.ParseInLocation("2006-01-02 15:04", timestamp, time.Local); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("지원하지 않는 시간 포맷: %s", timestamp)
}
