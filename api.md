
# 图寻 API 文档

> 基础路径：`/api`
>
> 统一响应格式（四个字段均为必返回）：
> ```json
> // 成功
> { "success": true, "resp": {}, "message": "", "code": 0 }
>
> // 失败
> { "success": false, "resp": null, "message": "错误描述", "code": 40001 }
> ```

---

## 鉴权

| Level | 说明 |
|-------|------|
| 0 | 客户端非登录用户，仅用于 tuxun-fe |
| 1 | 客户端普通登录用户，仅用于 tuxun-fe |
| 2 | 后台管理员，可进入 tuxun-admin-fe |
| 3 | 后台超级管理员，可进入 tuxun-admin-fe |

**Session Cookie**：`tz-sessions`，有效期 30 分钟，每次请求自动续期，退出后立即失效。
生产环境 `HttpOnly`、`Secure`、`SameSite=Lax`（防 CSRF）。

**HTTP 状态码规则**：

| 状态码 | 场景 |
|--------|------|
| 200 | 业务成功或业务类失败（参数错误、库存不足等） |
| 401 | 未登录或 Session 失效 |
| 403 | 权限不足 |
| 409 | 冲突（重复审核、重复核销、并发冲突） |

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

---

## 目录

- [用户认证](#用户认证)
- [活动](#活动)
- [图寻题目 (Photos)](#图寻题目-photos)
- [答题 (Attempts)](#答题-attempts)
- [评论 (Comments)](#评论-comments)
- [积分 (Score)](#积分-score)
- [奖品 (Goods)](#奖品-goods)
- [兑换 (Exchange)](#兑换-exchange)
- [消息通知 (Messages)](#消息通知-messages)
- [公告 (Notice)](#公告-notice)
- [反馈 (Feedback)](#反馈-feedback)
- [管理员](#管理员)

---

## 约定

- **分页**：`page` 默认 1（min=1），`page_size` 默认 10（min=1, max=20）。空列表返回 `[]`，不返回 `null`。
- **时间**：纯日期格式 `YYYY-MM-DD`；具体时间格式 ISO 8601 带时区 `2026-07-20T10:30:00+08:00`。
- **空值**：未设置的时间字段（如 `exchange_at`、`reviewed_at`）返回 `null`。
- **图片 URL**：统一返回以 `/uploads/` 开头的同域路径。
- **字段命名**：JSON 字段统一使用 `snake_case`；Query 参数统一使用 `snake_case`（如 `page_size`、`sort_by`、`activity_id`）。
- **上传文件**：仅支持 jpg/png，单文件 ≤20MB。反馈附件最多 3 个。
- **列表字段**：分页响应中数组字段统一命名为 `list`。

---

## 用户认证

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
    "avatar_url": "/uploads/avatars/avatar.jpg",
    "level": 1
  },
  "message": "",
  "code": 0
}
```

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
    "avatar_url": "/uploads/avatars/avatar.jpg",
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

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称（最长 20） |

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

### 6. 修改头像

```
PUT /api/user/avatar
```

**权限**：登录用户（Level ≥ 1）

**Content-Type**：`multipart/form-data`

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

---

## 活动

### 1. 当前活动

```
GET /api/activity/current
```

**权限**：无

**返回** `200`
- 有活动时：

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "title": "寻找校园角落",
    "cover_url": "/uploads/photos/cover.jpg",
    "description": "活动介绍",
    "is_active": true,
    "start_time": "2026-06-01",
    "end_time": "2026-06-30"
  },
  "message": "",
  "code": 0
}
```

- 无活动时：

```json
{
  "success": false,
  "resp": null,
  "message": "操作错误: 当前没有活动开放",
  "code": 5
}
```

---

### 2. 往期活动列表

```
GET /api/activity/history
```

**权限**：无

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
        "cover_url": "/uploads/photos/cover.jpg",
        "description": "活动介绍",
        "is_active": false,
        "start_time": "2026-05-01",
        "end_time": "2026-05-31"
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

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 所属活动 ID |
| title | string | 是 | 图片标题 |
| description | string | 否 | 图片描述/故事 |
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

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 是 | — | 所属活动 ID |
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
        "author": { "id": 1, "nickname": "张三", "avatar_url": "/uploads/avatars/avatar.jpg" },
        "title": "猜猜这是哪",
        "thumb_url": "/uploads/photos/thumb.jpg",
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
GET /api/photos/:id
```

**权限**：无

**说明**：`:id` 为图片（Photo）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "author": { "id": 1, "nickname": "张三", "avatar_url": "/uploads/avatars/avatar.jpg" },
    "activity": { "id": 1, "title": "寻找校园角落", "description": "活动介绍" },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image_url": "/uploads/photos/photo.jpg",
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
GET /api/photos/:id/image
```

**权限**：无

**说明**：直接返回图片二进制流（`Content-Type: image/jpeg` 或 `image/png`），适用于 `<img>` 标签 src。浏览器缓存 1 小时。不是 302 跳转。

---

### 5. 图片下载

```
GET /api/photos/:id/download
```

**权限**：无

**说明**：触发浏览器下载原图。返回二进制流 + `Content-Disposition: attachment` 头。不是 302 跳转。

---

### 6. 图片评论列表

```
GET /api/photos/:id/comments
```

**权限**：无

**说明**：`:id` 为图片（Photo）ID。

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
        "author": { "id": 1, "nickname": "张三", "avatar_url": "/uploads/avatars/avatar.jpg" },
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
POST /api/photos/:id/attempts
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| comment_text | string | 否 | 补充说明（最长 500） |
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
GET /api/photos/:id/attempts
```

**权限**：无

**说明**：`:id` 为图片（Photo）ID。

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
        "author": { "id": 2, "nickname": "李四", "avatar_url": "/uploads/avatars/avatar.jpg" },
        "photo_id": 1,
        "comment_text": "应该是东花园",
        "image_url": "/uploads/attempts/attempt.jpg",
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
GET /api/photos/:id/attempts/user
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。获取当前登录用户在该图片下的答题记录。

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
        "photo_id": 1,
        "comment_text": "应该是东花园",
        "image_url": "/uploads/attempts/attempt.jpg",
        "longitude": 108.123456,
        "latitude": 34.123456,
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

### 10. 切换图片点赞

```
POST /api/photos/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。已点赞则取消，未点赞则点赞。未登录返回 401。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 11 },
  "message": "",
  "code": 0
}
```

---

### 11. 获取图片点赞状态

```
GET /api/photos/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。未登录返回 401。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 11 },
  "message": "",
  "code": 0
}
```

---

### 12. 发表评论

```
POST /api/photos/:id/comments
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| comment_text | string | 是 | 评论内容（最长 500） |

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

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 是 | — | 所属活动 ID |
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |
| solved | bool | 否 | — | 筛选是否已破解 |
| sort_by | string | 否 | created_at | `created_at` / `likes_count` / `attempts_count` |

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
        "description": "校园神秘角落",
        "thumb_url": "/uploads/photos/thumb.jpg",
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
GET /api/photos/review/:id
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为图片（Photo）ID。只能查看自己的投稿（管理员可查看任意）。

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "activity": { "id": 1, "title": "寻找校园角落", "description": "活动介绍" },
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image_url": "/uploads/photos/photo.jpg",
    "longitude": 108.123456,
    "latitude": 34.123456,
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

### 1. 切换答题点赞

```
POST /api/attempts/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为答题记录（Attempt）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 5 },
  "message": "",
  "code": 0
}
```

---

### 2. 获取答题点赞状态

```
GET /api/attempts/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为答题记录（Attempt）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 5 },
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
    "total": 5,
    "list": [
      {
        "id": 1,
        "photo_id": 1,
        "comment_text": "应该是东花园",
        "image_url": "/uploads/attempts/attempt.jpg",
        "longitude": 108.123456,
        "latitude": 34.123456,
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

## 评论 (Comments)

### 1. 删除评论

```
DELETE /api/comments/:id
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为评论（Comment）ID。普通用户只能删除自己的评论，管理员可删除任意评论。

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

### 2. 切换评论点赞

```
POST /api/comments/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为评论（Comment）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 3 },
  "message": "",
  "code": 0
}
```

---

### 3. 获取评论点赞状态

```
GET /api/comments/:id/like
```

**权限**：登录用户（Level ≥ 1）

**说明**：`:id` 为评论（Comment）ID。

**返回** `200`

```json
{
  "success": true,
  "resp": { "is_like": true, "like_count": 3 },
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
        "thumb_url": "/uploads/goods/thumb.jpg",
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
GET /api/goods/:id
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
    "image_url": "/uploads/goods/good.jpg",
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

- 库存不足：

```json
{
  "success": false,
  "resp": null,
  "message": "参数错误: 奖品库存不足",
  "code": 3
}
```

- 积分不足：

```json
{
  "success": false,
  "resp": null,
  "message": "参数错误: 用户积分不足",
  "code": 3
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
        "good": { "id": 1, "name": "明信片套装", "thumb_url": "/uploads/goods/thumb.jpg", "need_score": 500, "stock": 20 },
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

## 消息通知 (Messages)

### 1. 消息列表

```
GET /api/messages/list
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "sender_id": 1,
        "title": "审核通过通知",
        "content": "您投稿的图片已通过审核",
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

### 2. 消息详情

```
GET /api/messages/:id
```

**权限**：登录用户（Level ≥ 1）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "user_id": 1,
    "sender_id": 1,
    "type": "review_approved",
    "title": "审核通过通知",
    "content": "您投稿的图片已通过审核",
    "related_id": 1,
    "related_type": "photo",
    "is_read": true,
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

---

### 3. 未读消息数

```
GET /api/messages/unread-count
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

### 4. 标记已读

```
PUT /api/messages/:id/read
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

## 公告 (Notice)

### 1. 公告列表

```
GET /api/messages/notice
```

**权限**：登录用户（Level ≥ 1）

**说明**：按活动 ID 查询管理员发布的公告。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| activity_id | int | 是 | — | 活动 ID |
| page | int | 否 | 1 | 页码（min=1） |
| page_size | int | 否 | 10 | 每页数量（min=1, max=20） |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "title": "系统公告",
        "content": "欢迎参与图寻挑战活动！",
        "activity_id": 1,
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
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
| title | string | 是 | 标题（最长 100） |
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

> 以下接口均需管理员权限（Level ≥ 2），标注「超级管理员」需 Level ≥ 3。

---

### 审核 — 图片

#### 待审核图片列表

```
GET /api/admin/photos/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | pending | `pending` / `approved` / `rejected` |

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "user_id": 1,
        "activity_id": 1,
        "title": "校园角落",
        "description": "猜猜这是哪里",
        "longitude": 108.123456,
        "latitude": 34.123456,
        "thumb_url": "/uploads/photos/thumb.jpg"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 审核图片

```
PUT /api/admin/photos/:id/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` / `reject` |
| reject_reason | string | 条件必填 | **action=reject 时必填**；action=approve 时不用传 |

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

#### 待审核答题列表

```
GET /api/admin/attempts/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | pending | `pending` / `solved` / `unsolved` |

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
        "guess_image_url": "/uploads/attempts/attempt.jpg",
        "guess_longitude": 108.5,
        "guess_latitude": 34.5,
        "thumb_url": "/uploads/photos/thumb.jpg",
        "longitude": 108.123,
        "latitude": 34.456,
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
PUT /api/admin/attempts/:id/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| solved | string | 是 | `solved` / `unsolved` |
| reject_reason | string | 否 | unsolved 的拒绝原因（不填时有默认文案） |

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

#### 待审核评论列表

```
GET /api/admin/comments/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| status | string | 否 | pending | `pending` / `approved` / `rejected` |

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
        "user": { "id": 1, "nickname": "张三", "avatar_url": "/uploads/avatars/avatar.jpg" },
        "comment": "评论内容",
        "created_at": "2026-06-01T12:00:00+08:00"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 审核评论

```
PUT /api/admin/comments/:id/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | `approve` / `reject` |
| reject_reason | string | 条件必填 | **action=reject 时必填**；action=approve 时不用传 |

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

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |
| keyword | string | 否 | — | 关键词搜索（最长 50） |

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
        "cover_url": "/uploads/photos/cover.jpg",
        "description": "活动介绍",
        "is_active": false,
        "start_time": "2026-05-01",
        "end_time": "2026-05-31"
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 活动详情

```
GET /api/admin/activity/:id
```

**权限**：管理员（Level ≥ 2）

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "title": "寻找校园角落",
    "cover_url": "/uploads/photos/cover.jpg",
    "description": "活动描述",
    "is_active": true,
    "start_time": "2026-07-01",
    "end_time": "2026-08-01",
    "photo_points": 50,
    "tiers": [
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

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 活动标题（最长 255） |
| cover_file | file | 否 | 封面图（jpg/png，≤20MB） |
| description | string | 是 | 活动描述 |
| start_time | string | 是 | 开始时间（`YYYY-MM-DD HH:MM:SS`） |
| end_time | string | 是 | 结束时间（必须晚于开始时间） |
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

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 活动 ID |
| title | string | 否 | 活动标题 |
| cover_file | file | 否 | 封面图（jpg/png，≤20MB） |
| description | string | 否 | 活动描述 |
| start_time | string | 否 | 开始时间 |
| end_time | string | 否 | 结束时间 |
| is_active | bool | 否 | 设为当前活动 |
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

#### 发布活动公告

```
POST /api/admin/activity/notice
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 活动 ID |
| title | string | 是 | 公告标题（最长 128） |
| content | string | 是 | 公告内容 |

**返回** `200`

```json
{
  "success": true,
  "resp": { "id": 1, "status": "success" },
  "message": "",
  "code": 0
}
```

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
| keyword | string | 否 | — | 关键词搜索（最长 50） |

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
        "thumb_url": "/uploads/goods/thumb.jpg",
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
GET /api/admin/goods/:id
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
    "image_url": "/uploads/goods/good.jpg",
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
| name | string | 是 | 奖品名称（最长 50） |
| description | string | 否 | 描述（最长 500） |
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
PUT /api/admin/goods/:id
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 奖品名称 |
| description | string | 否 | 描述 |
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
DELETE /api/admin/goods/:id
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
PUT /api/admin/goods/:id/status
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
PUT /api/admin/goods/:id/stock
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

**返回** `200`

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "list": [
      {
        "id": 1,
        "user": { "id": 1, "nickname": "张三", "avatar_url": "/uploads/avatars/avatar.jpg" },
        "good": { "id": 1, "name": "明信片套装", "thumb_url": "/uploads/goods/thumb.jpg", "need_score": 500, "stock": 20 },
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
GET /api/admin/feedback/:id
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
      { "id": 1, "url": "/uploads/feedbacks/xxx.jpg", "media_type": 1 }
    ],
    "created_at": "2026-06-01T12:00:00+08:00"
  },
  "message": "",
  "code": 0
}
```

#### 处理反馈

```
PUT /api/admin/feedback/:id
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

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| keyword | string | 否 | — | 按学号/姓名/昵称模糊搜索（最长 50） |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |

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
        "name": "张三",
        "nickname": "张三",
        "level": 1
      }
    ]
  },
  "message": "",
  "code": 0
}
```

#### 用户列表（超级管理员）

```
GET /api/admin/user
```

**权限**：超级管理员（Level ≥ 3）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| netid | string | 否 | — | 学号精确匹配 |
| name | string | 否 | — | 姓名精确匹配 |
| nickname | string | 否 | — | 昵称精确匹配 |
| page | int | 否 | 1 | 页码 |
| page_size | int | 否 | 10 | 每页数量（max=20） |

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
        "avatar_url": "/uploads/avatars/avatar.jpg",
        "level": 2
      }
    ]
  },
  "message": "",
  "code": 0
}
```

---

### 管理员等级管理（超级管理员）

```
PUT /api/admin/level
```

**权限**：超级管理员（Level ≥ 3）

**说明**：调整任意用户等级。`{id}` 为目标用户 ID。`target_level` 范围 0-3，不能超过操作者自身等级。只有 Level ≥ 3 可执行此操作。

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 目标用户 ID |
| target_level | int | 是 | 目标等级（0-3，min=0） |

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
| Message | is_read | `true` / `false` | false |
| Activity | is_active | `true` / `false` | false |

**可空字段**（未设置时返回 `null`）：`start_time`、`end_time`、`reviewed_at`、`exchange_at`、`reject_reason`。
