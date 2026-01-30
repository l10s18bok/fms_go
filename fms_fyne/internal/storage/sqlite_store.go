// Package storage는 데이터 저장소를 관리합니다.
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	// 장비 테이블
	createFirewallsTable := `
	CREATE TABLE IF NOT EXISTS firewalls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_name TEXT NOT NULL,
		device_ip TEXT,
		server_status TEXT DEFAULT '-',
		deploy_status TEXT DEFAULT '-',
		version TEXT DEFAULT '-',
		deploy_result TEXT,
		device_id TEXT,
		device_pw TEXT,
		device_ppk TEXT,
		last_checked_at TEXT,
		location TEXT,
		os_info TEXT,
		cpu_usage REAL DEFAULT 0,
		memory_usage REAL DEFAULT 0,
		disk_usage REAL DEFAULT 0,
		network_usage REAL DEFAULT 0,
		program_upload_path TEXT,
		program_update_scheme TEXT,
		program_update_path TEXT,
		program_update_port INTEGER DEFAULT 0,
		firewall_deploy_scheme TEXT,
		firewall_deploy_path TEXT,
		firewall_deploy_port INTEGER DEFAULT 0,
		device_info_scheme TEXT,
		device_info_path TEXT,
		device_info_port INTEGER DEFAULT 0
	);`

	// 장비 패키지 버전 테이블
	createFirewallProgramVersionsTable := `
	CREATE TABLE IF NOT EXISTS firewall_program_versions (
		firewall_id INTEGER NOT NULL,
		program_name TEXT NOT NULL,
		version TEXT,
		PRIMARY KEY (firewall_id, program_name),
		FOREIGN KEY (firewall_id) REFERENCES firewalls(id) ON DELETE CASCADE
	);`

	// 패키지 테이블
	createProgramsTable := `
	CREATE TABLE IF NOT EXISTS programs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		process_name TEXT NOT NULL,
		process_file_path TEXT NOT NULL,
		process_version TEXT NOT NULL,
		process_created_at TEXT
	);`

	// 인덱스 생성
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_history_type ON deploy_history(type);",
		"CREATE INDEX IF NOT EXISTS idx_history_device_ip ON deploy_history(device_ip);",
		"CREATE INDEX IF NOT EXISTS idx_history_timestamp ON deploy_history(timestamp);",
		"CREATE INDEX IF NOT EXISTS idx_rule_results_history_id ON rule_results(history_id);",
		"CREATE INDEX IF NOT EXISTS idx_firewalls_device_ip ON firewalls(device_ip);",
		"CREATE INDEX IF NOT EXISTS idx_firewalls_device_name ON firewalls(device_name);",
		"CREATE INDEX IF NOT EXISTS idx_programs_name ON programs(process_name);",
	}

	// 테이블 생성
	tables := []struct {
		name string
		sql  string
	}{
		{"deploy_history", createDeployHistoryTable},
		{"rule_results", createRuleResultsTable},
		{"firewalls", createFirewallsTable},
		{"firewall_program_versions", createFirewallProgramVersionsTable},
		{"programs", createProgramsTable},
	}

	for _, t := range tables {
		if _, err := s.db.Exec(t.sql); err != nil {
			return fmt.Errorf("%s 테이블 생성 실패: %w", t.name, err)
		}
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

// GetHistoryPage 페이지네이션 기반 배포 이력을 조회합니다. (필터/검색/LIMIT/OFFSET)
func (s *SQLiteStore) GetHistoryPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1단계: 필터/검색 조건으로 카운트 조회
	whereClause, whereArgs := buildHistoryWhereClause(req.Filter, req.Keyword, req.StartDate, req.EndDate)

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM deploy_history" + whereClause
	if err := s.db.QueryRow(countQuery, whereArgs...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("이력 카운트 조회 실패: %w", err)
	}

	result := &model.PageResult[model.DeployHistory]{
		TotalCount: totalCount,
	}

	if totalCount == 0 {
		result.Items = []*model.DeployHistory{}
		return result, nil
	}

	// 2단계: 해당 페이지 history 조회 (LIMIT/OFFSET)
	dataQuery := `SELECT id, type, timestamp, device_name, device_ip,
		template_version, program_name, program_version, message, status
		FROM deploy_history` + whereClause + ` ORDER BY id DESC LIMIT ? OFFSET ?`

	dataArgs := append(whereArgs, req.Limit, req.Offset)
	rows, err := s.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("이력 페이지 조회 실패: %w", err)
	}
	defer rows.Close()

	var histories []*model.DeployHistory
	var historyIDs []int
	for rows.Next() {
		h, err := s.scanHistory(rows)
		if err != nil {
			return nil, err
		}
		histories = append(histories, h)
		historyIDs = append(historyIDs, h.ID)
	}

	// 3단계: rule_results 배치 조회 (N+1 해결)
	if len(historyIDs) > 0 {
		ruleMap, err := s.getRuleResultsBatch(historyIDs)
		if err != nil {
			return nil, err
		}
		for _, h := range histories {
			if results, ok := ruleMap[h.ID]; ok {
				h.Results = results
			}
		}
	}

	result.Items = histories
	return result, nil
}

