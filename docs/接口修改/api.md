
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

## 鉴权

| Level | 说明 |
|-------|------|
| 0 | 未登录请求上下文，不是用户账号的 `level` 值 |
| 1 | 客户端普通登录用户，仅用于 tuxun-fe |
| 2 | 后台管理员，可进入 tuxun-admin-fe |
| 3 | 创世管理员，由部署侧管理，可存在多个 |

**Session Cookie**：`tz-sessions`，有效期 30 分钟，每次请求自动续期，退出后立即失效。
生产环境启用 `HttpOnly`、`Secure` 和 `SameSite=Lax`。`SameSite=Lax` 用于降低 CSRF 风险；后端仍应对写操作校验请求来源或采用等效的 CSRF 防护。

**登录态、权限等级与账号状态相互独立**：未登录是当前请求没有有效 Session，不写入用户表，也不把用户 `level` 设为 0。账号 `level` 只允许 `1`（普通用户）、`2`（管理员）、`3`（创世管理员）；`status` 独立为 `active` / `banned`。封禁或解封不改变 `level`，调整 `level` 也不改变 `status`。

| 操作者 | 可封禁 / 解封的目标 | 可调整权限等级的目标 |
|-------|---------------------|----------------------|
| Level 1 | 无 | 无 |
| Level 2 | 仅 Level 1 | 无 |
| Level 3 | Level 1、Level 2 | Level 1、Level 2，可在两级之间调整 |

Level 3 可以有多个，但任何 Level 3 都不能治理自己或其他 Level 3。Level 3 账号的创建、修改和删除，以及其状态或等级变更，只能通过部署侧受控流程完成；业务用户列表可只读展示其摘要。所有账号治理操作均应记录操作者、目标账号、变更前后值和时间。

**HTTP 状态码语义**：下表统一说明状态码含义；每个接口实际承诺返回的状态码以 `apifox-import.json` 对应 operation 的 `responses` 为准。

| 状态码 | 场景 |
|--------|------|
| 200 | 查询、更新、删除等操作成功 |
| 201 | 资源创建成功 |
| 204 | 查询成功但没有可返回内容（如当前没有有效全局公告） |
| 302 | 登录流程重定向 |
| 400 | 请求参数无效或业务前置条件不满足（如库存、积分不足） |
| 401 | 未登录或 Session 失效 |
| 403 | 权限不足 |
| 404 | 路径 `{id}` 定位的资源不存在（`code=5`）；请求体引用的对象不存在用 `400` |
| 409 | 冲突（重复审核、重复核销、幂等键冲突、并发冲突） |
| 429 | 超出接口声明的修改次数 / 频率限制（`code=9`） |
| 500 | 未预期的服务器内部错误 |

**错误码（`code`）**：

| code | message 前缀 | 说明 |
|------|-------------|------|
| 0 | — | 成功 |
| 3 | 参数错误 | 请求参数不合法 |
| 4 | 系统错误 | 服务器内部错误 |
| 5 | 操作错误 | 业务逻辑限制 |
| 6 | 鉴权错误 | 未登录 |
| 7 | 权限错误 | 权限不足 |
| 8 | 冲突错误 | 重复操作 / 并发冲突 |
| 9 | 频率限制 | 超出修改次数 / 频率限制 |

---

## 目录

