package repo

import "github.com/timickb/sagaflow/engine/internal/domain"

type DBSagaInstance struct {
}

func (si *DBSagaInstance) TableName() string {
	return "saga_instance"
}

func (si *DBSagaInstance) ToDomain() *domain.InstanceView {
	panic("implement me")
}