// getRuleResultsBatch 여러 이력의 규칙 결과를 한번에 배치 조회합니다. (N+1 해결)
func (s *SQLiteStore) getRuleResultsBatch(historyIDs []int) (map[int][]model.RuleResult, error) {
	if len(historyIDs) == 0 {
		return nil, nil
	}

	// IN 절 플레이스홀더 생성
	placeholders := make([]byte, 0, len(historyIDs)*2)
	args := make([]interface{}, len(historyIDs))
	for i, id := range historyIDs {
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
		args[i] = id
	}

	query := `SELECT history_id, rule, text, status, reason
		FROM rule_results
		WHERE history_id IN (` + string(placeholders) + `)
		ORDER BY id`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("규칙 결과 배치 조회 실패: %w", err)
	}
	defer rows.Close()

	ruleMap := make(map[int][]model.RuleResult)
	for rows.Next() {
		var historyID int
		var r model.RuleResult
		var text, reason sql.NullString

		if err := rows.Scan(&historyID, &r.Rule, &text, &r.Status, &reason); err != nil {
			return nil, fmt.Errorf("규칙 결과 배치 스캔 실패: %w", err)
		}

		if text.Valid {
			r.Text = text.String
		}
		if reason.Valid {
			r.Reason = reason.String
		}

		ruleMap[historyID] = append(ruleMap[historyID], r)
	}

	return ruleMap, nil
}

// buildHistoryWhereClause 이력 조회용 WHERE 절을 빌드합니다.
func buildHistoryWhereClause(filter, keyword, startDate, endDate string) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 유형 필터
	if filter != "" {
		if filter == model.HistoryTypeFirewall {
			// 레거시 타입 호환: "firewall", "template", ""
			conditions = append(conditions, "(type = ? OR type = ? OR type = ?)")
			args = append(args, model.HistoryTypeFirewall, "template", "")
		} else {
			conditions = append(conditions, "type = ?")
			args = append(args, filter)
		}
	}

	// 날짜 범위 검색
	if startDate != "" {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, startDate+" 00:00:00")
	}
	if endDate != "" {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, endDate+" 23:59:59")
	}

	// 키워드 검색 (LIKE)
	if keyword != "" {
		like := "%" + keyword + "%"
		conditions = append(conditions,
			`(device_name LIKE ? OR device_ip LIKE ? OR
			template_version LIKE ? OR program_version LIKE ? OR
			message LIKE ? OR type LIKE ?)`)
		args = append(args, like, like, like, like, like, like)
	}

	if len(conditions) == 0 {
		return "", nil
	}

	whereClause := " WHERE "
	for i, cond := range conditions {
		if i > 0 {
			whereClause += " AND "
		}
		whereClause += cond
	}
	return whereClause, args
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

// ========== 장비(Firewall) 관련 메서드 ==========

