package biz

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/captcha"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// CaptchaBiz 验证码业务
type CaptchaBiz struct {
	captchaRepo repository.Captcha
}

// NewCaptchaBiz 创建验证码业务
func NewCaptchaBiz(captchaRepo repository.Captcha) *CaptchaBiz {
	return &CaptchaBiz{
		captchaRepo: captchaRepo,
	}
}

// GenerateCaptcha 生成验证码
func (l *CaptchaBiz) GenerateCaptcha(ctx context.Context, params *bo.GenerateCaptchaParams) (*bo.CaptchaItem, error) {
	id, base64s, err := captcha.CreateCode(ctx, params.Type, params.Theme, params.Size...)
	if !types.IsNil(err) {
		log.Warnw("fun", "captcha.CreateCode", "err", err)
		return nil, merr.ErrorI18nAlertCaptchaErr(ctx)
	}
	// 过期时间
	duration := time.Minute * 1
	expireAt := time.Now().Add(duration).Unix()
	validateCaptchaItem := bo.ValidateCaptchaItem{
		ValidateCaptchaParams: bo.ValidateCaptchaParams{
			ID:    id,
			Value: captcha.GetCodeAnswer(id),
		},
		ExpireAt: expireAt,
	}
	// 存储验证码信息到缓存
	if err = l.captchaRepo.CreateCaptcha(ctx, &validateCaptchaItem, duration); !types.IsNil(err) {
		log.Warnw("fun", "captchaRepo.CreateCaptcha", "err", err)
		return nil, merr.ErrorI18nAlertCaptchaErr(ctx)
	}
	return &bo.CaptchaItem{
		ValidateCaptchaItem: validateCaptchaItem,
		Base64s:             base64s,
	}, nil
}

// VerifyCaptcha 验证验证码
func (l *CaptchaBiz) VerifyCaptcha(ctx context.Context, params *bo.ValidateCaptchaParams) error {
	// 获取验证码信息
	validateCaptchaItem, err := l.captchaRepo.GetCaptchaByID(ctx, params.ID)
	if !types.IsNil(err) {
		log.Warnw("fun", "captchaRepo.GetCaptchaByID", "err", err)
		return merr.ErrorI18nAlertCaptchaErr(ctx)
	}
	// 验证码是否过期
	if time.Now().Unix() > validateCaptchaItem.ExpireAt {
		return merr.ErrorI18nAlertCaptchaExpire(ctx)
	}
	if validateCaptchaItem.Value != params.Value {
		return merr.ErrorI18nAlertCaptchaErr(ctx)
	}
	return nil
}
