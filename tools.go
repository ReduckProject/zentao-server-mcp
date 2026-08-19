package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type parameterSpec struct {
	schema    map[string]interface{}
	normalize func(interface{}) (interface{}, error)
}

type strictToolDefinition struct {
	name        string
	description string
	params      map[string]parameterSpec
	required    []string
	handler     server.ToolHandlerFunc
}

type actionDefinition struct {
	name        string
	description string
	params      map[string]parameterSpec
	required    []string
	handler     server.ToolHandlerFunc
	mapArgs     func(map[string]interface{}) map[string]interface{}
}

type registeredTool struct {
	tool    mcp.Tool
	handler server.ToolHandlerFunc
}

func registerTools(s *server.MCPServer) {
	for _, item := range consolidatedTools() {
		s.AddTool(item.tool, item.handler)
	}
}

func consolidatedTools() []registeredTool {
	configure := strictToolDefinition{
		name:        "configure",
		description: "配置禅道服务器连接信息。仅首次使用或修改连接时调用。",
		params: map[string]parameterSpec{
			"base_url":        urlParam("禅道 REST API 地址，例如 http://localhost:8080/zentao/api.php/v1"),
			"account":         requiredStringParam("登录账号"),
			"password":        requiredStringParam("登录密码"),
			"token_expiry":    integerParam("Token 过期时间（秒），默认 86400", 60, 31536000),
			"default_product": productRefParam("默认产品 ID 或名称"),
		},
		required: []string{"base_url", "account", "password"},
		handler:  configureHandler,
	}

	tools := []registeredTool{{
		tool:    newStrictTool(configure),
		handler: newStrictToolHandler(configure),
	}}

	domains := []struct {
		name        string
		description string
		actions     []actionDefinition
	}{
		{
			name:        "zentao_user",
			description: "用户与动态查询。action 只能是 profile 或 dynamic，并且只传该 action 对应的参数。",
			actions: []actionDefinition{
				{name: "profile", description: "获取当前登录用户信息和脱敏连接状态", handler: getProfileHandler},
				{
					name:        "dynamic",
					description: "获取当前用户动态",
					params: map[string]parameterSpec{
						"time_range": enumStringParam("时间范围，默认 today", "today", "yesterday", "thisWeek", "lastWeek", "thisMonth", "lastMonth"),
					},
					handler: getTodayDynamicHandler,
				},
			},
		},
		{
			name:        "zentao_products",
			description: "产品管理。action 只能是 list、get 或 create，并且只传该 action 对应的参数。",
			actions: []actionDefinition{
				{
					name:        "list",
					description: "获取产品列表",
					params:      map[string]parameterSpec{"full": booleanParam("是否返回完整字段，默认 false")},
					handler:     getProductsHandler,
				},
				{
					name:        "get",
					description: "获取产品详情",
					params:      map[string]parameterSpec{"id": idStringParam("产品 ID")},
					required:    []string{"id"},
					handler:     getProductHandler,
				},
				{
					name:        "create",
					description: "创建产品",
					params: map[string]parameterSpec{
						"name":    requiredStringParam("产品名称"),
						"code":    requiredStringParam("产品代号"),
						"program": nonNegativeIntegerParam("所属项目集 ID；0 表示不指定"),
						"line":    nonNegativeIntegerParam("所属产品线 ID；0 表示不指定"),
						"PO":      optionalStringParam("产品负责人账号"),
						"QD":      optionalStringParam("测试负责人账号"),
						"RD":      optionalStringParam("发布负责人账号"),
						"type":    enumStringParam("产品类型", "normal", "branch", "platform"),
						"desc":    optionalStringParam("产品描述"),
						"acl":     enumStringParam("访问控制", "open", "private"),
					},
					required: []string{"name", "code"},
					handler:  createProductHandler,
				},
			},
		},
		{
			name:        "zentao_bugs",
			description: "Bug 管理。action 只能是 list、get、create、update、list_comments 或 add_comment，并且只传该 action 对应的参数。",
			actions:     bugActions(),
		},
		{
			name:        "zentao_builds",
			description: "版本管理。action 只能是 list、get、create 或 update，并且只传该 action 对应的参数。",
			actions:     buildActions(),
		},
		{
			name:        "zentao_stories",
			description: "需求管理。action 只能是 list_product、list_project、list_execution、get 或 create，并且只传该 action 对应的参数。",
			actions:     storyActions(),
		},
		{
			name:        "zentao_testcases",
			description: "测试用例管理。action 只能是 list、get 或 create，并且只传该 action 对应的参数。",
			actions:     testCaseActions(),
		},
	}

	for _, domain := range domains {
		tools = append(tools, registeredTool{
			tool:    newActionTool(domain.name, domain.description, domain.actions),
			handler: newActionToolHandler(domain.name, domain.actions),
		})
	}
	return tools
}

