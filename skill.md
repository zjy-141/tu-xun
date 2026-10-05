# User 模块代码风格与架构规范（以当前仓库实际代码为准）

> 本文档描述的是**当前仓库的真实实现**，不是目标设计。改动代码前请先比对本文档；
> 若实现与文档冲突，以代码为准并同步更新本文档。
>
> 上一版本文档描述的是「学号 + 密码 + argon2id」的旧方案，与本仓库现状不符，
> 已整体重写。**不要再参考旧的 `Password` / `Gender` / `Edulevel` / `UserForm` 等描述**
> ——这些字段和结构体在当前代码中并不存在（见文末「已废弃内容」）。

---

## 1. 核心数据流与主键逻辑

- **主键是自增 `ID`**：`model.BaseModel` 提供 `ID int64`（`primaryKey`），
  所有关联外键（`UserID` 等）指向它。
- **`NetID`（学号）是唯一业务标识**：`netid` 是 `VARCHAR(32) NOT NULL` 且带
  `uniqueIndex:idx_user_netid`，但**不是主键**。
- **认证方式为 OAuth（团委统一认证），系统内无密码**：
  `LoginCallback` → `code` 换 token → `Introspect` 校验 → `FetchUserinfo` 取用户信息 →
  按 `netid` 落库/读取 → 写入 Session。`model.User` **没有** `Password` 字段，
  **没有** argon2id 哈希钩子。
- **数据来源权威性**：用户基础信息以团委 OAuth 返回为准；本地仅维护
  `Nickname` / `AvatarURL` 等可编辑字段。

---

## 2. 结构体（Struct）放置与职责

| 结构体 | 所属包 | 关键标签约束 | 职责 |
| :--- | :--- | :--- | :--- |
| `model.User` | `model` | `gorm` 定义表结构；`netid` 带 `uniqueIndex` | 数据库表映射。`BeforeCreate` 校验学号非空并在 `Nickname` 为空时回退为 `Name` |
| `service.StudentOauthInfo` | `service` | `json` 标签对齐团委接口字段 | **仅 Service 层内部使用**，反序列化外部 HTTP 响应；**严禁透传给 Controller** |
| `service.UserSummary` | `service` | `json` 标签即对外返回字段 | **Service 返回给 Controller 的标准安全视图**。含 `ID`/`NetID`/`Username`/`Nickname`/`Avatar`/`Level`/`ScoreCount`/`Status` 及剩余修改次数 |
| `service.LoginResult` | `service` | 内嵌 `UserSummary` + `SessionID` | 登录成功响应 |
| `service.UpdateNicknameParams` | `service` | `ID` 标记 `json:"-"`；`Nickname` 带 `binding` | Controller 接收昵称更新。**`ID` 禁止前端传入** |
| `service.UploadAvatarParams` | `service` | `ID` 标记 `form:"-"`；`AvatarFile` 为文件 | Controller 接收头像上传。**`ID` 禁止前端传入** |
| `controller.UserSession` | `controller` | 会话存储结构 | **仅 Controller 层维护**，存 `ID`/`NetID`/`Username`/`Nickname`/`Level`。Service 只接受 `id int64` |

> **主键注入的正确姿势**：Controller 从 Session 取值后赋值给 DTO 中标记为 `-` 的字段
> （`params.ID = SessionGet(c, "user-session").(UserSession).ID`），再调用 Service。

---

## 3. Controller → Service 交互规范

### 3.1 会话（Session）管理权限

- **读写权**：`SessionGet` / `SessionSet` / `SessionClear` / `SessionDelete` / `SessionUpdate`
  **仅允许在 Controller 层调用**。`service` 包内不得出现这些调用（现状为零，请保持）。
- **身份注入**：Controller 从 Session 提取 `ID` 后，**必须显式赋值**给 DTO 中标 `-` 的字段，
  Service 完全信任传入的 `ID`。这能防止用户篡改请求体操作他人数据。