// GetAllFirewalls 모든 장비를 조회합니다.
func (s *SQLiteStore) GetAllFirewalls() ([]*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, device_name, device_ip, server_status, deploy_status,
		version, deploy_result, device_id, device_pw, device_ppk, last_checked_at,
		location, os_info, cpu_usage, memory_usage, disk_usage, network_usage,
		program_upload_path, program_update_scheme, program_update_path, program_update_port,
		firewall_deploy_scheme, firewall_deploy_path, firewall_deploy_port,
		device_info_scheme, device_info_path, device_info_port
		FROM firewalls ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("장비 조회 실패: %w", err)
	}
	defer rows.Close()

	var firewalls []*model.Firewall
	var firewallIDs []int
	for rows.Next() {
		fw, err := s.scanFirewall(rows)
		if err != nil {
			return nil, err
		}
		firewalls = append(firewalls, fw)
		firewallIDs = append(firewallIDs, fw.Index)
	}

	// ProgramVersions 배치 조회
	if len(firewallIDs) > 0 {
		pvMap, err := s.getProgramVersionsBatch(firewallIDs)
		if err != nil {
			return nil, err
		}
		for _, fw := range firewalls {
			if pv, ok := pvMap[fw.Index]; ok {
				fw.ProgramVersions = pv
			}
		}
	}

	return firewalls, nil
}

// GetFirewall ID로 장비를 조회합니다.
func (s *SQLiteStore) GetFirewall(id int) (*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, device_name, device_ip, server_status, deploy_status,
		version, deploy_result, device_id, device_pw, device_ppk, last_checked_at,
		location, os_info, cpu_usage, memory_usage, disk_usage, network_usage,
		program_upload_path, program_update_scheme, program_update_path, program_update_port,
		firewall_deploy_scheme, firewall_deploy_path, firewall_deploy_port,
		device_info_scheme, device_info_path, device_info_port
		FROM firewalls WHERE id = ?`, id)

	fw, err := s.scanFirewallRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("장비를 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("장비 조회 실패: %w", err)
	}

	// ProgramVersions 조회
	pv, err := s.getProgramVersions(fw.Index)
	if err != nil {
		return nil, err
	}
	fw.ProgramVersions = pv

	return fw, nil
}

// GetFirewallByIP IP로 장비를 조회합니다.
func (s *SQLiteStore) GetFirewallByIP(ip string) (*model.Firewall, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, device_name, device_ip, server_status, deploy_status,
		version, deploy_result, device_id, device_pw, device_ppk, last_checked_at,
		location, os_info, cpu_usage, memory_usage, disk_usage, network_usage,
		program_upload_path, program_update_scheme, program_update_path, program_update_port,
		firewall_deploy_scheme, firewall_deploy_path, firewall_deploy_port,
		device_info_scheme, device_info_path, device_info_port
		FROM firewalls WHERE device_ip = ?`, ip)

	fw, err := s.scanFirewallRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("장비를 찾을 수 없습니다: %s", ip)
		}
		return nil, fmt.Errorf("장비 조회 실패: %w", err)
	}

	// ProgramVersions 조회
	pv, err := s.getProgramVersions(fw.Index)
	if err != nil {
		return nil, err
	}
	fw.ProgramVersions = pv

	return fw, nil
}

