package ctxext

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
)

// DefaultSingle 默认反并发处理
//    按 qq 号反并发
//    并发时返回 "您有操作正在执行，请稍后再试!"
var DefaultSingle = single.New(
	single.WithKeyFn(func(ctx *zero.Ctx) interface{} {
		return ctx.Event.UserID
	}),
	single.WithPostFn(func(ctx *zero.Ctx) {
		ctx.Send("您有操作正在执行，请稍后再试!")
	}),
)

// defaultLimiterManager 默认限速器管理
//    每 10s 5次触发
var defaultLimiterManager = rate.NewManager(time.Second*10, 5)

// LimitByUser 默认限速器 每 10s 5次触发
//    按 qq 号限制
func LimitByUser(ctx *zero.Ctx) *rate.Limiter {
	return defaultLimiterManager.Load(ctx.Event.UserID)
}

// LimitByGroup 默认限速器 每 10s 5次触发
//    按群号限制
func LimitByGroup(ctx *zero.Ctx) *rate.Limiter {
	return defaultLimiterManager.Load(ctx.Event.GroupID)
}

// LimiterManager 自定义限速器管理
type LimiterManager struct {
	m *rate.LimiterManager
}

// NewLimiterManager 新限速器管理
func NewLimiterManager(interval time.Duration, burst int) (m LimiterManager) {
	m.m = rate.NewManager(interval, burst)
	return
}

// LimitByUser 自定义限速器
//    按 qq 号限制
func (m LimiterManager) LimitByUser(ctx *zero.Ctx) *rate.Limiter {
	return m.m.Load(ctx.Event.UserID)
}

// LimitByGroup 自定义限速器
//    按群号限制
func (m LimiterManager) LimitByGroup(ctx *zero.Ctx) *rate.Limiter {
	return m.m.Load(ctx.Event.GroupID)
}