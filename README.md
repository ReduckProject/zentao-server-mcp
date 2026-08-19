# Zentao MCP Server

禅道 MCP (Model Context Protocol) 服务器，为 AI 助手提供禅道 API 接口调用能力。

## 功能特性

- **Token 管理**: 自动获取和刷新 Token，支持过期时间配置（秒）
- **产品管理**: 获取产品列表、产品详情、创建产品
- **Bug 管理**: 创建 Bug、修改 Bug、获取 Bug 列表和详情、添加备注
- **版本管理**: 创建版本、修改版本、获取版本列表和详情
- **需求管理**: 创建需求、获取需求列表和详情
- **默认产品配置**: 支持配置默认产品（ID或名称），简化操作
- **精简输出**: 列表接口默认返回精简参数，可选返回完整数据
- **灵活配置**: 支持命令行参数指定配置文件路径

## 安装

```bash
go build -o zentao-mcp .
```

## 使用

```bash
# 使用默认配置文件（exe所在目录的zentao_config.json）
./zentao-mcp

# 指定配置文件路径
./zentao-mcp -c /path/to/config.json
# 或
./zentao-mcp -config /path/to/config.json
```

## 配置

### Claude Desktop 配置

在 Claude Desktop 配置文件中添加：

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "zentao": {
      "command": "path/to/zentao-mcp.exe"
    }
  }
}
```

指定配置文件：

```json
{
  "mcpServers": {
    "zentao": {
      "command": "zentao-mcp.exe",
      "args": ["-c", "/path/to/zentao_config.json"]
    }
  }
}
```

## 工具列表

公开工具已按领域从 24 个精简为 7 个。同一领域通过必填的 `action` 区分操作。

| 工具名称 | action | 功能 |
|---------|--------|------|
| `configure` | 无 | 配置禅道服务器连接信息 |
| `zentao_user` | `profile`, `dynamic` | 当前用户、连接状态和动态 |
| `zentao_products` | `list`, `get`, `create` | 产品查询与创建 |
| `zentao_bugs` | `list`, `get`, `create`, `update`, `list_comments`, `add_comment` | Bug 与备注管理 |
| `zentao_builds` | `list`, `get`, `create`, `update` | 版本管理 |
| `zentao_stories` | `list_product`, `list_project`, `list_execution`, `get`, `create` | 需求管理 |
| `zentao_testcases` | `list`, `get`, `create` | 测试用例管理 |

每个 action 都有独立的 JSON Schema：只允许该 action 对应的字段，并约束必填项、类型、枚举、数值范围、日期格式和数组结构。服务端还会再次执行同样的运行时校验，因此多传字段或错误类型不会发送到禅道 API。

## 使用示例

### 1. 首次配置

```
调用 configure 工具:
- base_url: http://your-zentao-server/zentao/api.php/v1 (禅道API地址)
- account: your_account
- password: your_password
- token_expiry: 86400 (可选，默认86400秒即24小时)
- default_product: 产品名称或ID (可选，设置默认产品)
```

### 2. 创建 Bug

```
调用 zentao_bugs 工具:
- action: create
- title: Bug标题
- severity: 严重程度(1-4)
- pri: 优先级(1-4)
- type: Bug类型
- product_id: 产品ID或名称 (可选，使用默认产品)
```

### 3. 给 Bug 添加备注

```
调用 zentao_bugs 工具:
- action: add_comment
- bug_id: Bug ID
- comment: 备注内容
- image_paths: 本地图片路径数组（可选，最多10张，单张不超过20MiB）
```

示例:
```json
{
  "action": "add_comment",
  "bug_id": 35,
  "comment": "问题截图如下：",
  "image_paths": ["C:/temp/error.png"]
}
```

图片会先上传到禅道，再以富文本图片插入备注。`zentao_bugs` 的 `list_comments` action 返回的每条备注包含纯文本 `content`、原始富文本 `html` 和图片数组 `images`。

### 4. 获取 Bug 列表

```
调用 zentao_bugs 工具:
- action: list
- product_id: 产品ID或名称 (可选，使用默认产品)
- status: Bug状态 (可选)
  - assigntome: 我的bug
  - all: 所有包括关闭的
  - 不传: 默认未关闭的
- limit: 每页数量 (可选)
- page: 页码 (可选)
- full: true/false (是否返回完整参数)
```

### 5. 获取 Bug 备注

```
调用 zentao_bugs 工具:
- action: list_comments
- bug_id: Bug ID
```

### 6. 获取产品列表

```
调用 zentao_products 工具:
- action: list
- full: true/false (是否返回完整参数)
```

## 默认产品功能

支持配置默认产品，简化操作：

1. 配置时指定 `default_product` 参数（可以是产品ID或名称）
2. 创建 Bug、需求时可不传 `product_id`，自动使用默认产品
3. 获取产品相关列表时可不传 `product_id`

## 数据结构

### Bug 类型

| 类型 | 说明 |
|-----|------|
| `codeerror` | 代码错误 |
| `config` | 配置相关 |
| `install` | 安装部署 |
| `security` | 安全相关 |
| `performance` | 性能问题 |
| `standard` | 标准规范 |
| `automation` | 测试脚本 |
| `designdefect` | 设计缺陷 |
| `others` | 其他 |

### 需求类型

| 类型 | 说明 |
|-----|------|
| `feature` | 功能 |
| `interface` | 接口 |
| `performance` | 性能 |
| `safe` | 安全 |
| `experience` | 体验 |
| `improve` | 改进 |
| `other` | 其他 |

### 测试用例类型

| 类型 | 说明 |
|-----|------|
| `feature` | 功能测试 |
| `performance` | 性能测试 |
| `config` | 配置相关 |
| `install` | 安装部署 |
| `security` | 安全相关 |
| `interface` | 接口测试 |
| `unit` | 单元测试 |
| `other` | 其他 |

### 测试用例适用阶段

| 阶段 | 说明 |
|-----|------|
| `unittest` | 单元测试阶段 |
| `feature` | 功能测试阶段 |
| `intergrate` | 集成测试阶段 |
| `system` | 系统测试阶段 |
| `smoke` | 冒烟测试阶段 |
| `bvt` | 版本验证阶段 |

## 开发

### 依赖

- Go 1.21+
- [mcp-go](https://github.com/mark3labs/mcp-go)

### 构建

```bash
go mod tidy
go build -o zentao-mcp .
```

## 许可证

MIT License
