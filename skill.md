收到！基于你提供的实际代码，我为你量身定制了这份 **User 模块代码风格与架构规范**。

这份规范不再泛泛而谈，而是精确锁定了你当前的代码模式（如 `NetID` 主键逻辑、OAuth 回调流程、Session 覆盖参数等），旨在统一团队协作口径，防止后续开发出现架构偏移。

---

# User 模块代码风格与架构规范（基于当前仓库实际代码）

## 1. 核心数据流与主键逻辑
- **唯一业务主键**：本模块**强依赖 `NetID`（学号）** 作为用户唯一标识，而非自增数字 ID。
- **数据来源权威性**：用户基础信息（姓名、性别、学历等）**以团委 OAuth 回调接口返回为准**。Service 层在登录回调时负责 `CreateOrUpdate`，但**仅更新 `Name`**（其他字段如 `Edulevel` 保持首次入库值不变，如需变更需走团委接口）。

---

## 2. 结构体（Struct）放置与职责铁律

| 结构体名称 | 所属包 | 核心标签约束 | 职责说明 |
| :--- | :--- | :--- | :--- |
| `model.User` | `model` | `gorm` 指定表结构；**`Password` 字段必须带 `json:"-"`** | 数据库表映射。包含 `BeforeCreate` / `BeforeUpdate` 钩子自动处理 `argon2id` 密码哈希。 |
| `service.StudentOauthInfo` | `service` | 包含团委接口返回的所有字段（`json` 标签） | **仅限 Service 层内部使用**，用于反序列化外部 HTTP 响应，严禁透传给 Controller。 |
| `service.UserForm` | `service` | `json` 标签明确返回字段 | **Service 层返回给 Controller 的标准安全视图**。严禁包含 `Password`、`Gender` 等敏感或无需前端感知的字段。 |
| `service.UserUpdateParams` | `service` | `NetID` 标记为 `json:"-"`；`Nickname` 带 `binding` 校验 | Controller 接收更新请求。**`NetID` 禁止由前端传入**，必须在 Controller 内部由 Session 覆盖赋值。 |
| `service.UserUploadAvatar` | `service` | `NetID` 标记为 `form:"-"`；`AvatarFile` 为文件类型 | Controller 接收头像上传。**`NetID` 同样禁止前端传入**，由 Session 覆盖。 |
| `controller.UserSession`（定义在 `init.go` 或 `controller` 包内） | `controller` | 映射 Session 存储字段 | **仅 Controller 层维护**。用于存储登录态（`NetID`, `Username`, `Nickname`, `Level`），Service 层只认 `string netID`。 |

---

## 3. Controller -> Service 交互规范（重点）

### 3.1 会话（Session）管理权限
- **读写权**：`SessionGet` / `SessionSet` / `SessionClear` **仅允许在 Controller 层调用**。
- **参数传递**：Controller 从 Session 中提取 `NetID` 后，**必须显式赋值给 DTO 中标记为 `-` 的字段**（如 `params.NetID = session.NetID`），再调用 Service。这能有效防止越权操作（用户无法通过篡改请求体修改他人信息）。

### 3.2 标准调用链路示例（更新昵称）
```go
// 1. Controller 绑定请求体（前端仅传 nickname）
var params service.UserUpdateParams
c.ShouldBind(&params)

// 2. Controller 强制注入身份（覆盖）
params.NetID = SessionGet(c, "user-session").(UserSession).NetID

// 3. 调用 Service（Service 完全信任传入的 NetID，只做业务逻辑）
err := srv.User.UserInfoUpdate(params)

// 4. 返回标准响应
c.JSON(http.StatusOK, ResponseNew(c, nil))
```

### 3.3 Service 层返回值约束
- **禁止返回 `model.User`**。所有数据返回给 Controller 前，**必须**转换为 自定义的 ** 结构体 **。
- **错误返回**：Service 返回 `error` 原始类型（如 `errors.New("...")` 或 `gorm.ErrRecordNotFound`），由 Controller 统一使用 `common.ErrNew` 进行包装和状态码映射。

---

## 4. 特殊安全机制与校验

### 4.1 密码处理（Model 钩子）
- **明文接收**：Controller 和 Service 层**从不**处理密码明文业务逻辑。
- **自动哈希**：`model.User` 的 `BeforeCreate` 和 `BeforeUpdate` 钩子会检测 `Password` 字段是否被赋值，若不为空则自动调用 `argon2id` 哈希入库。
- **校验逻辑**：验证密码是否正确时，使用 `model.User` 对象的方法 `CheckPassword(plaintext string) bool`，严禁在 Service 层手写哈希比对算法。

### 4.2 输入清洗
- **空值处理**：`BeforeCreate` 钩子中，若 `Nickname` 为空，自动回退为 `Name`。
- **边界裁剪**：Service 层 `UserInfoUpdate` 中对 `Nickname` 强制使用 `strings.TrimSpace`，去除首尾空格。

### 4.3 防越权设计
- 凡是涉及 `NetID` 作为查询/更新条件的 Service 方法，其 `NetID` 参数**必须来源于 Controller 的 Session 注入**，而非前端请求体（因此 DTO 中该字段标记为 `json:"-"`）。

---

## 5. 错误处理与日志规范
- **日志打印**：仅在**可能出现网络波动或外部依赖失败**的地方（如 `UploadAvatar` 调用 OSS 失败）使用 `logger.Errorf`。普通业务逻辑错误（如用户不存在）**不应**打印 Error 日志，直接返回 Error 给 Controller 处理即可。
- **统一响应**：Controller 所有正常返回必须使用 `ResponseNew(c, data)` 封装，错误返回必须使用 `c.Error(common.ErrNew(...))`。

---

## 6. 命名约定与新功能扩展清单

### 6.1 命名规范
- **接收器命名**：Controller 接收器统一为 `u`（如 `func (u *User) UserInfo`），Service 接收器统一为 `u`（如 `func (u *User) LoginCallback`）。
- **方法前缀**：涉及 OAuth 流程的用 `Login` / `Callback`；涉及普通数据操作用 `UserInfo` / `UserInfoUpdate`。

### 6.2 新增接口操作清单（必须按顺序执行）
1. **定义 DTO**：在 `service/struct.go` 中定义请求参数结构体（若涉及文件上传用 `form` 标签，普通 JSON 用 `json` 标签，`NetID` 一律标记为 `-`）。
2. **实现业务**：在 `service/user.go` 中实现业务方法（参数使用第 1 步的 DTO，返回 `UserForm` 或自定义结构）。
3. **编写控制器**：在 `controller/user.go` 中添加方法，绑定参数、注入 `NetID`、调用 `srv.User.新方法`。
4. **注册路由**：在 `router/router.go` 中添加对应路由。


---