func bugActions() []actionDefinition {
	common := func() map[string]parameterSpec {
		return map[string]parameterSpec{
			"product_id":   productRefParam("产品 ID 或名称；不传则使用默认产品"),
			"title":        requiredStringParam("Bug 标题"),
			"severity":     integerParam("严重程度 1-4", 1, 4),
			"pri":          integerParam("优先级 1-4", 1, 4),
			"type":         enumStringParam("Bug 类型", "codeerror", "config", "install", "security", "performance", "standard", "automation", "designdefect", "others"),
			"steps":        optionalStringParam("重现步骤"),
			"keywords":     optionalStringParam("关键词"),
			"branch":       nonNegativeIntegerParam("所属分支 ID；0 表示主干"),
			"module":       nonNegativeIntegerParam("所属模块 ID；0 表示不指定"),
			"execution":    nonNegativeIntegerParam("所属执行 ID；0 表示不指定"),
			"os":           optionalStringParam("操作系统"),
			"browser":      optionalStringParam("浏览器"),
			"task":         nonNegativeIntegerParam("相关任务 ID；0 表示不指定"),
			"story":        nonNegativeIntegerParam("相关需求 ID；0 表示不指定"),
			"deadline":     dateParam("截止日期，格式 YYYY-MM-DD"),
			"opened_build": stringArrayParam("影响版本数组，例如 [\"trunk\"]", 1, 100),
		}
	}

	createParams := common()
	updateParams := common()
	updateParams["id"] = idStringParam("Bug ID")

	return []actionDefinition{
		{
			name:        "list",
			description: "获取产品 Bug 列表",
			params: map[string]parameterSpec{
				"product_id": productRefParam("产品 ID 或名称；不传则使用默认产品"),
				"status":     enumStringParam("Bug 状态；不传时返回未关闭 Bug", "assigntome", "all"),
				"limit":      integerParam("每页数量，默认 20", 1, 100),
				"page":       integerParam("页码，从 1 开始", 1, 1000000),
				"full":       booleanParam("是否返回完整字段，默认 false"),
			},
			handler: getBugsHandler,
		},
		{
			name:        "get",
			description: "获取 Bug 详情",
			params:      map[string]parameterSpec{"id": idStringParam("Bug ID")},
			required:    []string{"id"},
			handler:     getBugHandler,
		},
		{
			name:        "create",
			description: "创建 Bug",
			params:      createParams,
			required:    []string{"title", "severity", "pri", "type"},
			handler:     createBugHandler,
		},
		{
			name:        "update",
			description: "修改 Bug",
			params:      updateParams,
			required:    []string{"id", "title", "severity", "pri", "type"},
			handler:     updateBugHandler,
		},
		{
			name:        "list_comments",
			description: "获取 Bug 备注列表",
			params:      map[string]parameterSpec{"bug_id": idStringParam("Bug ID")},
			required:    []string{"bug_id"},
			handler:     getBugCommentsHandler,
		},
		{
			name:        "add_comment",
			description: "给 Bug 添加备注，可上传本地图片",
			params: map[string]parameterSpec{
				"bug_id":      idStringParam("Bug ID"),
				"comment":     requiredStringParam("备注内容"),
				"image_paths": stringArrayParam("本地图片路径数组，最多 10 张", 1, 10),
			},
			required: []string{"bug_id", "comment"},
			handler:  addBugCommentHandler,
		},
	}
}

