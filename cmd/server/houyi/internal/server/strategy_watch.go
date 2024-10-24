package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"
)

const (
	// 策略任务执行间隔， 默认10s
	strategyWatchJobSpec = "@every 10s"
	// 执行任务超时时间
	strategyWatchJobTimeout = 10 * time.Second
)

func newStrategyWatch(c *houyiconf.Bootstrap, data *data.Data, alertService *service.AlertService) *StrategyWatch {
	cronInstance := cron.New()
	strategyConf := c.GetWatch().GetStrategy()
	interval := strategyWatchJobSpec
	if strategyConf.GetInterval() != "" {
		interval = strategyConf.GetInterval()
	}
	timeout := strategyWatchJobTimeout
	if strategyConf.GetTimeout().AsDuration() > 0 {
		timeout = strategyConf.GetTimeout().AsDuration()
	}
	return &StrategyWatch{
		data:         data,
		cronInstance: cronInstance,
		stopCh:       make(chan struct{}),
		entryIDMap:   make(map[string]cron.EntryID),
		alertService: alertService,
		interval:     interval,
		timeout:      timeout,
	}
}

var _ transport.Server = (*StrategyWatch)(nil)

// StrategyWatch 策略任务执行器
type StrategyWatch struct {
	data         *data.Data
	cronInstance *cron.Cron
	stopCh       chan struct{}
	entryIDMap   map[string]cron.EntryID
	interval     string
	timeout      time.Duration

	alertService *service.AlertService
}

// Start 启动策略任务执行器
func (s *StrategyWatch) Start(_ context.Context) error {
	if types.IsNil(s) || types.IsNil(s.cronInstance) {
		return merr.ErrorNotification("strategy watch is nil")
	}
	if types.IsNil(s.data) {
		return merr.ErrorNotification("data is nil")
	}
	if types.IsNil(s.data.GetStrategyQueue()) {
		return merr.ErrorNotification("strategy queue is nil")
	}
	go func() {
		defer after.RecoverX()
		for {
			select {
			case <-s.stopCh:
				return
			case msg, ok := <-s.data.GetStrategyQueue().Next():
				if !ok || !msg.GetTopic().IsStrategy() {
					continue
				}
				strategyMsg, ok := msg.GetData().(bo.IStrategy)
				if !ok {
					log.Warnf("strategy watch get data error: %v", msg.GetData())
					continue
				}
				if err := s.addJob(strategyMsg); err != nil {
					log.Errorw("add job err", err)
				}
			}
		}
	}()
	s.cronInstance.Start()
	log.Infof("[StrategyWatch] server started")
	return nil
}

// Stop 停止策略任务执行器
func (s *StrategyWatch) Stop(_ context.Context) error {
	defer log.Infof("[StrategyWatch] server stopped")
	close(s.stopCh)
	s.cronInstance.Stop()
	return nil
}

func (s *StrategyWatch) addJob(strategyMsg bo.IStrategy) error {
	// 删除策略任务
	if _, exist := s.entryIDMap[strategyMsg.Index()]; exist {
		log.Info("strategy watch remove job")
		s.cronInstance.Remove(s.entryIDMap[strategyMsg.Index()])
	}
	if !strategyMsg.GetStatus().IsEnable() {
		// 生成告警恢复事件（如果有告警发生过）
		return s.alertResolve(strategyMsg)
	}

	// 重新加入
	entryID, err := s.cronInstance.AddFunc(strategyMsg.GetInterval().CronTime(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		innerAlarm, err := s.alertService.InnerAlarm(ctx, strategyMsg)
		if err != nil {
			log.Warnw("inner alarm err", err)
			return
		}

		if err := s.data.GetAlertQueue().Push(innerAlarm.Message()); err != nil {
			log.Warnw("push inner alarm err", err)
			return
		}
	})
	if err != nil {
		return err
	}
	s.entryIDMap[strategyMsg.Index()] = entryID

	log.Infow("strategy watch add job", s.entryIDMap[strategyMsg.Index()])

	return nil
}

func (s *StrategyWatch) alertResolve(strategyMsg bo.IStrategy) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	innerAlarm, err := s.alertService.InnerAlarm(ctx, strategyMsg)
	if err != nil {
		log.Warnw("inner alarm err", err)
		return err
	}

	if err := s.data.GetAlertQueue().Push(innerAlarm.Message()); err != nil {
		log.Warnw("push inner alarm err", err)
		return err
	}
	return nil
}
