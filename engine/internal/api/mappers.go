package api

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/timickb/sagaflow/engine/internal/domain"
	pb "github.com/timickb/sagaflow/proto/gen/go/sagaflow"
)

func errorToStatus(err error) *pb.Status {
	if err == nil {
		return &pb.Status{Code: int32(codes.OK)}
	}

	st, ok := status.FromError(err)
	if !ok {
		return &pb.Status{
			Code:  int32(codes.Internal),
			Error: &pb.Error{Message: err.Error()},
		}
	}

	return &pb.Status{
		Code:  int32(st.Code()),
		Error: &pb.Error{Message: st.Message()},
	}
}

func instanceStatusToProto(s domain.InstanceStatus) pb.InstanceStatus {
	switch s {
	case domain.InstanceStatusPending:
		return pb.InstanceStatus_INSTANCE_STATUS_PENDING
	case domain.InstanceStatusRunning:
		return pb.InstanceStatus_INSTANCE_STATUS_RUNNING
	case domain.InstanceStatusCompleted:
		return pb.InstanceStatus_INSTANCE_STATUS_COMPLETED
	case domain.InstanceStatusFailed:
		return pb.InstanceStatus_INSTANCE_STATUS_FAILED
	case domain.InstanceStatusVerifying:
		return pb.InstanceStatus_INSTANCE_STATUS_VERIFYING
	case domain.InstanceStatusCompensating:
		return pb.InstanceStatus_INSTANCE_STATUS_COMPENSATING
	case domain.InstanceStatusCompensated:
		return pb.InstanceStatus_INSTANCE_STATUS_COMPENSATED
	case domain.InstanceStatusInconsistent:
		return pb.InstanceStatus_INSTANCE_STATUS_INCONSISTENT
	default:
		return pb.InstanceStatus_INSTANCE_STATUS_UNSPECIFIED
	}
}

func instanceViewToProto(v *domain.InstanceView) *pb.InstanceView {
	if v == nil {
		return nil
	}

	var initialCtx, runtimeCtx []byte
	if v.InitialContext != nil {
		initialCtx = v.InitialContext.GetRaw()
	}
	if v.RuntimeContext != nil {
		runtimeCtx = v.RuntimeContext.GetRaw()
	}

	return &pb.InstanceView{
		Status:         instanceStatusToProto(v.Status),
		SagaName:       v.SagaName,
		SagaVersion:    int32(v.SagaVersion),
		InitialContext: initialCtx,
		RuntimeContext: runtimeCtx,
	}
}

// === StartRequest mappers ===

// startRequestToDto converts proto StartRequest to domain InstanceStartDto.
func startRequestToDto(req *pb.StartRequest) (*domain.InstanceStartDto, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	initialCtx, err := domain.NewJsonInstanceContextFromRaw(req.InitialContext)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid initial context: %v", err)
	}

	dto := &domain.InstanceStartDto{
		SagaName:       req.SagaName,
		SagaVersion:    int(req.SagaVersion),
		InitialContext: initialCtx,
		StartStepName:  "", // значение вычисляется внутри юзкейса
	}

	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		key := *req.IdempotencyKey
		dto.IdempotencyKey = &key
	}

	return dto, nil
}