func buildActions() []actionDefinition {
	common := func() map[string]parameterSpec {
		return map[string]parameterSpec{
			"name":      requiredStringParam("版本名称"),
			"product":   positiveIntegerParam("所属产品 ID"),
			"execution": positiveIntegerParam("所属执行 ID"),
			"builder":   requiredStringParam("构建者账号"),
			"branch":    nonNegativeIntegerParam("所属分支 ID；0 表示主干"),
			"date":      dateParam("打包日期，格式 YYYY-MM-DD"),
			"scm_path":  optionalStringParam("源代码地址"),
			"file_path": optionalStringParam("下载地址"),
			"desc":      optionalStringParam("版本描述"),
		}
	}
	createParams := common()
	createParams["project_id"] = idStringParam("项目 ID")
	updateParams := common()
	updateParams["id"] = idStringParam("版本 ID")

	return []actionDefinition{
		{
			name:        "list",
			description: "获取项目版本列表",
			params: map[string]parameterSpec{
				"project_id": idStringParam("项目 ID"),
				"full":       booleanParam("是否返回完整字段，默认 false"),
			},
			required: []string{"project_id"},
			handler:  getBuildsHandler,
		},
		{
			name:        "get",
			description: "获取版本详情",
			params:      map[string]parameterSpec{"id": idStringParam("版本 ID")},
			required:    []string{"id"},
			handler:     getBuildHandler,
		},
		{
			name:        "create",
			description: "创建版本",
			params:      createParams,
			required:    []string{"project_id", "name", "product", "execution", "builder"},
			handler:     createBuildHandler,
		},
		{
			name:        "update",
			description: "修改版本",
			params:      updateParams,
			required:    []string{"id", "name", "product", "execution", "builder"},
			handler:     updateBuildHandler,
		},
	}
}

func storyActions() []actionDefinition {
	return []actionDefinition{
		{
			name:        "list_product",
			description: "获取产品需求列表",
			params: map[string]parameterSpec{
				"product_id": productRefParam("产品 ID 或名称；不传则使用默认产品"),
				"full":       booleanParam("是否返回完整字段，默认 false"),
			},
			handler: getProductStoriesHandler,
		},
		{
			name:        "list_project",
			description: "获取项目需求列表",
			params: map[string]parameterSpec{
				"project_id": idStringParam("项目 ID"),
				"full":       booleanParam("是否返回完整字段，默认 false"),
			},
			required: []string{"project_id"},
			handler:  getProjectStoriesHandler,
		},
		{
			name:        "list_execution",
			description: "获取执行需求列表",
			params: map[string]parameterSpec{
				"execution_id": idStringParam("执行 ID"),
				"full":         booleanParam("是否返回完整字段，默认 false"),
			},
			required: []string{"execution_id"},
			handler:  getExecutionStoriesHandler,
		},
		{
			name:        "get",
			description: "获取需求详情",
			params:      map[string]parameterSpec{"id": idStringParam("需求 ID")},
			required:    []string{"id"},
			handler:     getStoryHandler,
		},
		{
			name:        "create",
			description: "创建需求",
			params: map[string]parameterSpec{
				"title":       requiredStringParam("需求标题"),
				"product_id":  productRefParam("产品 ID 或名称；不传则使用默认产品"),
				"pri":         integerParam("优先级 1-4", 1, 4),
				"category":    enumStringParam("需求类型", "feature", "interface", "performance", "safe", "experience", "improve", "other"),
				"spec":        optionalStringParam("需求描述"),
				"verify":      optionalStringParam("验收标准"),
				"source":      enumStringParam("需求来源", "customer", "user", "po", "market"),
				"source_note": optionalStringParam("来源备注"),
				"estimate":    numberParam("预计工时，不得为负数", 0, 1000000),
				"keywords":    optionalStringParam("关键词"),
				"reviewer":    optionalStringParam("评审人账号"),
				"assigned_to": optionalStringParam("指派给账号"),
			},
			required: []string{"title", "pri", "category"},
			handler:  createStoryHandler,
			mapArgs: func(args map[string]interface{}) map[string]interface{} {
				if productID, exists := args["product_id"]; exists {
					args["product"] = productID
					delete(args, "product_id")
				}
				return args
			},
		},
	}
}

