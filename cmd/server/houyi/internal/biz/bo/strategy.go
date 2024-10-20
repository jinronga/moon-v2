package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*Strategy)(nil)

type (
	LabelNotices struct {
		// label key
		Key string `json:"key,omitempty"`
		// label value 支持正则
		Value string `json:"value,omitempty"`
		// 接收者 （告警组ID列表）
		ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
	}
	// Strategy 策略明细
	Strategy struct {
		// 接收者 （告警组ID列表）
		ReceiverGroupIDs []uint32 `json:"receiverGroupIDs,omitempty"`
		// 自定义接收者匹配对象
		LabelNotices []*LabelNotices `json:"labelNotices,omitempty"`
		// 策略ID
		ID uint32 `json:"id,omitempty"`
		// 策略等级ID
		LevelID uint32 `json:"levelId,omitempty"`
		// 策略名称
		Alert string `json:"alert,omitempty"`
		// 策略语句
		Expr string `json:"expr,omitempty"`
		// 策略持续时间
		For *types.Duration `json:"for,omitempty"`
		// 持续次数
		Count uint32 `json:"count,omitempty"`
		// 持续的类型
		SustainType vobj.Sustain `json:"sustainType,omitempty"`
		// 多数据源持续类型
		MultiDatasourceSustainType vobj.MultiDatasourceSustain `json:"multiDatasourceSustainType,omitempty"`
		// 策略标签
		Labels *vobj.Labels `json:"labels,omitempty"`
		// 策略注解
		Annotations vobj.Annotations `json:"annotations,omitempty"`
		// 执行频率
		Interval *types.Duration `json:"interval,omitempty"`
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
		// 策略采样率
		Step uint32 `json:"step,omitempty"`
		// 判断条件
		Condition vobj.Condition `json:"condition,omitempty"`
		// 阈值
		Threshold float64 `json:"threshold,omitempty"`
		// 团队ID
		TeamID uint32 `json:"teamId,omitempty"`
	}

	// Datasource 数据源明细
	Datasource struct {
		// 数据源类型
		Category vobj.DatasourceType `json:"category,omitempty"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storage_type,omitempty"`
		// 数据源配置 json
		Config map[string]string `json:"config,omitempty"`
		// 数据源地址
		Endpoint string `json:"endpoint,omitempty"`
		// 数据源ID
		ID uint32 `json:"id,omitempty"`
	}
)

func (s *Strategy) String() string {
	bs, _ := types.Marshal(s)
	return string(bs)
}

// Index 策略唯一索引
func (s *Strategy) Index() string {
	if types.IsNil(s) {
		return "houyi:strategy:0"
	}
	return fmt.Sprintf("houyi:strategy:%d:%d:%d", s.TeamID, s.ID, s.LevelID)
}

// Message 策略转消息
func (s *Strategy) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}
