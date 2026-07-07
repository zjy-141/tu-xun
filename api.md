
# 图寻 API 文档

> 基础路径：`/api`
>
> 统一响应格式：
> ```json
> { "success": true, "resp": {}, "message": "", "code": 0 }
> ```

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
- [反馈 (Feedback)](#反馈-feedback)
- [管理员](#管理员)

---

## 用户认证

### 1. 登录

```
GET /api/user/login
```

**权限**：无

**说明**：重定向到学校统一认证页面。已登录用户直接重定向到回调地址。

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

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "username": "张三",
    "nickname": "张三",
    "avatar_url": "avatar.jpg",
    "level": 1
  }
}
```

---

### 3. 登出

```
DELETE /api/user/logout
```

**权限**：无（清除 Session）

---

### 4. 获取个人信息

```
GET /api/user/info
```

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "netid": "20230001",
    "name": "张三",
    "nickname": "张三",
    "avatar_url": "avatar.jpg",
    "edulevel": "本科生",
    "score_count": 1000,
    "level": 1
  }
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
| nickname | string | 否 | 昵称（最长20） |

**返回**

```json
{
  "success": true,
  "resp": {
  }
}
```

---

### 6. 修改头像

```
PUT /api/user/avatar
```

**权限**：登录用户（Level ≥ 1）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| avatar | file | 是 | 新头像文件 |


**返回**

```json
{
  "success": true,
  "resp": {
  }
}
```


## 活动

### 1. 当前活动

```
GET /api/activity/current
```

**权限**：无

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "title": "寻找校园角落",
    "cover": "cover.jpg",
    "description": "活动介绍",
    "is_active": true,
    "start_time": "2026-06-01",
    "end_time": "2026-06-30"
  }
}
```

---

### 2. 往期活动列表

```
GET /api/activity/history
```

**权限**：无

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1（min=1） |
| limit | int | 否 | 每页数量，默认 10（min=1, max=20） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "activities": [
      {
        "id": 1,
        "title": "第一期活动",
        "cover": "cover.jpg",
        "description": "活动介绍",
        "is_active": false,
        "start_time": "2026-05-01",
        "end_time": "2026-05-31"
      }
    ]
  }
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
| image | file | 是 | 图片文件 |
| longitude | float | 是 | 经度 |
| latitude | float | 是 | 纬度 |
|coord_type|string|是|坐标系|

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1001,
    "status": "pending"
  }
}
```

---

### 2. 图片列表

```
GET /api/photos/list
```

**权限**：无

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 所属活动 ID |
| page | int | 否 | 页码（min=1） |
| limit | int | 否 | 每页数量（min=1, max=20） |
| solved | bool | 否 | 筛选是否已破解 |
| sort_by | string | 否 | 排序字段：`created_at` / `likes_count` / `attempts_count` |
| keyword | string | 否 | 关键词搜索（最长50） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 100,
    "photos": [
      {
        "id": 1,
        "author": { "id": 1, "nickname": "张三", "avatar_url": "avatar.jpg" },
        "title": "猜猜这是哪",
        "thumb_url": "thumb.jpg",
        "solved": false,
        "likes_count": 10
      }
    ]
  }
}
```

---

### 3. 图片详情

```
GET /api/photos/:id
```

**权限**：无

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "author": { "id": 1, "nickname": "张三", "avatar_url": "avatar.jpg" },
    "activity_id": 1,
    "title": "猜猜这是哪",
    "description": "一个神秘的角落",
    "image_url": "photo.jpg",
    "solved": false,
    "attempts_count": 5,
    "likes_count": 10,
    "created_at": "2026-06-01T12:00:00Z",
    "status": "approved"
  }
}
```

---

### 4. 图片流展示

```
GET /api/photos/:id/image
```

**权限**：无

**说明**：直接返回图片二进制流，可用于 `<img>` 标签 src 属性。浏览器缓存 1 小时。

---

### 5. 图片下载

```
GET /api/photos/:id/download
```

**权限**：无

**说明**：触发浏览器下载原图。

---

### 6. 图片评论列表

```
GET /api/photos/:id/comments
```

**权限**：无

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| sort_by | string | 否 | 排序：`created_at` / `likes_count` / `attempts_count` |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "comments": [
      {
        "id": 1,
        "author": { "id": 1, "nickname": "张三", "avatar_url": "avatar.jpg" },
        "photo_id": 1,
        "comment_text": "我知道这是哪里！",
        "likes_count": 3,
        "created_at": "2026-06-01T12:00:00Z"
      }
    ]
  }
}
```

---

### 7. 图片答题列表

```
GET /api/photos/:id/attempts
```