func testCaseActions() []actionDefinition {
	return []actionDefinition{
		{
			name:        "list",
			description: "获取产品用例列表",
			params: map[string]parameterSpec{
				"product_id": productRefParam("产品 ID 或名称；不传则使用默认产品"),
				"search":     optionalStringParam("按用例标题或关键词搜索"),
				"limit":      integerParam("每页数量，默认 20", 1, 100),
				"page":       integerParam("页码，从 1 开始", 1, 1000000),
				"full":       booleanParam("是否返回完整字段，默认 false"),
			},
			handler: getProductTestCasesHandler,
		},
		{
			name:        "get",
			description: "获取用例详情",
			params:      map[string]parameterSpec{"id": idStringParam("用例 ID")},
			required:    []string{"id"},
			handler:     getTestCaseHandler,
		},
		{
			name:        "create",
			description: "创建测试用例",
			params: map[string]parameterSpec{
				"product_id":   productRefParam("产品 ID 或名称；不传则使用默认产品"),
				"title":        requiredStringParam("用例标题"),
				"type":         enumStringParam("用例类型", "feature", "performance", "config", "install", "security", "interface", "unit", "other"),
				"steps":        testCaseStepsParam(),
				"pri":          integerParam("优先级 1-4，默认 1", 1, 4),
				"stage":        enumStringParam("适用阶段", "unittest", "feature", "intergrate", "system", "smoke", "bvt"),
				"precondition": optionalStringParam("前置条件"),
				"branch":       nonNegativeIntegerParam("所属分支 ID；0 表示主干"),
				"module":       nonNegativeIntegerParam("所属模块 ID；0 表示不指定"),
				"story":        nonNegativeIntegerParam("相关需求 ID；0 表示不指定"),
				"keywords":     optionalStringParam("关键词"),
			},
			required: []string{"title", "type", "steps"},
			handler:  createTestCaseHandler,
		},
	}
}

func newStrictTool(def strictToolDefinition) mcp.Tool {
	properties := make(map[string]interface{}, len(def.params))
	for name, spec := range def.params {
		properties[name] = spec.schema
	}
	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           properties,
		"required":             def.required,
		"additionalProperties": false,
	}
	return mcp.NewToolWithRawSchema(def.name, def.description, mustJSON(schema))
}

func newActionTool(name, description string, actions []actionDefinition) mcp.Tool {
	branches := make([]interface{}, 0, len(actions))
	actionNames := make([]string, 0, len(actions))
	topProperties := make(map[string]interface{})
	for _, action := range actions {
		actionNames = append(actionNames, action.name)
		properties := make(map[string]interface{}, len(action.params)+1)
		properties["action"] = map[string]interface{}{
			"type":        "string",
			"const":       action.name,
			"enum":        []string{action.name},
			"description": action.description,
		}
		for paramName, spec := range action.params {
			properties[paramName] = spec.schema
			if _, exists := topProperties[paramName]; !exists {
				topProperties[paramName] = spec.schema
			}
		}
		required := append([]string{"action"}, action.required...)
		branches = append(branches, map[string]interface{}{
			"title":                action.name,
			"description":          action.description,
			"type":                 "object",
			"properties":           properties,
			"required":             required,
			"additionalProperties": false,
		})
	}
	topProperties["action"] = map[string]interface{}{
		"type":        "string",
		"enum":        actionNames,
		"description": "要执行的操作；选择 action 后只能传该 action 分支声明的参数",
	}
	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           topProperties,
		"required":             []string{"action"},
		"additionalProperties": false,
		"oneOf":                branches,
	}
	return mcp.NewToolWithRawSchema(name, description, mustJSON(schema))
}

func newStrictToolHandler(def strictToolDefinition) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, err := validateArguments(def.name, request.Params.Arguments, def.params, def.required)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		request.Params.Arguments = args
		return def.handler(ctx, request)
	}
}

