# 图寻 API 文档

> 基础路径：`/api`

统一响应格式中的 `success`、`resp`、`message` 和 `code` 均为必返回字段。

成功响应：

```json
{ "success": true, "resp": {}, "message": "", "code": 0 }
```

失败响应：

```json
{ "success": false, "resp": null, "message": "参数错误: 请求参数不合法", "code": 3 }
```

---

## 目录

- [鉴权](#鉴权)
- [约定](#约定)
- [契约维护规则](#契约维护规则)
- [公共数据结构](#公共数据结构)
- [用户认证](#用户认证)
- [活动](#活动)
- [图寻题目 (Photos)](#图寻题目-photos)
- [作答 (Attempts)](#作答-attempts)
- [评论 (Comments)](#评论-comments)
- [我的记录 (My Records)](#我的记录-my-records)
- [积分 (Score)](#积分-score)
- [奖品 (Goods)](#奖品-goods)
- [兑换 (Exchange)](#兑换-exchange)
- [通知 (Notifications)](#通知-notifications)
- [内容位 (Contents)](#内容位-contents)
- [反馈 (Feedback)](#反馈-feedback)
- [管理员](#管理员)
- [状态枚举参考](#状态枚举参考)

---

## 鉴权

| Level | 说明                                 |
| ----- | ------------------------------------ |
| 1     | 客户端普通登录用户，仅用于 tuxun-fe  |
| 2     | 后台管理员，可进入 tuxun-admin-fe    |
| 3     | 超级管理员，由部署侧管理，可存在多个 |

**Session 鉴权与生命周期**：

- **双凭据模式**：网关与服务端同时支持 Cookie `tz-sessions`（Web H5 端）与请求头 `X-Session-Id: <session_id>`（小程序/跨域端）两种凭据透传方式。两者在 `apifox-import.json` 中分别登记为 `cookieAuth` 与 `sessionIdAuth`，需鉴权的 operation 同时声明两者（OR 关系，任一有效即可）；公开接口显式声明 `security: []`。
- **Cookie 安全防护**：生产环境 `tz-sessions` Cookie 强制启用 `HttpOnly`、`Secure` 和 `SameSite=Lax`。基于 Cookie 的写操作仍需 CSRF 防护；通过 `X-Session-Id` 头透传的请求不依赖浏览器 Cookie 机制，天然免疫 CSRF 攻击。
- **滑动续期机制**：Session 有效期为 14 天。当 Session 剩余有效时间低于 50%（7 天）时，随有效 API 请求自动重置 TTL 为 14 天；14 天内有任意使用即可延续登录态。
- **过期处理**：连续 14 天无请求导致 Session 过期的，返回 401（`code=6`），需重新登录。

**登录态、权限等级与账号状态相互独立**：未登录是当前请求没有有效 Session，不写入用户表，也没有 `level` 值（无需登录即可访问的公开接口，在各自的**权限**处标注为「无」）。账号 `level` 只允许 `1`（普通用户）、`2`（管理员）、`3`（超级管理员）；`status` 独立为 `active` / `banned`。封禁或解封不改变 `level`，调整 `level` 也不改变 `status`。

**用户治理仅限超级管理员（Level 3）**：用户列表、封禁/解封和权限等级调整接口都只有 Level 3 可以调用；普通管理员（Level 2）只负责内容审核与运营，不接触用户账号。

| 操作者           | 可封禁 / 解封的目标 | 可调整权限等级的目标               |
| ---------------- | ------------------- | ---------------------------------- |
| Level 1、Level 2 | 无                  | 无                                 |
| Level 3          | Level 1、Level 2    | Level 1、Level 2，可在两级之间调整 |

Level 3 可以有多个，但任何 Level 3 都不能治理自己或其他 Level 3。Level 3 账号的创建、修改和删除，以及其状态或等级变更，只能通过部署侧受控流程完成；业务用户列表可只读展示其摘要。所有账号治理操作均应记录操作者、目标账号、变更前后值和时间。

**HTTP 状态码语义**：下表统一说明状态码含义；每个接口实际承诺返回的状态码以 `apifox-import.json` 对应 operation 的 `responses` 为准。

| 状态码 | 场景                                                                     |
| ------ | ------------------------------------------------------------------------ |
| 200    | 查询、更新、删除等操作成功                                               |
| 201    | 资源创建成功                                                             |
| 400    | 请求参数无效或业务前置条件不满足（如库存、积分不足）                     |
| 401    | 未登录或 Session 失效                                                    |
| 403    | 权限不足                                                                 |
| 404    | 路径 `{id}` 定位的资源不存在（`code=5`）；请求体引用的对象不存在用 `400` |
| 409    | 冲突（重复审核、重复核销、幂等键冲突、并发冲突）                         |
| 429    | 超出接口声明的修改次数 / 频率限制（`code=9`）                            |
| 500    | 未预期的服务器内部错误                                                   |

**错误码（`code`）**：

| code | message 前缀 | 说明                    |
| ---- | ------------ | ----------------------- |
| 0    | —            | 成功                    |
| 3    | 参数错误     | 请求参数不合法          |
| 4    | 系统错误     | 服务器内部错误          |
| 5    | 操作错误     | 业务逻辑限制            |
| 6    | 鉴权错误     | 未登录                  |
| 7    | 权限错误     | 权限不足                |
| 8    | 冲突错误     | 重复操作 / 并发冲突     |
| 9    | 频率限制     | 超出修改次数 / 频率限制 |

---

## 约定

- 基础路径：默认使用当前部署同源的 `/api`；本地开发可使用 `http://127.0.0.1:8088/api`。
- 字段命名：JSON 字段与 Query 参数统一使用 `snake_case`（如 `page_size`、`activity_id`）。
- 统一响应：所有 JSON 响应固定包含 `success`、`resp`、`message`、`code` 四个必返字段。
- 分页规范：`page` 默认 1（min=1），`page_size` 默认 10（min=1, max=20）。分页响应中的 `total` 为经过权限、审核与筛选条件过滤后的实际匹配总条数；分页数组字段统一命名为 `list`，空列表返回 `[]`。
- 时间与空值：纯日期格式为 `YYYY-MM-DD`；具体时刻为 ISO 8601 带时区格式（如 `2026-07-20T10:30:00+08:00`）。所有可空字段（时间、可选配图、可选关联对象、可选附件等）一律返回 `null`，不省略字段，客户端只需判 `null`。唯一例外：`Media` 的 `origin_url` / `thumb_url` 与 `FeedbackMedia` 的 `thumb_url` 是「按接口下发口径省略」而非置 `null` 的字段（详见下一条），这三个字段均为非 `nullable` 的 `string`，下发 `null` 会违反 schema。客户端状态判断统一以 HTTP 响应头 `Date` 标头的服务器时间为基准；服务端必须配置 `Access-Control-Expose-Headers: Date` 响应头，确保 H5 跨域场景下 JavaScript 能够读取该标头。
- 媒体与图片规范：活动封面、题目等统一下发标准对象 `Media`，用户头像 `avatar` 保持字符串。`Media` 含原图 `origin_url` 与缩略图 `thumb_url`，每处至少下发其一，具体以各接口响应示例为准；不下发的那个整个字段缺省，不要下发 `null`。`width` / `height` 为原图真实像素，必返且 ≥ 1（`minimum: 1`，防止客户端算宽高比时除零）；缩略图须与原图等比，客户端用这一组宽高为两种规格预留占位框、消除加载抖动（CLS）。历史图片迁移时回读原图回填真实尺寸，无法读取才回退占位值。`FeedbackMedia` 同理：`thumb_url` 取不到时省略字段，而 `width` / `height` 是 `nullable`、视频取不到尺寸时返回 `null`。
- Multipart 参数宽松解析：`multipart/form-data` 请求中的数值与布尔表单项（如 `longitude`、`latitude`、`activity_id`），服务端统一进行宽松类型自动转换（如将 `"108.123"` 解析转换为 `float`）。
- 通用文件上传：图片上传位仅支持 `jpg/png` 格式，单文件大小 `≤ 20MB`。
- 通用字符串限制：所有请求中的 `title` 标题字段按 Unicode 码点计数最长 20 个字符；搜索关键词最长 50 个字符。
- 资源与引用不存在：所有以路径 `{id}` 定位资源的接口，目标不存在时统一返回 `404`（`code=5`）；请求体或参数引用的外键对象不存在时返回 `400`（`code=5`）。
- 未登录默认值：点赞状态 `liked`（题目 / 破解记录 / 评论）在未登录访问时恒为 `false`；题目列表与详情的个人字段 `solved`（本人是否已破解）、题目详情的 `user_attempts_count`（本人作答次数）未登录时分别为 `false`、`0`。以本人为口径的筛选参数（如 `GET /photos` 的 `solved`）在未登录时按该默认值处理，不额外报错。个人记录类接口（我的投稿 / 我的作答）不返回任何点赞字段。

---

## 契约维护规则

本节约束前后端双方对契约文档（`apifox-import.json` 与 `api.md`）的一切后续修改。

**事实源与流程**

1. `apifox-import.json` 是唯一权威契约，`api.md` 是其同步的人类可读版本。任何契约改动必须在同一次提交中同时更新两个文件，只改其一视为未完成。
2. 契约只在双方约定的对接分支上维护（当前为 `docs/api-contract-alignment`）；其他分支和各处副本仅作镜像，不作为对接依据。
3. 先改契约、后改代码。接口行为的增删改必须先落在契约文档中，再进行前后端实现，不允许实现先行、文档事后追认。
4. 提交说明须列出涉及的接口，并标明是否为破坏性改动。

**改动分级**

1. **非破坏性改动**——任一方可直接提交，提交后通知对方：新增接口；新增可选请求参数或请求字段；新增响应字段；放宽校验；修正描述、示例和笔误。
2. **破坏性改动**——必须先与对方沟通并确认，再提交：删除或重命名接口、参数、字段；修改字段类型、必填性、枚举取值；收紧校验；变更状态码、错误码或响应结构的语义。
3. 前端上线后，破坏性改动还须附旧客户端过渡方案；两端都未上线期间按可重写对待，不需要兼容。

**风格守则**（新增或修改接口时必须沿用既有约定，不引入第二种风格）

1. 字段与 Query 参数一律 `snake_case`；路径参数用 `{param}` 风格；分页一律 `page` / `page_size`，响应一律 `{ "list": [], "total": 0 }`；所有 JSON 响应必含 `success` / `resp` / `message` / `code`。
2. 状态码和错误码遵循本文档前面两张表；新增错误码必须先登记进错误码表。
3. 公共业务对象必须复用或扩展 `components.schemas` 中的既有 schema（分页列表用 `allOf` 组合 `PageBase`），不得内联重复定义；可枚举字段必须写 `enum`；必返字段必须列入 `required`；每个 operation 必须有唯一 `operationId`。
4. OpenAPI 3.0 的两个硬性写法（写错不报错，但语义会静默丢失）：
   - `$ref` 的**同级键一律被忽略**。要给引用挂 `description` / `nullable`，必须写成 `allOf: [{ "$ref": ... }]` 再在外层挂键，不能直接和 `$ref` 平级。
   - `nullable: true` **不会自动放行 `enum`**。可空的枚举字段必须把 `null` 显式写进 `enum` 数组（如 `enum: ["activity", null]`），否则校验器会把契约要求必返的 `null` 判为非法。
5. 时间格式和媒体 URL 遵循[约定](#约定)一节。

**校验**

1. 每次修改后、提交前，在 `tu-xun` 根目录运行 `python3 check_contract.py`（或 `uv run python check_contract.py`），全部通过才允许提交。脚本校验 JSON 合法性、`operationId` 唯一性、`$ref` 可解析性，以及 `api.md` 与 `apifox-import.json` 接口清单的双向一致。

---

## 公共数据结构

OpenAPI 中的公共结构统一定义在 `components.schemas`，各接口通过 `$ref` 或 `allOf` 引用，避免同一业务对象在不同接口中产生独立类型。

| Schema                                                    | 用途                                                                                                                                                                                                |
| --------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SuccessResponseBase` / `ErrorResponseBase`               | 统一 `success`、`message`、`code` 响应字段                                                                                                                                                          |
| `StandardErrorResponse`                                   | 标准错误响应，继承 `ErrorResponseBase` 且 `resp` 固定为 `null`                                                                                                                                      |
| `Media`                                                   | 标准图片对象：`origin_url`（原图）、`thumb_url`（缩略图）至少下发其一，`width` / `height` 为原图真实像素且缩略图与之等比；头像为字符串                                                              |
| `FeedbackMedia`                                           | 反馈附件媒体对象（`origin_url`、`thumb_url`、`width`、`height`、`media_type`），`media_type` 区分图片（1）与视频（2）；`width` / `height` 同样最小为 1（`minimum: 1`），视频取不到尺寸时可为 `null` |
| `UserBrief`                                               | 内容作者等简要用户信息                                                                                                                                                                              |
| `UserSummary`                                             | 登录和管理员用户列表使用的完整用户摘要（不含 `session_id`，防止批量泄漏）                                                                                                                           |
| `LoginResult`                                             | 登录与登录回调接口返回的结果，在 `UserSummary` 基础上增加 `session_id` 供跨端鉴权                                                                                                                   |
| `ActivityBrief`                                           | 图片关联的简要活动信息，包含 `id`、`title`、`start_time`、`end_time`                                                                                                                                |
| `ActivityCard`                                            | 进行中活动、历史活动和管理员活动列表共用的卡片结构，含 `photo_count` 题目数                                                                                                                         |
| `GoodBrief`                                               | 兑换记录中的奖品摘要                                                                                                                                                                                |
| `GoodItem`                                                | 客户端与管理端奖品列表共用的完整奖品项                                                                                                                                                              |
| `LikeResult`                                              | 题目、破解记录和评论点赞接口的统一结果                                                                                                                                                              |
| `AnnouncementListItem`                                    | 通知列表摘要项，含 `content_preview`                                                                                                                                                                |
| `Announcement`                                            | 通知详情结构，含完整正文、可选配图和可选活动关联                                                                                                                                                    |
| `InteractionMessage`                                      | 互动消息列表项，含发送者摘要、关联对象及 `photo_id`                                                                                                                                                 |
| `AdminPhotoListItem`                                      | 管理端题目列表项，包含所属活动、作者、审核状态、计数（作答/成功/点赞）和编辑所需字段                                                                                                                |
| `Location`                                                | 地理坐标对象，包含 `longitude`、`latitude`、`coord_type`                                                                                                                                            |
| `AttemptRecord`                                           | 我在本图的作答列表项（包含原图与缩略图的图片对象 `image`、猜测坐标 Location、判定状态 `status`、驳回原因 `reject_reason`、时间）                                                                    |
| `UserAttemptCard`                                         | 我的作答记录（跨题目聚合）列表项，包含该题我的已作答次数 `user_attempts_count`、作答判定状态与题目摘要 `photo`                                                                                      |
| `IdResult`                                                | 写操作回执中的公共 `id` 字段，各接口自行限制 `status` 枚举                                                                                                                                          |
| `ReviewActionRequest`                                     | 题目和评论审核的统一请求结构                                                                                                                                                                        |
| `GoodUpsertForm`                                          | 奖品新增和编辑的公共表单字段                                                                                                                                                                        |
| `PageBase`                                                | 所有分页结构共享的 `total` 字段                                                                                                                                                                     |
| `CreateAnnouncementRequest` / `UpdateAnnouncementRequest` | 管理员以 multipart 发布或编辑通知、可选配图与活动关联                                                                                                                                               |
| `ContentBlock` / `UpdateContentRequest`                   | 弹窗、积分规则、帮助中心共用的单例富文本内容位及其更新请求                                                                                                                                          |
| `AdminPhotoUpsertForm`                                    | 管理员新增或编辑题目的 multipart 表单                                                                                                                                                               |
| `SetLikeRequest`                                          | 题目、破解记录和评论共用的幂等点赞状态请求                                                                                                                                                          |
| `PhotoCardBase`                                           | 题目卡片公共字段：`id`、`title`、`image`、`created_at`                                                                                                                                              |
| `PhotoCard`                                               | 首页题目卡片，在 `PhotoCardBase` 上增加作者、点赞信息与个人破解状态 `solved`                                                                                                                        |
| `UserPhotoCard`                                           | 我的投稿卡片，在 `PhotoCardBase` 上增加审核状态（不含驳回原因，详见投稿详情）                                                                                                                       |
| `SolveRecord`                                             | 公开破解记录列表项：破解者、破解图片与点赞信息                                                                                                                                                      |
| `CommentItem`                                             | 公开评论列表项：评论者、正文与点赞信息                                                                                                                                                              |
| `ScoreLog`                                                | 积分流水列表项：变动值、变动后余额、变动原因与关联对象                                                                                                                                              |
| `ExchangeRecord` / `AdminExchangeRecord`                  | 客户端与管理端兑换记录列表项，管理端在客户端结构上增加兑换用户摘要                                                                                                                                  |
| `AdminFeedbackListItem`                                   | 管理端反馈列表项：提交者、标题、类型与处理状态                                                                                                                                                      |

分页结构以 `PageBase` 复用分页元数据，并按列表项类型分别定义 `ActivityCardPage`、`PhotoCardPage`、`UserPhotoCardPage`、`UserSummaryPage`、`UserAttemptCardPage`、`GoodItemPage`、`AnnouncementPage`、`InteractionMessagePage`、`AdminPhotoListPage`、`AdminAttemptListPage`、`AdminCommentListPage` 和 `AdminAnnouncementPage`。其他列表响应同样通过 `allOf` 组合 `PageBase` 和明确的 `list[]` schema，不使用会丢失列表项类型的无类型分页模型。

### Location

| 字段       | 类型           | 说明                                   |
| ---------- | -------------- | -------------------------------------- |
| longitude  | number(double) | 经度（-180.0 ~ 180.0）                 |
| latitude   | number(double) | 纬度（-90.0 ~ 90.0）                   |
| coord_type | string         | 坐标系类型：`wgs84` / `gcj02` / `bd09` |

### ActivityCard

| 字段        | 类型              | 说明                                              |
| ----------- | ----------------- | ------------------------------------------------- |
| id          | int               | 活动 ID                                           |
| title       | string            | 活动标题                                          |
| cover_image | Media             | 活动封面                                          |
| description | string            | 活动简介                                          |
| start_time  | string(date-time) | 必返非空的带时区开始时间                          |
| end_time    | string(date-time) | 必返非空的带时区结束时间，且必须晚于 `start_time` |
| photo_count | int               | 活动包含的题目数                                  |

活动契约不返回 `status` 或 `is_active` 等派生状态字段。后端必须使用服务器当前时间按以下唯一规则进行列表归类和写操作校验：

- `now < start_time`（`status=not_started`）：仅在管理端活动列表可见，不出现在客户端活动列表；题目不向客户端下发，也不允许投稿和作答；
- `start_time <= now < end_time`（`status=active`）：客户端与管理端均可见，允许投稿和作答；
- `now >= end_time`（`status=ended`）：客户端与管理端均可见，不再允许新投稿和作答。

`start_time` 和 `end_time` 在新建、更新及读取时均为必填非空字段，并且必须满足 `end_time > start_time`。时间比较按 ISO 8601 时刻进行。

活动一旦达到原 `end_time`，题目答案可能已经按活动结束规则公开；后续更新不得把 `end_time` 延长到服务器当前时间之后重新开放活动。需要提前结束时可以将 `end_time` 调到当前时间之前，但结束状态不可撤回。

---

## 用户认证

### 开发/测试登录

```
GET /api/test/login
```

**说明**：仅在开发和测试环境开启，生产环境必须关闭。允许通过 NetID 模拟指定账号一键登录。

**请求参数（Query）**

| 参数     | 类型   | 必填 | 说明               |
| -------- | ------ | ---- | ------------------ |
| netid    | string | 是   | 要登录用户的 NetID |
| password | string | 是   | 测试登录密码       |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "测试用户",
    "nickname": "测试用户",
    "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "level": 1,
    "status": "active",
    "session_id": "sess_abc123456789"
  },
  "message": "",
  "code": 0
}
```

**失败** `403`：账号已被封禁，拒绝建立会话（`code=7`）。

---

### 1. 登录回调

```
GET /api/user/logincallback
```

**权限**：无

**说明**：登录跳转由前端发起——前端重定向到学校统一身份认证页面，`redirect_uri` 指向自己的登录回调页；认证完成后认证服务携带一次性凭据 `code` 重定向回该页面，回调页再以 AJAX 调用本接口换取会话。后端不提供登录重定向接口。

后端用 `code` 与 `redirect_uri` 向学校认证服务换取用户身份，创建或更新本地用户记录并写入登录态（`tz-sessions` Cookie），返回 `LoginResult`（即 `UserSummary` 加上用于跨端鉴权的 `session_id`；`UserSummary` 本身不含该字段）。小程序可通过 `<web-view>` 登录获取该 `session_id` 并存入本地，后续请求通过 `X-Session-Id` Header 传递凭据。

`redirect_uri` 须与前端授权阶段传给认证服务的值完全一致，并命中后端环境变量配置的回调页白名单（C 端、B 端各一项）；不一致或不在白名单内返回 `400`，不执行换取。因两端回调页地址不同，该值由前端传入而非后端写死单一值，但可取范围仍受后端白名单约束，前端无法指定任意地址。`code` 为一次性凭据，换取成功后立即作废，有效期不超过 5 分钟（`code` 会出现在 URL 与网关访问日志中，避免被重放换取会话）。

**请求参数（Query）**

| 参数         | 类型   | 必填 | 说明                                             |
| ------------ | ------ | ---- | ------------------------------------------------ |
| code         | string | 是   | 学校统一认证回调返回的一次性凭据                 |
| redirect_uri | string | 是   | 授权阶段使用的回调页地址，须一致且在后端白名单内 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "张三",
    "nickname": "张三",
    "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "level": 1,
    "status": "active",
    "session_id": "sess_abc123456789"
  },
  "message": "",
  "code": 0
}
```

**失败** `400`：`code` 缺失、无效、已使用或已过期，或 `redirect_uri` 与授权阶段不一致、不在后端白名单内（`code=3`）。

**失败** `403`：账号已被封禁，拒绝建立会话（`code=7`）。

**跨域**：前端与 API 非同源时，调用本接口需携带凭据且服务端 CORS 须精确回显 `Origin` 并允许凭据；或改用响应中的 `session_id`，后续请求走 `X-Session-Id` 请求头（与小程序同一套机制）。

**部署前提**：两端回调页地址需同时登记在学校统一认证侧的合法 `redirect_uri` 与后端环境变量白名单中。两端均为 history 路由，认证服务回跳是对回调页地址的整页加载，静态服务须把未知路径回退到 `index.html`，否则回调页直接 404、登录链路断在最后一步。

---

### 2. 登出

```
DELETE /api/user/logout
```

**权限**：无（清除 Session）

**返回** `200`

```json
{
  "success": true,
  "resp": null,
  "message": "",
  "code": 0
}
```

---

### 3. 获取个人信息

```
GET /api/user/info
```

**权限**：登录用户（Level ≥ 1）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "张三",
    "nickname": "张三",
    "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "score_count": 1000,
    "level": 1,
    "nickname_edits_remaining": 3,
    "avatar_edits_remaining": 9
  },
  "message": "",
  "code": 0
}
```

**说明**：`nickname_edits_remaining` / `avatar_edits_remaining` 为本月剩余的昵称 / 头像修改次数（上限分别为 4 / 10，按服务器时区自然月初重置），供前端在资料页展示与禁用按钮；额度耗尽后对应修改接口返回 `429`（`code=9`）。

---

### 4. 修改昵称

```
PUT /api/user/nickname
```

**权限**：登录用户（Level ≥ 1）

**说明**：当前仅支持修改昵称 `nickname`（头像修改见 `PUT /api/user/avatar`）。昵称每自然月最多修改 **4 次**（按服务器时区，自然月初重置）。仅当请求携带 `nickname` 且与当前昵称不同时计 1 次，且成功修改才计数；提交相同昵称不计数、不报错。修改成功后返回新昵称 `nickname` 与本月剩余修改次数 `nickname_edits_remaining`。

**请求参数（JSON Body）**

| 参数     | 类型   | 必填 | 说明            |
| -------- | ------ | ---- | --------------- |
| nickname | string | 否   | 昵称（最长 10） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "nickname": "新昵称",
    "nickname_edits_remaining": 3
  },
  "message": "",
  "code": 0
}
```

超出本月修改次数时返回 `429`：

```json
{
  "success": false,
  "resp": null,
  "message": "频率限制: 本月昵称修改次数已达上限",
  "code": 9
}
```

---

### 5. 修改头像

```
PUT /api/user/avatar
```

**权限**：登录用户（Level ≥ 1）

**Content-Type**：`multipart/form-data`

**跨端传输重写**：因小程序与移动端（如 `uni.uploadFile`）无法直接发起 HTTP `PUT` 的文件上传，客户端在修改头像时可通过 HTTP `POST` 发起请求并携带请求头 `X-HTTP-Method-Override: PUT`，服务端按 `PUT` 语义处理。

**说明**：头像每自然月最多修改 **10 次**（按服务器时区，自然月初重置），成功修改才计数。文件类型或大小不符合要求时返回 `400`（`code=3`）。修改成功后返回新头像地址 `avatar` 与本月剩余修改次数 `avatar_edits_remaining`。

**请求参数**

| 参数   | 类型 | 必填 | 说明                       |
| ------ | ---- | ---- | -------------------------- |
| avatar | file | 是   | 头像文件（jpg/png，≤20MB） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "avatar_edits_remaining": 9
  },
  "message": "",
  "code": 0
}
```

超出本月修改次数时返回 `429`：

```json
{
  "success": false,
  "resp": null,
  "message": "频率限制: 本月头像修改次数已达上限",
  "code": 9
}
```

---

## 活动

### 活动列表

```
GET /api/activity
```

**权限**：无

**说明**：客户端活动列表，**只返回进行中（`active`）和已结束（`ended`）活动，不含未开始活动**——未开始活动仅在管理端活动列表可见，其题目也不向客户端下发、不允许投稿作答。接口不返回状态字段，客户端按 `start_time` / `end_time` 与当前时间判断；`status` 筛选由后端按服务器当前时间计算。默认按 `start_time` 倒序、`id` 倒序返回，保证稳定分页。活动结束时不迁移题目，题目始终保留原 `activity_id`。多个筛选条件按 AND 组合。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                                    |
| --------- | ------ | ---- | ------ | ------------------------------------------------------- |
| page      | int    | 否   | 1      | 页码（min=1）                                           |
| page_size | int    | 否   | 10     | 每页数量（min=1, max=20）                               |
| status    | string | 否   | —      | `active`（进行中）/ `ended`（已结束）；不传返回两者全部 |
| keyword   | string | 否   | —      | 按标题或描述文字模糊搜索（最长 50），不支持按 ID        |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 2,
    "list": [
      {
        "id": 1,
        "title": "寻找校园角落",
        "cover_image": {
          "thumb_url": "https://media.example.com/photos/cover.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "description": "活动介绍",
        "start_time": "2026-06-01T00:00:00+08:00",
        "end_time": "2026-06-30T23:59:59+08:00",
        "photo_count": 12
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

## 图寻题目 (Photos)

题目（图片）的浏览与投稿。可点赞对象全库仅三类：**题目、破解记录、评论**；个人记录类接口不返回点赞字段。

### 1. 题目列表

```
GET /api/photos
```

**权限**：无

**说明**：全站题目列表与搜索接口。不传任何活动筛选时，默认涵盖所有进行中与已结束活动的题目，支持搜索与展示往期题目。`activity_status` 按所属活动状态筛选（首页「进行中」场景传 `active`，往期场景传 `ended`）；`activity_id` 精确筛选单个活动；两者同时提供时按 AND 组合，若该活动不满足所选状态则返回空列表、不报错。活动状态由后端按服务器当前时间计算，口径与 `GET /api/activity` 一致（`start_time <= now < end_time` 为 `active`）；尚未开始的活动不对客户端暴露题目，任何筛选下都返回空列表。

**注意**：Query 参数 `solved` 与列表项字段 `solved` 语义一致，均指**当前登录用户本人是否已破解**该题，用于"只看我没做过的题"这类筛选。未登录时 `solved` 恒为 `false`，因此 `solved=true` 返回空列表、`solved=false` 等价于返回全部。管理端 `GET /admin/photos` 的同名参数是全站口径（`solved_count > 0`），两者不通用。

**排序口径**：`hot` 为复合加权热度分，默认公式 `likes_count × 2 + attempts_count × 1`，两项权重由后端配置，调整权重不算契约变更；客户端不感知具体权重，响应也不下发热度分字段。传入枚举外的值返回 `400`、`code=3`。

**请求参数（Query）**

| 参数            | 类型   | 必填 | 默认值     | 说明                                                                                                             |
| --------------- | ------ | ---- | ---------- | ---------------------------------------------------------------------------------------------------------------- |
| activity_id     | int    | 否   | —          | 按单个活动筛选；不传时涵盖所有进行中与已结束活动的题目                                                           |
| activity_status | string | 否   | —          | 按所属活动状态筛选：`active`（进行中）/ `ended`（已结束）；不传返回两者全部。与 `activity_id` 按 AND 组合        |
| solved          | bool   | 否   | —          | 按**当前登录用户本人**是否已破解筛选：`true` 我已破解、`false` 我未破解；不传返回全部。未登录时恒按 `false` 处理 |
| page            | int    | 否   | 1          | 页码（min=1）                                                                                                    |
| page_size       | int    | 否   | 10         | 每页数量（min=1, max=20）                                                                                        |
| sort_by         | string | 否   | created_at | `created_at`（创建时间）/ `hot`（热度）；均按降序排列，值相同时按 `id` 倒序保证稳定分页                          |
| keyword         | string | 否   | —          | 按题目标题、描述或作者昵称文字模糊搜索（最长 50），不支持按 ID                                                   |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "author": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "title": "猜猜这是哪",
        "image": {
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "likes_count": 10,
        "liked": false,
        "solved": false,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 题目详情

```
GET /api/photos/{id}
```

**权限**：无

**说明**：`{id}` 为题目（Photo）ID。答案坐标 `location`（`Location` 对象）在**题目所属活动已结束**（服务器时间 ≥ `end_time`）**或当前用户为该题作者（本人投稿）**时返回非空；其余情况（活动进行中/未开始且访问者非作者）`location` 固定为 `null`，答案不提前下发。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "author": {
      "id": 1,
      "nickname": "张三",
      "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
    },
    "activity": {
      "id": 1,
      "title": "寻找校园角落",
      "start_time": "2026-06-01T00:00:00+08:00",
      "end_time": "2026-06-30T23:59:59+08:00"
    },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image": {
      "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
      "width": 1080,
      "height": 720
    },
    "location": { "longitude": 108.123456, "latitude": 34.123456, "coord_type": "gcj02" },
    "attempts_count": 5,
    "user_attempts_count": 3,
    "solved_count": 2,
    "solved": false,
    "likes_count": 10,
    "liked": false,
    "created_at": "2026-06-01T12:00:00+08:00",
    "status": "approved"
  },
  "message": "",
  "code": 0
}
```

---

### 3. 上传投稿

```
POST /api/photos
```

**权限**：登录用户（Level ≥ 1）

**说明**：`activity_id` 必须对应满足 `start_time <= now < end_time` 的进行中活动。初始审核状态为 `pending`。经度 `longitude`（-180~~180）、纬度 `latitude`（-90~~90）、坐标系 `coord_type`（`wgs84`/`gcj02`/`bd09`）非法或越界返回 `400`、`code=5`。图片仅支持 jpg/png 格式，单文件 ≤20MB。

**活动约束**：`activity_id` 必须对应满足 `start_time <= now < end_time` 的进行中活动。后端必须根据服务器当前时间校验，不接受尚未开始或已结束活动的新投稿；校验失败返回 `400`、`code=5`。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数        | 类型   | 必填 | 说明                               |
| ----------- | ------ | ---- | ---------------------------------- |
| activity_id | int    | 是   | 所属活动 ID                        |
| title       | string | 是   | 题目标题（最长 20）                |
| description | string | 否   | 题目描述/故事（最长 50）           |
| image_file  | file   | 是   | 图片文件（jpg/png，≤20MB）         |
| longitude   | float  | 是   | 经度                               |
| latitude    | float  | 是   | 纬度                               |
| coord_type  | string | 是   | 坐标系：`wgs84` / `gcj02` / `bd09` |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1001, "status": "pending" },
  "message": "",
  "code": 0
}
```

---

### 4. 设置题目点赞状态

```
PUT /api/photos/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为题目（Photo）ID。只能对审核通过（`status=approved`）且所属活动为进行中或已结束的题目设置点赞状态；未审核、已驳回或尚未开始活动的题目设置点赞返回 `400`、`code=5`。接口为幂等写操作，重复提交相同状态不会反向切换。

**请求参数（JSON Body）**

| 参数  | 类型 | 必填 | 说明                          |
| ----- | ---- | ---- | ----------------------------- |
| liked | bool | 是   | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "liked": true, "likes_count": 11 },
  "message": "",
  "code": 0
}
```

---

## 作答 (Attempts)

作答（attempt）是对题目的一次猜测提交，`status` 为 `pending` / `solved` / `unsolved`；**破解（solve）不是独立实体，是 `status=solved` 的作答**。公开面只展示破解记录，个人在"我在本图的作答"里可见自己全部作答及判定。新作答提交后初始为 `pending`，其 `solved` / `unsolved` 判定规则见[审核作答](#审核作答)。

### 1. 提交作答

```
POST /api/photos/{id}/attempts
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为题目（Photo）ID。只有题目所属活动满足 `start_time <= now < end_time` 时才允许提交新作答；尚未开始或已结束时返回 `400`、`code=5`。作者不能作答自己投稿的题目，违反同样返回 `400`、`code=5`；已成功破解（`status=solved`）的题目不能再次作答，违反同样返回 `400`、`code=5`；作答次数达到上限的题目不能再次作答（每人对单题最多作答 **5** 次，即 `user_attempts_count` 上限为 5），违反同样返回 `400`、`code=5`。作答自动归属题目的活动，客户端不另行传递 `activity_id`。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数       | 类型   | 必填 | 说明                               |
| ---------- | ------ | ---- | ---------------------------------- |
| image_file | file   | 是   | 猜测的匹配照片（jpg/png，≤20MB）   |
| longitude  | float  | 是   | 猜测经度                           |
| latitude   | float  | 是   | 猜测纬度                           |
| coord_type | string | 是   | 坐标系：`wgs84` / `gcj02` / `bd09` |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 100, "status": "pending" },
  "message": "",
  "code": 0
}
```

---

### 2. 破解记录

```
GET /api/photos/{id}/solves
```

**权限**：无

**说明**：`{id}` 为题目（Photo）ID。公开展示本题**破解成功**（`status=solved`）的作答，按破解时间倒序，用于"谁破解了此题"的展示；不含未破解/待审核的作答，也不做状态或排序筛选。

**请求参数（Query）**

| 参数      | 类型 | 必填 | 默认值 | 说明               |
| --------- | ---- | ---- | ------ | ------------------ |
| page      | int  | 否   | 1      | 页码               |
| page_size | int  | 否   | 10     | 每页数量（max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "author": {
          "id": 2,
          "nickname": "李四",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "image": {
          "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "likes_count": 0,
        "liked": false,
        "created_at": "2026-06-02T10:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 3. 我在本图的作答

```
GET /api/photos/{id}/attempts/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为题目（Photo）ID。当前用户在本题的全部作答，按提交时间倒序，不做状态筛选。**题目已由路径限定，列表项不重复返回题目信息**；每条含该次作答的 `status`（`solved` / `unsolved` / `pending`，本人可见自己的判定结果）。若当前用户为该题作者（本人投稿），因作者不能作答自己投稿的题目，本列表恒为空。

**请求参数（Query）**

| 参数      | 类型 | 必填 | 默认值 | 说明                      |
| --------- | ---- | ---- | ------ | ------------------------- |
| page      | int  | 否   | 1      | 页码（min=1）             |
| page_size | int  | 否   | 10     | 每页数量（min=1, max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 3,
    "list": [
      {
        "id": 1,
        "image": {
          "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "location": { "longitude": 108.123456, "latitude": 34.123456, "coord_type": "gcj02" },
        "created_at": "2026-06-02T10:00:00+08:00",
        "status": "pending",
        "reject_reason": null
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 4. 设置破解点赞状态

```
PUT /api/solves/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为作答 ID，仅对破解记录（`status=solved` 的作答）有效。对待审核或判定未破解的作答点赞返回 `400`、`code=5`；目标作答不存在返回 `404`。接口为幂等写操作，重复提交相同状态不会反向切换。

**请求参数（JSON Body）**

| 参数  | 类型 | 必填 | 说明                          |
| ----- | ---- | ---- | ----------------------------- |
| liked | bool | 是   | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "liked": true, "likes_count": 5 },
  "message": "",
  "code": 0
}
```

---

## 评论 (Comments)

### 1. 评论列表

```
GET /api/photos/{id}/comments
```

**权限**：无

**说明**：`{id}` 为题目 ID。获取题目的公开评论列表，按评论时间 `created_at` 倒序排列。仅允许获取审核通过（`status=approved`）且所属活动为进行中或已结束的题目的评论列表。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值     | 说明                                                                         |
| --------- | ------ | ---- | ---------- | ---------------------------------------------------------------------------- |
| page      | int    | 否   | 1          | 页码                                                                         |
| page_size | int    | 否   | 10         | 每页数量（max=20）                                                           |
| sort_by   | string | 否   | created_at | `created_at` / `likes_count`；均按降序排列，值相同时按 `id` 倒序保证稳定分页 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "author": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "content": "我知道这是哪里！",
        "liked": false,
        "likes_count": 3,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 发表评论

```
POST /api/photos/{id}/comments
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为题目（Photo）ID，仅可对审核通过（`status=approved`）且所属活动为进行中或已结束的题目发表评论。评论正文 `content` 最长 140 字。提交后初始审核状态为 `pending`，经管理员审核通过后才在公开评论列表展示。

**请求参数（JSON Body）**

| 参数    | 类型   | 必填 | 说明                 |
| ------- | ------ | ---- | -------------------- |
| content | string | 是   | 评论内容（最长 140） |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 50, "status": "pending" },
  "message": "",
  "code": 0
}
```

---

### 3. 删除评论

```
DELETE /api/comments/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为评论（Comment）ID。仅允许评论作者本人删除自己的评论，非作者删除返回 `403`；管理员处理违规评论必须通过后台管理接口 `PUT /api/admin/comments/{id}/review` 进行。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "deleted" },
  "message": "",
  "code": 0
}
```

---

### 4. 设置评论点赞状态

```
PUT /api/comments/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为评论（Comment）ID。只能对未下架的有效评论点赞，已删除或下架的评论点赞返回 `404`。接口为幂等写操作，重复提交相同状态不会反向切换。

**请求参数（JSON Body）**

| 参数  | 类型 | 必填 | 说明                          |
| ----- | ---- | ---- | ----------------------------- |
| liked | bool | 是   | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "liked": true, "likes_count": 3 },
  "message": "",
  "code": 0
}
```

---

## 我的记录 (My Records)

当前用户的个人记录汇总。个人记录不返回点赞字段（点赞只发生在公开面的题目、破解记录、评论上）。

### 1. 我的投稿记录

```
GET /api/photos/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：可按单个活动或审核状态筛选；不传 `activity_id` 时返回当前用户的全部投稿。

列表按投稿时间 `created_at` 倒序、记录 `id` 倒序返回扁平分页列表。个人记录不返回点赞字段。

**请求参数（Query）**

| 参数        | 类型   | 必填 | 默认值 | 说明                                                              |
| ----------- | ------ | ---- | ------ | ----------------------------------------------------------------- |
| activity_id | int    | 否   | —      | 按活动筛选我的投稿；不传返回我在全部活动下的投稿                  |
| page        | int    | 否   | 1      | 页码（min=1）                                                     |
| page_size   | int    | 否   | 10     | 每页数量（min=1, max=20）                                         |
| status      | string | 否   | —      | 按审核状态筛选：`pending` / `approved` / `rejected`；不传返回全部 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "title": "猜猜这是哪",
        "image": {
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "created_at": "2026-06-01T12:00:00+08:00",
        "status": "approved"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 我的投稿详情

```
GET /api/photos/user/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为题目（Photo）ID。仅用于作者本人查看自己审核中（`pending`）或已驳回（`rejected`）的投稿；已通过公开（`approved`）的投稿、他人投稿或不存在的请求本接口一律遮蔽返回 `404`。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "activity": {
      "id": 1,
      "title": "寻找校园角落",
      "start_time": "2026-06-01T00:00:00+08:00",
      "end_time": "2026-06-30T23:59:59+08:00"
    },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image": {
      "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
      "width": 1080,
      "height": 720
    },
    "location": { "longitude": 108.123456, "latitude": 34.123456, "coord_type": "gcj02" },
    "created_at": "2026-06-01T12:00:00+08:00",
    "status": "rejected",
    "reject_reason": "画面模糊，无法辨认地点"
  },
  "message": "",
  "code": 0
}
```

---

### 3. 我的作答记录

```
GET /api/attempts/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：可按单个活动或作答判定筛选；不传 `activity_id` 时返回当前用户的全部作答记录。

列表按作答时间 `created_at` 倒序、作答记录 `id` 倒序返回扁平分页列表。每条包含关联题目摘要 `photo`（含题目标题与题目缩略图）、该题我的已作答次数 `user_attempts_count` 与最新一次作答判定 `status`（`pending` / `solved` / `unsolved`）。个人记录不返回点赞字段、作答图片与坐标。

**请求参数（Query）**

| 参数        | 类型   | 必填 | 默认值 | 说明                                                            |
| ----------- | ------ | ---- | ------ | --------------------------------------------------------------- |
| activity_id | int    | 否   | —      | 按活动筛选我的作答；不传返回我在全部活动下的作答                |
| page        | int    | 否   | 1      | 页码（min=1）                                                   |
| page_size   | int    | 否   | 10     | 每页数量（min=1, max=20）                                       |
| status      | string | 否   | —      | 按作答判定筛选：`pending` / `unsolved` / `solved`；不传返回全部 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "user_attempts_count": 3,
        "status": "pending",
        "created_at": "2026-06-02T10:00:00+08:00",
        "photo": {
          "id": 1,
          "title": "猜猜这是哪",
          "image": {
            "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
            "width": 1080,
            "height": 720
          }
        }
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

## 积分 (Score)

### 积分流水

```
GET /api/score/logs
```

**权限**：登录用户（Level ≥ 1）

**说明**：获取当前用户的积分变动历史记录（作答奖励、兑换消耗等），按时间倒序排列。

**请求参数（Query）**

| 参数      | 类型 | 必填 | 默认值 | 说明                      |
| --------- | ---- | ---- | ------ | ------------------------- |
| page      | int  | 否   | 1      | 页码（min=1）             |
| page_size | int  | 否   | 10     | 每页数量（min=1, max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "total_income": 1500,
    "total_expense": 500,
    "list": [
      {
        "id": 1,
        "delta": 10,
        "balance": 1000,
        "reason": "answer_correct",
        "related_id": 1,
        "related_type": "photo",
        "related_title": "猜猜这是哪",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

**total_income / total_expense**：均为必返非负整数，与 `total` 同级（不在列表项内）。`total_income` 为所有 `delta > 0` 的变动之和，`total_expense` 为所有 `delta < 0` 的变动取绝对值之和、恒为非负数（前端展示时自行加负号）。两者均为**全量口径**——统计当前用户的全部历史流水，不受分页影响，每页返回相同值。`admin_adjust` 按其 `delta` 正负分别计入收入或支出。满足恒等式 `total_income - total_expense = GET /api/user/info` 的 `score_count`（当前积分余额）。

**reason 类型**：`answer_correct`（答题正确得分）/ `review_pass`（投稿审核通过得分）/ `exchange`（兑换奖品扣分）/ `admin_adjust`（管理员人工调整）

**related_type / related_id / related_title**：`answer_correct`、`review_pass` 关联 `photo`（题目 ID 与题目标题）；`exchange` 关联 `exchange`（兑换记录 ID 与奖品名）；`admin_adjust` 无关联对象，`related_type` / `related_id` / `related_title` 均为 `null`。补充 `related_title`（string）避免列表渲染产生 N+1 请求。

---

## 奖品 (Goods)

### 奖品列表

```
GET /api/goods
```

**权限**：无

**说明**：客户端奖品列表，仅返回上架（`in_store`）奖品，返回内容不因调用者身份变化；下架奖品只在管理端列表可见。多个筛选条件按 AND 组合。兑换（`POST /api/exchange`）仍需登录。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                             |
| --------- | ------ | ---- | ------ | ------------------------------------------------ |
| page      | int    | 否   | 1      | 页码                                             |
| page_size | int    | 否   | 10     | 每页数量（max=20）                               |
| keyword   | string | 否   | —      | 按名称或描述文字模糊搜索（最长 50），不支持按 ID |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "name": "明信片套装",
        "description": "精美校园风景明信片",
        "image": {
          "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "score_price": 500,
        "stock": 20,
        "status": "in_store",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

不提供单独的奖品详情接口：列表项已含全部奖品字段（含 `description` 与图片 `image`），详情界面直接使用列表数据。

---

## 兑换 (Exchange)

**防伪核销码 `verify_code` 与线下核销流程**：每条兑换记录附带一个 `verify_code`，由后端生成并随记录固定不变（不进行定期轮换）——8-16 位大写字母与数字组合，须用加密安全随机源生成、全局唯一，不得由记录 ID 等可预测值推导（管理端只做扫码查询，猜中即可让他人的奖品被误核销）。

1. C 端把 `verify_code` **裸字符串**（如 `VX78A9B2`）编码成二维码展示，二维码内不放 URL、不放 JSON、不附带用户信息；
2. 管理员扫码得到该字符串，调 `GET /api/admin/exchange?verify_code=...` 拉取对应记录，弹窗展示用户、奖品、数量、状态供人工核对；
3. 核对无误后用查得的记录 `id` 调 `PUT /api/admin/exchange/{id}/verify` 完成核销或取消。

核销接口只接受记录 `id`，不接受 `verify_code`——多一步人工核对，避免"扫到即核销"的误操作。

### 1. 兑换奖品

```
POST /api/exchange
```

**权限**：登录用户（Level ≥ 1）

**说明**：扣减积分 `score_cost = score_price × quantity`。已下架奖品（`out_store`）、剩余库存不足（`stock < quantity`）或用户可用积分不足（`score_count < score_cost`）时均返回 `400`、`code=5`。

**幂等要求**：请求头必须携带 `Idempotency-Key`（16-128 字符）。同一用户使用同一键和相同请求内容重复提交时，后端返回首次请求结果，不得重复创建兑换记录、扣减库存或扣减积分；同一键对应不同请求内容时返回 `409`。兑换记录创建、库存扣减和积分扣减必须处于同一原子事务或具备等效的一致性保证。

| Header          | 类型   | 必填 | 说明                                |
| --------------- | ------ | ---- | ----------------------------------- |
| Idempotency-Key | string | 是   | 一次兑换操作的唯一键，建议使用 UUID |

**请求参数（JSON Body）**

| 参数     | 类型 | 必填 | 说明                                          |
| -------- | ---- | ---- | --------------------------------------------- |
| good_id  | int  | 是   | 奖品 ID                                       |
| quantity | int  | 是   | 兑换数量（min=1，且不得超过当前库存 `stock`） |

**返回** `201`

- 成功：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "pending", "verify_code": "VX78A9B2" },
  "message": "",
  "code": 0
}
```

- 参数无效返回 `400`（`code=3`）；库存不足或积分不足返回 `400`（`code=5`）

```json
{
  "success": false,
  "resp": null,
  "message": "操作错误: 奖品库存不足",
  "code": 5
}
```

- 同一幂等键对应不同请求内容：`409`

```json
{
  "success": false,
  "resp": null,
  "message": "冲突错误: Idempotency-Key 已用于不同请求内容",
  "code": 8
}
```

---

### 2. 兑换记录列表

```
GET /api/exchange
```

**权限**：登录用户（Level ≥ 1）

**说明**：获取当前用户的奖品兑换历史记录，按兑换时间倒序排列。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                               |
| --------- | ------ | ---- | ------ | -------------------------------------------------- |
| page      | int    | 否   | 1      | 页码                                               |
| page_size | int    | 否   | 10     | 每页数量（max=20）                                 |
| status    | string | 否   | —      | `pending` / `verified` / `cancelled`；不传返回全部 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "good": {
          "id": 1,
          "name": "明信片套装",
          "image": {
            "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
            "width": 1080,
            "height": 720
          },
          "score_price": 500
        },
        "quantity": 1,
        "score_cost": 500,
        "status": "pending",
        "verify_code": "VX78A9B2",
        "exchange_at": null,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

## 通知 (Notifications)

站内消息分两组接口：**通知**（`/announcements`，管理员发布）与**互动消息**（`/notifications`，点赞、评论业务事件自动生成）。首屏弹窗不属于通知，见[内容位](#内容位-contents)。

- 通知面向全部已登录用户，无过期时间；正文为**纯富文本** HTML（仅文本类标签，无图片/附件上传；后端按标签白名单过滤防 XSS），保留 `[image]` 占位标记，长度按去标签后的可见文本计，Unicode 码点最长 **2000**（原始 HTML 串 ≤10000）；可配一张图片，可选关联一个活动（**请求侧**：`related_type`/`related_id` 同时提供或同时省略；**响应侧**：两者与 `image` 均为必返可空字段，无值时为 `null`，见[约定](#约定)）。通知有独立的已读体系：**读通知详情即标记已读**，通知列表返回各条 `is_read` 和未读通知数 `unread_count`。
- 互动消息只投递给业务事件对应的接收人，带发送者摘要与关联对象；有独立未读数 `unread_count`，**拉取列表不改变已读状态**，通过「标记单条已读」或「一键已读」接口更新（互动消息无详情接口）。审核结果不生成互动消息：投稿与作答的审核状态在"我的投稿列表""我的作答列表"中自行查看。

### 1. 通知列表

```
GET /api/announcements
```

**权限**：登录用户（Level ≥ 1）

**说明**：列表返回 `content_preview` 摘要与各条 `is_read`，完整正文、配图与关联对象走通知详情接口；**拉取列表不改变已读状态，读详情才标记已读**。`unread_count` 为当前用户未读通知数（不受筛选影响），供角标展示。默认按 `created_at` 倒序、`id` 倒序稳定分页。

`content_preview` 由后端从完整 `content` 生成，规则固定为：

1. 剥离所有 HTML 标签，只保留可见文本；块级标签（如 `</p>`、`<br>`）之间按一个空格处理，避免相邻段落文字粘连；
2. 将所有字面量 `[image]` 图片占位标记替换为空格；
3. 将换行、制表符和连续 Unicode 空白合并为一个半角空格，并去除首尾空白；
4. 按 Unicode 码点而非字节截取前 50 个字符，不追加省略号；清洗后为空时返回空字符串。

`keyword` 搜索匹配**剥离 HTML 标签后的完整正文文本**（不搜 `content_preview`，以免 50 字以后内容无法命中；也不搜 HTML 源码，以免命中 `p`、`strong` 等标签名）。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                             |
| --------- | ------ | ---- | ------ | ------------------------------------------------ |
| page      | int    | 否   | 1      | 页码                                             |
| page_size | int    | 否   | 10     | 每页数量（max=20）                               |
| keyword   | string | 否   | —      | 按标题或正文文字模糊搜索（最长 50），不支持按 ID |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 3,
    "unread_count": 2,
    "list": [
      {
        "id": 10,
        "title": "系统维护通知",
        "content_preview": "今晚 23:00 至 23:30 进行系统维护。 请提前安排。",
        "is_read": false,
        "created_at": "2026-07-21T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 通知详情

```
GET /api/announcements/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：返回通知完整正文（纯富文本 HTML，保留 `[image]` 占位标记）、可选配图 `image` 与可选活动关联 `related_type`/`related_id`（三者均必返，无对应值时为 `null`；`related_type` 与 `related_id` 同时为 `null` 或同时非空）。**读取成功即由后端把该通知对当前用户标记为已读**（读即已读，幂等）；`is_read` 为本次读取前的状态。不设单独的标记已读接口。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 10,
    "title": "系统维护通知",
    "content": "<p>今晚 23:00 至 23:30 进行系统维护。</p>\n[image]\n<p>请提前安排。</p>",
    "image": {
      "origin_url": "https://media.example.com/notifications/maintenance.jpg?signature=example",
      "thumb_url": "https://media.example.com/notifications/thumb.jpg?signature=example",
      "width": 1080,
      "height": 720
    },
    "related_type": "activity",
    "related_id": 1,
    "is_read": false,
    "created_at": "2026-07-21T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

通知不存在时返回 `404`。

---

### 3. 互动消息列表

```
GET /api/notifications
```

**权限**：登录用户（Level ≥ 1）

**说明**：点赞、评论业务事件生成的互动消息，仅接收人可见。拉取列表不自动清空未读状态，未读通过标记接口或一键已读更新。`unread_count` 为当前未读消息总数，不受筛选条件影响，供小红点角标展示。`content` 由后端按事件生成单行文本。`type` 为可重复参数，多值之间 OR，不传返回全部。`related_type` / `related_id` 指向触发事件的对象：`like` 关联被点赞的 `photo`（题目）/ `solve`（破解记录）/ `comment`（评论）之一，`comment` 关联被评论的 `photo`（题目）；前端据此跳转。

**请求参数（Query）**

| 参数      | 类型           | 必填 | 默认值 | 说明                                                           |
| --------- | -------------- | ---- | ------ | -------------------------------------------------------------- |
| page      | int            | 否   | 1      | 页码                                                           |
| page_size | int            | 否   | 10     | 每页数量（max=20）                                             |
| type      | string(可重复) | 否   | —      | `like` / `comment`；重复参数 `type=like&type=comment`，多值 OR |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "unread_count": 3,
    "list": [
      {
        "id": 1,
        "type": "like",
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "related_type": "photo",
        "related_id": 1,
        "photo_id": 1001,
        "content": "张三 赞了你的投稿",
        "is_read": false,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 4. 标记单条互动消息已读

```
PUT /api/notifications/{id}/read
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为互动消息 ID。标记当前用户的指定互动消息为已读（`is_read=true`）。目标消息不存在或不属于当前用户返回 `404`。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "is_read": true },
  "message": "",
  "code": 0
}
```

---

### 5. 一键标记全部互动消息已读

```
PUT /api/notifications/read-all
```

**权限**：登录用户（Level ≥ 1）

**说明**：将当前用户的全部未读互动消息统一标记为已读（`is_read=true`），重置未读数。幂等操作。

**返回** `200`

```json
{
  "success": true,
  "resp": { "marked_count": 5 },
  "message": "",
  "code": 0
}
```

---

## 内容位 (Contents)

后台可编辑的单例富文本内容位，统一两个接口、按 `key` 寻址：`popup`（首屏弹窗）、`score_rules`（积分规则）、`help`（帮助中心）。"关于我们"由前端写死，不进契约。

- `content` 为**纯富文本** HTML 字符串（仅文本类标签，无图片/附件上传；后端按标签白名单过滤防 XSS）。长度限制针对**可见文本**：后端先剥离 HTML 标签，再按 Unicode 码点计数，最长 **2000**（标签本身不计入）；同时原始 HTML 串限制 **≤10000** 码点，防止超大标签负载。任一超限返回 `400`。
- `version` 每次后台保存自增（未编辑过为 0，`content` 为空字符串、`updated_at` 为 `null`）。
- **弹窗防打扰**：用户点击关闭后，前端在本地记录当前 `version`；仅当后台再次保存（`version` 增大）才重新弹出。服务端不记录任何用户的关闭状态。
- **弹窗关联引导**：`popup` 可选携带 `related_id`（关联的通知 ID）。前端"查看详情"跳转通知详情；未登录用户请求通知详情会得到 `401`，前端拦截并引导登录，登录后跳转，已登录用户直接跳转。

### 读取内容位

```
GET /api/contents/{key}
```

**权限**：无（弹窗是公开首屏提醒卡片，未登录可见）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "key": "popup",
    "content": "<p>欢迎参加校园图寻！</p>",
    "related_id": 10,
    "version": 3,
    "updated_at": "2026-07-26T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

`related_id` 必返，仅 `key=popup` 且已设置关联时为通知 ID，其余情况为 `null`。`key` 非法返回 `404`。

---

## 反馈 (Feedback)

### 提交反馈

```
POST /api/feedback
```

**权限**：登录用户（Level ≥ 1）

**说明**：提交用户反馈。可选上传 1 个附件文件 `media_file`（图片 ≤ 20MB 或视频 ≤ 50MB），适配小程序单文件上传限制。正文内容限制在 1 到 500 个字符之间；格式或大小超限返回 `400`、`code=5`。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数       | 类型   | 必填 | 说明                                                     |
| ---------- | ------ | ---- | -------------------------------------------------------- |
| title      | string | 是   | 标题（最长 20）                                          |
| content    | string | 是   | 内容（最长 500）                                         |
| type       | int    | 是   | 反馈类型：1-内容 2-玩法 3-技术 4-其他                    |
| phone      | string | 否   | 联系电话（最长 20）                                      |
| media_file | file   | 否   | 附件文件：图片（jpg/png，≤20MB）或视频（mp4/mov，≤50MB） |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "pending" },
  "message": "",
  "code": 0
}
```

---

## 管理员

> 以下接口均需管理员权限（Level ≥ 2）。

---

### 题目管理与审核

#### 题目列表

```
GET /api/admin/photos
```

**权限**：管理员（Level ≥ 2）

**说明**：这是管理端唯一的题目列表接口。不传筛选时返回所有活动下的全部未删除题目，覆盖 `not_started`、`active`、`ended` 活动和 `pending`、`approved`、`rejected` 审核状态；默认按 `created_at` 倒序、`id` 倒序稳定分页。客户端公开题目列表仍遵守活动时间和审核状态规则。

**请求参数（Query）**

| 参数         | 类型   | 必填 | 默认值 | 说明                                                                                                                                                                                          |
| ------------ | ------ | ---- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| page         | int    | 否   | 1      | 页码                                                                                                                                                                                          |
| page_size    | int    | 否   | 10     | 每页数量（max=20）                                                                                                                                                                            |
| status       | string | 否   | —      | `pending` / `approved` / `rejected`；不传返回全部审核状态                                                                                                                                     |
| activity_ids | int[]  | 否   | —      | 按多个活动 ID 筛选，使用重复参数，如 `activity_ids=1&activity_ids=2`；所选 ID 之间按 OR 匹配                                                                                                  |
| solved       | bool   | 否   | —      | 筛选该题是否已被**任何人**破解：`true` 即 `solved_count > 0`，`false` 即 `solved_count = 0`；不传返回全部。注意本参数为全站口径，与客户端 `GET /photos` 的 `solved`（本人是否已破解）语义不同 |
| keyword      | string | 否   | —      | 按题目 ID、标题或描述搜索（最长 50）                                                                                                                                                          |
| user_keyword | string | 否   | —      | 按作者 ID 或昵称搜索（最长 50）                                                                                                                                                               |

`status`、`activity_ids`、`solved`、`keyword` 同时提供时按 AND 组合；不存在的活动 ID 不单独返回 `404`，只是不产生匹配记录。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "activity": {
          "id": 1,
          "title": "寻找校园角落",
          "start_time": "2026-06-01T00:00:00+08:00",
          "end_time": "2026-06-30T23:59:59+08:00"
        },
        "author": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "title": "猜猜这是哪",
        "description": "校园神秘角落",
        "image": {
          "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "location": { "longitude": 108.123456, "latitude": 34.123456, "coord_type": "gcj02" },
        "solved_count": 2,
        "attempts_count": 5,
        "likes_count": 10,
        "status": "approved",
        "reject_reason": null,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 审核题目

```
PUT /api/admin/photos/{id}/review
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为题目 ID。管理员将 `pending` 题目审核为通过（`approved`）或驳回（`rejected`）。当判定为 `rejected` 时，驳回原因 `reject_reason` 为必填非空（最长 50 字符），为空拦截返回 `400`、`code=5`。**审核为终态，不可再改判**——对已 `approved` 或 `rejected` 的题目再次审核返回 `409`（`approved`↔`rejected` 两个方向都不允许）。

**请求参数（JSON Body）**

| 参数          | 类型   | 必填     | 说明                                                         |
| ------------- | ------ | -------- | ------------------------------------------------------------ |
| action        | string | 是       | `approve` / `reject`                                         |
| reject_reason | string | 条件必填 | **action=reject 时必填**（最长 50）；action=approve 时不用传 |

**返回** `200`

- 审核通过：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "approved" },
  "message": "",
  "code": 0
}
```

- 审核拒绝：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "rejected" },
  "message": "",
  "code": 0
}
```

- 已被他人审核（409）：

```json
{
  "success": false,
  "resp": null,
  "message": "冲突错误: 该题目已审核过",
  "code": 8
}
```

---

#### 新增题目

```
POST /api/admin/photos
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：管理员按表单字段 `activity_id` 指定所属活动，仅可给**未开始或进行中**活动（`now < end_time`）新增题目；已结束活动不可新增，返回 `400`、`code=5`。题目作者固定记录为官方账号——部署侧预置的一条真实用户记录（昵称如“图寻官方”），不随操作管理员变化；各接口返回的 `author` 即该官方账号的 `UserBrief`，客户端按普通作者统一展示，无需判定题目投稿渠道。管理员以个人身份投稿仍走客户端 `POST /photos` 普通流程并正常审核。新增后直接进入 `approved` 状态。`activity_id` 对应活动不存在返回 `400`。

| 参数        | 类型   | 必填 | 说明                       |
| ----------- | ------ | ---- | -------------------------- |
| activity_id | int    | 是   | 所属活动 ID                |
| title       | string | 是   | 题目标题（最长 20）        |
| description | string | 否   | 题目描述（最长 50）        |
| image_file  | file   | 是   | 题目图片（jpg/png，≤20MB） |
| longitude   | float  | 是   | 经度                       |
| latitude    | float  | 是   | 纬度                       |
| coord_type  | string | 是   | `wgs84` / `gcj02` / `bd09` |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1001, "status": "approved" },
  "message": "",
  "code": 0
}
```

#### 更新题目

```
PUT /api/admin/photos/{id}
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：`{id}` 为题目（Photo）ID。可编辑任意活动阶段的题目内容，不改变题目当前审核状态（审核为终态，编辑不会使其重新进入审核）。请求至少提供一个字段；如修改坐标，`longitude`、`latitude`、`coord_type` 必须同时提供。

| 参数        | 类型   | 必填     | 说明                         |
| ----------- | ------ | -------- | ---------------------------- |
| title       | string | 否       | 题目标题（最长 20）          |
| description | string | 否       | 题目描述（最长 50）          |
| image_file  | file   | 否       | 新题目图片（jpg/png，≤20MB） |
| longitude   | float  | 条件必填 | 与纬度、坐标系同时提供       |
| latitude    | float  | 条件必填 | 与经度、坐标系同时提供       |
| coord_type  | string | 条件必填 | `wgs84` / `gcj02` / `bd09`   |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1001, "status": "approved" },
  "message": "",
  "code": 0
}
```

题目不存在时返回 `404`。

---

### 审核 — 作答

#### 作答列表

```
GET /api/admin/attempts
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端作答记录列表，支持按作答 ID、题目、作答者及作答状态（`pending`/`solved`/`unsolved`）进行组合筛选。

**请求参数（Query）**

| 参数          | 类型   | 必填 | 默认值 | 说明                                                |
| ------------- | ------ | ---- | ------ | --------------------------------------------------- |
| page          | int    | 否   | 1      | 页码                                                |
| page_size     | int    | 否   | 10     | 每页数量（max=20）                                  |
| status        | string | 否   | —      | `pending` / `solved` / `unsolved`；不传返回全部状态 |
| keyword       | string | 否   | —      | 按作答 ID 搜索（最长 50）                           |
| photo_keyword | string | 否   | —      | 按题目 ID 或题目标题搜索（最长 50）                 |
| user_keyword  | string | 否   | —      | 按作答者 ID 或昵称搜索（最长 50）                   |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "photo": {
          "id": 1,
          "title": "原图标题",
          "image": {
            "origin_url": "https://media.example.com/photos/photo.jpg?signature=example",
            "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
            "width": 1080,
            "height": 720
          },
          "location": { "longitude": 108.123456, "latitude": 34.123456, "coord_type": "gcj02" }
        },
        "guess_image": {
          "origin_url": "https://media.example.com/attempts/attempt.jpg?signature=example",
          "thumb_url": "https://media.example.com/attempts/thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "guess_location": { "longitude": 108.5, "latitude": 34.5, "coord_type": "gcj02" },
        "status": "pending",
        "reject_reason": null,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 审核作答

```
PUT /api/admin/attempts/{id}/review
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为作答 ID。**判定机制为坐标自动 + 图片人工**：后端按猜测坐标与题目答案坐标的距离阈值（**默认 50 米，后端可配置**）自动判定坐标是否命中，提交的现场实拍猜测图片由管理员人工审核；管理员结合两者将 `pending` 作答确认为 `solved` 或 `unsolved`。**判定为终态，不可再改判**——对已 `solved` 或 `unsolved` 的作答再次审核返回 `409`（`solved`↔`unsolved` 两个方向都不允许）。当作答被判定为 `solved`（该用户在该题目上首次且唯一一次成功破解）时，系统自动更新该题目 `solved_count` 与作答者积分（`answer_correct`）。

**请求参数（JSON Body）**

| 参数          | 类型   | 必填 | 说明                                                                       |
| ------------- | ------ | ---- | -------------------------------------------------------------------------- |
| solved        | string | 是   | `solved` / `unsolved`                                                      |
| reject_reason | string | 否   | `unsolved` 的判定说明（最长 50）；不填时为空（`null`），后端不生成默认文案 |

**返回** `200`

- 答对：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "solved" },
  "message": "",
  "code": 0
}
```

- 未答对：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "unsolved" },
  "message": "",
  "code": 0
}
```

- 已被他人审核（409）：

```json
{
  "success": false,
  "resp": null,
  "message": "冲突错误: 该作答记录已审核过",
  "code": 8
}
```

---

### 审核 — 评论

#### 评论列表

```
GET /api/admin/comments
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端评论列表，支持按评论 ID、评论内容、题目及评论者进行组合筛选。

**请求参数（Query）**

| 参数          | 类型   | 必填 | 默认值 | 说明                                                  |
| ------------- | ------ | ---- | ------ | ----------------------------------------------------- |
| page          | int    | 否   | 1      | 页码                                                  |
| page_size     | int    | 否   | 10     | 每页数量（max=20）                                    |
| status        | string | 否   | —      | `pending` / `approved` / `rejected`；不传返回全部状态 |
| keyword       | string | 否   | —      | 按评论 ID 或评论内容搜索（最长 50）                   |
| photo_keyword | string | 否   | —      | 按题目 ID 或题目标题搜索（最长 50）                   |
| user_keyword  | string | 否   | —      | 按评论者 ID 或昵称搜索（最长 50）                     |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 8,
    "list": [
      {
        "id": 1,
        "photo": {
          "id": 1,
          "title": "图片标题"
        },
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "content": "评论内容",
        "status": "approved",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

列表项始终返回每条评论当前的 `status`；不传 `status` 时返回混合状态，不能用当前筛选条件推断单条记录状态。

#### 审核评论

```
PUT /api/admin/comments/{id}/review
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为评论 ID。管理员审核评论，审核通过（`approved`，公开可见）或驳回（`rejected`，不再展示）。`action=reject` 时须填 `reject_reason`。**评论审核可反复改判**：`approved` 与 `rejected` 之间可随时切换（与作答、题目审核的终态规则不同），不返回 `409`。

**请求参数（JSON Body）**

| 参数          | 类型   | 必填     | 说明                                                         |
| ------------- | ------ | -------- | ------------------------------------------------------------ |
| action        | string | 是       | `approve` / `reject`                                         |
| reject_reason | string | 条件必填 | **action=reject 时必填**（最长 50）；action=approve 时不用传 |

**返回** `200`

- 审核通过：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "approved" },
  "message": "",
  "code": 0
}
```

- 审核拒绝：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "rejected" },
  "message": "",
  "code": 0
}
```

---

### 活动管理

#### 活动列表

```
GET /api/admin/activity
```

**权限**：管理员（Level ≥ 2）

**说明**：返回全部活动（含未开始）。`status` 由后端按服务器当前时间计算。默认按 `start_time` 倒序、`id` 倒序稳定分页。不提供单独的活动详情接口——活动列表项（`ActivityCard`）已含全部活动字段，详情、编辑界面直接使用列表数据。因此本接口的 `cover_image` 与客户端活动列表不同：除列表展示用的 `thumb_url` 外，额外下发 `origin_url` 供编辑界面回填原图。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                             |
| --------- | ------ | ---- | ------ | ------------------------------------------------ |
| page      | int    | 否   | 1      | 页码                                             |
| page_size | int    | 否   | 10     | 每页数量（max=20）                               |
| status    | string | 否   | —      | `not_started` / `active` / `ended`；不传返回全部 |
| keyword   | string | 否   | —      | 按活动 ID、标题或描述搜索（最长 50）             |

**返回** `200`（`ActivityCardPage`）

```json
{
  "success": true,
  "resp": {
    "total": 3,
    "list": [
      {
        "id": 1,
        "title": "寻找校园角落",
        "cover_image": {
          "origin_url": "https://media.example.com/photos/cover.jpg?signature=example",
          "thumb_url": "https://media.example.com/photos/cover_thumb.jpg?signature=example",
          "width": 1080,
          "height": 720
        },
        "description": "活动介绍",
        "start_time": "2026-06-01T00:00:00+08:00",
        "end_time": "2026-06-30T23:59:59+08:00",
        "photo_count": 12
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 创建活动

```
POST /api/admin/activity
```

**权限**：管理员（Level ≥ 2）

**说明**：管理员发布新活动。活动开始时间必须早于结束时间（`start_time < end_time`），违反返回 `400`、`code=5`。活动封面图片仅支持 jpg/png 格式，单文件 ≤20MB。

**Content-Type**：`multipart/form-data`

**时间规则**：接口不接收 `status` 或 `is_active` 字段。活动是否进行中完全由 `start_time <= now < end_time` 判定。允许多个活动的时间范围重叠并同时进行。

**请求参数**

| 参数        | 类型              | 必填 | 说明                                                        |
| ----------- | ----------------- | ---- | ----------------------------------------------------------- |
| title       | string            | 是   | 活动标题（最长 20）                                         |
| cover_file  | file              | 是   | 封面图（jpg/png，≤20MB），必填以保证 `cover_image` 必返非空 |
| description | string            | 是   | 活动描述（最长 100）                                        |
| start_time  | string(date-time) | 是   | 带时区 ISO 8601 时间，例如 `2026-07-01T00:00:00+08:00`      |
| end_time    | string(date-time) | 是   | 带时区 ISO 8601 时间，且必须晚于开始时间                    |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

#### 更新活动

```
PUT /api/admin/activity/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为活动 ID。活动开始时间必须早于结束时间（`start_time < end_time`），违反返回 `400`、`code=5`。对于已经结束的活动（`now >= end_time`），禁止修改其 `start_time` 和 `end_time` 时间基准，违反拦截返回 `400`、`code=5`。

**Content-Type**：`multipart/form-data`

**时间规则**：修改 `start_time` / `end_time` 后，后端必须立即使用新时间范围进行列表归类和写操作校验。活动一旦达到旧 `end_time`，题目答案坐标可能已经公开，因此不得再把 `end_time` 延长到服务器当前时间之后使活动重新变为未结束；违反时返回 `400`。活动结束后，关联题目、投稿和作答记录保留原 `activity_id`，不进行数据迁移。允许与其他活动时间重叠。

**请求参数**

`{id}` 为活动 ID。

| 参数        | 类型              | 必填 | 说明                                                   |
| ----------- | ----------------- | ---- | ------------------------------------------------------ |
| title       | string            | 否   | 活动标题（最长 20）                                    |
| cover_file  | file              | 否   | 封面图（jpg/png，≤20MB）；不传则保留原封面，传了则替换 |
| description | string            | 否   | 活动描述（最长 100）                                   |
| start_time  | string(date-time) | 否   | 带时区 ISO 8601 开始时间                               |
| end_time    | string(date-time) | 否   | 带时区 ISO 8601 结束时间                               |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

活动管理不再提供按单个活动单独读取题目的接口。活动列表的“题目管理”入口应跳转到题目列表，并把当前活动 ID 作为 `activity_ids` 初始筛选值；活动候选项可复用活动列表的 `status` 与 `keyword` 做服务端预筛和名称/ID 实时搜索。

### 通知管理

管理端只写入通知（普通通知）；业务事件生成的互动消息（`like`/`comment`）不能在这里编辑或删除。

#### 通知列表

```
GET /api/admin/announcements
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端通知列表（`AdminAnnouncementPage`）。与客户端列表相比**不含任何面向用户的已读字段**——无页级 `unread_count`，列表项无 `is_read`；改为每条返回聚合已读人数 `read_count`（已读该通知的去重用户数），供运营查看触达效果。`content_preview` 为摘要。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                 |
| --------- | ------ | ---- | ------ | ------------------------------------ |
| page      | int    | 否   | 1      | 页码                                 |
| page_size | int    | 否   | 10     | 每页数量（max=20）                   |
| keyword   | string | 否   | —      | 按通知 ID、标题或正文搜索（最长 50） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 8,
    "list": [
      {
        "id": 10,
        "title": "系统维护通知",
        "content_preview": "今晚 23:00 至 23:30 进行系统维护。",
        "created_at": "2026-07-21T12:00:00+08:00",
        "read_count": 42
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 通知详情

```
GET /api/admin/announcements/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：读取通知完整正文、可选配图 `image` 与可选活动关联（无对应值时为 `null`），用于编辑前回填。返回 `AdminAnnouncement`：与客户端通知详情相比**去掉面向用户的 `is_read`**，改为返回聚合已读人数 `read_count`；**读取不标记已读**——管理员查看/编辑不消费面向用户的已读态。通知不存在返回 `404`。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 10,
    "title": "系统维护通知",
    "content": "<p>今晚 23:00 至 23:30 进行系统维护。</p>\n[image]\n<p>请提前安排。</p>",
    "image": {
      "origin_url": "https://media.example.com/notifications/maintenance.jpg?signature=example",
      "thumb_url": "https://media.example.com/notifications/thumb.jpg?signature=example",
      "width": 1080,
      "height": 720
    },
    "related_type": "activity",
    "related_id": 1,
    "created_at": "2026-07-21T12:00:00+08:00",
    "read_count": 42
  },
  "message": "",
  "code": 0
}
```

#### 发布通知

```
POST /api/admin/announcements
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：发布面向全部已登录用户的通知，可关联一个活动、可配一张图片。`related_id` 对应活动不存在返回 `400`。

| 参数         | 类型   | 必填     | 说明                                                                             |
| ------------ | ------ | -------- | -------------------------------------------------------------------------------- |
| title        | string | 是       | 通知标题（最长 20）                                                              |
| content      | string | 是       | 通知正文，纯富文本 HTML；去标签后可见文本按 Unicode 码点最长 2000，原始串 ≤10000 |
| image_file   | file   | 否       | 配图（jpg/png，≤20MB）                                                           |
| related_type | string | 条件可选 | 当前仅支持 `activity`；与 `related_id` 同时提供或同时省略                        |
| related_id   | int    | 条件可选 | 关联对象 ID                                                                      |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 10, "status": "published" },
  "message": "",
  "code": 0
}
```

#### 更新通知

```
PUT /api/admin/announcements/{id}
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：未提供字段保持不变。`remove_image` / `remove_relation` 分别移除已有配图和活动关联，不得与对应新值同传。更新不创建新通知。

| 参数            | 类型   | 必填     | 说明                                                                             |
| --------------- | ------ | -------- | -------------------------------------------------------------------------------- |
| title           | string | 否       | 通知标题（最长 20）                                                              |
| content         | string | 否       | 通知正文，纯富文本 HTML；去标签后可见文本按 Unicode 码点最长 2000，原始串 ≤10000 |
| image_file      | file   | 否       | 新配图（jpg/png，≤20MB）                                                         |
| remove_image    | bool   | 否       | `true` 时移除已有配图；不得与 `image_file` 同时提供                              |
| remove_relation | bool   | 否       | `true` 时移除已有活动关联；不得与 `related_type` / `related_id` 同时提供         |
| related_type    | string | 条件可选 | 当前仅支持 `activity`                                                            |
| related_id      | int    | 条件可选 | 关联对象 ID                                                                      |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 10, "status": "published" },
  "message": "",
  "code": 0
}
```

#### 删除通知

```
DELETE /api/admin/announcements/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：删除后通知不再出现在通知列表。互动消息不能通过本接口删除。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 10, "status": "deleted" },
  "message": "",
  "code": 0
}
```

参数或图片、关联组合无效时返回 `400`。

---

### 内容位管理

#### 更新内容位

```
PUT /api/admin/contents/{key}
```

**权限**：管理员（Level ≥ 2）

**说明**：编辑弹窗（`popup`）、积分规则（`score_rules`）、帮助中心（`help`）的富文本正文，全量覆盖。每次保存 `version` 自增——弹窗会因此对已关闭的用户重新弹出，属预期行为。`related_id` 仅 `key=popup` 允许（省略即清除关联，关联通知不存在返回 `400`；其他 `key` 提供该字段返回 `400`）。

**请求参数（JSON Body）**

| 参数       | 类型   | 必填 | 说明                                                                   |
| ---------- | ------ | ---- | ---------------------------------------------------------------------- |
| content    | string | 是   | 纯富文本 HTML；去标签后可见文本按 Unicode 码点最长 2000，原始串 ≤10000 |
| related_id | int    | 否   | 仅 `popup`：关联的通知 ID                                              |

**返回** `200`

```json
{
  "success": true,
  "resp": { "key": "popup", "version": 4, "status": "success" },
  "message": "",
  "code": 0
}
```

---

### 奖品管理

#### 奖品列表

```
GET /api/admin/goods
```

**权限**：管理员（Level ≥ 2）

**说明**：返回全部奖品（含下架 `out_store`）。列表项与客户端同为 `GoodItem`，已含全部管理字段和高清 `image`，不提供单独的奖品详情接口——详情、编辑和预览界面直接使用列表数据。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                   |
| --------- | ------ | ---- | ------ | -------------------------------------- |
| page      | int    | 否   | 1      | 页码                                   |
| page_size | int    | 否   | 10     | 每页数量（max=20）                     |
| status    | string | 否   | —      | `in_store` / `out_store`；不传返回全部 |
| keyword   | string | 否   | —      | 按奖品 ID、名称或描述搜索（最长 50）   |

**返回** `200`：结构与客户端奖品列表一致（`GoodItemPage`）。

#### 新增奖品

```
POST /api/admin/goods
```

**权限**：管理员（Level ≥ 2）

**说明**：管理员新增奖品上架。包含奖品名称、兑换所需积分、库存及缩略图。

**Content-Type**：`multipart/form-data`

| 参数        | 类型   | 必填 | 说明                                          |
| ----------- | ------ | ---- | --------------------------------------------- |
| name        | string | 是   | 奖品名称（最长 20）                           |
| description | string | 否   | 描述（最长 50）                               |
| score_price | int    | 是   | 所需积分（min=0）                             |
| stock       | int    | 是   | 库存（min=0）                                 |
| image_file  | file   | 是   | 奖品图片（jpg/png，≤20MB）                    |
| status      | string | 否   | `in_store`（上架，默认）/ `out_store`（下架） |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "in_store" },
  "message": "",
  "code": 0
}
```

#### 更新奖品

```
PUT /api/admin/goods/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为奖品 ID。管理员更新奖品信息（名称、积分、库存、上下架状态 `in_store`/`out_store`）。

**Content-Type**：`multipart/form-data`

| 参数        | 类型   | 必填 | 说明                       |
| ----------- | ------ | ---- | -------------------------- |
| name        | string | 否   | 奖品名称（最长 20）        |
| description | string | 否   | 描述（最长 50）            |
| score_price | int    | 否   | 所需积分                   |
| stock       | int    | 否   | 库存                       |
| image_file  | file   | 否   | 奖品图片（jpg/png，≤20MB） |
| status      | string | 否   | `in_store` / `out_store`   |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "in_store" },
  "message": "",
  "code": 0
}
```

#### 删除奖品

```
DELETE /api/admin/goods/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为奖品 ID。若该奖品已产生用户兑换记录，禁止物理删除，仅可进行下架（`out_store`）处理；尝试物理删除已有兑换账单的奖品拦截返回 `400`、`code=5`。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

### 兑换管理

#### 兑换列表

```
GET /api/admin/exchange
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端兑换记录列表，支持按兑换记录 ID、用户、奖品及防伪码进行组合搜索筛选。管理员扫码核销时，先用扫得的防伪码调本接口（`verify_code=VX78A9B2`）拉取待核销记录供人工核对，再调用核销接口完成核销。

**请求参数（Query）**

| 参数         | 类型   | 必填 | 默认值 | 说明                                                                                                           |
| ------------ | ------ | ---- | ------ | -------------------------------------------------------------------------------------------------------------- |
| page         | int    | 否   | 1      | 页码                                                                                                           |
| page_size    | int    | 否   | 10     | 每页数量（max=20）                                                                                             |
| status       | string | 否   | —      | `pending` / `verified` / `cancelled`；不传返回全部                                                             |
| keyword      | string | 否   | —      | 按兑换记录 ID 搜索（最长 50）                                                                                  |
| user_keyword | string | 否   | —      | 按用户 ID、学号、姓名或昵称搜索（最长 50）                                                                     |
| good_keyword | string | 否   | —      | 按奖品 ID 或奖品名称搜索（最长 50）                                                                            |
| verify_code  | string | 否   | —      | 按防伪核销码**精确匹配**（不区分大小写，服务端按大写归一后比对）；命中返回单条记录，无匹配返回空列表而非 `404` |

多个搜索字段和 `status` 同时提供时按 AND 组合。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "good": {
          "id": 1,
          "name": "明信片套装",
          "image": {
            "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
            "width": 1080,
            "height": 720
          },
          "score_price": 500
        },
        "quantity": 1,
        "score_cost": 500,
        "status": "pending",
        "verify_code": "VX78A9B2",
        "exchange_at": null,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 核销兑换

```
PUT /api/admin/exchange/{id}/verify
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为兑换记录 ID（int）。核销与取消统一走本接口，不接受防伪码——扫码核销时先用 `GET /api/admin/exchange?verify_code=...` 查到记录、弹窗人工核对，再用查得的 `id` 调本接口确认，避免"扫到即核销"的误操作。若该记录已核销或已取消（重复处理），返回 `409`、`code=8`；`action` 缺失或非法枚举返回 `400`、`code=3`。

**请求参数**

| 参数   | 位置 | 类型   | 必填 | 说明                                               |
| ------ | ---- | ------ | ---- | -------------------------------------------------- |
| id     | path | int    | 是   | 兑换记录 ID                                        |
| action | body | string | 是   | `verify`（核销）/ `cancel`（取消，退回积分和库存） |

**返回** `200`

- 核销成功：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "verified" },
  "message": "",
  "code": 0
}
```

- 取消成功：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "cancelled" },
  "message": "",
  "code": 0
}
```

- 已处理过（409）：

```json
{
  "success": false,
  "resp": null,
  "message": "冲突错误: 该兑换记录已处理",
  "code": 8
}
```

---

### 反馈管理

#### 反馈列表

```
GET /api/admin/feedback
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端反馈列表，支持按关键词、反馈类型及处理状态（`pending`/`resolved`）进行筛选。

**请求参数（Query）**

| 参数         | 类型   | 必填 | 默认值 | 说明                                 |
| ------------ | ------ | ---- | ------ | ------------------------------------ |
| page         | int    | 否   | 1      | 页码                                 |
| page_size    | int    | 否   | 10     | 每页数量（max=20）                   |
| type         | int    | 否   | —      | 1-内容 2-玩法 3-技术 4-其他          |
| status       | string | 否   | —      | `pending` / `resolved`；不传返回全部 |
| keyword      | string | 否   | —      | 按反馈 ID、标题或内容搜索（最长 50） |
| user_keyword | string | 否   | —      | 按提交者 ID 或昵称搜索（最长 50）    |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "user": {
          "id": 1,
          "nickname": "张三",
          "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
        },
        "title": "建议增加功能",
        "type": 2,
        "status": "pending",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 反馈详情

```
GET /api/admin/feedback/{id}
```

**权限**：管理员（Level ≥ 2）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "user": {
      "id": 1,
      "nickname": "张三",
      "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example"
    },
    "title": "建议增加功能",
    "content": "希望可以增加排行榜功能",
    "type": 2,
    "phone": "13800138000",
    "status": "pending",
    "media_file": {
      "origin_url": "https://media.example.com/feedbacks/attachment.jpg?signature=example",
      "thumb_url": "https://media.example.com/feedbacks/attachment_thumb.jpg?signature=example",
      "width": 1080,
      "height": 720,
      "media_type": 1
    },
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

附件 `media_file` 下发包含 `media_type` 的媒体对象（`origin_url`、`thumb_url`、`width`、`height`、`media_type`：`1` 图片、`2` 视频），供前端精准区分渲染组件；未上传附件时返回 `null`。

#### 处理反馈

```
PUT /api/admin/feedback/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为反馈 ID。管理员处理并标记反馈状态（如更新处理备注与状态为 `resolved`）。

**请求参数（JSON Body）**

| 参数   | 类型   | 必填 | 说明                   |
| ------ | ------ | ---- | ---------------------- |
| status | string | 是   | `pending` / `resolved` |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "resolved" },
  "message": "",
  "code": 0
}
```

---

### 工作台统计

```
GET /api/admin/stats
```

**权限**：管理员（Level ≥ 2）

**说明**：管理端工作台统计数字，一次请求返回全部计数，避免工作台多次请求。仅返回计数、不含任何明细（因此 Level 2 没有用户列表权限也可以展示用户总数）；后续新增统计字段属非破坏性改动。

| 字段                   | 说明                                        |
| ---------------------- | ------------------------------------------- |
| user_count             | 全站用户总数（含被封禁账号与 Level 2/3）    |
| pending_photo_count    | 待审核投稿数（`status=pending` 的题目投稿） |
| pending_attempt_count  | 待审核作答数（`status=pending` 的作答记录） |
| pending_comment_count  | 待审核评论数（`status=pending` 的评论）     |
| pending_feedback_count | 待处理反馈数（`status=pending` 的反馈）     |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "user_count": 1234,
    "pending_photo_count": 5,
    "pending_attempt_count": 8,
    "pending_comment_count": 3,
    "pending_feedback_count": 2
  },
  "message": "",
  "code": 0
}
```

---

### 用户管理

#### 用户列表

```
GET /api/admin/users
```

**权限**：超级管理员（Level 3）

**说明**：用户治理专用的统一查询接口，仅 Level 3 可调用（Level 2 返回 `403`）。`keyword` 按用户 ID、学号、姓名或昵称模糊搜索；`status`、`level` 为筛选项，与 `keyword` 按 AND 组合。所有筛选均省略时返回全部用户。工作台的用户总数展示用 `GET /admin/stats`，不依赖本接口。账号状态与权限等级相互独立。Level 3 记录只读展示，不提供业务侧治理操作。

**请求参数（Query）**

| 参数      | 类型   | 必填 | 默认值 | 说明                                       |
| --------- | ------ | ---- | ------ | ------------------------------------------ |
| keyword   | string | 否   | —      | 按用户 ID、学号、姓名或昵称搜索（最长 50） |
| page      | int    | 否   | 1      | 页码                                       |
| page_size | int    | 否   | 10     | 每页数量（max=20）                         |
| status    | string | 否   | —      | `active` / `banned`；不传返回全部          |
| level     | int    | 否   | —      | `1` 普通用户 / `2` 管理员 / `3` 超级管理员 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 2,
    "list": [
      {
        "id": 1,
        "netid": "20230001",
        "username": "张三",
        "nickname": "张三",
        "avatar": "https://media.example.com/avatars/avatar.jpg?signature=example",
        "level": 1,
        "status": "active"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 封禁/解封用户

```
PUT /api/admin/users/{id}/status
```

**权限**：超级管理员（Level 3）

**说明**：仅超级管理员（Level 3）可调用，可封禁或解封 Level 1、2 用户；禁止操作任何 Level 3 超级管理员账号，违反拦截返回 `403`。

**请求参数（JSON Body）**

| 参数   | 类型   | 必填 | 说明                          |
| ------ | ------ | ---- | ----------------------------- |
| status | string | 是   | `banned` 封禁 / `active` 解封 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 2, "status": "banned" },
  "message": "",
  "code": 0
}
```

---

#### 调整权限等级

```
PUT /api/admin/users/{id}/level
```

**权限**：超级管理员（Level 3）

**说明**：仅 Level 3 超级管理员可操作。目标等级只允许 1/2，用于普通用户与管理员之间调整，且不改变账号 `status`；禁止调整 Level 3 超级管理员等级，违反拦截返回 `403`。

**请求参数（JSON Body）**

| 参数         | 类型 | 必填 | 说明                      |
| ------------ | ---- | ---- | ------------------------- |
| target_level | int  | 是   | `1` 普通用户 / `2` 管理员 |

`{id}` 为目标用户 ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 2, "status": "success" },
  "message": "",
  "code": 0
}
```

---

## 状态枚举参考

| 实体                 | 字段    | 可选值                               | 默认值   |
| -------------------- | ------- | ------------------------------------ | -------- |
| Photo                | status  | `pending` / `approved` / `rejected`  | pending  |
| Attempt              | status  | `pending` / `solved` / `unsolved`    | pending  |
| Comment              | status  | `pending` / `approved` / `rejected`  | pending  |
| Good                 | status  | `in_store` / `out_store`             | in_store |
| Exchange             | status  | `pending` / `verified` / `cancelled` | pending  |
| Feedback             | status  | `pending` / `resolved`               | pending  |
| User                 | level   | `1` / `2` / `3`                      | 1        |
| User                 | status  | `active` / `banned`                  | active   |
| Activity list filter | status  | `not_started` / `active` / `ended`   | —        |
| InteractionMessage   | type    | `like` / `comment`                   | —        |
| InteractionMessage   | is_read | `true` / `false`，本次读取前状态     | false    |
| Announcement         | is_read | `true` / `false`，读详情即已读       | false    |

`User.level=3` 表示由部署侧管理的超级管理员，可存在多个。业务接口可以只读查询其摘要，但不能新增、修改、删除或改变其状态和等级。

写操作响应中的 `resp.status` 必须限制为对应业务状态：新建待处理记录为 `pending`，管理员新增题目为 `approved`，删除回执为 `deleted`，题目/评论审核为 `approved` / `rejected`，作答审核为 `solved` / `unsolved`，奖品为 `in_store` / `out_store`，兑换处理为 `verified` / `cancelled`，反馈为 `pending` / `resolved`，通知发布或更新为 `published`，用户封禁/解封为 `active` / `banned`。仅表示通用成功的接口固定为 `success`。

**可空字段**（未设置时返回 `null`）：`exchange_at`、`reject_reason`。
