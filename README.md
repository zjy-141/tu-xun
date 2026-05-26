# 图寻 API 文档 v3.0

> 更新：2026-05-26

---

## 1. 基础信息

| 项目 | 值 |
|------|-----|
| **Base URL** | `http://0.0.0.0:8088/api` |
| **Content-Type** | `application/json`；文件上传使用 `multipart/form-data` |
| **认证方式** | Session / Cookie（登录后服务端维护 `user-session`） |
| **静态资源** | `/uploads/` 目录直接暴露 |

### 1.1 统一响应格式

```json
{
  "success": true,
  "resp": {},
  "message": "",
  "code": 0
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `success` | bool | 是否成功 |
| `resp` | any | 业务数据 |
| `message` | string | 提示信息 |
| `code` | uint64 | 错误码（0 = 成功） |

### 1.2 错误码

| 常量 | 值 | HTTP | 说明 |
|------|----|------|------|
| `ParamErr` | 3 | 400 | 参数错误 |
| `SysErr` | 4 | 500 | 系统错误 |
| `OpErr` | 5 | 400 | 业务逻辑冲突 |
| `AuthErr` | 6 | 401 | 未登录或凭据无效 |
| `LevelErr` | 7 | 403 | 权限不足 |

### 1.3 权限等级

| Level | 角色 |
|-------|------|
| `0` | 普通用户 |
| `1` | 普通管理员 |
| `>= 2` | 高级管理员 |

### 1.4 分页参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `page` | int | 否 | 1 | 页码 |
| `limit` | int | 否 | 10 | 每页数量，最大 20 |

分页响应均含 `total` 字段。

---

## 2. 用户认证

### POST /auth/register — 注册

- **认证**：否

| 字段 | 类型 | 必填 | 校验 |
|------|------|------|------|
| `student_id` | string | 是 | 学号，唯一 |
| `name` | string | 是 | 昵称 |
| `password` | string | 是 | 6-20 位字母数字 |
| `phone` | string | 是 | 联系电话 |
| `email` | string | 否 | 邮箱格式 |
| `qq` | string | 否 | QQ 号 |
| `weixin` | string | 否 | 微信号 |

```json
{
  "student_id": "2023123456",
  "name": "张三",
  "password": "abc12345",
  "phone": "13800138000",
  "email": "zhangsan@stu.xjtu.edu.cn",
  "qq": "12345678",
  "weixin": "wxid_zhangsan"
}
```

**响应 201：**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "student_id": "2023123456",
    "name": "张三",
    "avatar_url": "",
    "email": "zhangsan@stu.xjtu.edu.cn",
    "phone": "13800138000",
    "level": 0,
    "qq": "12345678",
    "weixin": "wxid_zhangsan"
  }
}
```

> 注册成功自动登录，Cookie 写入 `user-session`。

---

### POST /auth/login — 登录

- **认证**：否

| 字段 | 类型 | 必填 |
|------|------|------|
| `student_id` | string | 是 |
| `password` | string | 是 |

```json
{ "student_id": "2023123456", "password": "abc12345" }
```

**响应 200：** 同注册响应体。

---

### DELETE /auth/logout — 登出

- **认证**：是（Level ≥ 0）

**响应 200：** `{ "success": true, "resp": null }`

---

### GET /auth/me — 当前用户信息

- **认证**：是

**响应 200：** 同注册响应体。

---

### PUT /auth/password — 修改密码

- **认证**：是

| 字段 | 类型 | 必填 | 校验 |
|------|------|------|------|
| `old_password` | string | 是 | 原密码 |
| `new_password` | string | 是 | 6-20 位字母数字 |

```json
{ "old_password": "abc12345", "new_password": "xyz67890" }
```

**响应 200：** `{ "success": true, "resp": null }`

---

### PUT /auth/profile — 修改个人信息

- **认证**：是

| 字段 | 类型 | 必填 |
|------|------|------|
| `name` | string | 否 |
| `phone` | string | 否 |
| `email` | string | 否 |
| `qq` | string | 否 |
| `weixin` | string | 否 |

```json
{ "name": "张三丰", "phone": "13900139000" }
```

**响应 200：** 返回更新后的 `UserForm`。

---

### PUT /auth/description — 修改个人简介

- **认证**：是

| 字段 | 类型 | 必填 |
|------|------|------|
| `description` | string | 是 |