**权限**：无

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| sort_by | string | 否 | 排序：`created_at` / `likes_count` / `attempts_count` |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "attempts": [
      {
        "id": 1,
        "author": { "id": 2, "nickname": "李四", "avatar_url": "avatar.jpg" },
        "photo_id": 1,
        "comment_text": "应该是东花园",
        "image_url": "attempt.jpg",
        "solved": false,
        "likes_count": 0,
        "created_at": "2026-06-02T10:00:00Z"
      }
    ]
  }
}
```

---

### 8. 切换图片点赞

```
POST /api/photos/:id/like
```

**权限**：无（公开接口，但需登录状态才有效）

**说明**：已点赞则取消，未点赞则点赞。

**返回**

```json
{
  "success": true,
  "resp": {
    "is_like": true,
    "like_count": 11
  }
}
```

---

### 9. 获取图片点赞状态

```
GET /api/photos/:id/like
```

**权限**：无（公开接口，但需登录状态才有效）

**返回**

```json
{
  "success": true,
  "resp": {
    "is_like": true,
    "like_count": 11
  }
}
```

---

## 答题 (Attempts)

### 1. 提交答案

```
POST /api/attempts/:id/attempts
```

**说明**：`:id` 为图片（Photo）ID

**权限**：登录用户（Level ≥ 1）

**Content-Type**：`multipart/form-data`

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| comment_text | string | 否 | 补充说明（最长500） |
| image_file | file | 是 | 猜测的匹配照片 |
| longitude | float | 是 | 猜测经度 |
| latitude | float | 是 | 猜测纬度 |
|coord_type|string|是|坐标系|

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 100,
    "status": "pending"
  }
}
```

---

### 2. 切换答题点赞

```
POST /api/attempts/:id/like
```

**说明**：`:id` 为答题记录（Attempt）ID

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "is_like": true,
    "like_count": 5
  }
}
```

---

### 3. 获取答题点赞状态

```
GET /api/attempts/:id/like
```

**说明**：`:id` 为答题记录（Attempt）ID

**权限**：登录用户（Level ≥ 1）

---

## 评论 (Comments)

### 1. 发表评论

```
POST /api/comments/:id/comments
```

**说明**：`:id` 为图片（Photo）ID

**权限**：登录用户（Level ≥ 1）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| comment_text | string | 否 | 评论内容（最长500） |

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 50,
    "status": "pending"
  }
}
```

---

### 2. 删除评论

```
DELETE /api/comments/:id
```

**说明**：`:id` 为评论（Comment）ID。普通用户只能删除自己的评论，管理员可删除任意评论。

**权限**：登录用户（Level ≥ 1）

---

### 3. 切换评论点赞

```
POST /api/comments/:id/like
```

**说明**：`:id` 为评论（Comment）ID

**权限**：登录用户（Level ≥ 1）

---

### 4. 获取评论点赞状态

```
GET /api/comments/:id/like
```

**说明**：`:id` 为评论（Comment）ID

**权限**：登录用户（Level ≥ 1）

---

## 积分 (Score)

### 1. 我的积分

```
GET /api/score
```

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "total_score": 1000
  }
}
```

---

### 2. 积分流水

```
GET /api/score/logs
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码（min=1） |
| limit | int | 否 | 每页数量（min=1, max=20） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 20,
    "score_logs": [
      {
        "id": 1,
        "delta": 10,
        "balance": 1000,
        "reason": "upload_photo",
        "related_id": 1,
        "related_type": "photo",
        "created_at": "2026-06-01T12:00:00Z"
      }
    ]
  }
}
```

> **reason 类型说明**：`upload_photo` / `answer_correct` / `like_photo` / `get_liked` / `comment` / `review_pass` / `daily_login` / `admin_adjust`

---

## 奖品 (Goods)

### 1. 奖品列表

```
GET /api/goods/list
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| available | bool | 否 | 仅看可兑换（库存 > 0） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "goods": [
      {
        "id": 1,
        "name": "明信片套装",
        "thumb_url": "thumb.jpg",
        "need_score": 500,
        "stock": 20
      }
    ]
  }
}
```

---

### 2. 奖品详情

```
GET /api/goods/:id
```

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "name": "明信片套装",
    "description": "精美校园风景明信片",
    "image_url": "good.jpg",
    "need_score": 500,
    "stock": 20
  }
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

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "status": "pending"
  }
}
```

---

### 2. 我的兑换记录

```
GET /api/exchange/list
```

**权限**：登录用户（Level ≥ 1）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| status | string | 否 | 筛选状态：`pending` / `verified` / `cancelled` |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "exchanges": [
      {
        "id": 1,
        "good": { "id": 1, "name": "明信片套装", "thumb_url": "thumb.jpg", "need_score": 500, "stock": 20 },
        "quantity": 1,
        "score_cost": 500,
        "status": "pending",
        "exchange_at": "",
        "created_at": "2026-06-01T12:00:00Z"
      }
    ]
  }
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

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 10,
    "messages": [
      {
        "id": 1,
        "sender_id": 0,
        "title": "审核通过通知",
        "is_read": false,
        "created_at": "2026-06-01T12:00:00Z"
      }
    ]
  }
}
```

---

### 2. 消息详情

```
GET /api/messages/:id
```

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "id": 1,
    "user_id": 1,
    "sender_id": 0,
    "type": "review_approved",
    "title": "审核通过通知",
    "content": "您投稿的图片已通过审核",
    "related_id": 1,
    "related_type": "photo",
    "is_read": true,
    "created_at": "2026-06-01T12:00:00Z"
  }
}
```

---

### 3. 未读消息数

