# 图寻 (Tu-Xun) — 校园图片定位挑战平台

基于 **Go + Gin** 的校园图片定位挑战平台后端服务。用户可以上传校园角落照片作为"图寻题目"，其他用户通过猜测拍摄地点来答题破解，赢取积分并兑换奖品。

---

## 技术栈

| 类别 | 技术 |
|------|------|
| 语言 | Go 1.25 |
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io/) (MySQL) |
| 认证 | Session（基于 Cookie）+ 学校统一认证 |
| 密码哈希 | [Argon2id](https://github.com/alexedwards/argon2id) |
| 图片存储 | 阿里云 OSS / 本地存储 |
| 图片处理 | [imaging](https://github.com/disintegration/imaging) |
| 日志 | [Logrus](https://github.com/sirupsen/logrus) + lumberjack 滚动切割 |
| 配置 | 环境变量 + `.env` 文件 ([godotenv](https://github.com/joho/godotenv)) |

---

## 快速开始

```bash
# 1. 克隆项目
git clone <repo-url>
cd tu-xun-be

# 2. 配置环境变量（拷贝示例文件并修改）
cp .env.example .env

# 3. 初始化数据库（执行 SQL 迁移文件）
# 使用 sql/tuxun-init.up.sql

# 4. 运行
go run main.go
```

服务默认监听 `0.0.0.0:8088`。

---

## 目录结构

```
├── common/           # 公共结构体与函数（错误处理、表单处理）
├── config/           # 配置管理（环境变量、CORS、Session、日志配置）
├── controller/       # HTTP 请求处理层（路由对应的 Handler）
├── logger/           # 日志封装（Gin/GORM/Stderr 日志）
├── middleware/        # 中间件（错误处理、日志、角色权限校验）
├── model/            # 数据模型（GORM 模型定义）
├── pkg/              # 工具包（数据绑定、通用函数）
├── router/           # 路由注册与 Server 初始化
├── service/          # 业务逻辑层
│   └── validator/    # 自定义数据校验器
├── sql/              # 数据库迁移脚本
├── uploads/          # 本地文件上传目录
│   ├── attempts/     # 答题附件
│   ├── avatars/      # 用户头像
│   ├── goods/        # 奖品图片
│   └── photos/       # 图寻图片
├── main.go           # 入口文件
└── go.mod
```

---

## 核心功能模块

### 用户系统
- 通过学校统一认证登录（OAuth 回调）
- 个人信息查看与修改（昵称、头像）
- 三级角色权限：普通用户(1) → 管理员(2) → 超级管理员(3)

### 活动管理
- 当前进行中的活动查询
- 往期活动列表（分页）
- 活动可配置图片奖励积分和答题奖励阶梯

### 图寻题目（Photo）
- 用户上传校园角落照片作为题目
- 支持经纬度坐标标记真实位置
- 图片浏览（流式传输）、下载
- 审核流程：待审核 → 通过/拒绝
- 点赞功能

### 答题挑战（Attempt）
- 用户提交答案（猜测照片拍摄地点）
- 管理员审核答题结果
- 破解成功获得积分奖励（支持多批次阶梯奖励）

### 评论系统
- 对图片发表评论
- 审核流程与点赞功能

### 积分系统
- 积分获取途径：上传照片、答题正确、点赞/被点赞、评论、审核通过、每日登录等
- 完整的积分流水记录

### 奖品兑换
- 奖品管理（CRUD、上架/下架、库存）
- 用户使用积分兑换奖品
- 管理员核销兑换记录

### 消息通知
- 系统消息推送（审核结果、公告等）
- 未读数统计、已读标记

### 反馈系统
- 用户提交反馈意见

---

## 环境变量

环境变量在 `config/config.go` 中定义，支持通过 `.env` 文件配置：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `APP_PROD` | (空) | 设为任意非空值开启生产模式 |
| `APP_SECRET` | `gin-example:secret` | 应用密钥（生产环境务必修改） |
| `APP_LANGUAGE` | `en` | 应用语言 |
| `APP_MYSQL_HOST` | `127.0.0.1` | MySQL 主机 |
| `APP_MYSQL_PORT` | `3306` | MySQL 端口 |
| `APP_MYSQL_NAME` | `static` | 数据库名 |
| `APP_MYSQL_USER` | `root` | 数据库用户 |
| `APP_MYSQL_PASS` | `123456` | 数据库密码 |
| `APP_ALLOW_ORIGINS` | `*` | CORS 允许来源 |
| `APP_ALLOW_HEADERS` | `Origin\|Content-Length\|Content-Type\|Authorization` | CORS 允许头 |
| `APP_LOG_LEVEL` | `info` | 日志级别（debug/info/warn/error） |
| `ONLINE_CALLBACK` | `127.0.0.1:8088` | 统一认证回调地址 |
| `OSS_ACCESS_KEY_ID` | `no` | 阿里云 OSS AccessKey |
| `OSS_ACCESS_KEY_SECRET` | (空) | 阿里云 OSS Secret |
| `OSS_REGION` | `cn-hangzhou` | OSS 区域 |
| `OSS_BUCKET_NAME` | (空) | OSS Bucket 名称 |
| `OSS_USE_LOCAL` | `true` | 是否使用本地存储 |
| `AUTO_APPROVAL` | `no` | 是否自动审核 |

---

## 日志系统

- 四种日志级别：`debug`、`info`、`warn`、`error`
- 生产模式下日志输出到 `log/` 目录，自动滚动切割
- Debug 模式下同步输出到 stdout
- 支持自定义日志钩子（如 TraceHook、RemoteHook）
- 记录内容包含 URL、Method、ClientIP 等信息

---

## 开发规范

- **Controller 层**：处理 HTTP 请求绑定与响应，不包含业务逻辑
- **Service 层**：承载核心业务逻辑
- **Model 层**：定义数据库模型，使用 `BaseModel` 嵌入公共字段
- **Scopes**：通用查询逻辑复用放在 `model/scopes.go`
- 对同一资源的处理方法绑定在同一结构体上，分别注册到 `controller/controller.go` 和 `service/service.go`

---

## API 文档

完整 API 文档见 [api.md](./api.md)。

### 接口概览

| 模块 | 路径前缀 | 说明 |
|------|----------|------|
| 用户认证 | `/api/user` | 登录/登出/个人信息 |
| 活动 | `/api/activity` | 当前活动/历史活动 |
| 图寻题目 | `/api/photos` | CRUD/图片流/点赞/评论/答题 |
| 答题 | `/api/attempts` | 提交答案/点赞 |
| 评论 | `/api/comments` | 发表/删除/点赞 |
| 积分 | `/api/score` | 积分查询/流水 |
| 奖品 | `/api/goods` | 奖品列表/详情 |
| 兑换 | `/api/exchange` | 兑换奖品 |
| 消息 | `/api/messages` | 消息列表/详情/未读数/已读 |
| 反馈 | `/api/feedback` | 提交反馈 |
| 管理员 | `/api/admin` | 审核管理/公告/商品管理/兑换核销 |

---

## License

待定


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