```json
{ "description": "热爱摄影的地理爱好者~" }
```

**响应 200：** `{ "success": true, "resp": null }`

---

### POST /auth/avatar — 上传头像

- **认证**：是
- **Content-Type**：`multipart/form-data`

| 字段 | 类型 | 必填 |
|------|------|------|
| `avatar` | file | 是 |

**响应 200：**

```json
{
  "success": true,
  "resp": { "avatar_url": "/uploads/avatars/xxx.jpg" }
}
```

---

## 3. 图片

### GET /photos — 图片列表

- **认证**：否

| Query | 类型 | 必填 | 说明 |
|-------|------|------|------|
| `page` | int | 否 | 默认 1 |
| `limit` | int | 否 | 默认 10 |
| `solved` | bool | 否 | 筛选已/未破解 |
| `sort_by` | string | 否 | `created_at` / `attempts_count` / `likes_count` |

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "total": 100,
    "photos": [
      {
        "id": 1,
        "title": "晨光中的图书馆",
        "description": "某个清晨的光影",
        "thumb_url": "/uploads/photos/xxx_thumb.jpg",
        "author": { "id": 1, "name": "张三", "avatar_url": "" },
        "solved": false,
        "attempts_count": 3,
        "likes_count": 12,
        "created_at": "2026-05-15T12:00:00+08:00"
      }
    ]
  }
}
```

> 仅返回已审核通过（`status = "approved"`）的图片。

---

### GET /photos/:id — 图片详情

- **认证**：否

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "title": "晨光中的图书馆",
    "description": "某个清晨的光影",
    "image_url": "/uploads/photos/xxx.jpg",
    "author": { "id": 1, "name": "张三", "avatar_url": "" },
    "solved": true,
    "attempts_count": 5,
    "created_at": "2026-05-15T12:00:00+08:00",
    "winner": {
      "id": 10,
      "image_url": "/uploads/attempts/xxx.jpg",
      "guessed_location": "主楼A座5楼",
      "likes_count": 5,
      "created_at": "2026-05-16T10:00:00+08:00",
      "user": { "id": 2, "name": "李四", "avatar_url": "" }
    }
  }
}
```

> `winner` 仅 `solved == true` 时返回。

---

### GET /photos/:id/image — 图片流式展示

- **认证**：否

直接返回图片二进制流（`Content-Type: image/jpeg`），用于 `<img>` 标签。

---

### GET /photos/:id/download — 图片下载

- **认证**：否

返回图片文件，含 `Content-Disposition: attachment` 头触发浏览器下载。

---

### POST /photos — 上传投稿

- **认证**：是
- **Content-Type**：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `image` | file | 是 | jpg/png |
| `title` | string | 是 | 图片标题 |
| `description` | string | 否 | 图片描述 |
| `location_secret` | string | 是 | 拍摄地点（仅管理员可见） |

**响应 201：**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "message": "投稿成功，等待管理员审核"
  }
}
```

> 新投稿状态为 `pending`，审核通过后公开。

---

### GET /photos/:id/comments — 图片评论列表

- **认证**：否

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "comments": [
      {
        "id": 1,
        "content": "好美的光影！",
        "likes_count": 3,
        "created_at": "2026-05-16T09:00:00+08:00",
        "user": { "id": 2, "name": "李四", "avatar_url": "" }
      }
    ]
  }
}
```

---

### GET /photos/:id/attempts — 图片答题列表

