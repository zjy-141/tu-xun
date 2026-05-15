

# 挑战西交图寻 API 文档 v1.0

## 1. 概述

本 API 服务于"挑战西交图寻"活动网站，支持用户发布校园隐藏机位的图片、猜图答题（需上传同一地点同一角度照片）、后台审核及奖品发放。系统强调图片的观赏性，即使不答题也能带来情感共鸣。

## 2. 基础信息

- **Base URL:** `https://api.xjtu-tuxun.com/api`
- **Content-Type:** `application/json`（文件上传使用 `multipart/form-data`）
- **认证方式:** Session / Cookie（登录后服务端维护会话，无需 JWT）

### 2.1 统一响应格式

所有接口返回统一的 JSON 响应体：

```json
{
  "success": true,
  "data": {},
  "message": "",
  "code": 0
}
```

| 字段    | 类型    | 说明                           |
|---------|---------|--------------------------------|
| success | bool    | 请求是否成功                   |
| data    | any     | 响应数据体（成功时返回）       |
| message | string  | 提示信息（失败时返回错误描述） |
| code    | uint64  | 错误码（成功时为 0）           |

### 2.2 错误码

错误码定义见 `common/error.go`：

| 常量      | 值 | 说明     |
|-----------|----|----------|
| ParamErr  | 3  | 参数错误 |
| SysErr    | 4  | 系统错误 |
| OpErr     | 5  | 操作错误 |
| AuthErr   | 6  | 鉴权错误 |
| LevelErr  | 7  | 权限错误 |

错误响应示例：

```json
{
  "success": false,
  "message": "学号或密码不能为空",
  "code": 3
}
```

### 2.3 分页参数

需要分页的接口统一使用以下 Query 参数（定义见 `common/form.go` 的 `PagerForm`）：

| 参数名 | 类型 | 必填 | 说明               |
|--------|------|------|--------------------|
| page   | int  | 否   | 页码，默认 1       |
| limit  | int  | 否   | 每页数量，最大 20  |

---

## 3. 用户认证

认证采用 Session 机制（`controller/session.go`），登录后服务端保存 `UserSession`（含 `ID`、`Username`、`Level`），`Level >= 1` 表示管理员。

### 3.1 用户注册

`POST /api/auth/register`

**请求体 (application/json):**

| 字段名     | 类型   | 必填 | 说明           |
|------------|--------|------|----------------|
| student_id | string | 是   | 学号，唯一     |
| name       | string | 是   | 昵称           |
| password   | string | 是   | 密码，6-20 位  |
| email      | string | 是   | 校园邮箱       |

**响应 (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "student_id": "2023123456",
    "name": "张三",
    "level": 0
  }
}
```

### 3.2 用户登录

`POST /api/auth/login`

**请求体:**

```json
{
  "student_id": "2023123456",
  "password": "******"
}
```

**响应 (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "name": "张三",
    "level": 0
  }
}
```

> 登录成功后，服务端通过 `SessionSet` 写入 `user-session`，后续请求自动携带 Cookie 维持会话。

### 3.3 用户登出

`POST /api/auth/logout`

需要登录。

**响应:**

```json
{
  "success": true,
  "data": null
}
```

> 调用 `SessionClear` 清除当前会话。

### 3.4 获取当前用户信息

`GET /api/auth/me`

需要登录。

**响应:**

```json
{
  "success": true,
  "data": {
    "id": 123,
    "student_id": "2023123456",
    "name": "张三",
    "level": 0,
    "prize_count": 1
  }
}
```

---

## 4. 图片（图寻题目）

### 4.1 上传图片（投稿）

`POST /api/photos`

需要登录。

**请求体 (multipart/form-data):**

