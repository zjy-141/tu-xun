# 图寻 API 文档 v2.0

> 版本：v2.0.29 | 更新：2026-05-15

---

## 1. 基础信息

| 项目 | 值 |
|------|-----|
| **Base URL** | `http://0.0.0.0:8088/api` |
| **Content-Type** | `application/json`；文件上传使用 `multipart/form-data` |
| **认证方式** | Session / Cookie（登录后服务端维护 `user-session`） |
| **静态资源** | `/uploads/` 目录直接暴露，如图片 `/uploads/photos/xxx.jpg` |

### 1.1 统一响应格式

```json
{
  "success": true,
  "data": {},
  "message": "",
  "code": 0
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | bool | 是否成功 |
| `data` | any | 业务数据（成功时返回） |
| `message` | string | 提示信息（失败时为错误描述） |
| `code` | uint64 | 错误码（成功时 `0`） |

### 1.2 错误码

| 常量 | 值 | HTTP 状态码 | 说明 |
|------|----|-----------|------|
| `ParamErr` | 3 | 400 | 参数错误 |
| `SysErr` | 4 | 500 | 系统错误 |
| `OpErr` | 5 | 400 | 操作错误（业务逻辑冲突） |
| `AuthErr` | 6 | 401 | 鉴权错误（未登录或凭据无效） |
| `LevelErr` | 7 | 403 | 权限不足 |

### 1.3 权限等级

| Level | 角色 |
|-------|------|
| `0` | 普通用户 |
| `>= 1` | 管理员 |

### 1.4 分页参数

需分页的接口统一使用以下 Query 参数：

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `page` | int | 否 | 1 | 页码 |
| `limit` | int | 否 | 10 | 每页数量，最大 20 |

分页响应均包含 `total` 字段（总数）。

---

## 2. 接口列表

### 2.1 用户认证

#### POST /auth/register — 注册

- **认证**：否
- **Content-Type**：`application/json`

**请求体：**

| 字段 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `student_id` | string | 是 | — | 学号，唯一 |
| `name` | string | 是 | — | 昵称 |
| `password` | string | 是 | 6-20 位 | 密码（bcrypt 加密存储） |
| `email` | string | 是 | 邮箱格式 | 校园邮箱 |

```json
{
  "student_id": "2023123456",
  "name": "张三",
  "password": "123456",
  "email": "zhangsan@stu.xjtu.edu.cn"
}
```

**响应 201：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "student_id": "2023123456",
    "name": "张三",
    "email": "zhangsan@stu.xjtu.edu.cn",
    "level": 0,
    "prize_count": 0
  }
}
```

> 注册成功自动登录，Cookie 中写入 `user-session`。

---

#### POST /auth/login — 登录

- **认证**：否
- **Content-Type**：`application/json`

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `student_id` | string | 是 | 学号 |
| `password` | string | 是 | 密码 |

```json
{
  "student_id": "2023123456",
  "password": "123456"
}
```

**响应 200：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "student_id": "2023123456",
    "name": "张三",
    "email": "zhangsan@stu.xjtu.edu.cn",
    "level": 0,
    "prize_count": 0
  }
}
```

**错误：** `AuthErr(6)` — 学号或密码错误

---

#### DELETE /auth/logout — 登出

- **认证**：是

**响应 200：**

```json
{
  "success": true,
  "data": null
}
```

---

#### GET /auth/me — 当前用户信息

- **认证**：是

**响应 200：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "student_id": "2023123456",
    "name": "张三",
    "email": "zhangsan@stu.xjtu.edu.cn",
    "level": 0,
    "prize_count": 1
  }
}
```

---

### 2.2 图片投稿

#### POST /photos — 上传投稿

- **认证**：是（`Level >= 0`）
- **Content-Type**：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `image` | file | 是 | jpg/png，≤20MB |
| `title` | string | 是 | 图片标题 |
| `description` | string | 否 | 图片描述/故事 |
| `location_secret` | string | 是 | 拍摄的具体地点（仅管理员可见） |