// SaveFirewall 장비를 저장합니다. (신규 생성 또는 업데이트)
func (s *SQLiteStore) SaveFirewall(fw *model.Firewall) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	// DeployResult JSON 직렬화
	var deployResultJSON sql.NullString
	if fw.DeployResult != nil {
		data, err := json.Marshal(fw.DeployResult)
		if err != nil {
			return fmt.Errorf("DeployResult 직렬화 실패: %w", err)
		}
		deployResultJSON = sql.NullString{String: string(data), Valid: true}
	}

	if fw.Index < 0 {
		// 신규 삽입
		result, err := tx.Exec(`INSERT INTO firewalls
			(device_name, device_ip, server_status, deploy_status, version, deploy_result,
			 device_id, device_pw, device_ppk, last_checked_at, location, os_info,
			 cpu_usage, memory_usage, disk_usage, network_usage,
			 program_upload_path, program_update_scheme, program_update_path, program_update_port,
			 firewall_deploy_scheme, firewall_deploy_path, firewall_deploy_port,
			 device_info_scheme, device_info_path, device_info_port)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			fw.DeviceName, fw.DeviceIP, fw.ServerStatus, fw.DeployStatus, fw.Version, deployResultJSON,
			fw.DeviceID, fw.DevicePW, fw.DevicePPK, fw.LastCheckedAt, fw.Location, fw.OSInfo,
			fw.CPUUsage, fw.MemoryUsage, fw.DiskUsage, fw.NetworkUsage,
			fw.ProgramUploadPath, fw.ProgramUpdateScheme, fw.ProgramUpdatePath, fw.ProgramUpdatePort,
			fw.FirewallDeployScheme, fw.FirewallDeployPath, fw.FirewallDeployPort,
			fw.DeviceInfoScheme, fw.DeviceInfoPath, fw.DeviceInfoPort)
		if err != nil {
			return fmt.Errorf("장비 삽입 실패: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("ID 조회 실패: %w", err)
		}
		fw.Index = int(id)
	} else {
		// 기존 업데이트
		_, err := tx.Exec(`UPDATE firewalls SET
			device_name = ?, device_ip = ?, server_status = ?, deploy_status = ?, version = ?, deploy_result = ?,
			device_id = ?, device_pw = ?, device_ppk = ?, last_checked_at = ?, location = ?, os_info = ?,
			cpu_usage = ?, memory_usage = ?, disk_usage = ?, network_usage = ?,
			program_upload_path = ?, program_update_scheme = ?, program_update_path = ?, program_update_port = ?,
			firewall_deploy_scheme = ?, firewall_deploy_path = ?, firewall_deploy_port = ?,
			device_info_scheme = ?, device_info_path = ?, device_info_port = ?
			WHERE id = ?`,
			fw.DeviceName, fw.DeviceIP, fw.ServerStatus, fw.DeployStatus, fw.Version, deployResultJSON,
			fw.DeviceID, fw.DevicePW, fw.DevicePPK, fw.LastCheckedAt, fw.Location, fw.OSInfo,
			fw.CPUUsage, fw.MemoryUsage, fw.DiskUsage, fw.NetworkUsage,
			fw.ProgramUploadPath, fw.ProgramUpdateScheme, fw.ProgramUpdatePath, fw.ProgramUpdatePort,
			fw.FirewallDeployScheme, fw.FirewallDeployPath, fw.FirewallDeployPort,
			fw.DeviceInfoScheme, fw.DeviceInfoPath, fw.DeviceInfoPort,
			fw.Index)
		if err != nil {
			return fmt.Errorf("장비 업데이트 실패: %w", err)
		}

		// 기존 ProgramVersions 삭제
		if _, err := tx.Exec("DELETE FROM firewall_program_versions WHERE firewall_id = ?", fw.Index); err != nil {
			return fmt.Errorf("패키지 버전 삭제 실패: %w", err)
		}
	}

	// ProgramVersions 삽입
	if len(fw.ProgramVersions) > 0 {
		for name, ver := range fw.ProgramVersions {
			if _, err := tx.Exec(`INSERT INTO firewall_program_versions (firewall_id, program_name, version)
				VALUES (?, ?, ?)`, fw.Index, name, ver); err != nil {
				return fmt.Errorf("패키지 버전 삽입 실패: %w", err)
			}
		}
	}

	return tx.Commit()
}

// DeleteFirewall 장비를 삭제합니다.
func (s *SQLiteStore) DeleteFirewall(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.Exec("DELETE FROM firewalls WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("장비 삭제 실패: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("장비를 찾을 수 없습니다: %d", id)
	}
	return nil
}

// ClearFirewalls 모든 장비를 삭제합니다.
func (s *SQLiteStore) ClearFirewalls() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.db.Exec("DELETE FROM firewalls"); err != nil {
		return fmt.Errorf("장비 전체 삭제 실패: %w", err)
	}
	return nil
}

// CountFirewalls 장비 수를 반환합니다.
func (s *SQLiteStore) CountFirewalls() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM firewalls").Scan(&count)
	return count
}

// GetFirewallPage 페이지네이션 기반 장비를 조회합니다.
func (s *SQLiteStore) GetFirewallPage(req model.PageRequest) (*model.PageResult[model.Firewall], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	whereClause, whereArgs := buildFirewallWhereClause(req.Keyword, req.StartDate, req.EndDate)

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM firewalls" + whereClause
	if err := s.db.QueryRow(countQuery, whereArgs...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("장비 카운트 조회 실패: %w", err)
	}

	result := &model.PageResult[model.Firewall]{TotalCount: totalCount}
	if totalCount == 0 {
		result.Items = []*model.Firewall{}
		return result, nil
	}

	dataQuery := `SELECT id, device_name, device_ip, server_status, deploy_status,
		version, deploy_result, device_id, device_pw, device_ppk, last_checked_at,
		location, os_info, cpu_usage, memory_usage, disk_usage, network_usage,
		program_upload_path, program_update_scheme, program_update_path, program_update_port,
		firewall_deploy_scheme, firewall_deploy_path, firewall_deploy_port,
		device_info_scheme, device_info_path, device_info_port
		FROM firewalls` + whereClause + ` ORDER BY id LIMIT ? OFFSET ?`

	dataArgs := append(whereArgs, req.Limit, req.Offset)
	rows, err := s.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("장비 페이지 조회 실패: %w", err)
	}
	defer rows.Close()

	var firewalls []*model.Firewall
	var firewallIDs []int
	for rows.Next() {
		fw, err := s.scanFirewall(rows)
		if err != nil {
			return nil, err
		}
		firewalls = append(firewalls, fw)
		firewallIDs = append(firewallIDs, fw.Index)
	}

	// ProgramVersions 배치 조회
	if len(firewallIDs) > 0 {
		pvMap, err := s.getProgramVersionsBatch(firewallIDs)
		if err != nil {
			return nil, err
		}
		for _, fw := range firewalls {
			if pv, ok := pvMap[fw.Index]; ok {
				fw.ProgramVersions = pv
			}
		}
	}

	result.Items = firewalls
	return result, nil
}

// scanFirewall rows에서 Firewall을 스캔합니다.
func (s *SQLiteStore) scanFirewall(rows *sql.Rows) (*model.Firewall, error) {
	var fw model.Firewall
	var deployResult, deviceID, devicePW, devicePPK, lastCheckedAt sql.NullString
	var location, osInfo sql.NullString
	var programUploadPath, programUpdateScheme, programUpdatePath sql.NullString
	var firewallDeployScheme, firewallDeployPath sql.NullString
	var deviceInfoScheme, deviceInfoPath sql.NullString

	err := rows.Scan(
		&fw.Index, &fw.DeviceName, &fw.DeviceIP, &fw.ServerStatus, &fw.DeployStatus,
		&fw.Version, &deployResult, &deviceID, &devicePW, &devicePPK, &lastCheckedAt,
		&location, &osInfo, &fw.CPUUsage, &fw.MemoryUsage, &fw.DiskUsage, &fw.NetworkUsage,
		&programUploadPath, &programUpdateScheme, &programUpdatePath, &fw.ProgramUpdatePort,
		&firewallDeployScheme, &firewallDeployPath, &fw.FirewallDeployPort,
		&deviceInfoScheme, &deviceInfoPath, &fw.DeviceInfoPort)
	if err != nil {
		return nil, fmt.Errorf("장비 스캔 실패: %w", err)
	}

	// NULL 처리
	if deployResult.Valid && deployResult.String != "" {
		var dr model.DeployResult
		if err := json.Unmarshal([]byte(deployResult.String), &dr); err == nil {
			fw.DeployResult = &dr
		}
	}
	if deviceID.Valid {
		fw.DeviceID = deviceID.String
	}
	if devicePW.Valid {
		fw.DevicePW = devicePW.String
	}
	if devicePPK.Valid {
		fw.DevicePPK = devicePPK.String
	}
	if lastCheckedAt.Valid {
		fw.LastCheckedAt = lastCheckedAt.String
	}
	if location.Valid {
		fw.Location = location.String
	}
	if osInfo.Valid {
		fw.OSInfo = osInfo.String
	}
	if programUploadPath.Valid {
		fw.ProgramUploadPath = programUploadPath.String
	}
	if programUpdateScheme.Valid {
		fw.ProgramUpdateScheme = programUpdateScheme.String
	}
	if programUpdatePath.Valid {
		fw.ProgramUpdatePath = programUpdatePath.String
	}
	if firewallDeployScheme.Valid {
		fw.FirewallDeployScheme = firewallDeployScheme.String
	}
	if firewallDeployPath.Valid {
		fw.FirewallDeployPath = firewallDeployPath.String
	}
	if deviceInfoScheme.Valid {
		fw.DeviceInfoScheme = deviceInfoScheme.String
	}
	if deviceInfoPath.Valid {
		fw.DeviceInfoPath = deviceInfoPath.String
	}

	return &fw, nil
}

// scanFirewallRow QueryRow에서 Firewall을 스캔합니다.
func (s *SQLiteStore) scanFirewallRow(row *sql.Row) (*model.Firewall, error) {
	var fw model.Firewall
	var deployResult, deviceID, devicePW, devicePPK, lastCheckedAt sql.NullString
	var location, osInfo sql.NullString
	var programUploadPath, programUpdateScheme, programUpdatePath sql.NullString
	var firewallDeployScheme, firewallDeployPath sql.NullString
	var deviceInfoScheme, deviceInfoPath sql.NullString

	err := row.Scan(
		&fw.Index, &fw.DeviceName, &fw.DeviceIP, &fw.ServerStatus, &fw.DeployStatus,
		&fw.Version, &deployResult, &deviceID, &devicePW, &devicePPK, &lastCheckedAt,
		&location, &osInfo, &fw.CPUUsage, &fw.MemoryUsage, &fw.DiskUsage, &fw.NetworkUsage,
		&programUploadPath, &programUpdateScheme, &programUpdatePath, &fw.ProgramUpdatePort,
		&firewallDeployScheme, &firewallDeployPath, &fw.FirewallDeployPort,
		&deviceInfoScheme, &deviceInfoPath, &fw.DeviceInfoPort)
	if err != nil {
		return nil, err
	}

	// NULL 처리 (scanFirewall과 동일)
	if deployResult.Valid && deployResult.String != "" {
		var dr model.DeployResult
		if err := json.Unmarshal([]byte(deployResult.String), &dr); err == nil {
			fw.DeployResult = &dr
		}
	}
	if deviceID.Valid {
		fw.DeviceID = deviceID.String
	}
	if devicePW.Valid {
		fw.DevicePW = devicePW.String
	}
	if devicePPK.Valid {
		fw.DevicePPK = devicePPK.String
	}
	if lastCheckedAt.Valid {
		fw.LastCheckedAt = lastCheckedAt.String
	}
	if location.Valid {
		fw.Location = location.String
	}
	if osInfo.Valid {
		fw.OSInfo = osInfo.String
	}
	if programUploadPath.Valid {
		fw.ProgramUploadPath = programUploadPath.String
	}
	if programUpdateScheme.Valid {
		fw.ProgramUpdateScheme = programUpdateScheme.String
	}
	if programUpdatePath.Valid {
		fw.ProgramUpdatePath = programUpdatePath.String
	}
	if firewallDeployScheme.Valid {
		fw.FirewallDeployScheme = firewallDeployScheme.String
	}
	if firewallDeployPath.Valid {
		fw.FirewallDeployPath = firewallDeployPath.String
	}
	if deviceInfoScheme.Valid {
		fw.DeviceInfoScheme = deviceInfoScheme.String
	}
	if deviceInfoPath.Valid {
		fw.DeviceInfoPath = deviceInfoPath.String
	}

	return &fw, nil
}

// getProgramVersions 특정 장비의 패키지 버전을 조회합니다.
func (s *SQLiteStore) getProgramVersions(firewallID int) (map[string]string, error) {
	rows, err := s.db.Query(`SELECT program_name, version FROM firewall_program_versions WHERE firewall_id = ?`, firewallID)
	if err != nil {
		return nil, fmt.Errorf("패키지 버전 조회 실패: %w", err)
	}
	defer rows.Close()

	pv := make(map[string]string)
	for rows.Next() {
		var name, version string
		if err := rows.Scan(&name, &version); err != nil {
			return nil, fmt.Errorf("패키지 버전 스캔 실패: %w", err)
		}
		pv[name] = version
	}
	return pv, nil
}

// getProgramVersionsBatch 여러 장비의 패키지 버전을 한번에 조회합니다.
func (s *SQLiteStore) getProgramVersionsBatch(firewallIDs []int) (map[int]map[string]string, error) {
	if len(firewallIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(firewallIDs))
	args := make([]interface{}, len(firewallIDs))
	for i, id := range firewallIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT firewall_id, program_name, version FROM firewall_program_versions
		WHERE firewall_id IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("패키지 버전 배치 조회 실패: %w", err)
	}
	defer rows.Close()

	pvMap := make(map[int]map[string]string)
	for rows.Next() {
		var firewallID int
		var name, version string
		if err := rows.Scan(&firewallID, &name, &version); err != nil {
			return nil, fmt.Errorf("패키지 버전 배치 스캔 실패: %w", err)
		}
		if pvMap[firewallID] == nil {
			pvMap[firewallID] = make(map[string]string)
		}
		pvMap[firewallID][name] = version
	}
	return pvMap, nil
}

// buildFirewallWhereClause 장비 조회용 WHERE 절을 빌드합니다.
func buildFirewallWhereClause(keyword, startDate, endDate string) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if keyword != "" {
		like := "%" + keyword + "%"
		conditions = append(conditions, "(device_name LIKE ? OR device_ip LIKE ? OR location LIKE ? OR os_info LIKE ?)")
		args = append(args, like, like, like, like)
	}
	if startDate != "" {
		conditions = append(conditions, "last_checked_at >= ?")
		args = append(args, startDate+" 00:00:00")
	}
	if endDate != "" {
		conditions = append(conditions, "last_checked_at <= ?")
		args = append(args, endDate+" 23:59:59")
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// ========== 패키지(Program) 관련 메서드 ==========

// GetAllPrograms 모든 패키지를 조회합니다.
func (s *SQLiteStore) GetAllPrograms() ([]*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, process_name, process_file_path, process_version, process_created_at
		FROM programs ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("패키지 조회 실패: %w", err)
	}
	defer rows.Close()

	var programs []*model.ProcessInfo
	for rows.Next() {
		p, err := s.scanProgram(rows)
		if err != nil {
			return nil, err
		}
		programs = append(programs, p)
	}
	return programs, nil
}

// GetProgram ID로 패키지를 조회합니다.
func (s *SQLiteStore) GetProgram(id int) (*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, process_name, process_file_path, process_version, process_created_at
		FROM programs WHERE id = ?`, id)

	p, err := s.scanProgramRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("패키지를 찾을 수 없습니다: %d", id)
		}
		return nil, fmt.Errorf("패키지 조회 실패: %w", err)
	}
	return p, nil
}

