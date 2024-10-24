package biz

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// NewConfigBiz 创建配置业务
func NewConfigBiz(c *rabbitconf.Bootstrap, cacheRepo repository.CacheRepo) *ConfigBiz {
	return &ConfigBiz{
		cacheRepo: cacheRepo,
		c:         c,
	}
}

// GetConfigData 获取配置数据
func GetConfigData() *Config {
	if types.IsNil(configData) {
		configData = &Config{
			Receivers: make(map[string]*conf.Receiver),
			Templates: make(map[string]string),
		}
	}
	return configData
}

var configData = &Config{
	Receivers: make(map[string]*conf.Receiver),
	Templates: make(map[string]string),
}

// GetReceivers 获取接收人
func (l *Config) GetReceivers() map[string]*conf.Receiver {
	if types.IsNil(l) {
		return map[string]*conf.Receiver{}
	}
	return GetConfigData().Receivers
}

// GetTemplates 获取模板
func (l *Config) GetTemplates() map[string]string {
	if types.IsNil(l) {
		return map[string]string{}
	}
	return GetConfigData().Templates
}

// ConfigBiz 配置业务
type ConfigBiz struct {
	cacheRepo repository.CacheRepo
	c         *rabbitconf.Bootstrap
}

// Config 配置数据
type Config struct {
	Receivers map[string]*conf.Receiver `json:"receivers"`
	Templates map[string]string         `json:"templates"`
	sync.RWMutex
}

// Bytes json 序列化
func (l *Config) Bytes() []byte {
	l.Lock()
	defer l.Unlock()

	data, _ := types.Marshal(l)
	return data
}

// Set 设置接收人
func (l *Config) Set(ctx context.Context, cache cache.ICacher, params *bo.CacheConfigParams) error {
	for k, v := range params.Receivers {
		l.Receivers[k] = v
	}
	for k, v := range params.Templates {
		l.Templates[k] = v
	}
	return cache.Set(ctx, bo.CacheConfigKey, string(l.Bytes()), time.Hour)
}

// Get 获取接收人
func (l *Config) Get() *bo.CacheConfigParams {
	return &bo.CacheConfigParams{
		Receivers: l.Receivers,
		Templates: l.Templates,
	}
}

// CacheConfig 缓存配置
func (b *ConfigBiz) CacheConfig(ctx context.Context, params *bo.CacheConfigParams) error {
	return configData.Set(ctx, b.cacheRepo.Cacher(), params)
}

// LoadConfig 加载配置
func (b *ConfigBiz) LoadConfig(ctx context.Context) error {
	defer log.Debug("加载配置完成")
	params := &bo.CacheConfigParams{
		Receivers: make(map[string]*conf.Receiver),
		Templates: make(map[string]string),
	}
	getJSONStr, _ := b.cacheRepo.Cacher().Get(ctx, bo.CacheConfigKey)
	if !types.TextIsNull(getJSONStr) {
		if err := types.Unmarshal([]byte(getJSONStr), &params); !types.IsNil(err) {
			return err
		}
	}

	configData = &Config{
		Receivers: params.Receivers,
		Templates: params.Templates,
	}

	// 加载配置文件配置进内存
	yamlConfig := &bo.CacheConfigParams{
		Receivers: b.c.GetReceivers(),
		Templates: b.c.GetTemplates(),
	}

	return configData.Set(ctx, b.cacheRepo.Cacher(), yamlConfig)
}