func newActionToolHandler(toolName string, actions []actionDefinition) server.ToolHandlerFunc {
	actionByName := make(map[string]actionDefinition, len(actions))
	actionNames := make([]string, 0, len(actions))
	for _, action := range actions {
		actionByName[action.name] = action
		actionNames = append(actionNames, action.name)
	}
	sort.Strings(actionNames)

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		rawAction, exists := request.Params.Arguments["action"]
		if !exists {
			return errorResult(fmt.Sprintf("%s 缺少必填参数 action；可选值: %s", toolName, strings.Join(actionNames, ", "))), nil
		}
		actionName, ok := rawAction.(string)
		if !ok || strings.TrimSpace(actionName) == "" {
			return errorResult(fmt.Sprintf("%s.action 必须是字符串；可选值: %s", toolName, strings.Join(actionNames, ", "))), nil
		}
		action, ok := actionByName[actionName]
		if !ok {
			return errorResult(fmt.Sprintf("%s.action 不支持 %q；可选值: %s", toolName, actionName, strings.Join(actionNames, ", "))), nil
		}

		input := make(map[string]interface{}, len(request.Params.Arguments)-1)
		for name, value := range request.Params.Arguments {
			if name != "action" {
				input[name] = value
			}
		}
		args, err := validateArguments(toolName+"."+actionName, input, action.params, action.required)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		if action.mapArgs != nil {
			args = action.mapArgs(args)
		}
		request.Params.Arguments = args
		return action.handler(ctx, request)
	}
}

func validateArguments(scope string, input map[string]interface{}, specs map[string]parameterSpec, required []string) (map[string]interface{}, error) {
	if input == nil {
		input = map[string]interface{}{}
	}
	unknown := make([]string, 0)
	for name := range input {
		if _, ok := specs[name]; !ok {
			unknown = append(unknown, name)
		}
	}
	if len(unknown) > 0 {
		sort.Strings(unknown)
		allowed := make([]string, 0, len(specs))
		for name := range specs {
			allowed = append(allowed, name)
		}
		sort.Strings(allowed)
		return nil, fmt.Errorf("%s 收到不支持的参数: %s；允许参数: %s", scope, strings.Join(unknown, ", "), strings.Join(allowed, ", "))
	}

	missing := make([]string, 0)
	for _, name := range required {
		if _, ok := input[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%s 缺少必填参数: %s", scope, strings.Join(missing, ", "))
	}

	normalized := make(map[string]interface{}, len(input))
	for name, value := range input {
		converted, err := specs[name].normalize(value)
		if err != nil {
			return nil, fmt.Errorf("%s.%s 参数错误: %w", scope, name, err)
		}
		normalized[name] = converted
	}
	return normalized, nil
}

func mustJSON(value interface{}) json.RawMessage {
	data, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("生成 MCP tool schema 失败: %v", err))
	}
	return data
}

func requiredStringParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "string", "minLength": 1, "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("必须是字符串")
			}
			if strings.TrimSpace(text) == "" {
				return nil, fmt.Errorf("不能为空")
			}
			return text, nil
		},
	}
}

func optionalStringParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "string", "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("必须是字符串")
			}
			return text, nil
		},
	}
}

func enumStringParam(description string, values ...string) parameterSpec {
	allowed := make(map[string]struct{}, len(values))
	for _, value := range values {
		allowed[value] = struct{}{}
	}
	return parameterSpec{
		schema: map[string]interface{}{"type": "string", "enum": values, "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("必须是字符串，可选值: %s", strings.Join(values, ", "))
			}
			if _, ok := allowed[text]; !ok {
				return nil, fmt.Errorf("不支持 %q，可选值: %s", text, strings.Join(values, ", "))
			}
			return text, nil
		},
	}
}

func booleanParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "boolean", "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			boolean, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("必须是布尔值 true 或 false")
			}
			return boolean, nil
		},
	}
}