```
GET /api/messages/unread-count
```

**权限**：登录用户（Level ≥ 1）

**返回**

```json
{
  "success": true,
  "resp": {
    "count": 3
  }
}
```

---

### 4. 标记已读

```
PUT /api/messages/:id/read
```

**权限**：登录用户（Level ≥ 1）

---

### 5. 公告列表

```
GET /api/messages/notice
```

**权限**：登录用户（Level ≥ 1）

**说明**：获取管理员发布的系统公告（消息类型为 `notice`，按活动关联筛选）。

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 活动 ID |
| page | int | 否 | 页码（min=1） |
| limit | int | 否 | 每页数量（min=1, max=20） |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "message_notices": [
      {
        "title": "系统公告",
        "content": "欢迎参与图寻挑战活动！",
        "related_id": 1,
        "related_type": "activity",
        "created_at": "2026-06-01T12:00:00Z"
      }
    ]
  }
}
```

---

## 反馈 (Feedback)

### 1. 提交反馈

```
POST /api/feedback
```

**权限**：登录用户（Level ≥ 1）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 否 | 标题（最长100） |
| comment_text | string | 否 | 内容（最长500） |

---

## 管理员

> 以下接口均需管理员权限（Level ≥ 2），标注「超级管理员」需 Level ≥ 3。

---

### 审核 - 图片

#### 待审核图片列表

```
GET /api/admin/photos/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| status | string | 否 | 筛选状态：`pending` / `approved` / `rejected` |

**返回**

```json
{
  "success": true,
  "resp": {
    "total": 5,
    "pending_photos": [
      {
        "id": 1,
        "user_id": 1,
        "activity_id": 1,
        "title": "校园角落",
        "description": "猜猜这是哪里",
        "longitude": 108.123456,
        "latitude": 34.123456,
        "thumb_url": "thumb.jpg"
      }
    ]
  }
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
| action | string | 是 | 审核操作：`approve` / `reject` |
| reject_reason | string | 否 | 拒绝原因 |

---

### 审核 - 答题

#### 待审核答题列表

```
GET /api/admin/attempts/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| status | string | 否 | 筛选状态：`pending` / `approved` / `rejected` |

#### 审核答题

```
PUT /api/admin/attempts/:id/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| solved | string | 是 | 是否破解成功：`solved` / `unsolved` |
| reject_reason | string | 否 | 拒绝原因 |

---

### 审核 - 评论

#### 待审核评论列表

```
GET /api/admin/comments/pending
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| status | string | 否 | 筛选状态：`pending` / `approved` / `rejected` |

#### 审核评论

```
PUT /api/admin/comments/:id/review
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| action | string | 是 | 审核操作：`approve` / `reject` |
| reject_reason | string | 否 | 拒绝原因 |

---

### 全服公告

```
POST /api/admin/notice
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 公告标题 |
| content | string | 是 | 公告内容 |
| relatedID | int | 否 | 关联实体 ID |
| relatedType | string | 否 | 关联实体类型 |

---

### 商品管理

#### 商品列表

```
GET /api/admin/goods/list
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| available | bool | 否 | 仅看可兑换 |
| status | string | 否 | 筛选：`inStore` / `outStore` |
| keyword | string | 否 | 关键词搜索（最长50） |

#### 商品详情

```
GET /api/admin/goods/:id
```

**权限**：管理员（Level ≥ 2）

#### 新增商品

```
POST /api/admin/goods/new
```

**权限**：管理员（Level ≥ 2）

**Content-Type**：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 奖品名称（最长50） |
| description | string | 否 | 描述（最长500） |
| need_score | int | 是 | 所需积分（min=0） |
| stock | int | 是 | 库存（min=0） |
| image | file | 是 | 商品图片 |
| status | string | 否 | `inStore`（上架）/ `outStore`（下架） |

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
| image | file | 否 | 商品图片 |
| status | string | 否 | `inStore` / `outStore` |

#### 删除商品

```
DELETE /api/admin/goods/:id
```

**权限**：管理员（Level ≥ 2）

#### 更新商品状态

```
PUT /api/admin/goods/:id/status
```

**权限**：管理员（Level ≥ 2）

#### 更新商品库存

```
PUT /api/admin/goods/:id/stock
```

**权限**：管理员（Level ≥ 2）

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stock | int | 是 | 新库存数量（min=0） |

---

### 兑换管理

#### 兑换记录列表

```
GET /api/admin/exchange/list
```

**权限**：管理员（Level ≥ 2）

**请求参数（Query）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| limit | int | 否 | 每页数量（max=20） |
| status | string | 否 | 筛选：`pending` / `verified` / `cancelled` |

#### 核销兑换

```
POST /api/admin/exchange/verify
```

**权限**：管理员（Level ≥ 2）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| exchange_id | int | 是 | 兑换记录 ID |
| action | string | 是 | `verify`（核销）/ `cancel`（取消） |

---

### 管理员等级管理（超级管理员）

```
PUT /api/admin/admins/:id/level
```

**权限**：超级管理员（Level ≥ 3）

**请求参数（JSON Body）**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_level | int | 是 | 目标等级（min=0） |