- **非 Cookie 宿主（如小程序）**：由 `middleware.XSessionID()` 读取 `X-Session-Id` 请求头，
  经 `controller.GetXSession` 查服务端映射后重建会话，后续 `CheckRole` 可正常读取。

### 3.2 标准调用链路示例（更新昵称）

```go
// 1. Controller 绑定请求体（前端仅传 nickname）
var params service.UpdateNicknameParams
if err := c.ShouldBind(&params); err != nil {
    c.Error(common.ErrNew(err, common.ParamErr))
    return
}

// 2. Controller 强制注入身份（覆盖）
params.ID = SessionGet(c, "user-session").(UserSession).ID

// 3. 调用 Service（Service 完全信任传入的 ID，只做业务逻辑）
resp, err := srv.UserSvc.UpdateNickname(params)
if err != nil {
    c.Error(err)
    return
}

// 4. 返回标准响应
c.JSON(http.StatusOK, ResponseNew(c, resp))
```

### 3.3 Service 层返回值与错误约束（**重点，易错**）

**（1）所有被 Controller 直接调用的 Service 方法一律使用具名返回值：**

```go
// 有业务返回值的
func (u *UserSvc) UpdateNickname(info UpdateNicknameParams) (resp UpdateNicknameResponse, err error)

// 无业务返回值的（只有 error）
func (m *MessageSvc) MarkRead(userID int64, id int64) (err error)

// 返回指针 / 基础类型时，具名参数用类型名，不要硬套 resp
// （因为 `return nil, ...` 这类语句要求具名参数可赋 nil）
func (s *AnnouncementSvc) GetByID(userID int64, id int64) (detail *AnnouncementDetail, err error)
```

**（2）错误包装职责在 Service 层，Controller 不再二次包装：**

```go
// ✅ 正确：来自 gorm / 标准库 / helper 的错误，在 Service 就地包装
if err := model.DB.Where("id = ?", id).First(&user).Error; err != nil {
    return resp, common.ErrNew(err, common.SysErr)
}

// ✅ 正确：业务错误用 errors.New，按语义选类型
if photo.Status != "approved" {
    return resp, common.ErrNew(errors.New("该图片尚未通过审核，暂不可答题"), common.OpErr)
}

// ✅ 正确：错误已经由另一个 Service 方法包装过 → 直传，不要二次包装
if _, err := scoreSvc.RegularScoreChange(tx, scoreParams); err != nil {
    tx.Rollback()
    return err
}
```

- **必须包装**：`gorm` 错误、标准库错误、包内 helper 返回的错误。
- **直接返回**：由其他 Service 导出方法返回的错误（它已在源头包装好）。
  二次包装会得到 `系统错误: 系统错误: xxx` 这样的重复前缀，污染日志与响应。
- **错误类型选择**（`common/error.go` 的 `ErrorType` 与 HTTP 状态码映射）：

  | 类型 | HTTP | 适用场景 |
  | :--- | :--- | :--- |
  | `common.ParamErr` | 400 | 参数缺失/非法 |
  | `common.OpErr` | 400 | 业务操作不允许（如状态不符） |
  | `common.AuthErr` | 401 | 未登录、身份无效 |
  | `common.LevelErr` | 403 | 权限不足 |
  | `common.ConflictErr` | 409 | 冲突（如重复审核） |
  | `common.RateLimitErr` | 429 | 频率限制 |
  | `common.SysErr` | 200（现状，见下） | 系统内部错误 |

  > **注意**：`common.HTTPStatus` 目前对 `SysErr` 落到 `default` 分支返回 **200**，
  > 只在 body 里用 `code` 表示错误。这与「只有 5xx 才记 Error 级」的日志分级不一致，
  > 属于已知问题；改动它属于**破坏性契约变更**（`api.md` 明确要求先沟通），
  > 需与前端确认后再动。

