package config

import (
	"fmt"
	
	"github.com/manteia/zhulong/pkg/storage"
)

// ConfigInterface 配置管理接口
type ConfigInterface interface {
	// 加载配置
	Load() error
	
	// 验证配置
	Validate() error
	
	// 获取存储配置
	GetStorageConfig() storage.Config
	
	// 获取服务器配置
	GetServerConfig() ServerConfig
	
	// 获取应用配置
	GetAppConfig() AppConfig
	
	// 重新加载配置
	Reload() error
}

// WatcherInterface 配置监听器接口
type WatcherInterface interface {
	// 启动监听
	Watch(changes chan<- *Config) error
	
	// 停止监听
	Stop() error
}

// Manager 配置管理器
type Manager struct {
	config     *Config
	configFile string
	watcher    WatcherInterface
}

// NewManager 创建配置管理器
func NewManager(configFile string) *Manager {
	return &Manager{
		configFile: configFile,
	}
}

// Load 加载配置
func (m *Manager) Load() error {
	config, err := LoadFromFile(m.configFile)
	if err != nil {
		return err
	}
	
	if err := config.Validate(); err != nil {
		return err
	}
	
	m.config = config
	return nil
}

// Validate 验证配置
func (m *Manager) Validate() error {
	if m.config == nil {
		return fmt.Errorf("配置未加载")
	}
	return m.config.Validate()
}

// GetStorageConfig 获取存储配置
func (m *Manager) GetStorageConfig() storage.Config {
	if m.config == nil {
		return nil
	}
	return m.config.GetStorageConfig()
}

// GetServerConfig 获取服务器配置
func (m *Manager) GetServerConfig() ServerConfig {
	if m.config == nil {
		return ServerConfig{}
	}
	return m.config.Server
}

// GetAppConfig 获取应用配置
func (m *Manager) GetAppConfig() AppConfig {
	if m.config == nil {
		return AppConfig{}
	}
	return m.config.App
}

// Reload 重新加载配置
func (m *Manager) Reload() error {
	return m.Load()
}

// StartWatching 启动配置监听
func (m *Manager) StartWatching(changes chan<- *Config) error {
	watcher, err := NewConfigWatcher(m.configFile)
	if err != nil {
		return err
	}
	
	m.watcher = watcher
	return watcher.Watch(changes)
}

// StopWatching 停止配置监听
func (m *Manager) StopWatching() error {
	if m.watcher == nil {
		return nil
	}
	return m.watcher.Stop()
}

// GetConfig 获取当前配置
func (m *Manager) GetConfig() *Config {
	return m.config
}