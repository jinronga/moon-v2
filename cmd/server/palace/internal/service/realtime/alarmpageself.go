package realtime

import (
	"context"

	pb "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/helper/middleware"
)

// AlarmPageSelfService is a service that implements the AlarmPageSelfServer.
type AlarmPageSelfService struct {
	pb.UnimplementedAlarmPageSelfServer

	alarmPageBiz *biz.AlarmPageBiz
}

// NewAlarmPageSelfService creates a new AlarmPageSelfService.
func NewAlarmPageSelfService(alarmPageBiz *biz.AlarmPageBiz) *AlarmPageSelfService {
	return &AlarmPageSelfService{
		alarmPageBiz: alarmPageBiz,
	}
}

// UpdateAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	if err := s.alarmPageBiz.UpdateAlarmPage(ctx, middleware.GetUserID(ctx), req.GetAlarmPageIds()); err != nil {
		return nil, err
	}
	return &pb.UpdateAlarmPageReply{}, nil
}

// ListAlarmPage implements AlarmPageSelfServer.
func (s *AlarmPageSelfService) ListAlarmPage(ctx context.Context, _ *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	alarmPageList, err := s.alarmPageBiz.ListAlarmPage(ctx, middleware.GetUserID(ctx))
	if err != nil {
		return nil, err
	}
	return &pb.ListAlarmPageReply{
		List: builder.NewParamsBuild().WithContext(ctx).RealtimeAlarmModuleBuilder().DoAlarmPageSelfBuilder().ToAPIs(alarmPageList),
	}, nil
}