- **禁止返回 `model.User` 等 ORM 实体**：所有返回给 Controller 的数据必须先转成
  `service` 包下的视图结构体（如 `UserSummary`）。

**（3）本规范覆盖范围**：所有**被 Controller 直接调用**的 Service 方法
（当前 58 个，分布于 16 个文件）。仅被 Service 内部互相调用的方法
（如 `RegularScoreChange`、`GetUserRank`、`IsActivityActive`）保持相同的错误包装风格，
但不强制具名返回。

---

## 4. 特殊安全机制与校验

### 4.1 认证与令牌校验

- **明文密码不存在**：不要新增 `Password` 字段或哈希逻辑；认证由团委 OAuth 承担。
- **登录链路**：`ExchangeCode`（code→access token）→ `Introspect`（校验 token 有效）→
  `FetchUserinfo`（取用户信息）。任一步失败即拒绝建立会话。
- **回调白名单**：`LoginCallback` 必须校验 `redirect_uri` 在 `FE_ORIGIN` / `ADMIN_ORIGIN`
  白名单内。
- **已知缺口**：H5 回调**未校验 `state`**（`service.LoginCallbackParams` 中 `State` 被注释掉），
  存在登录 CSRF 隐患；小程序侧 `state` 机制完整（`GenerateState` + 服务端存储 + 一次性消费 +
  15min TTL）。修复需前端配合。

### 4.2 输入清洗

- **空值处理**：`model.User.BeforeCreate` 中学号为空直接报错；`Nickname` 为空自动回退为 `Name`。
- **边界裁剪**：`UserSvc.UpdateNickname` 对昵称强制 `strings.TrimSpace` 去除首尾空格。

### 4.3 防越权设计

- 凡以 `ID` 作为查询/更新条件的 Service 方法，其 `ID` **必须来源于 Controller 的 Session 注入**，
  而非前端请求体（因此 DTO 中该字段标记为 `json:"-"` / `form:"-"`）。

---

## 5. 日志与错误响应规范

### 5.1 日志

日志按日期滚动写入**运行目录下的 `./log/`**，共三个 logrus 实例
（`config/logconf.go` 的 `initLogger`），均由 lumberjack 切割
（单文件 512MB、保留 5 个备份、7 天、压缩）：

| 文件 | 内容 |
| :--- | :--- |
| `gin.<日期>.log` | 访问日志与 panic 恢复（`middleware.GinLogger` / `GinRecovery`） |
| `database.<日期>.log` | GORM 执行的 SQL |
| `stderr.<日期>.log` | 进程 stderr 重定向捕获（`logger.RedirectStderr`） |

> **两个易踩的坑**：
> 1. 路径是**相对当前工作目录**的 `./log`，不是相对二进制所在目录。
>    从不同工作目录启动会产生多个 `log/` 文件夹；生产容器内为 `/runtime/log`，
>    由 compose 挂载到宿主 `/tz-master-data/app_log/<APP_NAME>`。
> 2. `logger.Out` 在 `AppMode == "debug"` 时会同时写 stdout，生产（`APP_PROD` 非空）只写文件。
>
> 另外注意：项目使用了 `gin.New()` 而非 `gin.Default()`，因为
> `middleware.GinLogger()` / `GinRecovery(true)` 已提供接入 logrus 的实现；
> 若改回 `gin.Default()` 会导致访问日志重复、Recovery 互相覆盖。

**打印原则**：

- 仅在**外部依赖可能失败**处打 Error 级日志（如 `service/oss.go` 中 OSS 上传失败）。
- **普通业务错误不应打 Error 级**（如「图片不存在」「次数已达上限」），
  直接 `return common.ErrNew(...)` 交给中间件与响应层即可。
- **现状偏差（待清理）**：`controller/` 下当前有 29 处生效的 `logger.Errorf`，其中包含大量
  参数绑定失败与业务错误（如 `controller attempt submit: %v`）。这会把业务错误记成
  Error 级、污染日志分级与告警。**后续新增代码请遵守上面的原则，存量待逐步清理。**

