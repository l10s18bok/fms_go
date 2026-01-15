package repository

import (
	"fms/internal/domain"
	"fms/internal/model"
	"fms/internal/storage"
)

// JSONFirewallRepository JSON 기반 장비 저장소 어댑터
type JSONFirewallRepository struct {
	store *storage.JSONStore
}

// NewJSONFirewallRepository 새로운 JSON 장비 저장소를 생성합니다.
func NewJSONFirewallRepository(store *storage.JSONStore) FirewallRepository {
	return &JSONFirewallRepository{store: store}
}

// GetAll 모든 장비를 조회합니다.
func (r *JSONFirewallRepository) GetAll() ([]*domain.Firewall, error) {
	firewalls, err := r.store.GetAllFirewalls()
	if err != nil {
		return nil, err
	}
	return r.toDomainList(firewalls), nil
}

// GetByIndex 인덱스로 장비를 조회합니다.
func (r *JSONFirewallRepository) GetByIndex(index int) (*domain.Firewall, error) {
	fw, err := r.store.GetFirewall(index)
	if err != nil {
		return nil, err
	}
	return r.toDomain(fw), nil
}

// GetByIP IP 주소로 장비를 조회합니다.
func (r *JSONFirewallRepository) GetByIP(ip string) (*domain.Firewall, error) {
	fw, err := r.store.GetFirewallByIP(ip)
	if err != nil {
		return nil, err
	}
	return r.toDomain(fw), nil
}

// Save 장비를 저장합니다.
func (r *JSONFirewallRepository) Save(fw *domain.Firewall) error {
	modelFw := r.toModel(fw)
	err := r.store.SaveFirewall(modelFw)
	if err != nil {
		return err
	}
	// Index 업데이트 (새로 생성된 경우)
	fw.Index = modelFw.Index
	return nil
}

// Delete 장비를 삭제합니다.
func (r *JSONFirewallRepository) Delete(index int) error {
	return r.store.DeleteFirewall(index)
}

// Clear 모든 장비를 삭제합니다.
func (r *JSONFirewallRepository) Clear() error {
	return r.store.ClearFirewalls()
}

// Count 장비 수를 반환합니다.
func (r *JSONFirewallRepository) Count() int {
	return r.store.CountFirewalls()
}

// toDomain model.Firewall을 domain.Firewall로 변환합니다.
func (r *JSONFirewallRepository) toDomain(fw *model.Firewall) *domain.Firewall {
	if fw == nil {
		return nil
	}

	domainFw := &domain.Firewall{
		Index:         fw.Index,
		DeviceName:    fw.DeviceName,
		DeviceIP:      fw.DeviceIP,
		ServerStatus:  fw.ServerStatus,
		DeployStatus:  fw.DeployStatus,
		Version:       fw.Version,
		DeviceID:      fw.DeviceID,
		DevicePW:      fw.DevicePW,
		DevicePPK:     fw.DevicePPK,
		LastCheckedAt: fw.LastCheckedAt,
		Processes:     []domain.DeviceProcessInfo{},
	}

	// DeployResult 변환
	if fw.DeployResult != nil {
		domainFw.DeployResult = &domain.DeployResult{
			IP:     fw.DeployResult.IP,
			Status: fw.DeployResult.Status,
		}
		if len(fw.DeployResult.Info) > 0 {
			domainFw.DeployResult.Info = make([]domain.ResultInfo, len(fw.DeployResult.Info))
			for i, info := range fw.DeployResult.Info {
				domainFw.DeployResult.Info[i] = domain.ResultInfo{
					Index:  info.Index,
					Rule:   info.Rule,
					Text:   info.Text,
					Status: info.Status,
					Reason: info.Reason,
				}
			}
		}
	}

	// ProgramVersions를 Processes로 변환
	if fw.ProgramVersions != nil {
		for name, version := range fw.ProgramVersions {
			domainFw.Processes = append(domainFw.Processes, domain.DeviceProcessInfo{
				Name:    name,
				Version: version,
			})
		}
	}

	return domainFw
}

// toDomainList model.Firewall 슬라이스를 domain.Firewall 슬라이스로 변환합니다.
func (r *JSONFirewallRepository) toDomainList(firewalls []*model.Firewall) []*domain.Firewall {
	result := make([]*domain.Firewall, len(firewalls))
	for i, fw := range firewalls {
		result[i] = r.toDomain(fw)
	}
	return result
}

// toModel domain.Firewall을 model.Firewall로 변환합니다.
func (r *JSONFirewallRepository) toModel(fw *domain.Firewall) *model.Firewall {
	if fw == nil {
		return nil
	}

	modelFw := &model.Firewall{
		Index:         fw.Index,
		DeviceName:    fw.DeviceName,
		DeviceIP:      fw.DeviceIP,
		ServerStatus:  fw.ServerStatus,
		DeployStatus:  fw.DeployStatus,
		Version:       fw.Version,
		DeviceID:      fw.DeviceID,
		DevicePW:      fw.DevicePW,
		DevicePPK:     fw.DevicePPK,
		LastCheckedAt: fw.LastCheckedAt,
	}

	// DeployResult 변환
	if fw.DeployResult != nil {
		modelFw.DeployResult = &model.DeployResult{
			IP:     fw.DeployResult.IP,
			Status: fw.DeployResult.Status,
		}
		if len(fw.DeployResult.Info) > 0 {
			modelFw.DeployResult.Info = make([]model.ResultInfo, len(fw.DeployResult.Info))
			for i, info := range fw.DeployResult.Info {
				modelFw.DeployResult.Info[i] = model.ResultInfo{
					Index:  info.Index,
					Rule:   info.Rule,
					Text:   info.Text,
					Status: info.Status,
					Reason: info.Reason,
				}
			}
		}
	}

	// Processes를 ProgramVersions로 변환
	if len(fw.Processes) > 0 {
		modelFw.ProgramVersions = make(map[string]string)
		for _, p := range fw.Processes {
			modelFw.ProgramVersions[p.Name] = p.Version
		}
	}

	return modelFw
}
