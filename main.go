package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 尝试加载已有配置
	if err := globalTokenManager.LoadConfig(); err != nil {
		log.Printf("加载配置文件失败（如果是首次使用，请先配置）: %v", err)
	}

	// 如果已配置，启动时自动获取token
	if globalTokenManager.IsConfigured() {
		if _, err := globalTokenManager.GetToken(); err != nil {
			log.Printf("启动时获取Token失败: %v", err)
		} else {
			log.Printf("启动时成功获取Token")
		}
	}

	s := server.NewMCPServer(
		"Zentao MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// 配置禅道连接
	configureTool := mcp.NewTool("configure",
		mcp.WithDescription("配置禅道服务器连接信息（首次使用或修改配置时调用）"),
		mcp.WithString("base_url",
			mcp.Required(),
			mcp.Description("禅道服务器地址，例如: http://localhost:8080"),
		),
		mcp.WithString("account",
			mcp.Required(),
			mcp.Description("登录账号"),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("登录密码"),
		),
		mcp.WithNumber("token_expiry",
			mcp.Description("Token过期时间（小时），默认24小时"),
		),
		mcp.WithString("default_product",
			mcp.Description("默认产品ID或名称（创建Bug、需求等时使用）"),
		),
	)

	// 获取Token
	getTokenTool := mcp.NewTool("get_token",
		mcp.WithDescription("获取禅道API Token（自动处理缓存和刷新）"),
	)

	// 刷新Token
	refreshTokenTool := mcp.NewTool("refresh_token",
		mcp.WithDescription("强制刷新Token（忽略缓存）"),
	)

	// 查看Token状态
	tokenStatusTool := mcp.NewTool("token_status",
		mcp.WithDescription("查看当前Token状态和配置信息"),
	)

	// 获取产品列表
	getProductsTool := mcp.NewTool("get_products",
		mcp.WithDescription("获取禅道产品列表"),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	// 获取产品详情
	getProductTool := mcp.NewTool("get_product",
		mcp.WithDescription("获取禅道产品详情"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("产品ID"),
		),
	)

	// 创建产品
	createProductTool := mcp.NewTool("create_product",
		mcp.WithDescription("创建禅道产品"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("产品名称"),
		),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("产品代号"),
		),
		mcp.WithNumber("program",
			mcp.Description("所属项目集ID"),
		),
		mcp.WithNumber("line",
			mcp.Description("所属产品线ID"),
		),
		mcp.WithString("PO",
			mcp.Description("产品负责人账号"),
		),
		mcp.WithString("QD",
			mcp.Description("测试负责人账号"),
		),
		mcp.WithString("RD",
			mcp.Description("发布负责人账号"),
		),
		mcp.WithString("type",
			mcp.Description("产品类型: normal(正常) | branch(多分支) | platform(多平台)"),
		),
		mcp.WithString("desc",
			mcp.Description("产品描述"),
		),
		mcp.WithString("acl",
			mcp.Description("访问控制: open(公开) | private(私有)"),
		),
	)

	// 创建Bug
	createBugTool := mcp.NewTool("create_bug",
		mcp.WithDescription("创建Bug"),
		mcp.WithString("product_id",
			mcp.Description("产品ID或名称（不传则使用默认产品）"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Bug标题"),
		),
		mcp.WithNumber("severity",
			mcp.Required(),
			mcp.Description("严重程度(1-4)"),
		),
		mcp.WithNumber("pri",
			mcp.Required(),
			mcp.Description("优先级(1-4)"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Bug类型: codeerror(代码错误) | config(配置相关) | install(安装部署) | security(安全相关) | performance(性能问题) | standard(标准规范) | automation(测试脚本) | designdefect(设计缺陷) | others(其他)"),
		),
		mcp.WithString("steps",
			mcp.Description("重现步骤"),
		),
		mcp.WithString("keywords",
			mcp.Description("关键词"),
		),
		mcp.WithNumber("branch",
			mcp.Description("所属分支ID"),
		),
		mcp.WithNumber("module",
			mcp.Description("所属模块ID"),
		),
		mcp.WithNumber("execution",
			mcp.Description("所属执行ID"),
		),
		mcp.WithString("os",
			mcp.Description("操作系统"),
		),
		mcp.WithString("browser",
			mcp.Description("浏览器"),
		),
		mcp.WithNumber("task",
			mcp.Description("相关任务ID"),
		),
		mcp.WithNumber("story",
			mcp.Description("相关需求ID"),
		),
		mcp.WithString("deadline",
			mcp.Description("截止日期(格式: YYYY-MM-DD)"),
		),
		mcp.WithArray("opened_build",
			mcp.Description("影响版本(数组，如: [\"trunk\"])"),
		),
	)

	// 修改Bug
	updateBugTool := mcp.NewTool("update_bug",
		mcp.WithDescription("修改Bug"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Bug ID"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Bug标题"),
		),
		mcp.WithNumber("severity",
			mcp.Required(),
			mcp.Description("严重程度(1-4)"),
		),
		mcp.WithNumber("pri",
			mcp.Required(),
			mcp.Description("优先级(1-4)"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Bug类型: codeerror(代码错误) | config(配置相关) | install(安装部署) | security(安全相关) | performance(性能问题) | standard(标准规范) | automation(测试脚本) | designdefect(设计缺陷) | others(其他)"),
		),
		mcp.WithString("steps",
			mcp.Description("重现步骤"),
		),
		mcp.WithString("keywords",
			mcp.Description("关键词"),
		),
		mcp.WithNumber("branch",
			mcp.Description("所属分支ID"),
		),
		mcp.WithNumber("module",
			mcp.Description("所属模块ID"),
		),
		mcp.WithNumber("execution",
			mcp.Description("所属执行ID"),
		),
		mcp.WithString("os",
			mcp.Description("操作系统"),
		),
		mcp.WithString("browser",
			mcp.Description("浏览器"),
		),
		mcp.WithNumber("task",
			mcp.Description("相关任务ID"),
		),
		mcp.WithNumber("story",
			mcp.Description("相关需求ID"),
		),
		mcp.WithString("deadline",
			mcp.Description("截止日期(格式: YYYY-MM-DD)"),
		),
		mcp.WithArray("opened_build",
			mcp.Description("影响版本(数组，如: [\"trunk\"])"),
		),
	)

	// 获取产品Bug列表
	getBugsTool := mcp.NewTool("get_bugs",
		mcp.WithDescription("获取产品Bug列表"),
		mcp.WithString("product_id",
			mcp.Description("产品ID或名称（不传则使用默认产品）"),
		),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	// 获取Bug详情
	getBugTool := mcp.NewTool("get_bug",
		mcp.WithDescription("获取Bug详情"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Bug ID"),
		),
	)

	// 创建版本
	createBuildTool := mcp.NewTool("create_build",
		mcp.WithDescription("创建版本"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("版本名称"),
		),
		mcp.WithNumber("product",
			mcp.Required(),
			mcp.Description("所属产品ID"),
		),
		mcp.WithNumber("execution",
			mcp.Required(),
			mcp.Description("所属执行ID"),
		),
		mcp.WithString("builder",
			mcp.Required(),
			mcp.Description("构建者账号"),
		),
		mcp.WithNumber("branch",
			mcp.Description("所属分支ID"),
		),
		mcp.WithString("date",
			mcp.Description("打包日期(格式: YYYY-MM-DD)"),
		),
		mcp.WithString("scm_path",
			mcp.Description("源代码地址"),
		),
		mcp.WithString("file_path",
			mcp.Description("下载地址"),
		),
		mcp.WithString("desc",
			mcp.Description("版本描述"),
		),
	)

	// 修改版本
	updateBuildTool := mcp.NewTool("update_build",
		mcp.WithDescription("修改版本"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("版本ID"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("版本名称"),
		),
		mcp.WithNumber("product",
			mcp.Required(),
			mcp.Description("所属产品ID"),
		),
		mcp.WithNumber("execution",
			mcp.Required(),
			mcp.Description("所属执行ID"),
		),
		mcp.WithString("builder",
			mcp.Required(),
			mcp.Description("构建者账号"),
		),
		mcp.WithNumber("branch",
			mcp.Description("所属分支ID"),
		),
		mcp.WithString("date",
			mcp.Description("打包日期(格式: YYYY-MM-DD)"),
		),
		mcp.WithString("scm_path",
			mcp.Description("源代码地址"),
		),
		mcp.WithString("file_path",
			mcp.Description("下载地址"),
		),
		mcp.WithString("desc",
			mcp.Description("版本描述"),
		),
	)

	// 获取版本详情
	getBuildTool := mcp.NewTool("get_build",
		mcp.WithDescription("获取版本详情"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("版本ID"),
		),
	)

	// 获取项目版本列表
	getBuildsTool := mcp.NewTool("get_builds",
		mcp.WithDescription("获取项目版本列表"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID"),
		),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	// 创建需求
	createStoryTool := mcp.NewTool("create_story",
		mcp.WithDescription("创建需求"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("需求标题"),
		),
		mcp.WithString("product",
			mcp.Description("产品ID或名称（不传则使用默认产品）"),
		),
		mcp.WithNumber("pri",
			mcp.Required(),
			mcp.Description("优先级(1-4)"),
		),
		mcp.WithString("category",
			mcp.Required(),
			mcp.Description("需求类型: feature(功能) | interface(接口) | performance(性能) | safe(安全) | experience(体验) | improve(改进) | other(其他)"),
		),
		mcp.WithString("spec",
			mcp.Description("需求描述"),
		),
		mcp.WithString("verify",
			mcp.Description("验收标准"),
		),
		mcp.WithString("source",
			mcp.Description("需求来源: customer(客户) | user(用户) | po(产品经理) | market(市场)"),
		),
		mcp.WithString("source_note",
			mcp.Description("来源备注"),
		),
		mcp.WithNumber("estimate",
			mcp.Description("预计工时"),
		),
		mcp.WithString("keywords",
			mcp.Description("关键词"),
		),
		mcp.WithString("reviewer",
			mcp.Description("评审人账号"),
		),
		mcp.WithString("assigned_to",
			mcp.Description("指派给账号"),
		),
	)

	// 获取需求详情
	getStoryTool := mcp.NewTool("get_story",
		mcp.WithDescription("获取需求详情"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("需求ID"),
		),
	)

	// 获取项目需求列表
	getProjectStoriesTool := mcp.NewTool("get_project_stories",
		mcp.WithDescription("获取项目需求列表"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("项目ID"),
		),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	// 获取产品需求列表
	getProductStoriesTool := mcp.NewTool("get_product_stories",
		mcp.WithDescription("获取产品需求列表"),
		mcp.WithString("product_id",
			mcp.Description("产品ID或名称（不传则使用默认产品）"),
		),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	// 获取执行需求列表
	getExecutionStoriesTool := mcp.NewTool("get_execution_stories",
		mcp.WithDescription("获取执行需求列表"),
		mcp.WithString("execution_id",
			mcp.Required(),
			mcp.Description("执行ID"),
		),
		mcp.WithBoolean("full",
			mcp.Description("是否返回完整参数，默认false返回精简参数"),
		),
	)

	s.AddTool(configureTool, configureHandler)
	s.AddTool(getTokenTool, getTokenHandler)
	s.AddTool(refreshTokenTool, refreshTokenHandler)
	s.AddTool(tokenStatusTool, tokenStatusHandler)
	s.AddTool(getProductsTool, getProductsHandler)
	s.AddTool(getProductTool, getProductHandler)
	s.AddTool(createProductTool, createProductHandler)
	s.AddTool(createBugTool, createBugHandler)
	s.AddTool(updateBugTool, updateBugHandler)
	s.AddTool(getBugsTool, getBugsHandler)
	s.AddTool(getBugTool, getBugHandler)
	s.AddTool(createBuildTool, createBuildHandler)
	s.AddTool(updateBuildTool, updateBuildHandler)
	s.AddTool(getBuildTool, getBuildHandler)
	s.AddTool(getBuildsTool, getBuildsHandler)
	s.AddTool(createStoryTool, createStoryHandler)
	s.AddTool(getStoryTool, getStoryHandler)
	s.AddTool(getProjectStoriesTool, getProjectStoriesHandler)
	s.AddTool(getProductStoriesTool, getProductStoriesHandler)
	s.AddTool(getExecutionStoriesTool, getExecutionStoriesHandler)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func configureHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	baseURL, ok := request.Params.Arguments["base_url"].(string)
	if !ok {
		return errorResult("base_url is required"), nil
	}

	account, ok := request.Params.Arguments["account"].(string)
	if !ok {
		return errorResult("account is required"), nil
	}

	password, ok := request.Params.Arguments["password"].(string)
	if !ok {
		return errorResult("password is required"), nil
	}

	tokenExpiry := 24
	if exp, ok := request.Params.Arguments["token_expiry"].(float64); ok {
		tokenExpiry = int(exp)
	}

	defaultProduct := ""
	if dp, ok := request.Params.Arguments["default_product"].(string); ok {
		defaultProduct = dp
	}

	config := &Config{
		BaseURL:        baseURL,
		Account:        account,
		Password:       password,
		TokenExpiry:    tokenExpiry,
		DefaultProduct: defaultProduct,
	}

	if err := globalTokenManager.SaveConfig(config); err != nil {
		return errorResult(fmt.Sprintf("保存配置失败: %v", err)), nil
	}

	// 配置保存后立即获取token
	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("配置成功，但获取Token失败: %v", err)), nil
	}

	result := map[string]interface{}{
		"success": true,
		"message": "配置成功并已获取Token",
		"token":   token,
	}
	if defaultProduct != "" {
		result["default_product"] = defaultProduct
	}

	data, _ := toJSON(result)
	return mcp.NewToolResultText(data), nil
}

func getTokenHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"token": "%s"}`, token)), nil
}

func refreshTokenHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	token, err := globalTokenManager.RefreshToken()
	if err != nil {
		return errorResult(fmt.Sprintf("刷新Token失败: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf(`{"success": true, "message": "Token已刷新", "token": "%s"}`, token)), nil
}

func tokenStatusHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info := globalTokenManager.GetTokenInfo()
	data, err := toJSON(info)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化状态信息失败: %v", err)), nil
	}
	return mcp.NewToolResultText(data), nil
}

func getProductsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	products, err := client.GetProducts(token)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		products, err = client.GetProducts(token)
		if err != nil {
			return errorResult(fmt.Sprintf("获取产品列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && products.Products != nil {
		simplified := make([]map[string]interface{}, len(products.Products))
		for i, p := range products.Products {
			simplified[i] = simplifyProduct(p)
		}
		result := map[string]interface{}{
			"total":    products.Total,
			"products": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化产品列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(products)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化产品列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getProductHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	productID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	product, err := client.GetProduct(token, productID)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		product, err = client.GetProduct(token, productID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取产品详情失败: %v", err)), nil
		}
	}

	data, err := toJSON(product)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化产品详情失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func createProductHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return errorResult("name is required"), nil
	}

	code, ok := request.Params.Arguments["code"].(string)
	if !ok {
		return errorResult("code is required"), nil
	}

	reqBody := &CreateProductRequest{
		Name: name,
		Code: code,
	}

	if program, ok := request.Params.Arguments["program"].(float64); ok {
		reqBody.Program = int(program)
	}
	if line, ok := request.Params.Arguments["line"].(float64); ok {
		reqBody.Line = int(line)
	}
	if po, ok := request.Params.Arguments["PO"].(string); ok {
		reqBody.PO = po
	}
	if qd, ok := request.Params.Arguments["QD"].(string); ok {
		reqBody.QD = qd
	}
	if rd, ok := request.Params.Arguments["RD"].(string); ok {
		reqBody.RD = rd
	}
	if typeVal, ok := request.Params.Arguments["type"].(string); ok {
		reqBody.Type = typeVal
	}
	if desc, ok := request.Params.Arguments["desc"].(string); ok {
		reqBody.Desc = desc
	}
	if acl, ok := request.Params.Arguments["acl"].(string); ok {
		reqBody.ACL = acl
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	product, err := client.CreateProduct(token, reqBody)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		product, err = client.CreateProduct(token, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("创建产品失败: %v", err)), nil
		}
	}

	data, err := toJSON(product)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化产品信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func createBugHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return errorResult("title is required"), nil
	}

	severity, ok := request.Params.Arguments["severity"].(float64)
	if !ok {
		return errorResult("severity is required"), nil
	}

	pri, ok := request.Params.Arguments["pri"].(float64)
	if !ok {
		return errorResult("pri is required"), nil
	}

	typeVal, ok := request.Params.Arguments["type"].(string)
	if !ok {
		return errorResult("type is required"), nil
	}

	reqBody := &BugRequest{
		Title:    title,
		Severity: int(severity),
		Pri:      int(pri),
		Type:     typeVal,
	}

	if branch, ok := request.Params.Arguments["branch"].(float64); ok {
		reqBody.Branch = int(branch)
	}
	if module, ok := request.Params.Arguments["module"].(float64); ok {
		reqBody.Module = int(module)
	}
	if execution, ok := request.Params.Arguments["execution"].(float64); ok {
		reqBody.Execution = int(execution)
	}
	if steps, ok := request.Params.Arguments["steps"].(string); ok {
		reqBody.Steps = steps
	}
	if keywords, ok := request.Params.Arguments["keywords"].(string); ok {
		reqBody.Keywords = keywords
	}
	if os, ok := request.Params.Arguments["os"].(string); ok {
		reqBody.OS = os
	}
	if browser, ok := request.Params.Arguments["browser"].(string); ok {
		reqBody.Browser = browser
	}
	if task, ok := request.Params.Arguments["task"].(float64); ok {
		reqBody.Task = int(task)
	}
	if story, ok := request.Params.Arguments["story"].(float64); ok {
		reqBody.Story = int(story)
	}
	if deadline, ok := request.Params.Arguments["deadline"].(string); ok {
		reqBody.Deadline = deadline
	}
	if openedBuild, ok := request.Params.Arguments["opened_build"].([]interface{}); ok {
		for _, v := range openedBuild {
			if s, ok := v.(string); ok {
				reqBody.OpenedBuild = append(reqBody.OpenedBuild, s)
			}
		}
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	// 解析产品ID（支持ID、名称或使用默认产品）
	productIDInput, _ := request.Params.Arguments["product_id"].(string)
	productID, err := resolveProductID(productIDInput, token)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	bug, err := client.CreateBug(token, productID, reqBody)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		bug, err = client.CreateBug(token, productID, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("创建Bug失败: %v", err)), nil
		}
	}

	data, err := toJSON(bug)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化Bug信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func updateBugHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	bugID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return errorResult("title is required"), nil
	}

	severity, ok := request.Params.Arguments["severity"].(float64)
	if !ok {
		return errorResult("severity is required"), nil
	}

	pri, ok := request.Params.Arguments["pri"].(float64)
	if !ok {
		return errorResult("pri is required"), nil
	}

	typeVal, ok := request.Params.Arguments["type"].(string)
	if !ok {
		return errorResult("type is required"), nil
	}

	reqBody := &BugRequest{
		Title:    title,
		Severity: int(severity),
		Pri:      int(pri),
		Type:     typeVal,
	}

	if branch, ok := request.Params.Arguments["branch"].(float64); ok {
		reqBody.Branch = int(branch)
	}
	if module, ok := request.Params.Arguments["module"].(float64); ok {
		reqBody.Module = int(module)
	}
	if execution, ok := request.Params.Arguments["execution"].(float64); ok {
		reqBody.Execution = int(execution)
	}
	if steps, ok := request.Params.Arguments["steps"].(string); ok {
		reqBody.Steps = steps
	}
	if keywords, ok := request.Params.Arguments["keywords"].(string); ok {
		reqBody.Keywords = keywords
	}
	if os, ok := request.Params.Arguments["os"].(string); ok {
		reqBody.OS = os
	}
	if browser, ok := request.Params.Arguments["browser"].(string); ok {
		reqBody.Browser = browser
	}
	if task, ok := request.Params.Arguments["task"].(float64); ok {
		reqBody.Task = int(task)
	}
	if story, ok := request.Params.Arguments["story"].(float64); ok {
		reqBody.Story = int(story)
	}
	if deadline, ok := request.Params.Arguments["deadline"].(string); ok {
		reqBody.Deadline = deadline
	}
	if openedBuild, ok := request.Params.Arguments["opened_build"].([]interface{}); ok {
		for _, v := range openedBuild {
			if s, ok := v.(string); ok {
				reqBody.OpenedBuild = append(reqBody.OpenedBuild, s)
			}
		}
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	bug, err := client.UpdateBug(token, bugID, reqBody)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		bug, err = client.UpdateBug(token, bugID, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("修改Bug失败: %v", err)), nil
		}
	}

	data, err := toJSON(bug)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化Bug信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getBugsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	// 解析产品ID（支持ID、名称或使用默认产品）
	productIDInput, _ := request.Params.Arguments["product_id"].(string)
	productID, err := resolveProductID(productIDInput, token)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	bugs, err := client.GetBugs(token, productID)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		bugs, err = client.GetBugs(token, productID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取Bug列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && bugs.Bugs != nil {
		simplified := make([]map[string]interface{}, len(bugs.Bugs))
		for i, b := range bugs.Bugs {
			simplified[i] = simplifyBug(b)
		}
		result := map[string]interface{}{
			"page":  bugs.Page,
			"total": bugs.Total,
			"bugs":  simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化Bug列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(bugs)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化Bug列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getBugHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	bugID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	bug, err := client.GetBug(token, bugID)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		bug, err = client.GetBug(token, bugID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取Bug详情失败: %v", err)), nil
		}
	}

	data, err := toJSON(bug)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化Bug详情失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func createBuildHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	projectID, ok := request.Params.Arguments["project_id"].(string)
	if !ok {
		return errorResult("project_id is required"), nil
	}

	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return errorResult("name is required"), nil
	}

	product, ok := request.Params.Arguments["product"].(float64)
	if !ok {
		return errorResult("product is required"), nil
	}

	execution, ok := request.Params.Arguments["execution"].(float64)
	if !ok {
		return errorResult("execution is required"), nil
	}

	builder, ok := request.Params.Arguments["builder"].(string)
	if !ok {
		return errorResult("builder is required"), nil
	}

	reqBody := &BuildRequest{
		Name:      name,
		Product:   int(product),
		Execution: int(execution),
		Builder:   builder,
	}

	if branch, ok := request.Params.Arguments["branch"].(float64); ok {
		reqBody.Branch = int(branch)
	}
	if date, ok := request.Params.Arguments["date"].(string); ok {
		reqBody.Date = date
	}
	if scmPath, ok := request.Params.Arguments["scm_path"].(string); ok {
		reqBody.ScmPath = scmPath
	}
	if filePath, ok := request.Params.Arguments["file_path"].(string); ok {
		reqBody.FilePath = filePath
	}
	if desc, ok := request.Params.Arguments["desc"].(string); ok {
		reqBody.Desc = desc
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	build, err := client.CreateBuild(token, projectID, reqBody)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		build, err = client.CreateBuild(token, projectID, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("创建版本失败: %v", err)), nil
		}
	}

	data, err := toJSON(build)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化版本信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func updateBuildHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	buildID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return errorResult("name is required"), nil
	}

	product, ok := request.Params.Arguments["product"].(float64)
	if !ok {
		return errorResult("product is required"), nil
	}

	execution, ok := request.Params.Arguments["execution"].(float64)
	if !ok {
		return errorResult("execution is required"), nil
	}

	builder, ok := request.Params.Arguments["builder"].(string)
	if !ok {
		return errorResult("builder is required"), nil
	}

	reqBody := &BuildRequest{
		Name:      name,
		Product:   int(product),
		Execution: int(execution),
		Builder:   builder,
	}

	if branch, ok := request.Params.Arguments["branch"].(float64); ok {
		reqBody.Branch = int(branch)
	}
	if date, ok := request.Params.Arguments["date"].(string); ok {
		reqBody.Date = date
	}
	if scmPath, ok := request.Params.Arguments["scm_path"].(string); ok {
		reqBody.ScmPath = scmPath
	}
	if filePath, ok := request.Params.Arguments["file_path"].(string); ok {
		reqBody.FilePath = filePath
	}
	if desc, ok := request.Params.Arguments["desc"].(string); ok {
		reqBody.Desc = desc
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	build, err := client.UpdateBuild(token, buildID, reqBody)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		build, err = client.UpdateBuild(token, buildID, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("修改版本失败: %v", err)), nil
		}
	}

	data, err := toJSON(build)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化版本信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getBuildHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	buildID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	build, err := client.GetBuild(token, buildID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		build, err = client.GetBuild(token, buildID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取版本详情失败: %v", err)), nil
		}
	}

	data, err := toJSON(build)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化版本详情失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getBuildsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	projectID, ok := request.Params.Arguments["project_id"].(string)
	if !ok {
		return errorResult("project_id is required"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	builds, err := client.GetBuilds(token, projectID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		builds, err = client.GetBuilds(token, projectID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取版本列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && builds.Builds != nil {
		simplified := make([]map[string]interface{}, len(builds.Builds))
		for i, b := range builds.Builds {
			simplified[i] = simplifyBuild(b)
		}
		result := map[string]interface{}{
			"total":  builds.Total,
			"builds": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化版本列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(builds)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化版本列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func createStoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return errorResult("title is required"), nil
	}

	pri, ok := request.Params.Arguments["pri"].(float64)
	if !ok {
		return errorResult("pri is required"), nil
	}

	category, ok := request.Params.Arguments["category"].(string)
	if !ok {
		return errorResult("category is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	// 解析产品ID（支持ID、名称或使用默认产品）
	productInput, _ := request.Params.Arguments["product"].(string)
	productID, err := resolveProductID(productInput, token)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	// 将产品ID转换为int
	var productInt int
	if _, err := fmt.Sscanf(productID, "%d", &productInt); err != nil {
		return errorResult(fmt.Sprintf("产品ID格式错误: %v", err)), nil
	}

	reqBody := &StoryRequest{
		Title:    title,
		Product:  productInt,
		Pri:      int(pri),
		Category: category,
	}

	if spec, ok := request.Params.Arguments["spec"].(string); ok {
		reqBody.Spec = spec
	}
	if verify, ok := request.Params.Arguments["verify"].(string); ok {
		reqBody.Verify = verify
	}
	if source, ok := request.Params.Arguments["source"].(string); ok {
		reqBody.Source = source
	}
	if sourceNote, ok := request.Params.Arguments["source_note"].(string); ok {
		reqBody.SourceNote = sourceNote
	}
	if estimate, ok := request.Params.Arguments["estimate"].(float64); ok {
		reqBody.Estimate = estimate
	}
	if keywords, ok := request.Params.Arguments["keywords"].(string); ok {
		reqBody.Keywords = keywords
	}
	if reviewer, ok := request.Params.Arguments["reviewer"].(string); ok {
		reqBody.Reviewer = []string{reviewer}
	}
	if assignedTo, ok := request.Params.Arguments["assigned_to"].(string); ok {
		reqBody.AssignedTo = assignedTo
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	story, err := client.CreateStory(token, reqBody)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		story, err = client.CreateStory(token, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("创建需求失败: %v", err)), nil
		}
	}

	data, err := toJSON(story)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化需求信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getStoryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	storyID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	story, err := client.GetStory(token, storyID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		story, err = client.GetStory(token, storyID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取需求详情失败: %v", err)), nil
		}
	}

	data, err := toJSON(story)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化需求详情失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getProjectStoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	projectID, ok := request.Params.Arguments["project_id"].(string)
	if !ok {
		return errorResult("project_id is required"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	stories, err := client.GetProjectStories(token, projectID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		stories, err = client.GetProjectStories(token, projectID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取项目需求列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && stories.Stories != nil {
		simplified := make([]map[string]interface{}, len(stories.Stories))
		for i, s := range stories.Stories {
			simplified[i] = simplifyStory(s)
		}
		result := map[string]interface{}{
			"page":    stories.Page,
			"total":   stories.Total,
			"stories": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(stories)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getProductStoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	// 解析产品ID（支持ID、名称或使用默认产品）
	productIDInput, _ := request.Params.Arguments["product_id"].(string)
	productID, err := resolveProductID(productIDInput, token)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	stories, err := client.GetProductStories(token, productID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		stories, err = client.GetProductStories(token, productID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取产品需求列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && stories.Stories != nil {
		simplified := make([]map[string]interface{}, len(stories.Stories))
		for i, s := range stories.Stories {
			simplified[i] = simplifyStory(s)
		}
		result := map[string]interface{}{
			"page":    stories.Page,
			"total":   stories.Total,
			"stories": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(stories)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getExecutionStoriesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	executionID, ok := request.Params.Arguments["execution_id"].(string)
	if !ok {
		return errorResult("execution_id is required"), nil
	}

	full := request.Params.Arguments["full"] == true

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	stories, err := client.GetExecutionStories(token, executionID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		stories, err = client.GetExecutionStories(token, executionID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取执行需求列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && stories.Stories != nil {
		simplified := make([]map[string]interface{}, len(stories.Stories))
		for i, s := range stories.Stories {
			simplified[i] = simplifyStory(s)
		}
		result := map[string]interface{}{
			"page":    stories.Page,
			"total":   stories.Total,
			"stories": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(stories)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化需求列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.TextContent{Type: "text", Text: msg}},
		IsError: true,
	}
}

// toJSON 将数据序列化为JSON，不转义Unicode字符
func toJSON(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(v); err != nil {
		return "", err
	}
	// 移除末尾的换行符
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

// simplifyProduct 精简产品信息
func simplifyProduct(p Product) map[string]interface{} {
	return map[string]interface{}{
		"id":     p.ID,
		"name":   p.Name,
		"code":   p.Code,
		"type":   p.Type,
		"status": p.Status,
		"acl":    p.ACL,
	}
}

// simplifyBug 精简Bug信息
func simplifyBug(b BugListItem) map[string]interface{} {
	result := map[string]interface{}{
		"id":       b.ID,
		"title":    b.Title,
		"severity": b.Severity,
		"pri":      b.Pri,
		"type":     b.Type,
		"status":   b.Status,
	}
	if b.OpenedBy != nil {
		result["openedBy"] = b.OpenedBy.Account
	}
	return result
}

// simplifyBuild 精简版本信息
func simplifyBuild(b Build) map[string]interface{} {
	return map[string]interface{}{
		"id":       b.ID,
		"name":     b.Name,
		"date":     b.Date,
		"builder":  b.Builder,
		"productName": b.ProductName,
	}
}

// simplifyStory 精简需求信息
func simplifyStory(s Story) map[string]interface{} {
	result := map[string]interface{}{
		"id":        s.ID,
		"title":     s.Title,
		"category":  s.Category,
		"pri":       s.Pri,
		"status":    s.Status,
		"stage":     s.Stage,
		"estimate":  s.Estimate,
	}
	if s.OpenedBy != nil {
		if u, ok := s.OpenedBy.(map[string]interface{}); ok {
			if acc, ok := u["account"].(string); ok {
				result["openedBy"] = acc
			}
		}
	}
	return result
}

// resolveProductID 解析产品ID，支持ID或名称，为空时使用默认产品
func resolveProductID(productID string, token string) (string, error) {
	if productID != "" {
		// 如果传入的是数字，直接返回
		if _, err := fmt.Sscanf(productID, "%d", new(int)); err == nil {
			return productID, nil
		}
		// 否则当作产品名称查找
		client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
		products, err := client.GetProducts(token)
		if err != nil {
			return "", fmt.Errorf("查找产品失败: %v", err)
		}
		for _, p := range products.Products {
			if fmt.Sprintf("%v", p.Name) == productID {
				return fmt.Sprintf("%v", p.ID), nil
			}
		}
		return "", fmt.Errorf("未找到名称为 '%s' 的产品", productID)
	}

	// 没有传入产品ID，使用默认产品
	config := globalTokenManager.GetConfig()
	if config.DefaultProduct == "" {
		return "", fmt.Errorf("未指定产品ID，且未配置默认产品")
	}

	// 默认产品也可能是ID或名称
	defaultProduct := config.DefaultProduct
	if _, err := fmt.Sscanf(defaultProduct, "%d", new(int)); err == nil {
		return defaultProduct, nil
	}

	// 当作名称查找
	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	products, err := client.GetProducts(token)
	if err != nil {
		return "", fmt.Errorf("查找默认产品失败: %v", err)
	}
	for _, p := range products.Products {
		if fmt.Sprintf("%v", p.Name) == defaultProduct {
			return fmt.Sprintf("%v", p.ID), nil
		}
	}
	return "", fmt.Errorf("未找到默认产品 '%s'", defaultProduct)
}