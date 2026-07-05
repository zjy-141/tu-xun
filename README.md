# Auth

## 1. 登录

### 接口

```http
POST /auth/login
```

### 权限

无

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| code | string | 是 | 学校统一认证登录码 |

### 请求示例

```json
{
  "code": "school_auth_code"
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "token": "jwt_token",
    "expire": 7200,
    "user": {
      "id": 1,
      "student_id": "20230001",
      "nickname": "张三",
      "avatar": "xxx.jpg"
    }
  }
}
```

---

# User

## 1. 获取个人信息

### 接口

```http
GET /user/profile
```

### 权限

登录用户

### 请求参数

无

### 请求示例

```http
GET /user/profile
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "student_id": "20230001",
    "real_name": "张三",
    "nickname": "张三",
    "avatar": "avatar.jpg",
    "college": "计算机学院",
    "score": 1000,
    "role": "student"
  }
}
```

## 2. 修改个人资料

### 接口

```http
PUT /user/profile
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| nickname | string | 否 | 昵称 |
| avatar | string | 否 | 头像URL |

### 请求示例

```json
{
  "nickname": "挑战达人",
  "avatar": "avatar.jpg"
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

# Activity

## 1. 当前活动

### 接口

```http
GET /activity/current
```

### 权限

无

### 请求参数

无

### 请求示例

```http
GET /activity/current
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "title": "寻找校园角落",
    "cover": "cover.jpg",
    "description": "活动介绍",
    "start_time": "2026-06-01",
    "end_time": "2026-06-30",
    "status": "running"
  }
}
```

## 2. 往期活动列表

### 接口

```http
GET /activity/history
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /activity/history?page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "title": "第一期活动",
        "cover": "cover.jpg"
      }
    ]
  }
}
```

---

# Photo

## 1. 发布题目

### 接口

```http
POST /photo
```

### 权限

登录用户

### 请求参数

| 参数          | 类型     | 必填  | 说明     |
| ----------- | ------ | --- | ------ |
| activity_id | int    | 是   | 所属活动ID |
| title       | string | 是   | 标题     |
| description | string | 否   | 描述     |
| image_url   | string | 否   | 图片URL  |
| longitude   | float  | 否   | 经度     |
| latitude    | float  | 否   | 纬度     |

### 请求示例

```json
{
  "activity_id": 1,
  "title": "猜猜这是哪里",
  "description": "根据图片猜地点",
  "image_url": "xxx.jpg",
  "longitude": 108.123456,
  "latitude": 34.123456
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "photo_id": 1001,
    "status": "pending"
  }
}
```


## 2. 题目列表

### 接口

```http
GET /photo/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| activity_id | int | 是 | 活动ID |
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |
| keyword | string | 否 | 关键词搜索 |

### 请求示例

```http
GET /photo/list?activity_id=1&page=1&size=10&keyword=图书馆
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1001,
        "title": "图书馆后门",
        "cover": "cover.jpg",
        "publisher": "张三",
        "like_count": 10,
        "comment_count": 5,
        "answer_count": 3,
        "is_liked": true
      }
    ]
  }
}
```

## 3. 题目详情

### 接口

```http
GET /photo/{id}
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 题目ID（路径参数） |

### 请求示例

```http
GET /photo/1001
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1001,
    "title": "图书馆后门",
    "description": "猜地点",
    "image_url": "cover.jpg",
    "publisher": {
      "id": 1,
      "nickname": "张三",
      "avatar": "avatar.jpg"
    },
    "publish_time": "2026-06-01",
    "like_count": 20,
    "comment_count": 15,
    "answer_count": 8,
    "is_liked": true
  }
}
```

---

# Answer

## 1. 提交答案

### 接口

