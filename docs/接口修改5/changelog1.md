# API 契约变更日志（第二轮）

> 本文档为简短说明，可能包含错误，主要以另外两个 API 文档（`apifox-import.json` / `api.md`）为主。
> 涉及接口：`GET /api/photos`、`GET /api/score/logs`、`GET /api/user/login`、`GET /api/user/logincallback`。
> **含破坏性改动**，详见第二节。

---

## 一、接口契约变更

### 1. `GET /api/photos`（题目列表）—— 排序枚举收敛为「创建时间 / 热度」**【破坏性】**

- Query 参数 `sort_by` 由 `created_at` / `likes_count` / `attempts_count` **三值收敛为两值**：`created_at`（创建时间）/ `hot`（热度）；同时补齐 `default: created_at`（原先仅 `api.md` 表格标注默认值，JSON 未声明）。
- `hot` 为**复合加权热度分**，默认公式 `likes_count × 2 + attempts_count × 1`。两项权重由后端配置，**调整权重不算契约变更**；客户端不感知具体权重，响应也**不下发**热度分字段。
- 两种排序均按降序排列，值相同时按 `id` 倒序保证稳定分页（口径不变）。
- 传入枚举外的值（含已废弃的 `likes_count`、`attempts_count`）返回 `400`、`code=3`。
- **不受影响**：评论列表 `GET /api/photos/{id}/comments` 的 `sort_by` 维持 `created_at` / `likes_count` 不变——评论没有作答维度，其「最热」就是点赞数本身。

### 2. `GET /api/score/logs`（积分流水）—— 新增累计总收入 / 总支出

响应 `resp` 中与 `total` **同级**（不在列表项内）新增两个**必返**非负整数字段：

| 字段            | 类型       | 说明                                                        |
| --------------- | ---------- | ----------------------------------------------------------- |
| `total_income`  | int（≥ 0） | 累计总收入：所有 `delta > 0` 的变动之和                     |
| `total_expense` | int（≥ 0） | 累计总支出：所有 `delta < 0` 的变动取**绝对值**之和，恒非负 |

- **全量口径**：统计当前用户的全部历史流水，不受分页影响，每页返回相同值（与 `unread_count` 的处理方式一致）。
- **符号约定**：`total_expense` 下发正数，前端展示时自行加负号，避免两端对符号理解不一致。
- `admin_adjust` 按其 `delta` 正负分别计入收入或支出，不单列。
- **对账恒等式**：`total_income - total_expense = GET /api/user/info` 的 `score_count`。
- 目的：积分页顶部的汇总展示，免去前端翻完所有分页再自行累加。

### 3. `GET /api/user/login`（登录）—— CAS 回调地址由后端改为指向前端页面

- **新增可选 Query 参数 `client`**（`enum: [fe, admin]`，默认 `fe`），指定认证完成后回跳的前端应用。回调地址映射如下，由部署侧配置，前端不可指定：

  | client  | 前端应用            | 登录回调页                             | 站点首页          |
  | ------- | ------------------- | -------------------------------------- | ----------------- |
  | `fe`    | C 端 tuxun-fe       | `{FE_ORIGIN}/subPages/auth/callback`   | `{FE_ORIGIN}/`    |
  | `admin` | B 端 tuxun-admin-fe | `{ADMIN_ORIGIN}/login/callback`        | `{ADMIN_ORIGIN}/` |

- **回调链路**：后端按 `client` 取出回调页地址作为 CAS 的 `service` → CAS 认证完成后携带一次性凭据 `guid` 重定向回**该前端页面** → 前端页面再以 **AJAX** 调用 `GET /api/user/logincallback?guid=...` 换取会话。
- 两端均为 history 路由（C 端由 hash 模式改为 history，URL 不再带 `#`），`guid` 以普通 query 参数追加到回调页地址上，即 `{FE_ORIGIN}/subPages/auth/callback?guid=xxx`；回调页原有 query 存在时用 `&` 拼接。
- 已有有效登录态时不走 CAS，直接重定向到**该 `client`** 的站点首页（原为固定的「应用首页」）。
- 明确两个响应：`302` 重定向；`400`（`code=3`）—— `client` 为枚举外的非法值，或该 `client` 未在部署侧配置回调地址，此时返回标准错误响应且**不执行任何重定向**。
- **不采用 `redirect_uri` 传任意 URL 的方案**：若允许前端传任意回跳地址，攻击者可构造 `/api/user/login?redirect_uri=https://evil.com/cb` 诱导用户完成认证，CAS 会把 `guid` 送到攻击者站点并被立即换成受害者会话，形成账号接管（`guid` 的一次性与 5 分钟有效期只能缩小窗口，挡不住即时兑换）。改用受限枚举加后端配置映射，从根上杜绝开放重定向；代价是新增端需要改后端配置。

### 4. `GET /api/user/logincallback`（登录回调）—— 补齐 400 声明与调用方式说明

- **补登 `400` 响应**（`StandardErrorResponse`）：`guid` 缺失、无效、已使用或已过期。原先该情况只写在接口描述里，`responses` 未声明，前端无法从契约得知需要处理该分支。
- 描述补充：本接口由前端登录回调页在拿到 CAS 回传的 `guid` 后**以 AJAX 方式调用**，不再由浏览器直接跳转访问。
- 补充跨域提示：前端与 API 非同源时，调用本接口需携带凭据且服务端 CORS 须精确回显 `Origin` 并允许凭据；或改用响应中的 `session_id`，后续请求走 `X-Session-Id` 请求头（与小程序同一套机制）。

---

## 二、破坏性改动清单

| 接口              | 改动                                                             | 影响                                                                 |
| ----------------- | ---------------------------------------------------------------- | -------------------------------------------------------------------- |
| `GET /api/photos` | `sort_by` 删除枚举值 `likes_count`、`attempts_count`，新增 `hot` | 旧客户端传这两个值将收到 `400`；两端均未上线，按可重写处理，不做兼容 |

其余三项均为**非破坏性**改动：新增可选请求参数（`client`）、新增响应字段（`total_income` / `total_expense`）、补充响应声明与描述。

---

## 三、后端实现与部署侧须知

1. **`client` → 回调页地址映射**需在部署侧配置（至少 `fe`、`admin` 两项）；未配置的 `client` 一律返回 `400`，**不得回退到默认地址后再跳转**。
2. **CAS 侧登记**：两个前端回调页地址需在学校统一认证侧登记为合法 `service`，否则 CAS 会拒绝该回调地址。若统一认证不支持登记多个 `service`，退化方案是后端保留自身回调地址、换取后再 `302` 到前端页面并附带 `guid`——链路对前端不变。
3. **前端 SPA 回退**：两端均为 history 路由，CAS 回跳是对回调页地址的整页加载，Nginx 等静态服务须把未知路径回退到 `index.html`，否则回调页直接 404、登录链路断在最后一步。
4. **CORS**：前端与 API 非同源时，`logincallback` 需允许携带凭据（精确回显 `Origin` + `Access-Control-Allow-Credentials`），或依赖 `session_id` + `X-Session-Id`。
5. **热度权重**做成配置项，不要硬编码；变更权重无需同步契约。
6. **汇总统计性能**：`total_income` / `total_expense` 是全量聚合，流水表增大后不宜每次请求全表扫描，建议维护用户级累计值或建立合适索引。
