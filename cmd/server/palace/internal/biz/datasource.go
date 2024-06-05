package biz

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
	"github.com/aide-family/moon/pkg/utils/after"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewDatasourceBiz(
	datasourceRepository repository.Datasource,
	datasourceMetricRepository repository.DatasourceMetric,
	datasourceMetricMicroRepository microrepository.DatasourceMetric,
	lock repository.Lock,
) *DatasourceBiz {
	return &DatasourceBiz{
		lock:                            lock,
		datasourceRepository:            datasourceRepository,
		datasourceMetricRepository:      datasourceMetricRepository,
		datasourceMetricMicroRepository: datasourceMetricMicroRepository,
	}
}

// DatasourceBiz .
type DatasourceBiz struct {
	lock                            repository.Lock
	datasourceRepository            repository.Datasource
	datasourceMetricRepository      repository.DatasourceMetric
	datasourceMetricMicroRepository microrepository.DatasourceMetric
}

// CreateDatasource 创建数据源
func (b *DatasourceBiz) CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error) {
	return b.datasourceRepository.CreateDatasource(ctx, datasource)
}

// UpdateDatasourceBaseInfo 更新数据源
func (b *DatasourceBiz) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	return b.datasourceRepository.UpdateDatasourceBaseInfo(ctx, datasource)
}

// UpdateDatasourceConfig 更新数据源配置
func (b *DatasourceBiz) UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error {
	return b.datasourceRepository.UpdateDatasourceConfig(ctx, datasource)
}

// ListDatasource 获取数据源列表
func (b *DatasourceBiz) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	return b.datasourceRepository.ListDatasource(ctx, params)
}

// DeleteDatasource 删除数据源
func (b *DatasourceBiz) DeleteDatasource(ctx context.Context, id uint32) error {
	return b.datasourceRepository.DeleteDatasourceByID(ctx, id)
}

// GetDatasource 获取数据源详情
func (b *DatasourceBiz) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	return b.datasourceRepository.GetDatasource(ctx, id)
}

// GetDatasourceSelect 获取数据源下拉列表
func (b *DatasourceBiz) GetDatasourceSelect(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bo.SelectOptionBo, error) {
	list, err := b.datasourceRepository.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return types.SliceTo(list, func(item *bizmodel.Datasource) *bo.SelectOptionBo {
		return bo.NewDatasourceOptionBuild(item).ToSelectOption()
	}), nil
}

// UpdateDatasourceStatus 更新数据源状态
func (b *DatasourceBiz) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	// 需要校验数据源是否被使用， 是否有权限
	return b.datasourceRepository.UpdateDatasourceStatus(ctx, status, ids...)
}

// SyncDatasourceMeta 同步数据源元数据
func (b *DatasourceBiz) SyncDatasourceMeta(ctx context.Context, id uint32) error {
	syncDatasourceMetaKey := "sync:datasource:meta:" + strconv.FormatUint(uint64(id), 10)
	if err := b.lock.Lock(ctx, syncDatasourceMetaKey, 10*time.Minute); !types.IsNil(err) {
		if errors.Is(err, merr.ErrorI18nLockFailedErr(ctx)) {
			return merr.ErrorI18nRetryLaterErr(ctx).WithMetadata(map[string]string{
				"retry": merr.GetI18nMessage(ctx, "DATASOURCE_SYNCING"),
			})
		}
		return err
	}
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return merr.ErrorI18nUnLoginErr(ctx)
	}
	go func() {
		defer after.RecoverX()
		defer b.lock.UnLock(context.Background(), syncDatasourceMetaKey)
		log.Debugw("sync", "datasource meta", "id", id)
		if err := b.syncDatasourceMeta(context.Background(), id, claims.GetTeam()); err != nil {
			log.Errorw("err", err)
			return
		}
	}()

	return nil
}

// Query 查询数据
func (b *DatasourceBiz) Query(ctx context.Context, params *bo.DatasourceQueryParams) ([]*bo.DatasourceQueryData, error) {
	if types.IsNil(params.Datasource) {
		// 查询数据源
		datasourceDetail, err := b.datasourceRepository.GetDatasource(ctx, params.DatasourceID)
		if !types.IsNil(err) {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, merr.ErrorI18nDatasourceNotFoundErr(ctx)
			}
			return nil, err
		}
		params.Datasource = datasourceDetail
	}

	return b.datasourceMetricMicroRepository.Query(ctx, params)
}

func (b *DatasourceBiz) syncDatasourceMeta(ctx context.Context, id, teamId uint32) error {
	// 获取数据源详情
	datasourceDetail, err := b.datasourceRepository.GetDatasourceNoAuth(ctx, id, teamId)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nDatasourceNotFoundErr(ctx)
		}
		return err
	}
	// 获取元数据
	metadata, err := b.datasourceMetricMicroRepository.GetMetadata(ctx, datasourceDetail)
	if !types.IsNil(err) {
		return err
	}
	// 创建元数据
	if err = b.datasourceMetricRepository.CreateMetricsNoAuth(ctx, teamId, metadata...); !types.IsNil(err) {
		return err
	}
	return nil
}