// GetProgramByName 이름으로 패키지를 조회합니다.
func (s *SQLiteStore) GetProgramByName(name string) (*model.ProcessInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`SELECT id, process_name, process_file_path, process_version, process_created_at
		FROM programs WHERE process_name = ?`, name)

	p, err := s.scanProgramRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("패키지를 찾을 수 없습니다: %s", name)
		}
		return nil, fmt.Errorf("패키지 조회 실패: %w", err)
	}
	return p, nil
}

// SaveProgram 패키지를 저장합니다. (신규 생성 또는 업데이트)
func (s *SQLiteStore) SaveProgram(p *model.ProcessInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p.ID < 0 {
		// 신규 삽입
		result, err := s.db.Exec(`INSERT INTO programs
			(process_name, process_file_path, process_version, process_created_at)
			VALUES (?, ?, ?, ?)`,
			p.ProcessName, p.ProcessFilePath, p.ProcessVersion, p.ProcessCreatedAt)
		if err != nil {
			return fmt.Errorf("패키지 삽입 실패: %w", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("ID 조회 실패: %w", err)
		}
		p.ID = int(id)
	} else {
		// 기존 업데이트
		_, err := s.db.Exec(`UPDATE programs SET
			process_name = ?, process_file_path = ?, process_version = ?, process_created_at = ?
			WHERE id = ?`,
			p.ProcessName, p.ProcessFilePath, p.ProcessVersion, p.ProcessCreatedAt, p.ID)
		if err != nil {
			return fmt.Errorf("패키지 업데이트 실패: %w", err)
		}
	}
	return nil
}

// DeleteProgram 패키지를 삭제합니다.
func (s *SQLiteStore) DeleteProgram(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.Exec("DELETE FROM programs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("패키지 삭제 실패: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("패키지를 찾을 수 없습니다: %d", id)
	}
	return nil
}

// ClearPrograms 모든 패키지를 삭제합니다.
func (s *SQLiteStore) ClearPrograms() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.db.Exec("DELETE FROM programs"); err != nil {
		return fmt.Errorf("패키지 전체 삭제 실패: %w", err)
	}
	return nil
}