- **认证**：否

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "total": 3,
    "attempts": [
      {
        "id": 1,
        "image_url": "/uploads/attempts/xxx.jpg",
        "guessed_location": "主楼A座5楼",
        "likes_count": 5,
        "created_at": "2026-05-16T10:00:00+08:00",
        "user": { "id": 2, "name": "李四", "avatar_url": "" }
      }
    ]
  }
}
```

---

## 4. 答题与评论

### POST /photos/:id/attempts — 提交答案

- **认证**：是
- **Content-Type**：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `image` | file | 是 | 同地点同角度拍摄的匹配照片 |
| `guessed_location` | string | 是 | 猜测的地点描述 |

**响应 201：**

```json
{
  "success": true,
  "resp": {
    "attempt_id": 1,
    "photo_id": 1,
    "status": "pending",
    "message": "已提交，等待管理员审核"
  }
}
```

> 同用户同图片已有 `pending` 记录时不可重复提交。

---

### POST /photos/:id/comments — 发表评论

- **认证**：是

| 字段 | 类型 | 必填 |
|------|------|------|
| `comment_text` | string | 是 |

```json
{ "comment_text": "拍得太好了！" }
```

**响应 201：**

```json
{
  "success": true,
  "resp": { "id": 1, "message": "评论发表成功" }
}
```

---

## 5. 点赞

> 点赞为 toggle 机制：已赞则取消，未赞则点赞。

### POST /photos/:id/like — 切换图片点赞

- **认证**：是

### GET /photos/:id/like — 图片点赞状态

- **认证**：是

### POST /comments/:id/like — 切换评论点赞

- **认证**：是

### GET /comments/:id/like — 评论点赞状态

- **认证**：是

### POST /attempts/:id/like — 切换答题点赞

- **认证**：是

### GET /attempts/:id/like — 答题点赞状态

- **认证**：是

**统一响应格式：**

```json
{
  "success": true,
  "resp": { "liked": true, "likes_count": 13 }
}
```

---

## 6. 用户主页 & 奖品

### GET /users/:id — 用户主页

- **认证**：否

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "name": "张三",
    "avatar_url": "/uploads/avatars/xxx.jpg",
    "level": 0,
    "description": "热爱摄影的地理爱好者~",
    "prize_count": 2,
    "photo_count": 5,
    "attempt_count": 12
  }
}
```

### GET /users/:id/photos — 用户图片列表

- **认证**：否
- **Query：** `page`、`limit`、`sort_by`

### GET /users/:id/attempts — 用户答题列表

- **认证**：否
- **Query：** `page`、`limit`、`sort_by`

### GET /users/:id/comments — 用户评论列表

- **认证**：否
- **Query：** `page`、`limit`、`sort_by`

### GET /users/me/prizes — 我的奖品

- **认证**：是

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "total": 2,
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

| `status` | 说明 |
|----------|------|
| `unclaimed` | 待领取 |
| `claimed` | 已发放 |

---

## 7. 管理员接口

> 以下接口需管理员权限（`Level ≥ 1`），标注"高级"的需 `Level ≥ 2`。

### 7.1 图片审核

#### GET /admin/photos/pending — 待审核图片

- **认证**：管理员
- **Query：** `page`、`limit`；高级管理员可传 `status` 筛选

#### PUT /admin/photos/:id/review — 审核图片

- **认证**：管理员

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `action` | string | 是 | `"approve"` / `"reject"` |
| `reject_reason` | string | reject 时必填 | 拒绝原因 |

### 7.2 答题审核

#### GET /admin/attempts/pending — 待审核答题

- **认证**：管理员
- **Query：** `page`、`limit`；高级管理员可传 `status`

#### PUT /admin/attempts/:id/review — 审核答题

- **认证**：管理员

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `action` | string | 是 | `"approve"` / `"reject"` |
| `reject_reason` | string | reject 时必填 | 拒绝原因 |
| `solved` | bool | 否 | 高级管理员标记为已破解 |

> approve + solved=true + 图片未破解 → 自动生成奖品、更新用户获奖次数。

### 7.3 评论审核

#### GET /admin/comments/pending — 待审核评论

- **认证**：管理员
- **Query：** `page`、`limit`

#### PUT /admin/comments/:id/review — 审核评论

- **认证**：管理员

| 字段 | 类型 | 必填 |
|------|------|------|
| `action` | string | 是（`approve` / `reject`） |
| `reject_reason` | string | reject 时必填 |

### 7.4 奖品发放

#### PUT /admin/prizes/:id/claim — 标记奖品已发放

- **认证**：管理员

### 7.5 高级管理员专属

#### PUT /admin/admins/:id/level — 调整管理员等级

- **认证**：高级管理员

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `target_level` | int | 是 | 目标等级（0 = 降为普通用户） |

```json
{ "target_level": 2 }
```

**响应 200：**

```json
{
  "success": true,
  "resp": {
    "user_id": 5,
    "name": "管理员张三",
    "old_level": 1,
    "new_level": 2,
    "message": "管理员升级成功"
  }
}
```

> 约束：目标等级 ≤ 自身等级，不可操作同级/上级。

---

## 8. 消息通知

### GET /messages — 通知列表

- **认证**：是
- **Query：** `page`、`limit`

### GET /messages/unread-count — 未读通知数