**响应 201：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "title": "晨光中的图书馆",
    "description": "某个清晨的光影",
    "image_url": "/uploads/photos/1712345678901234567.jpg",
    "thumb_url": "/uploads/photos/1712345678901234567.jpg",
    "status": "pending",
    "solved": false,
    "attempts_count": 0,
    "author": {
      "id": 1,
      "student_id": "2023123456",
      "name": "张三",
      "email": "zhangsan@stu.xjtu.edu.cn",
      "level": 0,
      "prize_count": 0
    }
  }
}
```

> 新上传的图片状态为 `pending`，需管理员审核通过（`approved`）后方公开可见。

---

#### GET /photos — 图片列表

- **认证**：否
- **Query 参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `page` | int | 否 | 默认 1 |
| `limit` | int | 否 | 默认 10，最大 20 |
| `solved` | bool | 否 | 筛选已/未破解（不传则全部） |

**响应 200：**

```json
{
  "success": true,
  "data": {
    "total": 100,
    "page": 1,
    "limit": 10,
    "items": [
      {
        "id": 1,
        "title": "晨光中的图书馆",
        "description": "某个清晨的光影",
        "image_url": "/uploads/photos/1712345678901234567.jpg",
        "author": { "id": 1, "name": "张三" },
        "solved": false,
        "attempts_count": 3,
        "created_at": "2026-05-15T12:00:00+08:00"
      }
    ]
  }
}
```

> 列表仅返回已审核通过（`status = "approved"`）的图片，`image_url` 为缩略图。

---

#### GET /photos/:id — 图片详情

- **认证**：否（登录后额外返回 `current_user_attempt`）

**路径参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int64 | 是 | 图片 ID（≥1） |

**响应 200：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "晨光中的图书馆",
    "description": "某个清晨的光影",
    "image_url": "/uploads/photos/1712345678901234567.jpg",
    "author": { "id": 1, "name": "张三" },
    "solved": true,
    "attempts_count": 5,
    "created_at": "2026-05-15T12:00:00+08:00",
    "winner": {
      "user_id": 2,
      "name": "李四",
      "created_at": "2026-05-16T10:00:00+08:00"
    },
    "current_user_attempt": {
      "id": 1,
      "status": "approved",
      "is_winner": false
    }
  }
}
```

| 字段 | 说明 |
|------|------|
| `winner` | 仅 `solved == true` 时返回，包含获奖者信息 |
| `current_user_attempt` | 仅登录用户已提交过答案时返回 |

---

### 2.3 答题

#### POST /photos/:id/attempts — 提交答案

- **认证**：是
- **Content-Type**：`multipart/form-data`

**路径参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int64 | 是 | 图片 ID |

**表单字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `image` | file | 是 | 同地点同角度拍摄的匹配照片 |
| `guessed_location` | string | 是 | 猜测的地点描述 |

**响应 201：**

```json
{
  "success": true,
  "data": {
    "attempt_id": 1,
    "photo_id": 1,
    "status": "pending",
    "message": "已提交，等待管理员审核。若审核通过且本题尚未被破解，您将获得奖品。"
  }
}
```

**业务规则：**
- 同一用户对同一图片已有 `pending` 状态的记录时，不可重复提交（返回 `OpErr`）
- 图片状态必须为 `approved`，否则返回 `OpErr`

---

#### GET /photos/:id/my-attempts — 我的答题记录

- **认证**：是

**路径参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int64 | 是 | 图片 ID |

**响应 200：**

```json
{
  "success": true,
  "data": {
    "photo_id": 1,
    "solved": true,
    "my_attempts": [
      {
        "id": 1,
        "image_url": "/uploads/attempts/1712345678901234568.jpg",
        "guessed_location": "主楼A座5楼东侧窗台",
        "status": "approved",
        "is_winner": false,
        "reviewed_at": "2026-05-17T14:00:00+08:00"
      }
    ]
  }
}
```

---

### 2.4 故事分享

#### POST /photos/:id/stories — 发布故事

- **认证**：是
- **Content-Type**：`application/json`