// CountPrograms 패키지 수를 반환합니다.
func (s *SQLiteStore) CountPrograms() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM programs").Scan(&count)
	return count
}

// GetProgramPage 페이지네이션 기반 패키지를 조회합니다.
func (s *SQLiteStore) GetProgramPage(req model.PageRequest) (*model.PageResult[model.ProcessInfo], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	whereClause, whereArgs := buildProgramWhereClause(req.Keyword, req.StartDate, req.EndDate)

	var totalCount int
	countQuery := "SELECT COUNT(*) FROM programs" + whereClause
	if err := s.db.QueryRow(countQuery, whereArgs...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("패키지 카운트 조회 실패: %w", err)
	}

	result := &model.PageResult[model.ProcessInfo]{TotalCount: totalCount}
	if totalCount == 0 {
		result.Items = []*model.ProcessInfo{}
		return result, nil
	}

	dataQuery := `SELECT id, process_name, process_file_path, process_version, process_created_at
		FROM programs` + whereClause + ` ORDER BY id LIMIT ? OFFSET ?`

	dataArgs := append(whereArgs, req.Limit, req.Offset)
	rows, err := s.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("패키지 페이지 조회 실패: %w", err)
	}
	defer rows.Close()

	var programs []*model.ProcessInfo
	for rows.Next() {
		p, err := s.scanProgram(rows)
		if err != nil {
			return nil, err
		}
		programs = append(programs, p)
	}

	result.Items = programs
	return result, nil
}

