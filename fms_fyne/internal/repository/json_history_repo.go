package repository

import (
	"fms/internal/domain"
	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/utils"
	"time"
)

// JSONHistoryRepository JSON 기반 배포 이력 저장소 어댑터
type JSONHistoryRepository struct {
	store *storage.JSONStore
}

// NewJSONHistoryRepository 새로운 JSON 이력 저장소를 생성합니다.
func NewJSONHistoryRepository(store *storage.JSONStore) HistoryRepository {
	return &JSONHistoryRepository{store: store}
}

// GetAll 모든 이력을 조회합니다.
func (r *JSONHistoryRepository) GetAll() ([]*domain.DeployHistory, error) {
	histories, err := r.store.GetAllHistory()
	if err != nil {
		return nil, err
	}
	return r.toDomainList(histories), nil
}

// GetByID ID로 이력을 조회합니다.
func (r *JSONHistoryRepository) GetByID(id int) (*domain.DeployHistory, error) {
	h, err := r.store.GetHistory(id)
	if err != nil {
		return nil, err
	}
	return r.toDomain(h), nil
}

// GetByType 이력 유형으로 조회합니다.
func (r *JSONHistoryRepository) GetByType(historyType string) ([]*domain.DeployHistory, error) {
	histories, err := r.store.GetAllHistory()
	if err != nil {
		return nil, err
	}

	var filtered []*model.DeployHistory
	for _, h := range histories {
		if h.Type == historyType {
			filtered = append(filtered, h)
		}
	}

	return r.toDomainList(filtered), nil
}

// Save 이력을 저장합니다.
func (r *JSONHistoryRepository) Save(h *domain.DeployHistory) error {
	modelH := r.toModel(h)
	err := r.store.SaveHistory(modelH)
	if err != nil {
		return err
	}
	// ID 업데이트 (새로 생성된 경우)
	h.ID = modelH.ID
	return nil
}

// Delete 이력을 삭제합니다.
func (r *JSONHistoryRepository) Delete(id int) error {
	return r.store.DeleteHistory(id)
}

// Clear 모든 이력을 삭제합니다.
func (r *JSONHistoryRepository) Clear() error {
	return r.store.ClearHistory()
}

// Count 이력 수를 반환합니다.
func (r *JSONHistoryRepository) Count() int {
	return r.store.CountHistory()
}

// toDomain model.DeployHistory를 domain.DeployHistory로 변환합니다.
func (r *JSONHistoryRepository) toDomain(h *model.DeployHistory) *domain.DeployHistory {
	if h == nil {
		return nil
	}

	domainH := &domain.DeployHistory{
		ID:          h.ID,
		Timestamp:   domain.JSONTime(time.Time(h.Timestamp)),
		DeviceName:  h.DeviceName,
		DeviceIP:    h.DeviceIP,
		Type:        h.Type,
		Status:      h.Status,
		TemplateVer: h.TemplateVer,
		ProgramName: h.ProgramName,
		Message:     h.Message,
		Results:     []domain.RuleResult{},
	}

	// Version 설정 (Type에 따라)
	if h.Type == model.HistoryTypeProgram {
		domainH.Version = h.ProgramVer
	} else {
		domainH.Version = h.TemplateVer
	}

	// Results 변환
	if len(h.Results) > 0 {
		domainH.Results = make([]domain.RuleResult, len(h.Results))
		for i, r := range h.Results {
			domainH.Results[i] = domain.RuleResult{
				Rule:   r.Rule,
				Text:   r.Text,
				Status: r.Status,
				Reason: r.Reason,
			}
		}
	}

	return domainH
}

// toDomainList model.DeployHistory 슬라이스를 domain.DeployHistory 슬라이스로 변환합니다.
func (r *JSONHistoryRepository) toDomainList(histories []*model.DeployHistory) []*domain.DeployHistory {
	result := make([]*domain.DeployHistory, len(histories))
	for i, h := range histories {
		result[i] = r.toDomain(h)
	}
	return result
}

// toModel domain.DeployHistory를 model.DeployHistory로 변환합니다.
func (r *JSONHistoryRepository) toModel(h *domain.DeployHistory) *model.DeployHistory {
	if h == nil {
		return nil
	}

	modelH := &model.DeployHistory{
		ID:          h.ID,
		Timestamp:   utils.JSONTime(time.Time(h.Timestamp)),
		DeviceName:  h.DeviceName,
		DeviceIP:    h.DeviceIP,
		Type:        h.Type,
		Status:      h.Status,
		TemplateVer: h.TemplateVer,
		ProgramName: h.ProgramName,
		Message:     h.Message,
		Results:     []model.RuleResult{},
	}

	// Version 설정 (Type에 따라)
	if h.Type == domain.HistoryTypeProgram {
		modelH.ProgramVer = h.Version
	} else {
		modelH.TemplateVer = h.Version
	}

	// Results 변환
	if len(h.Results) > 0 {
		modelH.Results = make([]model.RuleResult, len(h.Results))
		for i, r := range h.Results {
			modelH.Results[i] = model.RuleResult{
				Rule:   r.Rule,
				Text:   r.Text,
				Status: r.Status,
				Reason: r.Reason,
			}
		}
	}

	return modelH
}
