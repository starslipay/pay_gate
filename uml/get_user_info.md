# get_user_info 接口时序图

## 访问链路

```
用户 --> pay_gate --> user_mgr
```

## 时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant PG as pay_gate
    participant AM as AuthInterceptor<br/>(鉴权中间件)
    participant UM as user_mgr

    U->>PG: POST /api/pay_gate/get_user_info<br/>{user_id, user_token}

    Note over PG: 进入 AuthInterceptor 中间件

    PG->>AM: 解析请求参数
    AM->>AM: 提取 user_id / user_token
    alt user_token 为空
        AM-->>U: {code:101005, msg:"user_token is missing", data:null}
    else user_id 为空
        AM-->>U: {code:101007, msg:"user_id is missing", data:null}
    else 参数完整
        AM->>UM: CheckUserToken(user_id, user_token)
        alt RPC 调用失败
            UM-->>AM: error
            AM-->>U: {code:101006, msg:"user_token is invalid", data:null}
        else token 无效 (valid_status != 1)
            UM-->>AM: {valid_status: 0}
            AM-->>U: {code:101006, msg:"user_token is invalid", data:null}
        else token 有效
            UM-->>AM: {valid_status: 1}
            AM->>PG: 放行请求
        end
    end

    Note over PG: 进入 GetUserInfoHandler

    PG->>UM: GetUserInfo(user_id)
    alt RPC 调用失败
        UM-->>PG: error
        PG-->>U: {code:101001, msg:"RPC_ERROR:...", data:null}
    else 调用成功
        UM-->>PG: {user_id, name, gender, age, address, phone, email, id_type, id_card}
        PG-->>U: {code:0, msg:"success", data:{user_id, name, ...}}
    end
```

## 流程说明

| 步骤 | 参与者 | 操作 | 说明 |
|------|--------|------|------|
| 1 | 用户 | 发送 HTTP 请求 | POST `/api/pay_gate/get_user_info`，携带 `user_id` 和 `user_token` |
| 2 | pay_gate | 进入鉴权中间件 | `AuthInterceptor` 拦截请求 |
| 3 | 中间件 | 解析请求参数 | 从 JSON body 或表单中提取 `user_id`、`user_token` |
| 4 | 中间件 | 参数校验 | 校验 `user_token` 和 `user_id` 是否存在 |
| 5 | 中间件→user_mgr | 鉴权 RPC | 调用 `CheckUserToken` 校验 token 有效性 |
| 6 | 中间件 | 放行/拦截 | token 有效则放行，无效则返回错误 |
| 7 | pay_gate→user_mgr | 业务 RPC | 调用 `GetUserInfo` 获取用户信息 |
| 8 | pay_gate | 组装响应 | 将 user_mgr 返回的数据包装为统一格式 |
| 9 | pay_gate→用户 | 返回响应 | 返回 `{code, msg, data}` 格式的 JSON |

## 错误码说明

| 错误码 | 含义 | 触发场景 |
|--------|------|----------|
| 101005 | user_token 缺失 | 请求中未携带 `user_token` |
| 101007 | user_id 缺失 | 请求中未携带 `user_id` |
| 101006 | user_token 无效 | token 校验失败（RPC失败或 valid_status != 1） |
| 101001 | RPC 错误 | user_mgr 服务调用失败 |