func integerParam(description string, minValue, maxValue int64) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{
			"type":        "integer",
			"minimum":     minValue,
			"maximum":     maxValue,
			"description": description,
		},
		normalize: func(value interface{}) (interface{}, error) {
			integer, err := toInteger(value)
			if err != nil {
				return nil, err
			}
			if integer < minValue || integer > maxValue {
				return nil, fmt.Errorf("必须在 %d 到 %d 之间", minValue, maxValue)
			}
			return float64(integer), nil
		},
	}
}

func positiveIntegerParam(description string) parameterSpec {
	return integerParam(description, 1, math.MaxInt32)
}

func nonNegativeIntegerParam(description string) parameterSpec {
	return integerParam(description, 0, math.MaxInt32)
}

func idStringParam(description string) parameterSpec {
	spec := integerParam(description, 1, math.MaxInt32)
	spec.normalize = func(value interface{}) (interface{}, error) {
		integer, err := toInteger(value)
		if err != nil {
			return nil, err
		}
		if integer < 1 || integer > math.MaxInt32 {
			return nil, fmt.Errorf("必须是大于 0 的整数 ID")
		}
		return strconv.FormatInt(integer, 10), nil
	}
	return spec
}

func productRefParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{
			"oneOf": []interface{}{
				map[string]interface{}{"type": "integer", "minimum": 1, "maximum": math.MaxInt32},
				map[string]interface{}{"type": "string", "minLength": 1},
			},
			"description": description,
		},
		normalize: func(value interface{}) (interface{}, error) {
			if text, ok := value.(string); ok {
				if strings.TrimSpace(text) == "" {
					return nil, fmt.Errorf("产品 ID 或名称不能为空")
				}
				return text, nil
			}
			integer, err := toInteger(value)
			if err != nil || integer < 1 || integer > math.MaxInt32 {
				return nil, fmt.Errorf("必须是大于 0 的整数产品 ID 或非空产品名称")
			}
			return strconv.FormatInt(integer, 10), nil
		},
	}
}

func numberParam(description string, minValue, maxValue float64) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "number", "minimum": minValue, "maximum": maxValue, "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			number, err := toNumber(value)
			if err != nil {
				return nil, err
			}
			if number < minValue || number > maxValue {
				return nil, fmt.Errorf("必须在 %g 到 %g 之间", minValue, maxValue)
			}
			return number, nil
		},
	}
}

func dateParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "string", "format": "date", "pattern": `^\d{4}-\d{2}-\d{2}$`, "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("必须是 YYYY-MM-DD 格式的字符串")
			}
			if _, err := time.Parse("2006-01-02", text); err != nil {
				return nil, fmt.Errorf("必须是有效日期，格式为 YYYY-MM-DD")
			}
			return text, nil
		},
	}
}

func urlParam(description string) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{"type": "string", "format": "uri", "minLength": 1, "description": description},
		normalize: func(value interface{}) (interface{}, error) {
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("必须是 URL 字符串")
			}
			parsed, err := url.ParseRequestURI(text)
			if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
				return nil, fmt.Errorf("必须是包含主机名的 http 或 https URL")
			}
			return strings.TrimRight(text, "/"), nil
		},
	}
}

func stringArrayParam(description string, minItems, maxItems int) parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{
			"type":        "array",
			"minItems":    minItems,
			"maxItems":    maxItems,
			"uniqueItems": true,
			"items":       map[string]interface{}{"type": "string", "minLength": 1},
			"description": description,
		},
		normalize: func(value interface{}) (interface{}, error) {
			items, err := interfaceSlice(value)
			if err != nil {
				return nil, fmt.Errorf("必须是字符串数组")
			}
			if len(items) < minItems || len(items) > maxItems {
				return nil, fmt.Errorf("数组长度必须在 %d 到 %d 之间", minItems, maxItems)
			}
			normalized := make([]interface{}, 0, len(items))
			seen := make(map[string]struct{}, len(items))
			for index, item := range items {
				text, ok := item.(string)
				if !ok || strings.TrimSpace(text) == "" {
					return nil, fmt.Errorf("第 %d 项必须是非空字符串", index+1)
				}
				if _, exists := seen[text]; exists {
					return nil, fmt.Errorf("不能包含重复项 %q", text)
				}
				seen[text] = struct{}{}
				normalized = append(normalized, text)
			}
			return normalized, nil
		},
	}
}

