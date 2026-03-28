package backoffice

import (
	api "github.com/timickb/sagaflow/engine/api/backoffice"
	"github.com/timickb/sagaflow/engine/internal/domain"
	"github.com/timickb/sagaflow/engine/pkg/utils"
)

func mapGetFeedDtoFromApi(src api.GetFeedRequestObject) *domain.GetFeedDto {
	result := new(domain.GetFeedDto)
	result.Count = int64(src.Params.Count)

	if src.Params.Statuses != nil {
		list := *src.Params.Statuses
		result.Statuses = utils.MapSlice(list, mapSagaStatusFromApi)
	}
	return result
}

func mapSagaStatusFromApi(src api.SagaStatus) domain.InstanceStatus {
	switch src {
	case api.COMPENSATED:
		return domain.InstanceStatusCompensated
	case api.COMPLETED:
		return domain.InstanceStatusCompleted
	case api.FAILED:
		return domain.InstanceStatusFailed
	case api.PENDING:
		return domain.InstanceStatusPending
	case api.RUNNING:
		return domain.InstanceStatusRunning
	case api.COMPENSATING:
		return domain.InstanceStatusCompensating
	case api.INCONSISTENT:
		return domain.InstanceStatusInconsistent
	default:
		return ""
	}
}