// scanProgram rows에서 ProcessInfo를 스캔합니다.
func (s *SQLiteStore) scanProgram(rows *sql.Rows) (*model.ProcessInfo, error) {
	var p model.ProcessInfo
	var createdAt sql.NullString
	err := rows.Scan(&p.ID, &p.ProcessName, &p.ProcessFilePath, &p.ProcessVersion, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("패키지 스캔 실패: %w", err)
	}
	if createdAt.Valid {
		p.ProcessCreatedAt = createdAt.String
	}
	return &p, nil
}

// scanProgramRow QueryRow에서 ProcessInfo를 스캔합니다.
func (s *SQLiteStore) scanProgramRow(row *sql.Row) (*model.ProcessInfo, error) {
	var p model.ProcessInfo
	var createdAt sql.NullString
	err := row.Scan(&p.ID, &p.ProcessName, &p.ProcessFilePath, &p.ProcessVersion, &createdAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		p.ProcessCreatedAt = createdAt.String
	}
	return &p, nil
}

// buildProgramWhereClause 패키지 조회용 WHERE 절을 빌드합니다.
func buildProgramWhereClause(keyword, startDate, endDate string) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if keyword != "" {
		like := "%" + keyword + "%"
		conditions = append(conditions, "(process_name LIKE ? OR process_version LIKE ? OR process_file_path LIKE ?)")
		args = append(args, like, like, like)
	}
	if startDate != "" {
		conditions = append(conditions, "process_created_at >= ?")
		args = append(args, startDate+" 00:00:00")
	}
	if endDate != "" {
		conditions = append(conditions, "process_created_at <= ?")
		args = append(args, endDate+" 23:59:59")
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// ========== 내부 헬퍼 메서드 ==========

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
