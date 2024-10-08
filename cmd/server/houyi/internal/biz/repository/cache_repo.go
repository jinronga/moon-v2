package repository

import (
	"github.com/aide-family/moon/pkg/plugin/cache"
)

// CacheRepo 换成统一repo
type CacheRepo interface {
	// Cacher 获取缓存实现
	Cacher() cache.ICacher
}
