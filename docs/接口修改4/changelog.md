# API 契约变更日志

> 本文档为简短说明，可能包含错误，主要以另外两个API文档为主。

---

## 一、 全局约定与传输层变更

1. **HTTP 方法重写**：`PUT /api/user/avatar` 接口增加对 `X-HTTP-Method-Override: PUT` 请求头的支持，兼容小程序/移动端 `uni.uploadFile` 传输限制。
2. **跨端鉴权头与 Session 规范**：
   - 新增请求头 `X-Session-Id: <session_id>` 鉴权凭据透传方式，适配小程序 `<web-view>` 等非 Cookie 宿主环境。`apifox-import.json` 中登记 `cookieAuth`（cookie `tz-sessions`）与 `sessionIdAuth`（header `X-Session-Id`）两个 securityScheme；54 个需鉴权 operation 同时声明两者（OR 关系，任一有效即可），10 个公开 operation 显式声明 `security: []`。
   - Session 生存期由 30 分钟调整为 14 天；采用按需滑动策略（剩余有效时间低于 50%/7 天时自动重置 14 天 TTL）。
   - 明确生产环境 `tz-sessions` 强制启用 `HttpOnly`、`Secure`、`SameSite=Lax`；基于 Cookie 的写操作仍需 CSRF 防护，经 `X-Session-Id` 透传的请求不依赖 Cookie，不受 CSRF 影响。
3. **服务器时间与跨域响应头**：[约定] 补充规定服务端必须配置 `Access-Control-Expose-Headers: Date` 响应头，保证 H5 跨域场景下客户端可读取服务器时间。
4. **Multipart 参数解析**：[约定] 补充规定服务端对 `multipart/form-data` 中的数值与布尔表单项统一进行宽松类型自动转换。
5. **分页 `total` 口径**：[约定] 与 `PageBase.total` 描述明确 `total` 为**经过权限、审核与筛选条件过滤后的实际匹配总条数**，不得返回过滤前的总数。
6. **可空字段与缺省规则**：[约定] 补充规定所有可空字段一律返回 `null`，不省略字段；仅 `Media` 的 `origin_url` / `thumb_url` 与 `FeedbackMedia` 的 `thumb_url` 是按接口口径/资源情况「缺省整个字段」而非置 `null` 的非 `nullable` 特例属性。

---

## 二、 公共数据结构 (Schemas) 变更

1. **媒体结构重构 (`Media`) 与字段命名**：
   - 弃用原 `MediaUrl` 结构；升级为包含 `origin_url`（高清原图完整地址）、`thumb_url`（缩略图完整地址）与 `width` / `height` 真实像素尺寸的标准图片对象 `Media`。根据接口场景下发 `origin_url`、`thumb_url` 或两者皆有。
   - 媒体字段摒弃 `_url` 冗余后缀与零散字段：用户头像由 `avatar_url` 重命名为 `avatar` (`string`)；活动封面由 `cover_url` 重命名为 `cover_image` (`Media`)；题目/奖品/作答/通知等图片的散乱字段（`thumb_url` / `image_url`）统一重构重命名为 **`image`** (`Media`)。
   - `width` / `height` 必返非空 `int` 且 `minimum: 1`，前端据此预留 CSS 占位框实现 0 布局抖动（CLS）。
2. **新增反馈媒体结构 (`FeedbackMedia`)**：
   - 新增 `FeedbackMedia` 对象（包含 `{ origin_url, thumb_url, width, height, media_type }`，`media_type`: `1` 图片、`2` 视频），专用于反馈详情中的附件展示；`width` / `height` 约束 `minimum: 1`，视频取不到尺寸时允许为 `null`（仅 `origin_url`、`media_type` 必返，`thumb_url` 无缩略图时省略）。
3. **新增登录结果结构 (`LoginResult`)**：
   - 新增 `LoginResult` 结构（包含 `UserSummary` + `session_id`），专供登录与回调接口 (`GET /user/logincallback`, `GET /test/login`) 返回；`UserSummary` 不包含 `session_id`。
4. **`PhotoCard` 结构**：
   - 补充 `solved` (`bool`) 属性，指示当前登录用户是否已破解该题（未登录时恒为 `false`）。
5. **`ActivityBrief` 与 `ActivityCard` 结构**：
   - `ActivityBrief` 补充 `start_time` 与 `end_time` (`ISO 8601`) 字段；`ActivityCard` 补充 `photo_count` (`int`) 题目数量，封面字段对应调整为 `cover_image` (`Media`)。
6. **`InteractionMessage` 结构**：
   - 补充 `photo_id` (`int`) 属性，支持消息直达关联题目。
7. **`AttemptRecord` / `GoodItem` / `AdminPhotoListItem` / `Announcement` 结构**：
   - 图片字段统一重命名为 `image` (`Media`)，对象内部使用 `origin_url` 与/或 `thumb_url`。

---

## 三、 接口契约变更

1. **`GET /api/score/logs`（积分流水）**：
   - 响应列表项新增 `related_title` (`string`，可为 `null`) 属性，与 `related_id` / `related_type` 同为必返可空字段；
   - `related_type` 补充 `enum: [photo, exchange]`（`admin_adjust` 时为 `null`）。
2. **`POST /api/feedback`（提交反馈）**：
   - 请求表单附件参数调整为单文件 `media_file`（图片 ≤ 20MB 或视频 ≤ 50MB）；`GET /api/admin/feedback/{id}` 响应对应返回单个 `media_file` (`FeedbackMedia`)，未上传时为 `null`。
3. **兑换记录新增防伪码 `verify_code` 与线下核销流程**：
   - `POST /api/exchange`、`GET /api/exchange`、`GET /api/admin/exchange` 响应新增 `verify_code` (`string`, `^[A-Z0-9]{8,16}$`)；
   - 核销流程：扫码 ➔ `GET /api/admin/exchange?verify_code=...` 拉取记录 ➔ 用记录 `id` 调 `PUT /api/admin/exchange/{id}/verify` 完成核销。
4. **`GET /api/admin/exchange`（管理端兑换列表）**：
   - 新增 Query 参数 `verify_code`，按防伪码**精确匹配**。
5. **`PUT /api/admin/exchange/{id}/verify`（核销兑换）**：
   - 路径参数 `{id}` 维持 `int`（兑换记录 ID）。
6. **`GET /api/user/logincallback`（登录回调）**：
   - 响应类型由 `UserSummary` 改为 `LoginResult`。
7. **`GET /api/photos`（题目列表）**：
   - Query 参数 `solved` 语义调整为**当前登录用户本人是否已破解**。
8. **全局角色称谓**：
   - 文档中原 `创世管理员` 统一更名为 `超级管理员`。