### 5.2 统一响应

- Controller 正常返回一律 `c.JSON(status, ResponseNew(c, data))`。
- Controller 错误返回一律 `c.Error(err)`，其中 `err` 应为 Service 已包装的错误
  （或 Controller 自行产生的 `common.ErrNew(err, common.ParamErr)`），
  由 `middleware.Error` + `common.HTTPStatus` 统一映射状态码与响应体。

---

## 6. 命名约定与新功能扩展清单

### 6.1 命名规范

- **Service 接收器**：按服务类型命名，如 `func (u *UserSvc) UserInfo(...)`、
  `func (a *AttemptSvc) Create(...)`。**不要**再使用旧的 `func (u *User) ...` 写法。
- **Controller 接收器**：按业务实体命名，如 `func (u *User) UserInfo(c *gin.Context)`、
  `func (ctr *Admin) ...`（历史遗留两种风格并存，新代码请与同文件保持一致）。
- **方法命名**：OAuth 流程用 `LoginCallback` / `ExchangeCode` / `Introspect` / `FetchUserinfo`；
  普通数据操作用 `UserInfo` / `UpdateNickname` / `UploadAvatar`。

### 6.2 新增接口操作清单（必须按顺序执行）

1. **定义 DTO**：在 `service/struct.go` 中定义请求参数结构体，涉及文件上传用 `form` 标签、
   普通 JSON 用 `json` 标签；**身份字段一律标记为 `-`**（`json:"-"` 或 `form:"-"`）。
2. **实现业务**：在对应 `service/*.go` 中实现方法，使用**具名返回**，
   并按 §3.3 规则包装错误。
3. **编写控制器**：在 `controller/*.go` 中添加方法 → `ShouldBind` → 注入 Session 身份 →
   调用 Service → `ResponseNew` 返回；错误 `c.Error(err)`。
4. **注册路由**：在 `router/router.go` 中添加路由，注意分组级的 `middleware.CheckRole(n)`。
5. **同步契约**：若涉及接口增删改或字段/枚举/状态码变化，必须同步
   `api.md` 与 `apifox-import.json`，并运行 `python3 docs/check_contract.py` 校验一致性。

---

## 7. 已废弃内容（旧版文档残留，勿再遵循）

以下描述来自上一版规范，**与当前代码不符**，已在本文档中移除对应要求：

| 旧文档要求 | 当前实际 |
| :--- | :--- |
| `model.User` 必须有 `Password` 字段且带 `json:"-"`，由 `BeforeCreate`/`BeforeUpdate` 自动 argon2id 哈希 | **无 Password 字段、无哈希钩子**；认证走 OAuth |
| 用 `CheckPassword(plaintext) bool` 校验密码 | 不存在该方法 |
| `NetID` 为唯一业务主键（替代自增 ID） | 主键是自增 `ID`；`netid` 仅为 `uniqueIndex` |
| 结构体 `service.UserForm` / `service.UserUpdateParams` / `service.UserUploadAvatar` | 实际为 `UserSummary` / `UpdateNicknameParams` / `UploadAvatarParams` |
| `Edulevel`（学历）、`Gender`（性别）字段 | 两字段已注释移除，表中不存在 |
| Service 返回**原始 error**，由 Controller 统一 `common.ErrNew` 包装 | **相反**：Service 负责包装，Controller 直传（见 §3.3） |
| Service 接收器统一为 `u`（`func (u *User) ...`） | 接收器为 `UserSvc` 等服务类型 |
| `service.User.UserInfoUpdate(params)` | 实际为 `srv.UserSvc.UpdateNickname(params)` |

> 历史遗留且**尚未清理**的项：`service/struct.go` 中的 `TestLoginParams` 与
> `controller/test.go` 的测试登录接口、`model/user.go` 中被注释掉的字段、
> `pkg/sensitive` 的 fail-open 分支等，均属待整改项，另行跟踪。
