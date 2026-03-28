package backoffice

import (
	"context"

	api "github.com/timickb/sagaflow/engine/api/backoffice"
	"github.com/timickb/sagaflow/engine/internal/domain"
)

type Handler struct {
	usecase domain.InstanceUsecase
}

func NewHandler(usecase domain.InstanceUsecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) GetFeed(ctx context.Context, req api.GetFeedRequestObject) (api.GetFeedResponseObject, error) {
	// TODO: implement
	panic("implement me")
}

func (h *Handler) GedFeedCount(ctx context.Context, req api.GedFeedCountRequestObject) (api.GedFeedCountResponseObject, error) {
	// TODO: implement
	panic("implement me")
}

func (h *Handler) GetFeedPaging(ctx context.Context, req api.GetFeedPagingRequestObject) (api.GetFeedPagingResponseObject, error) {
	// TODO: implement
	panic("implement me")
}

func (h *Handler) StartSaga(ctx context.Context, req api.StartSagaRequestObject) (api.StartSagaResponseObject, error) {
	// TODO: implement
	panic("implement me")
}

func (h *Handler) GetSagaStatus(ctx context.Context, req api.GetSagaStatusRequestObject) (api.GetSagaStatusResponseObject, error) {
	// TODO: implement
	panic("implement me")
}
