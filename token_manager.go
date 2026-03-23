package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 禅道配置
type Config struct {
	BaseURL        string `json:"base_url"`
	Account        string `json:"account"`
	Password       string `json:"password"`
	TokenExpiry    int    `json:"token_expiry"`    // Token过期时间（秒），默认86400秒（24小时）
	DefaultProduct string `json:"default_product"` // 默认产品ID或名称
}

// TokenCache Token缓存
type TokenCache struct {
	Token      string    `json:"token"`
	ExpireTime time.Time `json:"expire_time"`
}

// TokenManager Token管理器
type TokenManager struct {
	config     *Config
	cache      *TokenCache
	client     *ZentaoClient
	mu         sync.RWMutex
	configPath string
}

// NewTokenManager 创建Token管理器
func NewTokenManager() *TokenManager {
	return &TokenManager{}
}

// SetConfigPath 设置配置文件路径
func (tm *TokenManager) SetConfigPath(configPath string) {
	if configPath != "" {
		tm.configPath = configPath
	} else {
		// 未指定则使用 exe 所在目录
		execPath, err := os.Executable()
		if err != nil {
			execPath, _ = os.Getwd()
		}
		configDir := filepath.Dir(execPath)
		tm.configPath = filepath.Join(configDir, "zentao_config.json")
	}
}

// GetConfigPath 获取配置文件路径
func (tm *TokenManager) GetConfigPath() string {
	return tm.configPath
}

// LoadConfig 加载配置
func (tm *TokenManager) LoadConfig() error {
	data, err := os.ReadFile(tm.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w (路径: %s)", err, tm.configPath)
	}

	if err := json.Unmarshal(data, &tm.config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认过期时间
	if tm.config.TokenExpiry <= 0 {
		tm.config.TokenExpiry = 86400
	}

	tm.client = NewZentaoClient(tm.config.BaseURL)
	return nil
}

// SaveConfig 保存配置
func (tm *TokenManager) SaveConfig(config *Config) error {
	tm.config = config
	if tm.config.TokenExpiry <= 0 {
		tm.config.TokenExpiry = 86400
	}

	data, err := json.MarshalIndent(tm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(tm.configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	tm.client = NewZentaoClient(tm.config.BaseURL)
	return nil
}

// GetConfig 获取当前配置
func (tm *TokenManager) GetConfig() *Config {
	return tm.config
}

// IsConfigured 检查是否已配置
func (tm *TokenManager) IsConfigured() bool {
	return tm.config != nil && tm.config.BaseURL != "" && tm.config.Account != "" && tm.config.Password != ""
}

// GetToken 获取Token（自动刷新）
func (tm *TokenManager) GetToken() (string, error) {
	tm.mu.RLock()
	if tm.cache != nil && time.Now().Before(tm.cache.ExpireTime) {
		token := tm.cache.Token
		tm.mu.RUnlock()
		return token, nil
	}
	tm.mu.RUnlock()

	// 需要刷新Token
	return tm.RefreshToken()
}

// RefreshToken 强制刷新Token
func (tm *TokenManager) RefreshToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.IsConfigured() {
		return "", fmt.Errorf("禅道未配置，请先调用 configure 工具设置服务器地址和账号密码")
	}

	token, err := tm.client.GetToken(tm.config.Account, tm.config.Password)
	if err != nil {
		return "", err
	}

	// 计算过期时间，提前10分钟刷新
	expiry := time.Duration(tm.config.TokenExpiry) * time.Second
	tm.cache = &TokenCache{
		Token:      token,
		ExpireTime: time.Now().Add(expiry - 10*time.Minute),
	}

	return token, nil
}

// GetTokenInfo 获取Token信息
func (tm *TokenManager) GetTokenInfo() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := map[string]interface{}{
		"configured": tm.IsConfigured(),
	}

	if tm.config != nil {
		result["base_url"] = tm.config.BaseURL
		result["account"] = tm.config.Account
		result["token_expiry_seconds"] = tm.config.TokenExpiry
	}

	if tm.cache != nil {
		result["has_token"] = true
		result["expire_time"] = tm.cache.ExpireTime.Format(time.RFC3339)
		result["expired"] = time.Now().After(tm.cache.ExpireTime)
	} else {
		result["has_token"] = false
	}

	return result
}

// 全局Token管理器
var globalTokenManager = NewTokenManager()