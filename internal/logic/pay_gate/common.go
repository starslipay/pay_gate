package pay_gate

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/starslipay/pay_gate/internal/svc"
	"github.com/starslipay/pay_gate/internal/xerr"
	"github.com/starslipay/user_mgr/user_mgr_pb"

	"github.com/zeromicro/go-zero/core/logx"
)

// RedisKeyRelationPrefix relation 缓存 key 前缀, 与 user_mgr 写入侧保持一致: relation|{user_id}
// value 为 URL 键值对串(含 uid), 形如: uid=10000019235
const RedisKeyRelationPrefix = "relation|"

// getRelationCached 获取 relation 映射(user_id -> uid): 优先读 Redis 缓存,
// 未命中或 Redis 不可用时降级调用 UserMgr.GetRelation 查 DB, 并将结果回填缓存
func getRelationCached(ctx context.Context, svcCtx *svc.ServiceContext, userId string) (*user_mgr_pb.GetRelationRsp, error) {
	if uid, ok := getRelationFromCache(ctx, svcCtx, userId); ok {
		return &user_mgr_pb.GetRelationRsp{
			UserId: userId,
			Uid:    uid,
		}, nil
	}

	// 缓存未命中或 Redis 不可用, 降级查 DB
	relationRsp, err := svcCtx.UserMgr.GetRelation(ctx, &user_mgr_pb.GetRelationReq{
		UserId: userId,
	})
	if err != nil {
		return nil, xerr.HandleRPCError(err, "UserMgr.GetRelation")
	}

	// 回源成功后回填缓存, 避免后续请求持续穿透到 UserMgr
	setRelationCache(ctx, svcCtx, userId, relationRsp.Uid)
	return relationRsp, nil
}

// getRelationFromCache 读取 Redis relation 缓存, 命中返回 uid;
// key 不存在、value 解析失败或 Redis 不可用时返回 ok=false, 由调用方降级查 DB
func getRelationFromCache(ctx context.Context, svcCtx *svc.ServiceContext, userId string) (uid int64, ok bool) {
	cached, err := svcCtx.Redis.Get(ctx, RedisKeyRelationPrefix+userId).Result()
	if err != nil {
		// key 不存在是正常未命中(不记错误); 其余错误说明 Redis 不可用, 记录后降级
		if !errors.Is(err, redis.Nil) {
			logx.WithContext(ctx).Errorf("getRelationFromCache get relation cache failed, user_id: %s, err: %v", userId, err)
		}
		return 0, false
	}

	v, err := url.ParseQuery(cached)
	if err != nil {
		logx.WithContext(ctx).Errorf("getRelationFromCache parse relation cache failed, user_id: %s, val: %s, err: %v", userId, cached, err)
		return 0, false
	}
	uid, err = strconv.ParseInt(v.Get("uid"), 10, 64)
	if err != nil {
		logx.WithContext(ctx).Errorf("getRelationFromCache parse relation cache uid failed, user_id: %s, val: %s, err: %v", userId, cached, err)
		return 0, false
	}
	return uid, true
}

// setRelationCache 回填 relation 缓存(与 user_mgr 写入侧同格式: uid={uid} 的 URL 键值对串);
// 回填失败不影响主流程, 仅记录错误日志
func setRelationCache(ctx context.Context, svcCtx *svc.ServiceContext, userId string, uid int64) {
	values := url.Values{}
	values.Set("uid", strconv.FormatInt(uid, 10))
	if err := svcCtx.Redis.Set(ctx, RedisKeyRelationPrefix+userId, values.Encode(), 0).Err(); err != nil {
		logx.WithContext(ctx).Errorf("setRelationCache write relation cache failed, user_id: %s, uid: %d, err: %v", userId, uid, err)
	}
}