**路径参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int64 | 是 | 图片 ID |

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `content` | string | 是 | 故事内容 |
| `media_url` | string | 否 | 可选的媒体 URL（视频/图片） |

```json
{
  "content": "那天黄昏偶然走过那条小路……",
  "media_url": "/uploads/stories/1712345678901234569.mp4"
}
```

**响应 201：**

```json
{
  "success": true,
  "data": {
    "id": 1,
    "photo_id": 1,
    "user_id": 2,
    "content": "那天黄昏偶然走过那条小路……",
    "media_url": "/uploads/stories/1712345678901234569.mp4",
    "likes": 0,
    "user": {
      "id": 2,
      "student_id": "2023123457",
      "name": "李四",
      "email": "lisi@stu.xjtu.edu.cn",
      "level": 0,
      "prize_count": 1
    }
  }
}
```

---

#### POST /stories/media — 上传故事媒体

- **认证**：是
- **Content-Type**：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file` | file | 是 | jpg/png，≤20MB |

**响应 201：**

```json
{
  "success": true,
  "data": {
    "media_url": "/uploads/stories/1712345678901234569.mp4"
  }
}
```

> 先调此接口上传媒体拿到 URL，再调 `POST /photos/:id/stories` 传入 `media_url`。

---

#### GET /photos/:id/stories — 故事列表

- **认证**：否

**路径参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | int64 | 是 | 图片 ID |

**响应 200：**

```json
{
  "success": true,
  "data": {
    "stories": [
      {
        "id": 1,
        "user_name": "李四",
        "content": "那天黄昏偶然走过那条小路……",
        "media_url": "/uploads/stories/1712345678901234569.mp4",
        "likes": 3,
        "created_at": "2026-05-16T09:00:00+08:00"
      }
    ]
  }
}
```

> 按创建时间倒序排列。

---

### 2.5 奖品

#### GET /users/me/prizes — 我的奖品

- **认证**：是

**响应 200：**

```json
{
  "success": true,
  "data": {
    "prizes": [
      {
        "id": 1,
        "photo_id": 1,
        "photo_title": "晨光中的图书馆",
        "status": "unclaimed",
        "prize_type": "明信片套装",
        "awarded_at": "2026-05-17T14:00:00+08:00"
      }
    ]
  }
}
```

| `status` 值 | 说明 |
|------------|------|
| `unclaimed` | 待领取 |
| `claimed` | 已发放 |

---

### 2.6 管理员接口

> 以下接口需要管理员权限（`Level >= 1`），通过 `CheckRole(1)` 中间件校验。

#### GET /admin/photos/pending — 待审核图片

- **认证**：管理员
- **Query：** `page`、`limit`

**响应 200：**

```json
{
  "success": true,
  "data": {
    "total": 5,
    "items": [
      {
        "id": 2,
        "title": "隐蔽的小径",
        "location_secret": "东花园西南角灌木丛后",
        "author": { "id": 3, "name": "王五" },
        "created_at": "2026-05-16T08:00:00+08:00"
      }
    ]
  }
}
```

> `location_secret` 仅管理员可见。

---

#### PUT /admin/photos/:id/review — 审核图片

- **认证**：管理员
- **Content-Type**：`application/json`

**路径参数：** `id` — 图片 ID

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `action` | string | 是 | `"approve"` 或 `"reject"` |
| `reject_reason` | string | `action=reject` 时必填 | 拒绝原因 |

```json
{
  "action": "approve"
}
```

**响应 200：**

```json
{
  "success": true,
  "data": {
    "id": 2,
    "status": "approved",
    "message": "图片已通过审核，现已公开"
  }
}
```

拒绝时：
```json
{
  "success": true,
  "data": {
    "id": 2,
    "status": "rejected",
    "message": "图片已拒绝: 图片模糊，无法辨认"
  }
}
```

---

#### GET /admin/attempts/pending — 待审核答题

- **认证**：管理员
- **Query：** `page`、`limit`

**响应 200：**

```json
{
  "success": true,
  "data": {
    "total": 3,
    "items": [
      {
        "attempt_id": 1,
        "photo_id": 1,
        "photo_title": "晨光中的图书馆",
        "user": { "id": 2, "name": "李四" },
        "image_url": "/uploads/attempts/1712345678901234568.jpg",
        "guessed_location": "主楼A座5楼东侧窗台",
        "submitted_at": "2026-05-17T12:00:00+08:00"
      }
    ]
  }
}
```

---

#### PUT /admin/attempts/:id/review — 审核答题

- **认证**：管理员
- **Content-Type**：`application/json`

**路径参数：** `id` — 答题记录 ID

**请求体：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `action` | string | 是 | `"approve"` 或 `"reject"` |
| `reject_reason` | string | `action=reject` 时必填 | 拒绝原因 |

```json
{
  "action": "approve"
}
```

**响应 200（通过且为首位答对者）：**

```json
{
  "success": true,
  "data": {
    "attempt_id": 1,
    "status": "approved",
    "is_winner": true,
    "photo_solved": true,
    "message": "审核通过，恭喜答对！将为您发放纪念奖品。"
  }
}
```

**响应 200（通过但非首位）：**

```json
{
  "success": true,
  "data": {
    "attempt_id": 1,
    "status": "approved",
    "is_winner": false,
    "photo_solved": true,
    "message": "正确答案，但奖品已被领走。感谢您的参与！"
  }
}
```

**审核自动逻辑：**
- `approve` 时：若该图 `solved == false` → 标记 `is_winner = true`、`photo.solved = true`，**自动生成奖品记录**；否则仅标记通过
- `reject` 时：需填写 `reject_reason`

---

#### PUT /admin/prizes/:id/claim — 标记奖品已发放

- **认证**：管理员

**路径参数：** `id` — 奖品 ID

**响应 200：**

```json
{
  "success": true,
  "data": {
    "prize_id": 1,
    "status": "claimed"
  }
}
```

---

## 3. 接口速查表

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/auth/register` | — | 注册 |
| POST | `/auth/login` | — | 登录 |
| DELETE | `/auth/logout` | 用户 | 登出 |
| GET | `/auth/me` | 用户 | 当前用户信息 |
| POST | `/photos` | 用户 | 上传图片投稿 |
| GET | `/photos` | — | 图片列表（公共浏览） |
| GET | `/photos/:id` | — | 图片详情 |
| POST | `/photos/:id/attempts` | 用户 | 提交答题 |
| GET | `/photos/:id/my-attempts` | 用户 | 我的答题记录 |
| POST | `/photos/:id/stories` | 用户 | 发布故事 |
| GET | `/photos/:id/stories` | — | 故事列表 |
| POST | `/stories/media` | 用户 | 上传故事媒体 |
| GET | `/users/me/prizes` | 用户 | 我的奖品 |
| GET | `/admin/photos/pending` | 管理员 | 待审核图片 |
| PUT | `/admin/photos/:id/review` | 管理员 | 审核图片 |
| GET | `/admin/attempts/pending` | 管理员 | 待审核答题 |
| PUT | `/admin/attempts/:id/review` | 管理员 | 审核答题 |
| PUT | `/admin/prizes/:id/claim` | 管理员 | 标记奖品已发放 |

---

## 4. 业务规则摘要

| 规则 | 说明 |
|------|------|
| 图片审核 | 新投稿 `status = "pending"`，管理员审核后变为 `approved`/`rejected`，仅 `approved` 公开可见 |
| 答题唯一性 | 同一用户对同一图片只能有一个 `pending` 答题，不可重复提交 |
| 首位获奖 | 第一个被审核通过的答案自动获奖，生成奖品记录；后续通过的仅标记通过不获奖 |
| 奖品类型 | 固定为"明信片套装" |
| 故事媒体 | 先调 `/stories/media` 上传拿到 URL，再发故事 |
| 图片限制 | jpg/png，≤20MB |
| 密码安全 | bcrypt 加密，响应中永不返回 `password` 字段 |
| 管理员 | `Level >= 1`，需在数据库中手动设置 |