- **认证**：是

```json
{ "success": true, "resp": { "unread_count": 3 } }
```

### PUT /messages/:id/read — 标记已读

- **认证**：是

---

## 9. 会话（私信）

### GET /conversations — 会话列表

- **认证**：是

返回当前用户的所有私信会话（类似微信首页）。

### GET /conversations/:id — 会话详情

- **认证**：是
- **Query：** `page`、`limit`

返回与指定用户的对话记录。

### POST /conversations/:id — 发送消息

- **认证**：是

| 字段 | 类型 | 必填 |
|------|------|------|
| `content` | string | 是 |

```json
{ "content": "你好，请问这个机位怎么找？" }
```

---

## 10. 接口速查表

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/auth/register` | — | 注册 |
| POST | `/auth/login` | — | 登录 |
| DELETE | `/auth/logout` | 用户 | 登出 |
| GET | `/auth/me` | 用户 | 当前用户信息 |
| PUT | `/auth/password` | 用户 | 修改密码 |
| PUT | `/auth/profile` | 用户 | 修改个人信息 |
| PUT | `/auth/description` | 用户 | 修改个人简介 |
| POST | `/auth/avatar` | 用户 | 上传头像 |
| GET | `/photos` | — | 图片列表 |
| GET | `/photos/:id` | — | 图片详情 |
| GET | `/photos/:id/image` | — | 图片流式展示 |
| GET | `/photos/:id/download` | — | 图片下载 |
| GET | `/photos/:id/comments` | — | 图片评论列表 |
| GET | `/photos/:id/attempts` | — | 图片答题列表 |
| POST | `/photos` | 用户 | 上传投稿 |
| POST | `/photos/:id/attempts` | 用户 | 提交答案 |
| POST | `/photos/:id/comments` | 用户 | 发表评论 |
| POST | `/photos/:id/like` | 用户 | 切换图片点赞 |
| GET | `/photos/:id/like` | 用户 | 图片点赞状态 |
| POST | `/comments/:id/like` | 用户 | 切换评论点赞 |
| GET | `/comments/:id/like` | 用户 | 评论点赞状态 |
| POST | `/attempts/:id/like` | 用户 | 切换答题点赞 |
| GET | `/attempts/:id/like` | 用户 | 答题点赞状态 |
| GET | `/users/:id` | — | 用户主页 |
| GET | `/users/:id/photos` | — | 用户图片列表 |
| GET | `/users/:id/attempts` | — | 用户答题列表 |
| GET | `/users/:id/comments` | — | 用户评论列表 |
| GET | `/users/me/prizes` | 用户 | 我的奖品 |
| GET | `/admin/photos/pending` | 管理员 | 待审核图片 |
| PUT | `/admin/photos/:id/review` | 管理员 | 审核图片 |
| GET | `/admin/attempts/pending` | 管理员 | 待审核答题 |
| PUT | `/admin/attempts/:id/review` | 管理员 | 审核答题 |
| GET | `/admin/comments/pending` | 管理员 | 待审核评论 |
| PUT | `/admin/comments/:id/review` | 管理员 | 审核评论 |
| PUT | `/admin/prizes/:id/claim` | 管理员 | 标记奖品已发放 |
| PUT | `/admin/admins/:id/level` | 高级管理员 | 调整管理员等级 |
| GET | `/messages` | 用户 | 通知列表 |
| GET | `/messages/unread-count` | 用户 | 未读通知数 |
| PUT | `/messages/:id/read` | 用户 | 标记已读 |
| GET | `/conversations` | 用户 | 会话列表 |
| GET | `/conversations/:id` | 用户 | 会话详情 |
| POST | `/conversations/:id` | 用户 | 发送私信 |

---

## 11. 业务规则摘要

| 规则 | 说明 |
|------|------|
| 图片审核 | 新投稿 `pending` → 管理员审核后 `approved`/`rejected`，仅 `approved` 公开 |
| 答题唯一性 | 同用户同图片仅一个 `pending` 答题 |
| 首位获奖 | 首个审核通过 + 标记破解 → 自动获奖并生成奖品 |
| 密码安全 | argon2id 加密，响应永不返回 `password` |
| 管理员等级 | Level 1 = 普通管理员，Level ≥ 2 = 高级管理员 |
| 等级调整 | 高级管理员可调整下级等级，不可越权操作 |
