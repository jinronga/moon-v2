package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameStrategyGroup = "strategy_group"

// StrategyGroup 告警规则组
type StrategyGroup struct {
	model.AllFieldModel
	Name       string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:规则组名称"`
	Status     vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用"`
	Remark     string      `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`
	Strategies []*Strategy `gorm:"foreignKey:GroupID"`
	Categories []*SysDict  `gorm:"many2many:strategy_group_categories" json:"categories"`
}

// UnmarshalBinary redis存储实现
func (c *StrategyGroup) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *StrategyGroup) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName Strategy's table name
func (*StrategyGroup) TableName() string {
	return tableNameStrategyGroup
}
