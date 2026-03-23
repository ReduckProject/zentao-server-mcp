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
      "command": "path/to/zentao-mcp.exe",
      "args": ["-c", "/path/to/zentao_config.json"]
    }
  }
}
```

## 工具列表

### 配置与 Token

| 工具名称 | 描述 |
|---------|------|
| `configure` | 配置禅道服务器连接信息（首次使用必须调用） |
| `get_token` | 获取禅道 API Token（自动处理缓存和刷新） |
| `refresh_token` | 强制刷新 Token（忽略缓存） |
| `token_status` | 查看当前 Token 状态和配置信息 |

### 产品管理

| 工具名称 | 描述 |
|---------|------|
| `get_products` | 获取禅道产品列表 |
| `get_product` | 获取禅道产品详情 |
| `create_product` | 创建禅道产品 |

### Bug 管理

| 工具名称 | 描述 |
|---------|------|
| `create_bug` | 创建 Bug |
| `update_bug` | 修改 Bug |
| `get_bugs` | 获取产品 Bug 列表 |
| `get_bug` | 获取 Bug 详情 |
| `add_bug_comment` | 给 Bug 添加备注 |

### 版本管理

| 工具名称 | 描述 |
|---------|------|
| `create_build` | 创建版本 |
| `update_build` | 修改版本 |
| `get_builds` | 获取项目版本列表 |
| `get_build` | 获取版本详情 |

### 需求管理

| 工具名称 | 描述 |
|---------|------|
| `create_story` | 创建需求 |
| `get_story` | 获取需求详情 |
| `get_project_stories` | 获取项目需求列表 |
| `get_product_stories` | 获取产品需求列表 |
| `get_execution_stories` | 获取执行需求列表 |

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
调用 create_bug 工具:
- title: Bug标题
- severity: 严重程度(1-4)
- pri: 优先级(1-4)
- type: Bug类型
- product_id: 产品ID或名称 (可选，使用默认产品)
```

### 3. 给 Bug 添加备注

```
调用 add_bug_comment 工具:
- bug_id: Bug ID
- comment: 备注内容
```

### 4. 获取产品列表

```
调用 get_products 工具:
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