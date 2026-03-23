package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TokenRequest 获取Token的请求体
type TokenRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// TokenResponse 获取Token的响应
type TokenResponse struct {
	Token string `json:"token"`
}

// User 用户信息
type User struct {
	ID       interface{} `json:"id"`
	Account  string      `json:"account"`
	Avatar   string      `json:"avatar"`
	Realname string      `json:"realname"`
}

// Product 产品信息
type Product struct {
	ID             interface{}    `json:"id"`
	Program        interface{}    `json:"program"`
	Name           string         `json:"name"`
	Code           string         `json:"code"`
	Line           interface{}    `json:"line"`
	PO             *User          `json:"PO,omitempty"`
	QD             *User          `json:"QD,omitempty"`
	RD             *User          `json:"RD,omitempty"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	Desc           string         `json:"desc"`
	ACL            string         `json:"acl"`
	Whitelist      []User         `json:"whitelist,omitempty"`
	CreatedBy      *User          `json:"createdBy,omitempty"`
	CreatedDate    string         `json:"createdDate"`
	CreatedVersion string         `json:"createdVersion"`
	ProgramName    string         `json:"programName"`
	Stories        map[string]any `json:"stories,omitempty"`
	Plans          interface{}    `json:"plans"`
	Releases       interface{}    `json:"releases"`
	Bugs           interface{}    `json:"bugs"`
	Progress       interface{}    `json:"progress"`
}

// ProductListResponse 产品列表响应
type ProductListResponse struct {
	Total    int       `json:"total"`
	Products []Product `json:"products"`
}

// ProductDetail 产品详情
type ProductDetail struct {
	ID             interface{}    `json:"id"`
	Program        interface{}    `json:"program"`
	Name           string         `json:"name"`
	Code           string         `json:"code"`
	Bind           string         `json:"bind"`
	Line           interface{}    `json:"line"`
	Type           string         `json:"type"`
	Status         string         `json:"status"`
	SubStatus      string         `json:"subStatus"`
	Desc           string         `json:"desc"`
	PO             *User          `json:"PO,omitempty"`
	QD             *User          `json:"QD,omitempty"`
	RD             *User          `json:"RD,omitempty"`
	ACL            string         `json:"acl"`
	Whitelist      []interface{}  `json:"whitelist"`
	Reviewer       string         `json:"reviewer"`
	CreatedBy      *User          `json:"createdBy,omitempty"`
	CreatedDate    string         `json:"createdDate"`
	CreatedVersion string         `json:"createdVersion"`
	Order          interface{}    `json:"order"`
	Deleted        string         `json:"deleted"`
	Stories        map[string]any `json:"stories,omitempty"`
	Plans          interface{}    `json:"plans"`
	Releases       interface{}    `json:"releases"`
	Builds         interface{}    `json:"builds"`
	Cases          interface{}    `json:"cases"`
	Projects       interface{}    `json:"projects"`
	Executions     interface{}    `json:"executions"`
	Bugs           interface{}    `json:"bugs"`
	Docs           interface{}    `json:"docs"`
	Progress       interface{}    `json:"progress"`
	CaseReview     bool           `json:"caseReview"`
}

// CreateProductRequest 创建产品请求
type CreateProductRequest struct {
	Name      string `json:"name"`
	Program   int    `json:"program,omitempty"`
	Code      string `json:"code"`
	Line      int    `json:"line,omitempty"`
	PO        string `json:"PO,omitempty"`
	QD        string `json:"QD,omitempty"`
	RD        string `json:"RD,omitempty"`
	Type      string `json:"type,omitempty"`
	Desc      string `json:"desc,omitempty"`
	ACL       string `json:"acl,omitempty"`
	Whitelist []int  `json:"whitelist,omitempty"`
}

// CreateProductResponse 创建产品响应
type CreateProductResponse struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Program      string         `json:"program"`
	Code         string         `json:"code"`
	Line         string         `json:"line"`
	Type         string         `json:"type"`
	Status       string         `json:"status"`
	PO           interface{}    `json:"PO"`
	QD           interface{}    `json:"QD"`
	RD           interface{}    `json:"RD"`
	ACL          string         `json:"acl"`
	Whitelist    []interface{}  `json:"whitelist"`
	CreatedBy    *User          `json:"createdBy"`
	CreatedDate  string         `json:"createdDate"`
	Desc         string         `json:"desc"`
}

// ErrorResponse API错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// BugRequest 创建/修改Bug请求
type BugRequest struct {
	Branch      int      `json:"branch,omitempty"`
	Module      int      `json:"module,omitempty"`
	Execution   int      `json:"execution,omitempty"`
	Title       string   `json:"title"`
	Keywords    string   `json:"keywords,omitempty"`
	Severity    int      `json:"severity"`
	Pri         int      `json:"pri"`
	Type        string   `json:"type"`
	OS          string   `json:"os,omitempty"`
	Browser     string   `json:"browser,omitempty"`
	Steps       string   `json:"steps,omitempty"`
	Task        int      `json:"task,omitempty"`
	Story       int      `json:"story,omitempty"`
	Deadline    string   `json:"deadline,omitempty"`
	OpenedBuild []string `json:"openedBuild,omitempty"`
}

// BugStatus Bug状态
type BugStatus struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// BugListItem Bug列表项
type BugListItem struct {
	ID             interface{} `json:"id"`
	Product        interface{} `json:"product"`
	Branch         interface{} `json:"branch"`
	Module         interface{} `json:"module"`
	Project        interface{} `json:"project"`
	Execution      interface{} `json:"execution"`
	Plan           interface{} `json:"plan"`
	Story          interface{} `json:"story"`
	StoryVersion   interface{} `json:"storyVersion"`
	Task           interface{} `json:"task"`
	ToTask         interface{} `json:"toTask"`
	ToStory        interface{} `json:"toStory"`
	Title          string      `json:"title"`
	Keywords       string      `json:"keywords"`
	Severity       interface{} `json:"severity"`
	Pri            interface{} `json:"pri"`
	Type           string      `json:"type"`
	OS             string      `json:"os"`
	Browser        string      `json:"browser"`
	Hardware       string      `json:"hardware"`
	Found          string      `json:"found"`
	Steps          string      `json:"steps"`
	Status         interface{} `json:"status"`
	SubStatus      string      `json:"subStatus"`
	Color          string      `json:"color"`
	Confirmed      interface{} `json:"confirmed"`
	ActivatedCount interface{} `json:"activatedCount"`
	ActivatedDate  string      `json:"activatedDate"`
	FeedbackBy     string      `json:"feedbackBy"`
	NotifyEmail    string      `json:"notifyEmail"`
	Mailto         interface{} `json:"mailto"`
	OpenedBy       *User       `json:"openedBy,omitempty"`
	OpenedDate     string      `json:"openedDate"`
	OpenedBuild    interface{} `json:"openedBuild"`
	AssignedTo     *User       `json:"assignedTo,omitempty"`
	AssignedDate   string      `json:"assignedDate"`
	Deadline       interface{} `json:"deadline"`
	ResolvedBy     *User       `json:"resolvedBy,omitempty"`
	Resolution     string      `json:"resolution"`
	ResolvedBuild  string      `json:"resolvedBuild"`
	ResolvedDate   string      `json:"resolvedDate"`
	ClosedBy       *User       `json:"closedBy,omitempty"`
	ClosedDate     string      `json:"closedDate"`
	DuplicateBug   interface{} `json:"duplicateBug"`
	LinkBug        string      `json:"linkBug"`
	Case           interface{} `json:"case"`
	CaseVersion    interface{} `json:"caseVersion"`
	Result         interface{} `json:"result"`
	Repo           interface{} `json:"repo"`
	Entry          string      `json:"entry"`
	Lines          string      `json:"lines"`
	V1             string      `json:"v1"`
	V2             string      `json:"v2"`
	RepoType       string      `json:"repoType"`
	Testtask       interface{} `json:"testtask"`
	LastEditedBy   *User       `json:"lastEditedBy,omitempty"`
	LastEditedDate string      `json:"lastEditedDate"`
	Deleted        interface{} `json:"deleted"`
	NeedConfirm    interface{} `json:"needconfirm"`
}

// BugListResponse Bug列表响应
type BugListResponse struct {
	Page  int          `json:"page"`
	Total int          `json:"total"`
	Limit int          `json:"limit"`
	Bugs  []BugListItem `json:"bugs"`
}

// Bug Bug信息
type Bug struct {
	ID             interface{} `json:"id"`
	Product        interface{} `json:"product"`
	Branch         interface{} `json:"branch"`
	Module         interface{} `json:"module"`
	Project        interface{} `json:"project"`
	Execution      interface{} `json:"execution"`
	Plan           interface{} `json:"plan"`
	Story          interface{} `json:"story"`
	StoryVersion   interface{} `json:"storyVersion"`
	Task           interface{} `json:"task"`
	ToTask         interface{} `json:"toTask"`
	ToStory        interface{} `json:"toStory"`
	Title          string      `json:"title"`
	Keywords       string      `json:"keywords"`
	Severity       interface{} `json:"severity"`
	Pri            interface{} `json:"pri"`
	Type           string      `json:"type"`
	OS             string      `json:"os"`
	Browser        string      `json:"browser"`
	Hardware       string      `json:"hardware"`
	Found          string      `json:"found"`
	Steps          string      `json:"steps"`
	Status         string      `json:"status"`
	SubStatus      string      `json:"subStatus"`
	Color          string      `json:"color"`
	Confirmed      interface{} `json:"confirmed"`
	ActivatedCount interface{} `json:"activatedCount"`
	ActivatedDate  string      `json:"activatedDate"`
	FeedbackBy     string      `json:"feedbackBy"`
	NotifyEmail    string      `json:"notifyEmail"`
	Mailto         []string    `json:"mailto"`
	OpenedBy       *User       `json:"openedBy,omitempty"`
	OpenedDate     string      `json:"openedDate"`
	OpenedBuild    interface{} `json:"openedBuild"`
	AssignedTo     *User       `json:"assignedTo,omitempty"`
	AssignedDate   string      `json:"assignedDate"`
	Deadline       interface{} `json:"deadline"`
	ResolvedBy     *User       `json:"resolvedBy,omitempty"`
	Resolution     string      `json:"resolution"`
	ResolvedBuild  string      `json:"resolvedBuild"`
	ResolvedDate   interface{} `json:"resolvedDate"`
	ClosedBy       *User       `json:"closedBy,omitempty"`
	ClosedDate     string      `json:"closedDate"`
	DuplicateBug   interface{} `json:"duplicateBug"`
	LinkBug        string      `json:"linkBug"`
	Case           interface{} `json:"case"`
	CaseVersion    interface{} `json:"caseVersion"`
	Result         interface{} `json:"result"`
	Repo           interface{} `json:"repo"`
	Entry          string      `json:"entry"`
	Lines          string      `json:"lines"`
	V1             string      `json:"v1"`
	V2             string      `json:"v2"`
	RepoType       string      `json:"repoType"`
	Testtask       interface{} `json:"testtask"`
	LastEditedBy   *User       `json:"lastEditedBy,omitempty"`
	LastEditedDate string      `json:"lastEditedDate"`
	Deleted        interface{} `json:"deleted"`
	ExecutionName  string      `json:"executionName"`
	StoryTitle     string      `json:"storyTitle"`
	StoryStatus    string      `json:"storyStatus"`
	TaskName       interface{} `json:"taskName"`
	PlanName       interface{} `json:"planName"`
	ProjectName    string      `json:"projectName"`
	ToCases        []string    `json:"toCases"`
	Files          []string    `json:"files"`
}

// ZentaoClient 禅道客户端
type ZentaoClient struct {
	baseURL    string
	httpClient *http.Client
}

// apiPrefix API路径前缀
const apiPrefix = "/api.php/v1"

// NewZentaoClient 创建禅道客户端
func NewZentaoClient(baseURL string) *ZentaoClient {
	return &ZentaoClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken 获取禅道API Token
func GetToken(baseURL, account, password string) (string, error) {
	client := NewZentaoClient(baseURL)
	return client.GetToken(account, password)
}

// GetToken 获取禅道API Token
func (c *ZentaoClient) GetToken(account, password string) (string, error) {
	url := fmt.Sprintf("%s%s/tokens", c.baseURL, apiPrefix)

	reqBody := TokenRequest{
		Account:  account,
		Password: password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if tokenResp.Token == "" {
		return "", fmt.Errorf("响应中未包含token")
	}

	return tokenResp.Token, nil
}

// GetProducts 获取产品列表
func (c *ZentaoClient) GetProducts(token string) (*ProductListResponse, error) {
	url := fmt.Sprintf("%s%s/products", c.baseURL, apiPrefix)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var productResp ProductListResponse
	if err := json.Unmarshal(body, &productResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &productResp, nil
}

// GetProduct 获取产品详情
func (c *ZentaoClient) GetProduct(token string, productID string) (*ProductDetail, error) {
	url := fmt.Sprintf("%s%s/products/%s", c.baseURL, apiPrefix, productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var product ProductDetail
	if err := json.Unmarshal(body, &product); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &product, nil
}

// CreateProduct 创建产品
func (c *ZentaoClient) CreateProduct(token string, reqBody *CreateProductRequest) (*CreateProductResponse, error) {
	url := fmt.Sprintf("%s%s/products", c.baseURL, apiPrefix)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var productResp CreateProductResponse
	if err := json.Unmarshal(body, &productResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &productResp, nil
}

// CreateBug 创建Bug
func (c *ZentaoClient) CreateBug(token string, productID string, reqBody *BugRequest) (*Bug, error) {
	url := fmt.Sprintf("%s%s/products/%s/bugs", c.baseURL, apiPrefix, productID)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var bug Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &bug, nil
}

// UpdateBug 修改Bug
func (c *ZentaoClient) UpdateBug(token string, bugID string, reqBody *BugRequest) (*Bug, error) {
	url := fmt.Sprintf("%s%s/bugs/%s", c.baseURL, apiPrefix, bugID)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var bug Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &bug, nil
}
// GetBugs 获取产品Bug列表
func (c *ZentaoClient) GetBugs(token string, productID string) (*BugListResponse, error) {
	url := fmt.Sprintf("%s%s/products/%s/bugs", c.baseURL, apiPrefix, productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var bugList BugListResponse
	if err := json.Unmarshal(body, &bugList); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &bugList, nil
}

// GetBug 获取Bug详情
func (c *ZentaoClient) GetBug(token string, bugID string) (*Bug, error) {
	url := fmt.Sprintf("%s%s/bugs/%s", c.baseURL, apiPrefix, bugID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查是否是错误响应
	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var bug Bug
	if err := json.Unmarshal(body, &bug); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &bug, nil
}

// BuildRequest 创建/修改版本请求
type BuildRequest struct {
	Execution int    `json:"execution"`
	Product   int    `json:"product"`
	Branch    int    `json:"branch,omitempty"`
	Name      string `json:"name"`
	Builder   string `json:"builder"`
	Date      string `json:"date,omitempty"`
	ScmPath   string `json:"scmPath,omitempty"`
	FilePath  string `json:"filePath,omitempty"`
	Desc      string `json:"desc,omitempty"`
}

// Build 版本信息
type Build struct {
	ID            interface{}    `json:"id"`
	Project       interface{}    `json:"project"`
	Product       interface{}    `json:"product"`
	Branch        interface{}    `json:"branch"`
	Execution     interface{}    `json:"execution"`
	Name          string         `json:"name"`
	ScmPath       string         `json:"scmPath"`
	FilePath      string         `json:"filePath"`
	Date          string         `json:"date"`
	Stories       interface{}    `json:"stories"`
	Bugs          interface{}    `json:"bugs"`
	Builder       interface{}    `json:"builder"`
	Desc          string         `json:"desc"`
	Deleted       interface{}    `json:"deleted"`
	ExecutionName string         `json:"executionName"`
	ExecutionID   interface{}    `json:"executionID"`
	ProductName   string         `json:"productName"`
	ProductType   string         `json:"productType"`
	BranchName    string         `json:"branchName"`
	Files         []interface{}  `json:"files"`
}

// BuildListResponse 版本列表响应
type BuildListResponse struct {
	Total  int     `json:"total"`
	Builds []Build `json:"builds"`
}

// CreateBuild 创建版本
func (c *ZentaoClient) CreateBuild(token string, projectID string, reqBody *BuildRequest) (*Build, error) {
	url := fmt.Sprintf("%s%s/projects/%s/builds", c.baseURL, apiPrefix, projectID)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var build Build
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &build, nil
}

// UpdateBuild 修改版本
func (c *ZentaoClient) UpdateBuild(token string, buildID string, reqBody *BuildRequest) (*Build, error) {
	url := fmt.Sprintf("%s%s/builds/%s", c.baseURL, apiPrefix, buildID)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var build Build
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &build, nil
}

// GetBuild 获取版本详情
func (c *ZentaoClient) GetBuild(token string, buildID string) (*Build, error) {
	url := fmt.Sprintf("%s%s/builds/%s", c.baseURL, apiPrefix, buildID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var build Build
	if err := json.Unmarshal(body, &build); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &build, nil
}

// GetBuilds 获取项目版本列表
func (c *ZentaoClient) GetBuilds(token string, projectID string) (*BuildListResponse, error) {
	url := fmt.Sprintf("%s%s/projects/%s/builds", c.baseURL, apiPrefix, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var buildList BuildListResponse
	if err := json.Unmarshal(body, &buildList); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &buildList, nil
}

// StoryRequest 创建需求请求
type StoryRequest struct {
	Title      string   `json:"title"`
	Product    int      `json:"product"`
	Pri        int      `json:"pri"`
	Category   string   `json:"category"`
	Spec       string   `json:"spec,omitempty"`
	Verify     string   `json:"verify,omitempty"`
	Source     string   `json:"source,omitempty"`
	SourceNote string   `json:"sourceNote,omitempty"`
	Estimate   float64  `json:"estimate,omitempty"`
	Keywords   string   `json:"keywords,omitempty"`
	Reviewer   []string `json:"reviewer,omitempty"`
	AssignedTo string   `json:"assignedTo,omitempty"`
}

// Story 需求信息
type Story struct {
	ID             interface{}   `json:"id"`
	Parent         interface{}   `json:"parent"`
	Product        interface{}   `json:"product"`
	Branch         interface{}   `json:"branch"`
	Module         interface{}   `json:"module"`
	Plan           interface{}   `json:"plan"`
	Source         string        `json:"source"`
	SourceNote     string        `json:"sourceNote"`
	FromBug        interface{}   `json:"fromBug"`
	Title          string        `json:"title"`
	Keywords       string        `json:"keywords"`
	Type           string        `json:"type"`
	Category       string        `json:"category"`
	Pri            interface{}   `json:"pri"`
	Estimate       interface{}   `json:"estimate"`
	Status         string        `json:"status"`
	SubStatus      string        `json:"subStatus"`
	Color          string        `json:"color"`
	Stage          string        `json:"stage"`
	StagedBy       string        `json:"stagedBy"`
	Mailto         interface{}   `json:"mailto"`
	OpenedBy       interface{}   `json:"openedBy"`
	OpenedDate     string        `json:"openedDate"`
	AssignedTo     interface{}   `json:"assignedTo"`
	AssignedDate   interface{}   `json:"assignedDate"`
	LastEditedBy   interface{}   `json:"lastEditedBy"`
	LastEditedDate interface{}   `json:"lastEditedDate"`
	ReviewedBy     interface{}   `json:"reviewedBy"`
	ReviewedDate   interface{}   `json:"reviewedDate"`
	ClosedBy       interface{}   `json:"closedBy"`
	ClosedDate     interface{}   `json:"closedDate"`
	ClosedReason   string        `json:"closedReason"`
	ToBug          interface{}   `json:"toBug"`
	ChildStories   string        `json:"childStories"`
	LinkStories    string        `json:"linkStories"`
	DuplicateStory interface{}   `json:"duplicateStory"`
	Version        interface{}   `json:"version"`
	URChanged      string        `json:"URChanged"`
	Deleted        interface{}   `json:"deleted"`
	Spec           string        `json:"spec"`
	Verify         string        `json:"verify"`
	Executions     []interface{} `json:"executions"`
	Tasks          []interface{} `json:"tasks"`
	Stages         []interface{} `json:"stages"`
	Children       []interface{} `json:"children"`
	PlanTitle      string        `json:"planTitle"`
}

// StoryListResponse 需求列表响应
type StoryListResponse struct {
	Page    int     `json:"page"`
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Stories []Story `json:"stories"`
}

// CreateStory 创建需求
func (c *ZentaoClient) CreateStory(token string, reqBody *StoryRequest) (*Story, error) {
	url := fmt.Sprintf("%s%s/stories", c.baseURL, apiPrefix)

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("JSON编码失败: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var story Story
	if err := json.Unmarshal(body, &story); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &story, nil
}

// GetStory 获取需求详情
func (c *ZentaoClient) GetStory(token string, storyID string) (*Story, error) {
	url := fmt.Sprintf("%s%s/stories/%s", c.baseURL, apiPrefix, storyID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var story Story
	if err := json.Unmarshal(body, &story); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &story, nil
}

// GetProjectStories 获取项目需求列表
func (c *ZentaoClient) GetProjectStories(token string, projectID string) (*StoryListResponse, error) {
	url := fmt.Sprintf("%s%s/projects/%s/stories", c.baseURL, apiPrefix, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var storyList StoryListResponse
	if err := json.Unmarshal(body, &storyList); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &storyList, nil
}

// GetProductStories 获取产品需求列表
func (c *ZentaoClient) GetProductStories(token string, productID string) (*StoryListResponse, error) {
	url := fmt.Sprintf("%s%s/products/%s/stories", c.baseURL, apiPrefix, productID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var storyList StoryListResponse
	if err := json.Unmarshal(body, &storyList); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &storyList, nil
}

// GetExecutionStories 获取执行需求列表
func (c *ZentaoClient) GetExecutionStories(token string, executionID string) (*StoryListResponse, error) {
	url := fmt.Sprintf("%s%s/executions/%s/stories", c.baseURL, apiPrefix, executionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var errResp ErrorResponse
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return nil, fmt.Errorf("API错误: %s", errResp.Error)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var storyList StoryListResponse
	if err := json.Unmarshal(body, &storyList); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &storyList, nil
}
