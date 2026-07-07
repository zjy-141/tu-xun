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
| 测试 | `/api/tenzor/tiaozhan/test` | 测试登录 |
| 用户认证 | `/api/user` | 登录/登出/个人信息/修改头像 |
| 活动 | `/api/activity` | 当前活动/历史活动 |
| 图寻题目 | `/api/photos` | CRUD/图片流/点赞/评论/答题 |
| 答题 | `/api/attempts` | 提交答案/点赞 |
| 评论 | `/api/comments` | 发表/删除/点赞 |
| 积分 | `/api/score` | 积分查询/流水 |
| 奖品 | `/api/goods` | 奖品列表/详情 |
| 兑换 | `/api/exchange` | 兑换奖品/兑换记录 |
| 消息 | `/api/messages` | 消息列表/详情/未读数/已读/公告查询 |
| 反馈 | `/api/feedback` | 提交反馈 |
| 管理员 | `/api/admin` | 审核管理/活动管理/公告/商品管理/兑换核销/权限管理 |

---

## License

待定