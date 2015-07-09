package app

import (
	json "github.com/orofarne/strict-json"
)

type CommonConfig struct {
	MetaSize          uint           // 8 by default
	TileSize          uint           // AxA tile
	MonitoringPlugins []PluginConfig // Monitoring exporter
}

type PluginConfig struct {
	Plugin       string          // Plugin name
	PluginConfig json.RawMessage // Plugin JSON configuration
}

type RenderPoolConfig struct {
	Cmd         		[]string // Render slave binary
	MinZoom     		uint
	MaxZoom     		uint
	Tags        		[]string
	PoolSize    		uint
	HPQueueSize 		uint
	LPQueueSize 		uint
	RenderTTL   		uint
	ExecutionTimeout	string
}

type RenderPoolsConfig struct {
	RenderPools []RenderPoolConfig
}
