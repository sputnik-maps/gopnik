package gopnikprerenderlib

import (
	"app"

	json "github.com/orofarne/strict-json"
)

type PrerenderConfig struct {
	Threads       int              // Use >= n threads (-1 for NumCPU)
	NodeQueueSize int              // Number of parallel requests per node
	DebugAddr     string           // Address for statistics
	UIAddr        string           // Bind WebUI addr
	Logging       json.RawMessage  // see loghelper.go
	PerfLog       string           // Performance log file
	Slaves        app.PluginConfig // Cluster of slaves
}

type PrerenderGlobalConfig struct {
	Prerender             PrerenderConfig  // Prerender config
	CachePlugin           app.PluginConfig //
	app.CommonConfig                       //
	app.RenderPoolsConfig                  //
	json.OtherKeys                         //
}