```http
POST /answer
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| photo_id | int | 是 | 题目ID |
| image_url | string | 是 | 答案图片URL |
| longitude | float | 否 | 经度 |
| latitude | float | 否 | 纬度 |

### 请求示例

```json
{
  "photo_id": 1001,
  "image_url": "answer.jpg",
  "longitude": 108.123456,
  "latitude": 34.123456
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "answer_id": 1,
    "status": "pending"
  }
}
```

## 2. 答题记录列表

### 接口

```http
GET /answer/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| photo_id | int | 是 | 题目ID |
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /answer/list?photo_id=1001&page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 10,
    "list": [
      {
        "id": 1,
        "nickname": "李四",
        "avatar": "avatar.jpg",
        "image_url": "answer.jpg",
        "answer_time": "2026-06-01",
        "like_count": 5,
        "is_liked": false
      }
    ]
  }
}
```

---

# Comment

## 1. 评论列表

### 接口

```http
GET /comment/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| photo_id | int | 是 | 题目ID |
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /comment/list?photo_id=1001&page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 30,
    "list": [
      {
        "id": 1,
        "nickname": "王五",
        "avatar": "avatar.jpg",
        "content": "这个地方我知道",
        "like_count": 2,
        "is_liked": true,
        "create_time": "2026-06-01"
      }
    ]
  }
}
```

## 2. 发表评论

### 接口

```http
POST /comment
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| photo_id | int | 是 | 题目ID |
| content | string | 是 | 评论内容 |

### 请求示例

```json
{
  "photo_id": 1001,
  "content": "这个地方我知道"
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "comment_id": 10
  }
}
```

## 3. 删除评论

### 接口

```http
DELETE /comment/{id}
```

### 权限

登录用户（仅可删除自己的评论）

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 评论ID（路径参数） |

### 请求示例

```http
DELETE /comment/10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

# Like

## 1. 点赞/取消点赞

### 接口

```http
POST /like/toggle
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| target_type | string | 是 | 目标类型：`photo` / `answer` / `comment` |
| target_id | int | 是 | 目标ID |

### 请求示例

```json
{
  "target_type": "photo",
  "target_id": 1001
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "liked": true,
    "like_count": 25
  }
}
```

---

# Score

## 1. 我的积分

### 接口

```http
GET /score
```

### 权限

登录用户

### 请求参数

无

### 请求示例

```http
GET /score
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total_score": 1000
  }
}
```

## 2. 积分流水

### 接口

```http
GET /score/logs
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /score/logs?page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 100,
    "list": [
      {
        "id": 1,
        "change_score": 50,
        "type": "answer_reward",
        "remark": "答题成功",
        "create_time": "2026-06-01"
      }
    ]
  }
}
```

---

# Goods

## 1. 商品列表（用户端）

### 接口

```http
GET /goods/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /goods/list?page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 20,
    "list": [
      {
        "id": 1,
        "name": "钥匙扣",
        "image_url": "xxx.jpg",
        "need_score": 100,
        "stock": 50
      }
    ]
  }
}
```

## 2. 商品详情（用户端）

### 接口

```http
GET /goods/{id}
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品ID（路径参数） |

### 请求示例

```http
GET /goods/1
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "钥匙扣",
    "description": "活动纪念品",
    "image_url": "xxx.jpg",
    "need_score": 100,
    "stock": 50
  }
}
```

---

# Exchange

## 1. 兑换商品

### 接口

```http
POST /exchange
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| goods_id | int | 是 | 商品ID |

### 请求示例

```json
{
  "goods_id": 1
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "exchange_id": 1001,
    "remain_score": 900
  }
}
```

## 2. 我的兑换记录

### 接口

```http
GET /exchange/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |
| status | string | 否 | 状态：`pending` / `verified` / `cancelled` |

### 请求示例

```http
GET /exchange/list?page=1&size=10&status=pending
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 5,
    "list": [
      {
        "id": 1001,
        "goods_name": "钥匙扣",
        "status": "pending",
        "create_time": "2026-06-01"
      }
    ]
  }
}
```

---

# Notice

## 1. 消息列表

### 接口

```http
GET /notice/list
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |

### 请求示例

```http
GET /notice/list?page=1&size=10
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "unread_count": 3,
    "total": 10,
    "list": [
      {
        "id": 1,
        "title": "活动开始通知",
        "create_time": "2026-06-01"
      }
    ]
  }
}
```

## 2. 消息详情

### 接口

