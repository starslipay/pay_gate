// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package pay_gate

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/starslipay/account_mgr/account_mgr_pb"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/types"
	"github.com/starslipay/pay_gate/internal/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	QryModeSlave  = 1 // 从库查询
	QryModeMaster = 2 // 主库查询
)

// RedisKeyBalancePrefix 余额缓存 key 前缀, 与 account_mgr 写入侧保持一致: balance|{uid}
// value 形如: balance=1000000&balance_modify_time_ms=886&balance_modify_time_s=1792133820&cur_type=1
const RedisKeyBalancePrefix = "balance|"

type Get_user_balance_infoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGet_user_balance_infoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Get_user_balance_infoLogic {
	return &Get_user_balance_infoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *Get_user_balance_infoLogic) Get_user_balance_info(req *types.GetUserBalanceInfoReq) (resp *types.GetUserBalanceInfoRsp, err error) {
	// 优先读 Redis relation 缓存, 未命中或 Redis 不可用时降级 UserMgr 查 DB
	relationRsp, err := getRelationCached(l.ctx, l.svcCtx, req.UserId)
	if err != nil {
		return nil, err
	}

	// 优先查 Redis 余额缓存, 命中则直接返回, 省去一次 AccountMgr RPC + DB 查询
	// 缓存降级开关(Cache.Degrade)打开时跳过 Redis, 直接走下方 DB 查询
	if !l.svcCtx.Config.Cache.Degrade {
		if balance, curType, ok := l.getBalanceFromCache(relationRsp.Uid); ok {
			return &types.GetUserBalanceInfoRsp{
				UserId:  req.UserId,
				Balance: balance,
				CurType: curType,
			}, nil
		}
	}

	// 缓存未命中/Redis 不可用/降级开关打开, 降级查 DB
	getUserBalanceInfoRsp, err := l.svcCtx.AccountMgr.GetUserBalanceInfo(l.ctx, &account_mgr_pb.GetUserBalanceInfoReq{
		Uid:     relationRsp.Uid,
		QryMode: QryModeSlave, // 查询从库
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "AccountMgr.GetUserBalanceInfo")
	}

	return &types.GetUserBalanceInfoRsp{
		UserId:  req.UserId,
		Balance: getUserBalanceInfoRsp.Balance,
		CurType: getUserBalanceInfoRsp.CurType,
	}, nil
}

// getBalanceFromCache 读取 Redis 余额缓存, 命中返回余额与币种;
// key 不存在、value 解析失败或 Redis 不可用时返回 ok=false, 由调用方降级查 DB
func (l *Get_user_balance_infoLogic) getBalanceFromCache(uid int64) (balance int64, curType int32, ok bool) {
	key := RedisKeyBalancePrefix + strconv.FormatInt(uid, 10)
	cached, err := l.svcCtx.Redis.Get(l.ctx, key).Result()
	if err != nil {
		// key 不存在是正常未命中(不记错误); 其余错误说明 Redis 不可用, 记录后降级
		if !errors.Is(err, redis.Nil) {
			logx.WithContext(l.ctx).Errorf("getBalanceFromCache get balance cache failed, uid: %d, err: %v", uid, err)
		}
		return 0, 0, false
	}

	v, err := url.ParseQuery(cached)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("getBalanceFromCache parse balance cache failed, uid: %d, val: %s, err: %v", uid, cached, err)
		return 0, 0, false
	}
	balance, err = strconv.ParseInt(v.Get("balance"), 10, 64)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("getBalanceFromCache parse balance cache balance failed, uid: %d, val: %s, err: %v", uid, cached, err)
		return 0, 0, false
	}
	curTypeVal, _ := strconv.ParseInt(v.Get("cur_type"), 10, 32)
	return balance, int32(curTypeVal), true
}
