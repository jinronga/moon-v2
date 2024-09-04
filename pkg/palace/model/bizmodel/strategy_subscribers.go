package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategySubscribers = "strategy_subscribers"

// StrategySubscribers 策略订阅者信息
type StrategySubscribers struct {
	model.AllFieldModel
	Strategy        *Strategy       `gorm:"foreignKey:StrategyID" json:"strategy"`
	AlarmNoticeType vobj.NotifyType `gorm:"column:notice_type;type:int;not null;comment:通知类型;" json:"alarm_notice_type"`
	UserID          uint32          `gorm:"column:user_id;type:int;not null;comment:订阅人id;uniqueIndex:idx__strategy_subscriber_user_id,priority:1" json:"user_id"`
	StrategyID      uint32          `gorm:"column:strategy_id;type:int;comment:告警分组id;uniqueIndex:idx__strategy_subscriber_user_id,priority:2" json:"strategy_id"`
}

// UnmarshalBinary redis存储实现
func (c *StrategySubscribers) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategySubscribers) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName StrategySubscribers  table name
func (*StrategySubscribers) TableName() string {
	return tableNameStrategySubscribers
}