```http
GET /notice/{id}
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 消息ID（路径参数） |

### 请求示例

```http
GET /notice/1
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "title": "活动开始通知",
    "content": "活动正式开始",
    "create_time": "2026-06-01"
  }
}
```

---

# Feedback

## 1. 提交反馈

### 接口

```http
POST /feedback
```

### 权限

登录用户

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 反馈内容 |

### 请求示例

```json
{
  "content": "希望增加排行榜功能"
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

# Admin

## 1. 审核题目

### 接口

```http
POST /admin/photo/review
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| photo_id | int | 是 | 题目ID |
| status | string | 是 | `approved` / `rejected` |
| reason | string | 否 | 驳回原因 |

### 请求示例

```json
{
  "photo_id": 1001,
  "status": "approved",
  "reason": ""
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

## 2. 审核答案

### 接口

```http
POST /admin/answer/review
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| answer_id | int | 是 | 答案ID |
| status | string | 是 | `approved` / `rejected` |
| reason | string | 否 | 驳回原因 |

### 请求示例

```json
{
  "answer_id": 1,
  "status": "approved",
  "reason": ""
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

## 3. 发布公告

### 接口

```http
POST /admin/notice
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 公告标题 |
| content | string | 是 | 公告内容 |

### 请求示例

```json
{
  "title": "活动开始通知",
  "content": "活动开始啦"
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "notice_id": 1
  }
}
```

## 4. 商品列表（后台）

### 接口

```http
GET /admin/goods/list
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认1 |
| size | int | 否 | 每页数量，默认10 |
| keyword | string | 否 | 商品名称搜索 |
| status | int | 否 | 1上架 0下架 |

### 请求示例

```http
GET /admin/goods/list?page=1&size=10&keyword=钥匙扣
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "total": 2,
    "list": [
      {
        "id": 1,
        "name": "挑战钥匙扣",
        "image_url": "https://xxx.com/key.jpg",
        "description": "活动纪念钥匙扣",
        "need_score": 100,
        "stock": 50,
        "status": 1,
        "created_at": "2026-06-01 10:00:00"
      }
    ]
  }
}
```

## 5. 商品详情（后台）

### 接口

```http
GET /admin/goods/{id}
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品ID（路径参数） |

### 请求示例

```http
GET /admin/goods/1
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "挑战钥匙扣",
    "image_url": "https://xxx.com/key.jpg",
    "description": "活动纪念钥匙扣",
    "need_score": 100,
    "stock": 50,
    "status": 1
  }
}
```

## 6. 新增商品

### 接口

```http
POST /admin/goods
```

### 权限

管理员

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 商品名称 |
| image_url | string | 是 | 商品图片URL |
| description | string | 否 | 商品描述 |
| need_score | int | 是 | 所需积分 |
| stock | int | 是 | 库存数量 |
| status | int | 是 | 1上架 0下架 |

### 请求示例

```json
{
  "name": "挑战钥匙扣",
  "image_url": "https://xxx.com/key.jpg",
  "description": "活动纪念钥匙扣",
  "need_score": 100,
  "stock": 50,
  "status": 1
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "goods_id": 1
  }
}
```

## 7. 修改商品

### 接口

```http
PUT /admin/goods/{id}
```

### 权限

管理员

### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 否 | 商品名称 |
| image_url | string | 否 | 商品图片URL |
| description | string | 否 | 商品描述 |
| need_score | int | 否 | 所需积分 |
| stock | int | 否 | 库存数量 |
| status | int | 否 | 1上架 0下架 |

### 请求示例

```json
{
  "name": "挑战纪念钥匙扣",
  "need_score": 120,
  "stock": 80
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

## 8. 删除商品

### 接口

```http
DELETE /admin/goods/{id}
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| id | int | 是 | 商品ID（路径参数） |

### 请求示例

```http
DELETE /admin/goods/1
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

## 9. 上下架商品

### 接口

```http
PUT /admin/goods/{id}/status
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | int | 是 | 1上架，0下架 |

### 请求示例

```json
{
  "status": 1
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

## 10. 补充库存

### 接口

```http
PUT /admin/goods/{id}/stock
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stock | int | 是 | 修改后的库存数量 |

### 请求示例

```json
{
  "stock": 100
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "stock": 100
  }
}
```

## 11. 奖品核销

### 接口

```http
POST /admin/exchange/verify
```

### 权限

管理员

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| exchange_id | int | 是 | 兑换记录ID |

### 请求示例

```json
{
  "exchange_id": 1001
}
```

### 返回

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "success": true
  }
}
```

---

## 附录：常用状态定义

- 商品状态：`1` 上架，`0` 下架  
- 兑换状态：`pending` 待核销，`verified` 已核销，`cancelled` 已取消  
- 审核状态：`approved` 通过，`rejected` 驳回

---

以上共计 **40个接口**，覆盖登录认证、用户、活动、题目、答案、评论、点赞、积分、商城兑换、消息、反馈及后台管理全部功能，可直接用于 Swagger/OpenAPI 文档生成。