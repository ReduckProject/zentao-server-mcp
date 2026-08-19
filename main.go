package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// 命令行参数
var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "配置文件路径（不指定则使用exe所在目录的zentao_config.json）")
	flag.StringVar(&configPath, "c", "", "配置文件路径（简写）")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 设置配置文件路径
	globalTokenManager.SetConfigPath(configPath)

	// 尝试加载已有配置
	if err := globalTokenManager.LoadConfig(); err != nil {
		log.Printf("加载配置文件失败（如果是首次使用，请先配置）: %v", err)
	} else {
		log.Printf("使用配置文件: %s", globalTokenManager.GetConfigPath())
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

	registerTools(s)

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

	tokenExpiry := 86400
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
	_, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("配置成功，但获取Token失败: %v", err)), nil
	}

	result := map[string]interface{}{
		"success":           true,
		"message":           "配置成功并已验证连接",
		"connection_status": globalTokenManager.GetConnectionStatus(),
	}
	if defaultProduct != "" {
		result["default_product"] = defaultProduct
	}

	data, _ := toJSON(result)
	return mcp.NewToolResultText(data), nil
}

func getProfileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	profile, err := client.GetUserProfile(token)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		profile, err = client.GetUserProfile(token)
		if err != nil {
			return errorResult(fmt.Sprintf("获取用户信息失败: %v", err)), nil
		}
	}

	connectionStatus := globalTokenManager.GetConnectionStatus()
	connectionStatus["connected"] = true
	result := map[string]interface{}{
		"profile":           profile.Profile,
		"connection_status": connectionStatus,
	}

	data, err := toJSON(result)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化用户信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getTodayDynamicHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	// 先获取用户信息以获取用户ID
	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	profile, err := client.GetUserProfile(token)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		profile, err = client.GetUserProfile(token)
		if err != nil {
			return errorResult(fmt.Sprintf("获取用户信息失败: %v", err)), nil
		}
	}

	// 获取时间范围参数
	timeRange, _ := request.Params.Arguments["time_range"].(string)

	userID := fmt.Sprintf("%v", profile.Profile.ID)
	dynamics, err := client.GetTodayDynamic(userID, timeRange)
	if err != nil {
		return errorResult(fmt.Sprintf("获取动态失败: %v", err)), nil
	}

	data, err := toJSON(map[string]interface{}{
		"user_id":    userID,
		"time_range": timeRange,
		"dynamic":    dynamics,
	})
	if err != nil {
		return errorResult(fmt.Sprintf("序列化动态数据失败: %v", err)), nil
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

	// 解析可选参数
	var opts GetBugsOptions
	if status, ok := request.Params.Arguments["status"].(string); ok {
		opts.Status = status
	}
	if limit, ok := request.Params.Arguments["limit"].(float64); ok {
		opts.Limit = int(limit)
	}
	if page, ok := request.Params.Arguments["page"].(float64); ok {
		opts.Page = int(page)
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
	bugs, err := client.GetBugs(token, productID, &opts)
	if err != nil {
		// Token可能过期，尝试刷新后重试
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		bugs, err = client.GetBugs(token, productID, &opts)
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

func getBugCommentsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	bugID, ok := request.Params.Arguments["bug_id"].(string)
	if !ok {
		return errorResult("bug_id is required"), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	comments, err := client.GetBugComments(bugID)
	if err != nil {
		return errorResult(fmt.Sprintf("获取Bug备注失败: %v", err)), nil
	}

	data, err := toJSON(map[string]interface{}{
		"bug_id":   bugID,
		"comments": comments,
		"total":    len(comments),
	})
	if err != nil {
		return errorResult(fmt.Sprintf("序列化备注列表失败: %v", err)), nil
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

func getProductTestCasesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	full := request.Params.Arguments["full"] == true

	var limit, page int
	if l, ok := request.Params.Arguments["limit"].(float64); ok {
		limit = int(l)
	}
	if p, ok := request.Params.Arguments["page"].(float64); ok {
		page = int(p)
	}
	search, _ := request.Params.Arguments["search"].(string)

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
	cases, err := client.GetProductTestCases(token, productID, limit, page, search)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		cases, err = client.GetProductTestCases(token, productID, limit, page, search)
		if err != nil {
			return errorResult(fmt.Sprintf("获取用例列表失败: %v", err)), nil
		}
	}

	// 如果不需要完整参数，返回精简数据
	if !full && cases.Testcases != nil {
		simplified := make([]map[string]interface{}, len(cases.Testcases))
		for i, tc := range cases.Testcases {
			simplified[i] = simplifyTestCase(tc)
		}
		result := map[string]interface{}{
			"page":      cases.Page,
			"total":     cases.Total,
			"testcases": simplified,
		}
		data, err := toJSON(result)
		if err != nil {
			return errorResult(fmt.Sprintf("序列化用例列表失败: %v", err)), nil
		}
		return mcp.NewToolResultText(data), nil
	}

	data, err := toJSON(cases)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化用例列表失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func getTestCaseHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	caseID, ok := request.Params.Arguments["id"].(string)
	if !ok {
		return errorResult("id is required"), nil
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	testcase, err := client.GetTestCase(token, caseID)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		testcase, err = client.GetTestCase(token, caseID)
		if err != nil {
			return errorResult(fmt.Sprintf("获取用例详情失败: %v", err)), nil
		}
	}

	data, err := toJSON(testcase)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化用例详情失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func createTestCaseHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	title, ok := request.Params.Arguments["title"].(string)
	if !ok {
		return errorResult("title is required"), nil
	}

	typeVal, ok := request.Params.Arguments["type"].(string)
	if !ok {
		return errorResult("type is required"), nil
	}

	// 解析步骤
	stepsRaw, ok := request.Params.Arguments["steps"].([]interface{})
	if !ok {
		return errorResult("steps is required"), nil
	}
	var steps []TestCaseStepRequest
	for _, s := range stepsRaw {
		if stepMap, ok := s.(map[string]interface{}); ok {
			step := TestCaseStepRequest{}
			if desc, ok := stepMap["desc"].(string); ok {
				step.Desc = desc
			}
			if expect, ok := stepMap["expect"].(string); ok {
				step.Expect = expect
			}
			steps = append(steps, step)
		}
	}

	reqBody := &CreateTestCaseRequest{
		Title: title,
		Type:  typeVal,
		Steps: steps,
	}

	if pri, ok := request.Params.Arguments["pri"].(float64); ok {
		reqBody.Pri = int(pri)
	}
	if stage, ok := request.Params.Arguments["stage"].(string); ok {
		reqBody.Stage = stage
	}
	if precondition, ok := request.Params.Arguments["precondition"].(string); ok {
		reqBody.Precondition = precondition
	}
	if branch, ok := request.Params.Arguments["branch"].(float64); ok {
		reqBody.Branch = int(branch)
	}
	if module, ok := request.Params.Arguments["module"].(float64); ok {
		reqBody.Module = int(module)
	}
	if story, ok := request.Params.Arguments["story"].(float64); ok {
		reqBody.Story = int(story)
	}
	if keywords, ok := request.Params.Arguments["keywords"].(string); ok {
		reqBody.Keywords = keywords
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
	testcase, err := client.CreateTestCase(token, productID, reqBody)
	if err != nil {
		token, refreshErr := globalTokenManager.RefreshToken()
		if refreshErr != nil {
			return errorResult(fmt.Sprintf("刷新Token失败: %v", refreshErr)), nil
		}
		testcase, err = client.CreateTestCase(token, productID, reqBody)
		if err != nil {
			return errorResult(fmt.Sprintf("创建用例失败: %v", err)), nil
		}
	}

	data, err := toJSON(testcase)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化用例信息失败: %v", err)), nil
	}

	return mcp.NewToolResultText(data), nil
}

func addBugCommentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !globalTokenManager.IsConfigured() {
		return errorResult("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码"), nil
	}

	bugID, ok := request.Params.Arguments["bug_id"].(string)
	if !ok {
		return errorResult("bug_id is required"), nil
	}

	comment, ok := request.Params.Arguments["comment"].(string)
	if !ok {
		return errorResult("comment is required"), nil
	}
	if strings.TrimSpace(comment) == "" {
		return errorResult("comment cannot be empty"), nil
	}

	var imagePaths []string
	if rawImagePaths, exists := request.Params.Arguments["image_paths"]; exists {
		items, ok := rawImagePaths.([]interface{})
		if !ok {
			return errorResult("image_paths must be an array of strings"), nil
		}
		imagePaths = make([]string, 0, len(items))
		for _, item := range items {
			imagePath, ok := item.(string)
			if !ok || strings.TrimSpace(imagePath) == "" {
				return errorResult("image_paths must contain only non-empty strings"), nil
			}
			imagePaths = append(imagePaths, imagePath)
		}
	}

	token, err := globalTokenManager.GetToken()
	if err != nil {
		return errorResult(fmt.Sprintf("获取Token失败: %v", err)), nil
	}

	client := NewZentaoClient(globalTokenManager.GetConfig().BaseURL)
	result, err := client.AddBugComment(token, bugID, comment, imagePaths)
	if err != nil {
		return errorResult(fmt.Sprintf("添加备注失败: %v", err)), nil
	}

	data, err := toJSON(result)
	if err != nil {
		return errorResult(fmt.Sprintf("序列化响应失败: %v", err)), nil
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
		"id":          b.ID,
		"name":        b.Name,
		"date":        b.Date,
		"builder":     b.Builder,
		"productName": b.ProductName,
	}
}

// simplifyTestCase 精简用例信息
func simplifyTestCase(tc TestCaseListItem) map[string]interface{} {
	result := map[string]interface{}{
		"id":     tc.ID,
		"title":  tc.Title,
		"pri":    tc.Pri,
		"type":   tc.Type,
		"status": tc.Status,
	}
	if tc.OpenedBy != nil {
		result["openedBy"] = tc.OpenedBy.Account
	}
	return result
}

// simplifyStory 精简需求信息
func simplifyStory(s Story) map[string]interface{} {
	result := map[string]interface{}{
		"id":       s.ID,
		"title":    s.Title,
		"category": s.Category,
		"pri":      s.Pri,
		"status":   s.Status,
		"stage":    s.Stage,
		"estimate": s.Estimate,
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