| 字段名          | 类型   | 必填 | 说明                                            |
|-----------------|--------|------|-------------------------------------------------|
| image           | file   | 是   | 图片文件（jpg/png，≤20MB，建议长边≥1920px）     |
| title           | string | 是   | 图片标题，≤50 字符                               |
| description     | string | 否   | 图片描述 / 背后的故事                            |
| location_secret | string | 是   | 具体地点（仅管理员可见，如"主楼A座5楼东侧窗台"） |

**响应 (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": 1001,
    "title": "晨光中的钱学森图书馆",
    "description": "某个清晨的光影",
    "image_url": "https://cdn.xjtu-tuxun.com/photos/1001.jpg",
    "status": "pending",
    "solved": false,
    "created_at": "2025-03-15T12:00:00Z"
  }
}
```

### 4.2 获取图片列表（公共浏览）

`GET /api/photos`

不需要登录。返回已审核通过的所有图片，**不包含具体地点**。

**Query 参数:**

| 参数名 | 类型 | 必填 | 说明                                  |
|--------|------|------|---------------------------------------|
| page   | int  | 否   | 页码，默认 1                          |
| limit  | int  | 否   | 每页数量，默认 10，最大 20            |
| solved | bool | 否   | 筛选已/未破解的图片，不传则返回所有   |

**响应 (200 OK):**

```json
{
  "success": true,
  "data": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "items": [
      {
        "id": 1001,
        "title": "晨光中的钱学森图书馆",
        "description": "某个清晨的光影",
        "image_url": "https://cdn.../1001_thumb.jpg",
        "author": { "id": 123, "name": "张三" },
        "solved": false,
        "attempts_count": 3,
        "created_at": "2025-03-15T12:00:00Z"
      }
    ]
  }
}
```

### 4.3 获取图片详情

`GET /api/photos/:id`

使用 `IDUriForm`（`uri:"id" binding:"min=1"`）。

**响应:**

```json
{
  "success": true,
  "data": {
    "id": 1001,
    "title": "晨光中的钱学森图书馆",
    "description": "某个清晨的光影",
    "image_url": "https://cdn.../1001.jpg",
    "author": { "id": 123, "name": "张三" },
    "solved": true,
    "winner": {
      "user_id": 456,
      "name": "李四",
      "created_at": "2025-03-20T10:00:00Z"
    },
    "attempts_count": 5,
    "current_user_attempt": {
      "id": 202,
      "status": "approved",
      "is_winner": false
    },
    "created_at": "2025-03-15T12:00:00Z"
  }
}
```

> `current_user_attempt` 仅在已登录且已提交过答案时返回。当 `solved` 为 `true` 时，该图片不再发放奖品，但用户仍可继续提交答案（仅作交流）。

---

## 5. 答题（提交匹配照片）

### 5.1 提交答案

`POST /api/photos/:id/attempts`

需要登录。用户需上传在**同一地点同一角度**拍摄的照片，并描述猜测的地点（管理员将比对与 `location_secret` 是否一致及角度相似度）。

**请求体 (multipart/form-data):**

| 字段名           | 类型   | 必填 | 说明                                         |
|------------------|--------|------|----------------------------------------------|
| image            | file   | 是   | 用户拍摄的匹配照片                            |
| guessed_location | string | 是   | 用户认为的地点描述（如"主楼A座5楼东侧窗台"） |

**响应 (201 Created):**

```json
{
  "success": true,
  "data": {
    "attempt_id": 5001,
    "photo_id": 1001,
    "status": "pending",
    "message": "已提交，等待管理员审核。若审核通过且本题尚未被破解，您将获得奖品。"
  }
}
```

> **答题唯一性**：同一用户对同一张图片已有"待审核"状态的答题记录时，重复提交返回错误（code: 5，操作错误）。

### 5.2 获取某图片下我的所有答题记录

`GET /api/photos/:id/my-attempts`

需要登录。

**响应:**

```json
{
  "success": true,
  "data": {
    "photo_id": 1001,
    "solved": true,
    "my_attempts": [
      {
        "id": 5001,
        "image_url": "https://cdn.../attempt_5001.jpg",
        "guessed_location": "主楼A座5楼东侧窗台",
        "status": "approved",
        "is_winner": false,
        "reviewed_at": "2025-03-22T14:00:00Z"
      }
    ]
  }
}
```

---

## 6. 管理员审核接口

以下接口需要管理员权限（`Level >= 1`），通过中间件 `middleware.CheckRole(1)` 校验（见 `middleware/role.go`）。

### 6.1 获取待审核图片列表

`GET /api/admin/photos/pending`

**Query 参数:** `page`, `limit`

**响应:**

```json
{
  "success": true,
  "data": {
    "total": 5,
    "items": [
      {
        "id": 1002,
        "title": "隐蔽的小径",
        "location_secret": "东花园西南角灌木丛后",
        "author": { "id": 789, "name": "王五" },
        "created_at": "2025-03-21T08:00:00Z"
      }
    ]
  }
}
```

### 6.2 审核图片（通过 / 拒绝）

`PUT /api/admin/photos/:id/review`

**请求体:**

```json
{
  "action": "approve",
  "reject_reason": "图片模糊，无法辨认"
}
```

| 字段名        | 类型   | 必填                    | 说明                    |
|---------------|--------|-------------------------|-------------------------|
| action        | string | 是                      | `"approve"` 或 `"reject"` |
| reject_reason | string | action=reject 时必填    | 拒绝原因                |

**响应 (200):**

```json
{
  "success": true,
  "data": {
    "id": 1002,
    "status": "approved",
    "message": "图片已通过审核，现已公开"
  }
}
```

### 6.3 获取待审核答题记录

`GET /api/admin/attempts/pending`

**响应:**

```json
{
  "success": true,
  "data": {
    "total": 3,
    "items": [
      {
        "attempt_id": 5001,
        "photo_id": 1001,
        "photo_title": "晨光中的钱学森图书馆",
        "user": { "id": 456, "name": "李四" },
        "image_url": "https://cdn.../attempt_5001.jpg",
        "guessed_location": "主楼A座5楼东侧窗台",
        "submitted_at": "2025-03-20T12:00:00Z"
      }
    ]
  }
}
```

### 6.4 审核答题记录

`PUT /api/admin/attempts/:id/review`

**请求体:**

```json
{
  "action": "approve",
  "reject_reason": "照片拍摄角度偏差较大"
}
```

管理员审核通过时，系统会**自动判断**该题目是否已被其他人答对：

- 若 `photo.solved == false` → 标记该 attempt 为 `is_winner=true`，同时将 `photo.solved` 设为 `true`，并生成奖品领取记录。
- 若 `photo.solved == true` → 标记 attempt 为 `is_winner=false`，回复用户"正确答案，但奖品已被领走"。

**响应:**

```json
{
  "success": true,
  "data": {
    "attempt_id": 5001,
    "status": "approved",
    "is_winner": true,
    "photo_solved": true,
    "message": "审核通过，恭喜答对！将为您发放纪念奖品。"
  }
}
```

---

## 7. 奖品管理

### 7.1 获取我的奖品

`GET /api/users/me/prizes`

需要登录。列出已获得但未领取 / 已领取的奖品。

**响应:**

```json
{
  "success": true,
  "data": {
    "prizes": [
      {
        "id": 3001,
        "photo_id": 1001,
        "photo_title": "晨光中的钱学森图书馆",
        "status": "unclaimed",
        "prize_type": "明信片套装",
        "awarded_at": "2025-03-22T14:00:00Z"
      }
    ]
  }
}
```

### 7.2 （管理员）标记奖品已发放

`PUT /api/admin/prizes/:id/claim`

管理员在线下发放奖品后调用此接口。

**响应:**

```json
{
  "success": true,
  "data": {
    "prize_id": 3001,
    "status": "claimed"
  }
}
```

---

## 8. 故事分享（扩展）

答对者可（或任何用户）分享发现该角落的经历，增强社区情感体验。

### 8.1 发布故事

`POST /api/photos/:id/stories`

需要登录。

**请求体 (application/json):**

```json
{
  "content": "那天黄昏偶然走过那条小路，回头一看竟与这张图片一模一样，感觉时间都静止了……",
  "media_url": "https://...optional-video-or-image"
}
```

| 字段名    | 类型   | 必填 | 说明                         |
|-----------|--------|------|------------------------------|
| content   | string | 是   | 故事内容                     |
| media_url | string | 否   | 可选，支持短视频或照片的 URL |

**响应 (201):**

```json
{
  "success": true,
  "data": {
    "story_id": 9001,
    "user": { "id": 456, "name": "李四" },
    "content": "那天黄昏偶然走过...",
    "likes": 0,
    "created_at": "2025-03-23T09:00:00Z"
  }
}
```

### 8.2 获取图片下的故事列表

`GET /api/photos/:id/stories`

**响应:**

```json
{
  "success": true,
  "data": {
    "stories": [
      {
        "id": 9001,
        "user_name": "李四",
        "content": "那天黄昏偶然走过...",
        "media_url": null,
        "likes": 3,
        "created_at": "2025-03-23T09:00:00Z"
      }
    ]
  }
}
```

---

## 9. 接口汇总

| 方法   | 路径                              | 认证     | 说明                 |
|--------|-----------------------------------|----------|----------------------|
| POST   | `/api/auth/register`              | 否       | 用户注册             |
| POST   | `/api/auth/login`                 | 否       | 用户登录（Session）  |
| POST   | `/api/auth/logout`                | 是       | 用户登出             |
| GET    | `/api/auth/me`                    | 是       | 获取当前用户信息     |
| POST   | `/api/photos`                     | 是       | 上传图片（投稿）     |
| GET    | `/api/photos`                     | 否       | 图片列表（公共浏览） |
| GET    | `/api/photos/:id`                 | 否       | 图片详情             |
| POST   | `/api/photos/:id/attempts`        | 是       | 提交答案             |
| GET    | `/api/photos/:id/my-attempts`     | 是       | 我的答题记录         |
| GET    | `/api/admin/photos/pending`       | 管理员   | 待审核图片列表       |
| PUT    | `/api/admin/photos/:id/review`    | 管理员   | 审核图片             |
| GET    | `/api/admin/attempts/pending`     | 管理员   | 待审核答题记录       |
| PUT    | `/api/admin/attempts/:id/review`  | 管理员   | 审核答题记录         |
| GET    | `/api/users/me/prizes`            | 是       | 我的奖品             |
| PUT    | `/api/admin/prizes/:id/claim`     | 管理员   | 标记奖品已发放       |
| POST   | `/api/photos/:id/stories`         | 是       | 发布故事             |
| GET    | `/api/photos/:id/stories`         | 否       | 故事列表             |

---

## 10. 补充说明

- **图片处理**：上传原图后自动生成缩略图用于列表展示，原图用于详情页。
- **管理员权限**：通过 `middleware.CheckRole(1)` 中间件控制，`UserSession.Level >= 1` 即为管理员。
- **答题唯一性**：同一用户对同一张图片只能有一笔"待审核"状态的答题记录，重复提交返回操作错误（code: 5）。
- **奖品发放**：通过审核并获得 `is_winner=true` 的用户，需在规定时间内线下凭校园卡领取，管理员手动标记 `claimed`。
- **合规性**：所有图片须为原创或已获得授权，不得含有人像敏感信息或违规内容。
- **Session 密钥**：生产环境中通过环境变量 `APP_SECRET` 设置，见 `config/config.go`。

---

> 本 API 文档对应于"挑战西交图寻"网站的后端设计，遵循项目 tz-gin 开发规范（Session 认证、Controller/Service 分层、统一错误码等）。如有扩展需求（如点赞、评论、排行榜）可在此基础上增加对应端点。