func testCaseStepsParam() parameterSpec {
	return parameterSpec{
		schema: map[string]interface{}{
			"type":     "array",
			"minItems": 1,
			"maxItems": 100,
			"items": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"desc":   map[string]interface{}{"type": "string", "minLength": 1, "description": "步骤描述"},
					"expect": map[string]interface{}{"type": "string", "minLength": 1, "description": "期望结果"},
				},
				"required":             []string{"desc", "expect"},
				"additionalProperties": false,
			},
			"description": "用例步骤数组，每项必须包含 desc 和 expect",
		},
		normalize: func(value interface{}) (interface{}, error) {
			items, err := interfaceSlice(value)
			if err != nil {
				return nil, fmt.Errorf("必须是步骤对象数组")
			}
			if len(items) < 1 || len(items) > 100 {
				return nil, fmt.Errorf("步骤数量必须在 1 到 100 之间")
			}
			normalized := make([]interface{}, 0, len(items))
			for index, item := range items {
				step, ok := item.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("第 %d 项必须是对象", index+1)
				}
				if len(step) != 2 {
					return nil, fmt.Errorf("第 %d 项只能包含 desc 和 expect", index+1)
				}
				desc, descOK := step["desc"].(string)
				expect, expectOK := step["expect"].(string)
				if !descOK || strings.TrimSpace(desc) == "" {
					return nil, fmt.Errorf("第 %d 项的 desc 必须是非空字符串", index+1)
				}
				if !expectOK || strings.TrimSpace(expect) == "" {
					return nil, fmt.Errorf("第 %d 项的 expect 必须是非空字符串", index+1)
				}
				normalized = append(normalized, map[string]interface{}{"desc": desc, "expect": expect})
			}
			return normalized, nil
		},
	}
}

func interfaceSlice(value interface{}) ([]interface{}, error) {
	switch items := value.(type) {
	case []interface{}:
		return items, nil
	case []string:
		result := make([]interface{}, len(items))
		for index, item := range items {
			result[index] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("not an array")
	}
}

func toInteger(value interface{}) (int64, error) {
	switch number := value.(type) {
	case int:
		return int64(number), nil
	case int8:
		return int64(number), nil
	case int16:
		return int64(number), nil
	case int32:
		return int64(number), nil
	case int64:
		return number, nil
	case uint:
		if uint64(number) > math.MaxInt64 {
			return 0, fmt.Errorf("整数过大")
		}
		return int64(number), nil
	case uint8:
		return int64(number), nil
	case uint16:
		return int64(number), nil
	case uint32:
		return int64(number), nil
	case uint64:
		if number > math.MaxInt64 {
			return 0, fmt.Errorf("整数过大")
		}
		return int64(number), nil
	case float32:
		return floatToInteger(float64(number))
	case float64:
		return floatToInteger(number)
	case json.Number:
		integer, err := number.Int64()
		if err != nil {
			return 0, fmt.Errorf("必须是整数")
		}
		return integer, nil
	default:
		return 0, fmt.Errorf("必须是整数")
	}
}

func floatToInteger(number float64) (int64, error) {
	if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number || number < math.MinInt64 || number > math.MaxInt64 {
		return 0, fmt.Errorf("必须是整数")
	}
	return int64(number), nil
}

func toNumber(value interface{}) (float64, error) {
	switch number := value.(type) {
	case float64:
		if math.IsNaN(number) || math.IsInf(number, 0) {
			return 0, fmt.Errorf("必须是有限数字")
		}
		return number, nil
	case float32:
		return float64(number), nil
	case json.Number:
		parsed, err := number.Float64()
		if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return 0, fmt.Errorf("必须是数字")
		}
		return parsed, nil
	default:
		integer, err := toInteger(value)
		if err != nil {
			return 0, fmt.Errorf("必须是数字")
		}
		return float64(integer), nil
	}
}