- [契约维护规则](#契约维护规则)
- [公共数据结构](#公共数据结构)
- [用户认证](#用户认证)
- [活动](#活动)
- [图寻题目 (Photos)](#图寻题目-photos)
- [答题 (Attempts)](#答题-attempts)
- [评论 (Comments)](#评论-comments)
- [积分 (Score)](#积分-score)
- [奖品 (Goods)](#奖品-goods)
- [兑换 (Exchange)](#兑换-exchange)
- [通知 (Notifications)](#通知-notifications)
- [反馈 (Feedback)](#反馈-feedback)
- [管理员](#管理员)

---

## 约定

- **基础路径**：默认使用当前部署同源的 `/api`；本地开发可使用 `http://127.0.0.1:8088/api`。
- **分页**：`page` 默认 1（min=1），`page_size` 默认 10（min=1, max=20）。空列表返回 `[]`，不返回 `null`。
- **时间**：纯日期格式 `YYYY-MM-DD`；具体时间格式 ISO 8601 带时区 `2026-07-20T10:30:00+08:00`。
- **空值**：未设置的时间字段（如 `exchange_at`、`reviewed_at`）返回 `null`。
- **媒体 URL**：`avatar_url`、`cover_url`、`image_url`、`thumb_url`、`guess_image_url` 和附件 `url` 均由后端返回可直接访问的完整 URL。生产环境可能返回有有效期的私有对象存储签名 URL。客户端必须将 URL 视为不透明字符串，不得自行拼接、解析或长期持久化；失效后重新请求对应业务接口获取最新地址。文档中的 `media.example.com` 仅为示例域名。
- **字段命名**：JSON 字段统一使用 `snake_case`；Query 参数统一使用 `snake_case`（如 `page_size`、`sort_by`、`activity_id`）。
- **上传文件**：仅支持 jpg/png，单文件 ≤20MB。反馈附件最多 3 个。
- **标题长度**：所有请求中的 `title` 字段（题目、活动、通知、反馈）按 Unicode 码点计数，最长 20 个字符，超长返回 `400`。
- **列表字段**：分页响应中数组字段统一命名为 `list`。
- **管理端搜索**：所有管理后台分页列表均提供服务端搜索；字符串搜索参数最长 50 个字符，省略或空字符串表示不做该项搜索。关键词和状态、类型等筛选同时提供时按 AND 组合。详情页、编辑页和工作台等非列表界面不强制增加无意义的搜索框。
- **接口复用**：接口权限表示最低 Level，Level 2/3 可直接复用数据范围和行为相同的普通用户接口，不因调用者是管理员而重复建接口。只有管理端需要查看非公开状态、额外管理字段或执行不同写入规则时，才使用独立的 `/admin` 接口。
- **资源不存在**：所有以路径 `{id}` 定位资源的接口，目标不存在时返回 `404`（`code=5`）；请求体引用的对象（如 `good_id`、`exchange_id`）不存在时返回 `400`。各接口的 `responses` 已逐一声明。

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
4. 时间格式和媒体 URL 遵循[约定](#约定)一节。

**校验**

1. 每次修改后、提交前，在 `tu-xun` 根目录运行 `python3 check_contract.py`（或 `uv run python check_contract.py`），全部通过才允许提交。脚本校验 JSON 合法性、`operationId` 唯一性、`$ref` 可解析性，以及 `api.md` 与 `apifox-import.json` 接口清单的双向一致。

---

## 公共数据结构

OpenAPI 中的公共结构统一定义在 `components.schemas`，各接口通过 `$ref` 或 `allOf` 引用，避免同一业务对象在不同接口中产生独立类型。

| Schema | 用途 |
|------|------|
| `SuccessResponseBase` / `ErrorResponseBase` | 统一 `success`、`message`、`code` 响应字段 |
| `StandardErrorResponse` | 标准错误响应，继承 `ErrorResponseBase` 且 `resp` 固定为 `null` |
| `MediaUrl` | 头像、封面、图片、缩略图和附件的可访问 URL |
| `UserBrief` | 内容作者等简要用户信息 |
| `UserSummary` | 登录和管理员用户列表使用的完整用户摘要 |
| `ActivityBrief` | 图片关联的简要活动信息，仅包含 `id`、`title` |
| `ActivityCard` | 进行中活动、历史活动和管理员活动列表共用的卡片结构 |
| `ActivityDetail` | 在 `ActivityCard` 基础上扩展积分和奖励阶梯 |
| `GoodBrief` | 商品列表和兑换记录中的商品摘要 |
| `LikeResult` | 图片、答题和评论点赞接口的统一结果 |
| `NotificationListItem` | 通知列表摘要，只包含标题、正文预览、已读状态和时间等列表必需字段 |
| `Notification` | 通知详情结构；管理员发布的普通通知可包含完整正文和图片 |
| `AdminPhotoListItem` | 管理端统一题目池列表项，包含所属活动、作者、审核状态和编辑所需字段 |
| `UserAttemptItem` | 两个“我的答题”列表共用的列表项 |
| `IdResult` | 写操作回执中的公共 `id` 字段，各接口自行限制 `status` 枚举 |
| `ReviewActionRequest` | 图片和评论审核的统一请求结构 |
| `GoodUpsertForm` | 商品新增和编辑的公共表单字段 |
| `PageBase` | 所有分页结构共享的 `total` 字段 |
| `CreateNotificationRequest` / `UpdateNotificationRequest` | 管理员以 multipart 发布或编辑一般通知、全局公告及可选图片 |
| `AdminPhotoUpsertForm` | 管理员新增或编辑活动题目的 multipart 表单 |
| `SetLikeRequest` | 图片、答题和评论共用的幂等点赞状态请求 |
| `PhotoCardBase` | 题目卡片的活动、标题、介绍、缩略图和统计等公共字段 |
| `PhotoCard` | 首页题目卡片，在 `PhotoCardBase` 上增加作者 |
| `UserPhotoCard` | 我的投稿卡片，在 `PhotoCardBase` 上增加审核状态 |

分页结构以 `PageBase` 复用分页元数据，并按列表项类型分别定义 `ActivityCardPage`、`PhotoCardPage`、`UserPhotoCardPage`、`UserSummaryPage`、`GoodBriefPage`、`UserAttemptPage`、`NotificationPage` 和 `AdminPhotoListPage`。其他列表响应同样通过 `allOf` 组合 `PageBase` 和明确的 `list[]` schema，不使用会丢失列表项类型的无类型分页模型。

### ActivityCard

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 活动 ID |
| title | string | 活动标题 |
| cover_url | MediaUrl | 活动封面 |
| description | string | 活动简介 |
| start_time | string(date-time) | 必返非空的带时区开始时间 |
| end_time | string(date-time) | 必返非空的带时区结束时间，且必须晚于 `start_time` |

`ActivityDetail` 继承以上字段，并增加 `photo_points` 和数组类型的 `reward_tiers`。

活动契约不返回 `status` 或 `is_active` 等派生状态字段。后端必须使用服务器当前时间按以下唯一规则进行列表归类和写操作校验：

- `now < start_time`：仅在管理端活动列表中存在，不出现在客户端进行中或往期列表；
- `start_time <= now < end_time`：出现在 `/activity/active`，允许投稿和答题；
- `now >= end_time`：出现在 `/activity/history`，不再允许新投稿和答题。

`start_time` 和 `end_time` 在新建、更新及读取时均为必填非空字段，并且必须满足 `end_time > start_time`。时间比较按 ISO 8601 时刻进行。

活动一旦达到原 `end_time`，题目答案可能已经按活动结束规则公开；后续更新不得把 `end_time` 延长到服务器当前时间之后重新开放活动。需要提前结束时可以将 `end_time` 调到当前时间之前，但结束状态不可撤回。

---

## 用户认证

### 开发/测试登录

```
GET /api/test/login
```

**环境范围**：仅在开发和测试环境开启，生产环境必须关闭。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| netid | string | 否 | 测试用户学号 |
| username | string | 否 | 测试用户姓名 |
| password | string | 是 | 测试登录密码 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "测试用户",
    "nickname": "测试用户",
    "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "level": 1,
    "status": "active"
  },
  "message": "",
  "code": 0
}
```

**失败** `403`：账号已被封禁，拒绝建立会话（`code=7`）。

---

### 1. 登录

```
GET /api/user/login
```

**权限**：无

**说明**：重定向到学校统一认证页面。已登录用户直接重定向到回调地址。

**响应**：302 重定向

---

### 2. 登录回调

```
GET /api/user/logincallback
```

**权限**：无

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| guid | string | 是 | 学校统一认证返回的 GUID |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "张三",
    "nickname": "张三",
    "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "level": 1,
    "status": "active"
  },
  "message": "",
  "code": 0
}
```

**失败** `403`：账号已被封禁，拒绝建立会话（`code=7`）。

---

### 3. 登出

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

### 4. 获取个人信息

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
    "name": "张三",
    "nickname": "张三",
    "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example",
    "score_count": 1000,
    "level": 1
  },
  "message": "",
  "code": 0
}
```

---

### 5. 修改个人资料

```
PUT /api/user/info
```

**权限**：登录用户（Level ≥ 1）

**说明**：昵称每自然月最多修改 **4 次**（按服务器时区，自然月初重置）。仅当请求携带 `nickname` 且与当前昵称不同时计 1 次，且成功修改才计数；提交相同昵称不计数、不报错。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称（最长 10） |

**返回** `200`

```json
{
  "success": true,
  "resp": null,
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

### 6. 修改头像

```
PUT /api/user/avatar
```

**权限**：登录用户（Level ≥ 1）

**Content-Type**：`multipart/form-data`

**说明**：头像每自然月最多修改 **10 次**（按服务器时区，自然月初重置），成功修改才计数。文件类型或大小不符合要求时返回 `400`（`code=3`）。

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| avatar | file | 是 | 头像文件（jpg/png，≤20MB） |

**返回** `200`

```json
{
  "success": true,
  "resp": null,
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

### 1. 进行中活动列表

```
GET /api/activity/active
```

**权限**：无

**说明**：返回当前时间处于 `start_time`（含）与 `end_time`（不含）之间的全部活动，允许多个活动同时进行。没有进行中活动时返回空列表。默认按 `start_time` 倒序、`id` 倒序返回，保证稳定分页。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |

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
        "cover_url": "https://media.example.com/photos/cover.jpg?signature=example",
        "description": "活动介绍",
        "start_time": "2026-06-01T00:00:00+08:00",
        "end_time": "2026-06-30T23:59:59+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 往期活动列表

```
GET /api/activity/history
```

**权限**：无

**说明**：仅返回 `end_time <= 当前时间` 的已结束活动，不包含尚未开始的活动。默认按 `end_time` 倒序、`id` 倒序返回，保证稳定分页。活动进入往期时不迁移题目，题目始终保留原 `activity_id`。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "title": "第一期活动",
        "cover_url": "https://media.example.com/photos/cover.jpg?signature=example",
        "description": "活动介绍",
        "start_time": "2026-05-01T00:00:00+08:00",
        "end_time": "2026-05-31T23:59:59+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

## 图寻题目 (Photos)

### 1. 上传投稿

```
POST /api/photos
```

**权限**：登录用户（Level ≥ 1）

**活动约束**：`activity_id` 必须对应满足 `start_time <= now < end_time` 的进行中活动。后端必须根据服务器当前时间校验，不接受尚未开始或已结束活动的新投稿；校验失败返回 `400`、`code=5`。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 所属活动 ID |
| title | string | 是 | 图片标题（最长 20） |
| description | string | 否 | 图片描述/故事（最长 50） |
| image_file | file | 是 | 图片文件（jpg/png，≤20MB） |
| longitude | float | 是 | 经度 |
| latitude | float | 是 | 纬度 |
| coord_type | string | 是 | 坐标系：`wgs84` / `gcj02` / `bd09` |

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

### 2. 图片列表

```
GET /api/photos/list
```

**权限**：无

**说明**：作为首页题目流使用。不传 `activity_id` 时聚合所有进行中活动的题目；每道题仍只属于一个活动，卡片通过 `activity.title` 展示 `#活动名`。传入 `activity_id` 时可查询进行中或已结束活动的题目，用于活动主页和往期详情；尚未开始的活动不对客户端暴露题目，返回空列表。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 否 | — | 按单个活动筛选；不传时聚合所有进行中活动的题目 |
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |
| solved | bool | 否 | — | 筛选是否已破解 |
| sort_by | string | 否 | created_at | `created_at` / `likes_count` / `attempts_count` |
| keyword | string | 否 | — | 关键词搜索（最长 50） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "author": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "activity": { "id": 1, "title": "寻找校园角落" },
        "title": "猜猜这是哪",
        "description": "校园神秘角落",
        "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
        "solved": false,
        "likes_count": 10,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 3. 图片详情

```
GET /api/photos/{id}
```

**权限**：无

**说明**：`{id}` 为图片（Photo）ID。**仅当题目所属活动已结束**（服务器时间 ≥ `end_time`）时，响应才包含答案坐标 `longitude`、`latitude`、`coord_type`（投稿时提交的原始值与坐标系）；活动进行中或未开始时不返回这三个字段，答案不提前下发。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "author": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
    "activity": { "id": 1, "title": "寻找校园角落" },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image_url": "https://media.example.com/photos/photo.jpg?signature=example",
    "longitude": 108.123456,
    "latitude": 34.123456,
    "coord_type": "gcj02",
    "solved": false,
    "attempts_count": 5,
    "likes_count": 10,
    "created_at": "2026-06-01T12:00:00+08:00",
    "status": "approved"
  },
  "message": "",
  "code": 0
}
```

---

### 4. 图片流展示

```
GET /api/photos/{id}/image
```

**权限**：无

**响应**：`200`，后端从私有对象存储读取并返回图片二进制流。OpenAPI content 为 `image/*`，schema 为 `type: string, format: binary`；实际 `Content-Type` 为 `image/jpeg` 或 `image/png`。客户端无需处理对象存储鉴权。

---

### 5. 图片下载

```
GET /api/photos/{id}/download
```

**权限**：无

**响应**：`200`，后端从私有对象存储读取并返回 `application/octet-stream` 二进制流；schema 为 `type: string, format: binary`，并返回 `Content-Disposition: attachment`。客户端无需处理对象存储鉴权。

---

### 6. 图片评论列表

```
GET /api/photos/{id}/comments
```

**权限**：无

**说明**：`{id}` 为图片（Photo）ID。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| sort_by | string | 否 | created_at | `created_at` / `likes_count` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "author": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "photo_id": 1,
        "comment_text": "我知道这是哪里！",
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

### 7. 提交答题

```
POST /api/photos/{id}/attempts
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。只有题目所属活动满足 `start_time <= now < end_time` 时才允许提交新答题；尚未开始或已结束时返回 `400`、`code=5`。答题自动归属题目的活动，客户端不另行传递 `activity_id`。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| image_file | file | 是 | 猜测的匹配照片（jpg/png，≤20MB） |
| longitude | float | 是 | 猜测经度 |
| latitude | float | 是 | 猜测纬度 |
| coord_type | string | 是 | 坐标系：`wgs84` / `gcj02` / `bd09` |

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

### 8. 图片答题列表

```
GET /api/photos/{id}/attempts
```

**权限**：无

**说明**：`{id}` 为图片（Photo）ID。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| sort_by | string | 否 | created_at | `created_at` / `likes_count` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "author": { "id": 2, "nickname": "李四", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "photo_id": 1,
        "image_url": "https://media.example.com/attempts/attempt.jpg?signature=example",
        "solved": false,
        "likes_count": 0,
        "created_at": "2026-06-02T10:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 9. 用户在某图片下的答题列表

```
GET /api/photos/{id}/attempts/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。获取当前登录用户在该图片下的答题记录。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |
| status | string | 否 | — | `pending` / `unsolved` / `solved` |
| sort_by | string | 否 | created_at | `created_at` / `likes_count` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 3,
    "list": [
      {
        "id": 1,
        "photo": {
          "id": 1,
          "title": "猜猜这是哪",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "activity": { "id": 1, "title": "寻找校园角落" }
        },
        "image_url": "https://media.example.com/attempts/attempt.jpg?signature=example",
        "longitude": 108.123456,
        "latitude": 34.123456,
        "coord_type": "gcj02",
        "likes_count": 0,
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

### 10. 设置图片点赞状态

```
PUT /api/photos/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。接口为幂等写操作，重复提交相同状态不会反向切换。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_like | bool | 是 | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 11 },
  "message": "",
  "code": 0
}
```

---

### 11. 获取图片点赞状态

```
GET /api/photos/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。未登录返回 401。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 11 },
  "message": "",
  "code": 0
}
```

---

### 12. 发表评论

```
POST /api/photos/{id}/comments
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| comment_text | string | 是 | 评论内容（最长 140） |

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

### 13. 我的投稿列表

```
GET /api/photos/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：可按单个活动筛选；不传 `activity_id` 时返回当前用户的全部投稿。

列表保持扁平分页结构，但后端必须保证同一活动的记录连续返回：活动分段按关联活动的 `start_time` 倒序排列（开始时间相同时按 `activity.id` 倒序保证稳定顺序），每个活动分段内再按 `sort_by` 排序，默认为 `created_at` 倒序，相同时按记录 `id` 倒序。

分页以投稿记录数量计算，一个活动分段可能跨页。前端必须使用 `activity.id` 合并相邻分页中的同一分段，并使用 `activity.title` 作为分段标题；不得以可能重名的活动标题作为分组键。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 否 | — | 按单个活动筛选；不传时返回当前用户的全部投稿 |
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |
| solved | bool | 否 | — | 筛选是否已破解 |
| sort_by | string | 否 | created_at | 活动分段内排序：`created_at` / `likes_count` / `attempts_count` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "activity": { "id": 1, "title": "寻找校园角落" },
        "title": "猜猜这是哪",
        "description": "校园神秘角落",
        "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
        "solved": false,
        "likes_count": 10,
        "created_at": "2026-06-01T12:00:00+08:00",
        "status": "approved",
        "reject_reason": null
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 14. 我的投稿详情

```
GET /api/photos/review/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为图片（Photo）ID。只能查看自己的投稿（管理员可查看任意）。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "activity": { "id": 1, "title": "寻找校园角落" },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image_url": "https://media.example.com/photos/photo.jpg?signature=example",
    "longitude": 108.123456,
    "latitude": 34.123456,
    "coord_type": "gcj02",
    "solved": false,
    "likes_count": 10,
    "attempts_count": 5,
    "created_at": "2026-06-01T12:00:00+08:00",
    "status": "approved",
    "reject_reason": null
  },
  "message": "",
  "code": 0
}
```

---

## 答题 (Attempts)

### 1. 设置答题点赞状态

```
PUT /api/attempts/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为答题记录（Attempt）ID。接口为幂等写操作。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_like | bool | 是 | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 5 },
  "message": "",
  "code": 0
}
```

---

### 2. 获取答题点赞状态

```
GET /api/attempts/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为答题记录（Attempt）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 5 },
  "message": "",
  "code": 0
}
```

---

### 3. 我的答题列表

```
GET /api/attempts/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：可按单个活动筛选；不传 `activity_id` 时返回当前用户的全部答题记录。

列表保持扁平分页结构，但后端必须保证同一活动的记录连续返回：活动分段按关联活动的 `start_time` 倒序排列（开始时间相同时按 `photo.activity.id` 倒序保证稳定顺序），每个活动分段内再按 `sort_by` 排序，默认为 `created_at` 倒序，相同时按答题记录 `id` 倒序。

分页以答题记录数量计算，一个活动分段可能跨页。前端必须使用 `photo.activity.id` 合并相邻分页中的同一分段，并使用 `photo.activity.title` 作为分段标题。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 否 | — | 按单个活动筛选；不传时返回当前用户的全部答题 |
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |
| status | string | 否 | — | `pending` / `unsolved` / `solved` |
| sort_by | string | 否 | created_at | 活动分段内排序：`created_at` / `likes_count` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "photo": {
          "id": 1,
          "title": "猜猜这是哪",
          "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
          "activity": { "id": 1, "title": "寻找校园角落" }
        },
        "image_url": "https://media.example.com/attempts/attempt.jpg?signature=example",
        "longitude": 108.123456,
        "latitude": 34.123456,
        "coord_type": "gcj02",
        "likes_count": 0,
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

用户答题记录中的 `longitude` / `latitude` 与其对应的 `coord_type` 成对使用；管理端审核列表另外返回 `guess_coord_type`，用于解释答题者提交的猜测坐标。

---

## 评论 (Comments)

### 1. 删除评论

```
DELETE /api/comments/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为评论（Comment）ID。普通用户只能删除自己的评论，管理员可删除任意评论。

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

### 2. 设置评论点赞状态

```
PUT /api/comments/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为评论（Comment）ID。接口为幂等写操作。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| is_like | bool | 是 | `true` 点赞，`false` 取消点赞 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 3 },
  "message": "",
  "code": 0
}
```

---

### 3. 获取评论点赞状态

```
GET /api/comments/{id}/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`{id}` 为评论（Comment）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "likes_count": 3 },
  "message": "",
  "code": 0
}
```

---

## 积分 (Score)

### 1. 我的积分

```
GET /api/score
```

**权限**：登录用户（Level ≥ 1）

**返回** `200`

```json
{
  "success": true,
  "resp": { "total_score": 1000 },
  "message": "",
  "code": 0
}
```

---

### 2. 积分流水

```
GET /api/score/logs
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "delta": 10,
        "balance": 1000,
        "reason": "upload_photo",
        "related_id": 1,
        "related_type": "photo",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

**reason 类型**：`upload_photo` / `answer_correct` / `like_photo` / `get_liked` / `comment` / `review_pass` / `daily_login` / `admin_adjust` / `exchange`

---

## 奖品 (Goods)

### 1. 奖品列表

```
GET /api/goods/list
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| available | bool | 否 | — | 仅看可兑换（库存 > 0） |

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
        "thumb_url": "https://media.example.com/goods/thumb.jpg?signature=example",
        "need_score": 500,
        "stock": 20
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 2. 奖品详情

```
GET /api/goods/{id}
```

**权限**：登录用户（Level ≥ 1）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "name": "明信片套装",
    "description": "精美校园风景明信片",
    "image_url": "https://media.example.com/goods/good.jpg?signature=example",
    "need_score": 500,
    "stock": 20
  },
  "message": "",
  "code": 0
}
```

---

## 兑换 (Exchange)

### 1. 兑换奖品

```
POST /api/exchange/claim
```

**权限**：登录用户（Level ≥ 1）

**幂等要求**：请求头必须携带 `Idempotency-Key`（16-128 字符）。同一用户使用同一键和相同请求内容重复提交时，后端返回首次请求结果，不得重复创建兑换记录、扣减库存或扣减积分；同一键对应不同请求内容时返回 `409`。兑换记录创建、库存扣减和积分扣减必须处于同一原子事务或具备等效的一致性保证。

| Header | 类型 | 必填 | 说明 |
|------|------|------|------|
| Idempotency-Key | string | 是 | 一次兑换操作的唯一键，建议使用 UUID |

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| good_id | int | 是 | 奖品 ID |
| quantity | int | 是 | 兑换数量 |

**返回** `201`

- 成功：

```json
{
  "success": true,
  "resp": { "id": 1, "status": "pending" },
  "message": "",
  "code": 0
}
```

- 请求参数无效、库存不足或积分不足：`400`

```json
{
  "success": false,
  "resp": null,
  "message": "参数错误: 奖品库存不足",
  "code": 3
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
GET /api/exchange/list
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `pending` / `verified` / `cancelled` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "good": { "id": 1, "name": "明信片套装", "thumb_url": "https://media.example.com/goods/thumb.jpg?signature=example", "need_score": 500, "stock": 20 },
        "quantity": 1,
        "score_cost": 500,
        "status": "pending",
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

通知统一包含普通通知和互动消息。普通通知由管理员发布，包含一般通知和全局公告；互动消息由点赞、评论、审核等业务事件自动生成。

`category=normal` 时 `type` 只能为 `general` 或 `global_announcement`；`category=interaction` 时 `type` 只能为 `like`、`comment` 或 `review`。

全局公告是普通通知的特殊类型：登录用户进入应用时通过专用接口读取当前未读且未过期的公告并弹窗展示，同时仍出现在通知列表中。

| category | type | 产生方式 | 条件字段 | 客户端行为 |
|------|------|------|------|------|
| `normal` | `general` | 管理员发布 | 可选 `image_url`；`related_type` / `related_id` 同时提供或同时省略；不返回 `expires_at` | 通知列表展示，可展示图片并按关联对象跳转 |
| `normal` | `global_announcement` | 管理员发布 | 可选 `image_url`；必返 `expires_at`；不返回关联对象 | 已登录用户进入应用时弹窗展示，同时保留在通知列表中 |
| `interaction` | `like` / `comment` / `review` | 业务事件自动生成 | 按事件返回 `sender_id`、`related_type`、`related_id`；不返回图片和 `expires_at` | 通知列表展示并跳转到相关业务对象 |

通知接口不向客户端返回接收人的 `user_id`。接收人关系和已读记录由后端内部维护；响应中的 `is_read` 始终针对当前 Session 用户计算。管理员发布的通知没有图片时，详情中省略 `image_url`。

一般通知和全局公告属于全局发布记录，面向全部已登录用户；因此 Level 2/3 使用 `category=normal` 查询时可以读取全部未删除发布记录，包括早于当前管理员账号创建时间的记录。互动消息只投递给业务事件对应的目标用户。访客公告不属于 `Notification`，未来如有需要应单独设计。

### 1. 通知列表

```
GET /api/notifications
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| category | string | 否 | — | `normal` / `interaction`，不传返回全部 |
| type | string | 否 | — | `general` / `global_announcement` / `like` / `comment` / `review` |
| related_type | string | 否 | — | `activity` / `photo` / `attempt` / `comment` / `feedback`；与 `related_id` 同时使用 |
| related_id | int | 否 | — | 关联对象 ID；与 `related_type` 同时使用 |
| keyword | string | 否 | — | 按通知 ID、标题或完整正文搜索（最长 50），不只匹配摘要 |

`related_type` 和 `related_id` 只提供其中一个，或 `category` 与 `type` 组合不合法时返回 `400`。多个筛选条件按 AND 组合。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "category": "interaction",
        "type": "review",
        "title": "审核通过通知",
        "content_preview": "您投稿的图片已通过审核",
        "is_read": false,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

列表项固定只返回 `id`、`category`、`type`、`title`、`content_preview`、`is_read`、`created_at`，不返回完整 `content`、`image_url`、关联对象或发送者字段。页面不得为补摘要而逐条请求详情。

`content_preview` 由后端从完整 `content` 生成，规则固定为：

1. 将所有字面量 `[image]` 图片占位标记替换为空格；
2. 将换行、制表符和连续 Unicode 空白合并为一个半角空格，并去除首尾空白；
3. 按 Unicode 码点而非字节截取前 50 个字符，不追加省略号；清洗后为空时返回空字符串。

`keyword` 搜索仍匹配原始完整正文，不对 `content_preview` 搜索，以免 50 字以后内容无法命中。

---

### 2. 当前全局公告

```
GET /api/notifications/global-announcement
```

**权限**：登录用户（Level ≥ 1）。

**说明**：用户登录并恢复 Session 后调用。返回当前用户最新一条未读、未过期的 `global_announcement`。如果存在多条，按 `created_at` 倒序、`id` 倒序返回第一条。用户关闭弹窗时调用 `/notifications/{id}/read` 标记已读，因此后续登录和跨设备访问都不会重复弹出该公告。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 10,
    "category": "normal",
    "type": "global_announcement",
    "title": "系统维护公告",
    "content": "今晚 23:00 至 23:30 进行系统维护。",
    "image_url": "https://media.example.com/notifications/maintenance.jpg?signature=example",
    "expires_at": "2026-07-31T23:59:59+08:00",
    "is_read": false,
    "created_at": "2026-07-21T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

当前用户没有未读且未过期的全局公告时返回 `204`，不返回响应体。未登录或权限不足分别返回 `401`、`403`。

---

### 3. 通知详情

```
GET /api/notifications/{id}
```

**权限**：登录用户（Level ≥ 1）

**说明**：普通通知对全部已登录用户可读；互动消息仅对对应接收人可读。管理后台读取一般通知和全局公告详情时复用本接口，不另设管理端只读详情接口。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "category": "normal",
    "type": "general",
    "title": "活动通知",
    "content": "新活动已经开始。\n[image]\n欢迎参加。",
    "image_url": "https://media.example.com/notifications/notice.jpg?signature=example",
    "related_id": 1,
    "related_type": "activity",
    "is_read": true,
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

`sender_id`、`image_url`、`related_id`、`related_type` 和 `expires_at` 仅在有对应值时返回。详情中的 `content` 保留完整正文和 `[image]` 占位标记，客户端可结合 `image_url` 按业务排版渲染。

---

### 4. 未读通知数

```
GET /api/notifications/unread-count
```

**权限**：登录用户（Level ≥ 1）

**返回** `200`

```json
{
  "success": true,
  "resp": { "count": 3 },
  "message": "",
  "code": 0
}
```

---

### 5. 标记已读

```
PUT /api/notifications/{id}/read
```

**权限**：登录用户（Level ≥ 1）

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

## 反馈 (Feedback)

### 1. 提交反馈

```
POST /api/feedback
```

**权限**：登录用户（Level ≥ 1）

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 标题（最长 20） |
| content | string | 是 | 内容（最长 500） |
| type | int | 是 | 反馈类型：1-内容 2-玩法 3-技术 4-其他 |
| phone | string | 否 | 联系电话（最长 20） |
| image_file1 | file | 否 | 附件 1（jpg/png，≤20MB） |
| image_file2 | file | 否 | 附件 2 |
| image_file3 | file | 否 | 附件 3 |

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

> 以下接口均需管理员权限（Level ≥ 2），标注「创世管理员」仅允许 Level 3。

---

### 题目管理与图片审核

#### 管理端统一题目池

```
GET /api/admin/photos
```

**权限**：管理员（Level ≥ 2）

**说明**：这是管理端唯一的题目列表接口。不传筛选时返回所有活动下的全部未删除题目，覆盖 `not_started`、`active`、`ended` 活动和 `pending`、`approved`、`rejected` 审核状态；默认按 `created_at` 倒序、`id` 倒序稳定分页。客户端公开题目列表仍遵守活动时间和审核状态规则。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `pending` / `approved` / `rejected`；不传返回全部审核状态 |
| activity_ids | int[] | 否 | — | 按多个活动 ID 筛选，使用重复参数，如 `activity_ids=1&activity_ids=2`；所选 ID 之间按 OR 匹配 |
| keyword | string | 否 | — | 按题目 ID、标题、描述、作者 ID 或昵称搜索（最长 50） |

`status`、`activity_ids`、`keyword` 同时提供时按 AND 组合；不存在的活动 ID 不单独返回 `404`，只是不产生匹配记录。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "activity": { "id": 1, "title": "寻找校园角落" },
        "author": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "title": "猜猜这是哪",
        "description": "校园神秘角落",
        "image_url": "https://media.example.com/photos/photo.jpg?signature=example",
        "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
        "longitude": 108.123456,
        "latitude": 34.123456,
        "coord_type": "gcj02",
        "solved": false,
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

#### 审核图片

```
PUT /api/admin/photos/{id}/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` / `reject` |
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
  "resp": { "id": 1, "status": "approved" },
  "message": "冲突错误: 该图片已审核过",
  "code": 8
}
```

---

### 审核 — 答题

#### 管理端答题列表

```
GET /api/admin/attempts
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `pending` / `solved` / `unsolved`；不传返回全部状态 |
| keyword | string | 否 | — | 按答题 ID、题目 ID 或题目标题搜索（最长 50） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "attempt_id": 1,
        "photo_id": 1,
        "photo_title": "原图标题",
        "guess_image_url": "https://media.example.com/attempts/attempt.jpg?signature=example",
        "guess_longitude": 108.5,
        "guess_latitude": 34.5,
        "guess_coord_type": "gcj02",
        "thumb_url": "https://media.example.com/photos/thumb.jpg?signature=example",
        "longitude": 108.123,
        "latitude": 34.456,
        "coord_type": "gcj02",
        "status": "pending",
        "submitted_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 审核答题

```
PUT /api/admin/attempts/{id}/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| solved | string | 是 | `solved` / `unsolved` |
| reject_reason | string | 否 | unsolved 的拒绝原因（最长 50，不填时有默认文案） |

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
  "resp": { "id": 1, "status": "solved" },
  "message": "冲突错误: 该答题记录已审核过",
  "code": 8
}
```

---

### 审核 — 评论

#### 管理端评论列表

```
GET /api/admin/comments
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `pending` / `approved` / `rejected`；不传返回全部状态 |
| keyword | string | 否 | — | 按评论 ID、题目 ID、题目标题、评论内容或用户昵称搜索（最长 50） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 8,
    "list": [
      {
        "comment_id": 1,
        "photo_id": 1,
        "photo_title": "图片标题",
        "user": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "comment": "评论内容",
        "status": "pending",
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

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` / `reject` |
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
  "resp": { "id": 1, "status": "approved" },
  "message": "冲突错误: 该评论已审核过",
  "code": 8
}
```

---

### 活动管理

#### 活动列表

```
GET /api/admin/activity/list
```

**权限**：管理员（Level ≥ 2）

**说明**：返回全部活动，包括尚未开始、进行中和已结束。接口不返回状态字段，管理端根据 `start_time` / `end_time` 与当前时间显示状态标签；`status` 筛选由后端使用服务器当前时间计算。不提供手动设置“当前活动”的字段。默认按 `start_time` 倒序、`id` 倒序返回。多个筛选条件按 AND 组合。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| keyword | string | 否 | — | 按活动 ID、标题或描述搜索（最长 50） |
| status | string | 否 | — | `not_started`（未开始）/ `active`（进行中）/ `ended`（已结束） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "title": "第一期活动",
        "cover_url": "https://media.example.com/photos/cover.jpg?signature=example",
        "description": "活动介绍",
        "start_time": "2026-05-01T00:00:00+08:00",
        "end_time": "2026-05-31T23:59:59+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 活动详情

```
GET /api/admin/activity/{id}
```

**权限**：管理员（Level ≥ 2）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "title": "寻找校园角落",
    "cover_url": "https://media.example.com/photos/cover.jpg?signature=example",
    "description": "活动描述",
    "start_time": "2026-07-01T00:00:00+08:00",
    "end_time": "2026-08-01T23:59:59+08:00",
    "photo_points": 50,
    "reward_tiers": [
      { "batch": 1, "rank_limit": 5, "attempt_points": 100 }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 创建活动

```
POST /api/admin/activity/create
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**时间规则**：接口不接收 `status` 或 `is_active` 字段。活动是否进行中完全由 `start_time <= now < end_time` 判定。允许多个活动的时间范围重叠并同时进行。

**奖励隔离规则**：`photo_points` 和 `reward_tiers` 只对所属 `activity_id` 对应的活动生效。投稿、答题排名、批次与积分发放必须按 `activity_id` 独立计算，不得在同时进行的多个活动之间合并统计。

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 活动标题（最长 20） |
| cover_file | file | 否 | 封面图（jpg/png，≤20MB） |
| description | string | 是 | 活动描述（最长 100） |
| start_time | string(date-time) | 是 | 带时区 ISO 8601 时间，例如 `2026-07-01T00:00:00+08:00` |
| end_time | string(date-time) | 是 | 带时区 ISO 8601 时间，且必须晚于开始时间 |
| photo_points | int | 是 | 图片奖励积分（min=0） |
| reward_tiers | string | 否 | 奖励阶梯 JSON：`[{"batch":1,"rank_limit":5,"attempt_points":100}]` |

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
POST /api/admin/activity/update
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**时间规则**：修改 `start_time` / `end_time` 后，后端必须立即使用新时间范围进行列表归类和写操作校验。活动一旦达到旧 `end_time`，题目答案坐标可能已经公开，因此不得再把 `end_time` 延长到服务器当前时间之后使活动重新变为未结束；违反时返回 `400`。活动结束后，关联题目、投稿和答题记录保留原 `activity_id`，不进行数据迁移。允许与其他活动时间重叠。

**奖励隔离规则**：奖励配置的更新只影响该 `activity_id`。已发放积分是否追溯调整必须由后端保持明确且一致的业务规则，不得因其他活动的配置变更而受影响。

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 活动 ID |
| title | string | 否 | 活动标题（最长 20） |
| cover_file | file | 否 | 封面图（jpg/png，≤20MB） |
| description | string | 否 | 活动描述（最长 100） |
| start_time | string(date-time) | 否 | 带时区 ISO 8601 开始时间 |
| end_time | string(date-time) | 否 | 带时区 ISO 8601 结束时间 |
| photo_points | int | 否 | 图片奖励积分（min=0） |
| reward_tiers | string | 否 | 奖励阶梯（传空数组 `[]` 清空，不传则不变） |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

活动管理不再提供按单个活动单独读取题目的接口。活动列表的“题目管理”入口应跳转到统一题目池，并把当前活动 ID 作为 `activity_ids` 初始筛选值；活动候选项可复用活动列表的 `status` 与 `keyword` 做服务端预筛和名称/ID 实时搜索。

#### 新增活动题目

```
POST /api/admin/activity/{activity_id}/photos
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：管理员可在任意活动阶段新增题目，不受活动时间限制。题目作者固定记录为官方账号——部署侧预置的一条真实用户记录（昵称如“图寻官方”），不随操作管理员变化；各接口返回的 `author` 即该官方账号的 `UserBrief`，客户端按普通作者统一展示，无需判定题目投稿渠道。管理员以个人身份投稿仍走客户端 `POST /photos` 普通流程并正常审核。新增后直接进入 `approved` 状态。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 题目标题（最长 20） |
| description | string | 否 | 题目描述（最长 50） |
| image_file | file | 是 | 题目图片（jpg/png，≤20MB） |
| longitude | float | 是 | 经度 |
| latitude | float | 是 | 纬度 |
| coord_type | string | 是 | `wgs84` / `gcj02` / `bd09` |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1001, "status": "approved" },
  "message": "",
  "code": 0
}
```

#### 更新活动题目

```
PUT /api/admin/activity/{activity_id}/photos/{photo_id}
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：可编辑任意活动阶段的题目内容，不改变题目当前审核状态；编辑 `rejected` 题目也不会重新进入审核或恢复公开，重新上架需另行设计。状态变更继续使用图片审核接口。请求至少提供一个字段；如修改坐标，`longitude`、`latitude`、`coord_type` 必须同时提供。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 否 | 题目标题（最长 20） |
| description | string | 否 | 题目描述（最长 50） |
| image_file | file | 否 | 新题目图片（jpg/png，≤20MB） |
| longitude | float | 条件必填 | 与纬度、坐标系同时提供 |
| latitude | float | 条件必填 | 与经度、坐标系同时提供 |
| coord_type | string | 条件必填 | `wgs84` / `gcj02` / `bd09` |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1001, "status": "approved" },
  "message": "",
  "code": 0
}
```

题目不存在或不属于路径中的活动时返回 `404`。

---

### 通知管理

管理端只写入管理员发布的 `category=normal` 通知；按用户业务事件生成的 `interaction` 消息不能在这里编辑或删除。读取不再重复建设 `/admin` 接口：

- 列表复用 `GET /api/notifications`，固定传 `category=normal`，并按需传 `type`、`keyword`、`page`、`page_size`；列表只使用 `content_preview`，不得下载完整正文或逐条补详情；
- 查看或编辑前复用 `GET /api/notifications/{id}` 获取完整 `content`、可选 `image_url`、关联对象和失效时间；
- `is_read` 只表示当前管理员账号自己的已读状态，不是通知阅读率，管理后台可忽略该字段。

#### 发布通知

```
POST /api/admin/notifications
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：管理员发布一般通知或全局公告，可附带一张图片。互动消息由业务事件自动生成，不通过本接口发布。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 是 | `general` / `global_announcement` |
| title | string | 是 | 通知标题（最长 20） |
| content | string | 是 | 通知内容（最长 2000） |
| image_file | file | 否 | 通知图片（jpg/png，≤20MB） |
| related_type | string | 条件可选 | 一般通知的关联类型，当前仅支持 `activity`；与 `related_id` 同时提供或同时省略 |
| related_id | int | 条件可选 | 一般通知的关联对象 ID |
| expires_at | string(date-time) | 条件必填 | `global_announcement` 必填且必须晚于服务器当前时间；一般通知不得提供 |

`type=global_announcement` 时不得提供 `related_type` 或 `related_id`。新发布的未过期全局公告会覆盖旧公告的弹窗优先级，但旧公告仍保留在通知列表中。

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "published" },
  "message": "",
  "code": 0
}
```

#### 更新通知

```
PUT /api/admin/notifications/{id}
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

**说明**：未提供字段保持不变，最终字段组合仍须满足通知类型约束。切换为 `global_announcement` 时后端自动清除活动关联且本次必须提供 `expires_at`；切换为 `general` 时后端自动清除 `expires_at`。更新不创建新通知，也不重置用户已读状态；如需让全局公告再次弹出，应新建通知。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| type | string | 否 | `general` / `global_announcement` |
| title | string | 否 | 通知标题（最长 20） |
| content | string | 否 | 通知内容（最长 2000） |
| image_file | file | 否 | 新通知图片（jpg/png，≤20MB） |
| remove_image | bool | 否 | `true` 时移除已有图片；不得与 `image_file` 同时提供 |
| remove_relation | bool | 否 | `true` 时移除已有活动关联；不得与 `related_type` / `related_id` 同时提供 |
| related_type | string | 条件可选 | 当前仅支持 `activity` |
| related_id | int | 条件可选 | 关联对象 ID |
| expires_at | string(date-time) | 条件可选 | 全局公告失效时间 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "published" },
  "message": "",
  "code": 0
}
```

#### 删除通知

```
DELETE /api/admin/notifications/{id}
```

**权限**：管理员（Level ≥ 2）

**说明**：删除后不再出现在客户端通知列表，也不再参与全局公告弹窗选择。互动消息不能通过本接口删除。

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "deleted" },
  "message": "",
  "code": 0
}
```

通知类型、条件字段或图片参数组合无效时返回 `400`。

---

### 商品管理

#### 商品列表

```
GET /api/admin/goods/list
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| available | bool | 否 | — | 仅看可兑换 |
| status | string | 否 | — | `inStore` / `outStore` |
| keyword | string | 否 | — | 按商品 ID、名称或描述搜索（最长 50） |

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
        "thumb_url": "https://media.example.com/goods/thumb.jpg?signature=example",
        "need_score": 500,
        "stock": 20,
        "status": "inStore",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 商品详情

```
GET /api/admin/goods/{id}
```

**权限**：管理员（Level ≥ 2）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "name": "明信片套装",
    "description": "精美校园风景明信片",
    "image_url": "https://media.example.com/goods/good.jpg?signature=example",
    "need_score": 500,
    "stock": 20,
    "status": "inStore",
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

#### 新增商品

```
POST /api/admin/goods/new
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 奖品名称（最长 20） |
| description | string | 否 | 描述（最长 50） |
| need_score | int | 是 | 所需积分（min=0） |
| stock | int | 是 | 库存（min=0） |
| image | file | 是 | 商品图片（jpg/png，≤20MB） |
| status | string | 否 | `inStore`（上架，默认）/ `outStore`（下架） |

**返回** `201`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "inStore" },
  "message": "",
  "code": 0
}
```

#### 更新商品

```
PUT /api/admin/goods/{id}
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 奖品名称（最长 20） |
| description | string | 否 | 描述（最长 50） |
| need_score | int | 否 | 所需积分 |
| stock | int | 否 | 库存 |
| image | file | 否 | 商品图片（jpg/png，≤20MB） |
| status | string | 否 | `inStore` / `outStore` |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "inStore" },
  "message": "",
  "code": 0
}
```

#### 删除商品

```
DELETE /api/admin/goods/{id}
```

**权限**：管理员（Level ≥ 2）

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

#### 更新商品状态

```
PUT /api/admin/goods/{id}/status
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 是 | `inStore`（上架）/ `outStore`（下架） |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "inStore" },
  "message": "",
  "code": 0
}
```

#### 更新商品库存

```
PUT /api/admin/goods/{id}/stock
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stock | int | 是 | 新库存数量（min=0） |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "stock": 50 },
  "message": "",
  "code": 0
}
```

---

### 兑换管理

#### 兑换记录列表

```
GET /api/admin/exchange/list
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `pending` / `verified` / `cancelled` |
| exchange_id | int | 否 | — | 兑换记录 ID 精确匹配 |
| user_keyword | string | 否 | — | 按用户 ID、学号、姓名或昵称搜索（最长 50） |
| good_keyword | string | 否 | — | 按商品 ID 或商品名称搜索（最长 50） |

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
        "user": { "id": 1, "nickname": "张三", "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example" },
        "good": { "id": 1, "name": "明信片套装", "thumb_url": "https://media.example.com/goods/thumb.jpg?signature=example", "need_score": 500, "stock": 20 },
        "quantity": 1,
        "score_cost": 500,
        "status": "pending",
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
POST /api/admin/exchange/verify
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| exchange_id | int | 是 | 兑换记录 ID |
| action | string | 是 | `verify`（核销）/ `cancel`（取消，退回积分和库存） |

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
  "resp": { "id": 1, "status": "verified" },
  "message": "冲突错误: 该兑换记录已处理",
  "code": 8
}
```

---

### 反馈管理

#### 反馈列表

```
GET /api/admin/feedback/list
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| type | int | 否 | — | 1-内容 2-玩法 3-技术 4-其他 |
| status | string | 否 | — | `pending` / `resolved` |
| keyword | string | 否 | — | 按反馈 ID、标题、内容或用户 ID 搜索（最长 50） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
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
    "user_id": 1,
    "title": "建议增加功能",
    "content": "希望可以增加排行榜功能",
    "type": 2,
    "phone": "13800138000",
    "status": "pending",
    "medias": [
      { "id": 1, "url": "https://media.example.com/feedbacks/attachment.jpg?signature=example", "media_type": 1 }
    ],
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

#### 处理反馈

```
PUT /api/admin/feedback/{id}
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 是 | `pending` / `resolved` |

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

### 用户管理

#### 搜索用户

```
GET /api/admin/users
```

**权限**：管理员（Level ≥ 2）

**说明**：Level ≥ 2 管理员使用的模糊搜索接口。与仅限 Level 3 的 `/admin/user` 精确筛选接口权限和查询语义不同。`keyword` 省略或为空时不做关键词筛选；所有筛选均省略时返回全部用户，此时 `total` 即全站用户总数，供工作台等场景使用。账号状态与权限等级相互独立，多个条件按 AND 组合。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| keyword | string | 否 | — | 按学号/姓名/昵称模糊搜索（最长 50） |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `active` / `banned` |
| level | int | 否 | — | `1` 普通用户 / `2` 管理员 / `3` 创世管理员 |

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
        "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example",
        "level": 1,
        "status": "active"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 用户列表（创世管理员）

```
GET /api/admin/user
```

**权限**：创世管理员（Level 3）

**说明**：仅限 Level 3 的精确筛选接口，用于账号治理场景；公共用户字段与 `/admin/users` 一致。精确字段、账号状态和权限等级可组合并按 AND 匹配。Level 3 记录只读，不提供业务侧治理操作。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| netid | string | 否 | — | 学号精确匹配（最长 50） |
| name | string | 否 | — | 姓名精确匹配（最长 50） |
| nickname | string | 否 | — | 昵称精确匹配（最长 50） |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | — | `active` / `banned` |
| level | int | 否 | — | `1` 普通用户 / `2` 管理员 / `3` 创世管理员 |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 1,
    "list": [
      {
        "id": 1,
        "netid": "20230001",
        "username": "张三",
        "nickname": "张三",
        "avatar_url": "https://media.example.com/avatars/avatar.jpg?signature=example",
        "level": 2,
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

**权限**：管理员（Level ≥ 2）

**说明**：`{id}` 为目标用户 ID，接口按操作者与目标等级校验权限：Level 2 只能封禁或解封 Level 1；Level 3 可以操作 Level 1、2；任何人都不能操作 Level 3。状态设置为幂等操作，且不会改变账号 `level`。`banned` 生效时立即吊销目标用户全部 Session 并拒绝其重新登录（登录回调返回 `403`）；`active` 后可重新登录并恢复原等级权限。无权操作目标返回 `403`（`code=7`），目标不存在返回 `404`（`code=5`）。封禁不影响该用户既有内容，内容处理走既有审核接口。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 是 | `banned` 封禁 / `active` 解封 |

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

### 用户权限等级管理（创世管理员）

```
PUT /api/admin/level
```

**权限**：创世管理员（Level 3）

**说明**：调整 Level 1、2 用户的权限等级，且不改变账号 `status`。`target_level` 只允许 1 或 2；Level 0 仅表示未登录请求上下文，Level 3 只能由部署侧管理，两者都不能通过本接口写入。不能修改任何 Level 3 账号；无权操作目标返回 `403`（`code=7`）。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 目标用户 ID |
| target_level | int | 是 | `1` 普通用户 / `2` 管理员 |

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

| 实体 | 字段 | 可选值 | 默认值 |
|------|------|--------|--------|
| Photo | status | `pending` / `approved` / `rejected` | pending |
| Attempt | status | `pending` / `solved` / `unsolved` | pending |
| Comment | status | `pending` / `approved` / `rejected` | pending |
| Good | status | `inStore` / `outStore` | inStore |
| Exchange | status | `pending` / `verified` / `cancelled` | pending |
| Feedback | status | `pending` / `resolved` | pending |
| User | level | `1` / `2` / `3` | 1 |
| User | status | `active` / `banned` | active |
| Activity list filter | status | `not_started` / `active` / `ended` | — |
| Notification | category | `normal` / `interaction` | — |
| Notification | type | `general` / `global_announcement` / `like` / `comment` / `review` | — |
| Notification | is_read | `true` / `false` | false |

`User.level=3` 表示由部署侧管理的创世管理员，可存在多个。业务接口可以只读查询其摘要，但不能新增、修改、删除或改变其状态和等级。

写操作响应中的 `resp.status` 必须限制为对应业务状态：新建待处理记录为 `pending`，管理员新增活动题目为 `approved`，删除回执为 `deleted`，图片/评论审核为 `approved` / `rejected`，答题审核为 `solved` / `unsolved`，商品为 `inStore` / `outStore`，兑换处理为 `verified` / `cancelled`，反馈为 `pending` / `resolved`，通知发布或更新为 `published`，用户封禁/解封为 `active` / `banned`。仅表示通用成功的接口固定为 `success`。

**可空字段**（未设置时返回 `null`）：`reviewed_at`、`exchange_at`、`reject_reason`。
