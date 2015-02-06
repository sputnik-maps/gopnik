package gopnikprerenderlib

import (
	"app"

	json "github.com/orofarne/strict-json"
)

type PrerenderSlaveConfig struct {
	SaverPoolSize int             // Render pool size
	Threads       int             // Use >= n threads (-1 for NumCPU)
	RPCAddr       string          // Bind RPC addr
	DebugAddr     string          // Address for statistics
	Logging       json.RawMessage // see loghelper.go
	PerfLog       string          // Performance log file
}

type PrerenderConfig struct {
	Threads   int              // Use >= n threads (-1 for NumCPU)
	DebugAddr string           // Address for statistics
	UIAddr    string           // Bind WebUI addr
	Logging   json.RawMessage  // see loghelper.go
	PerfLog   string           // Performance log file
	Slaves    app.PluginConfig // Cluster of slaves
}

type PrerenderGlobalConfig struct {
	Prerender             PrerenderConfig      // Prerender config
	PrerenderSlave        PrerenderSlaveConfig // Prerender slave config
	CachePlugin           app.PluginConfig     //
	app.CommonConfig                           //
	app.RenderPoolsConfig                      //
	json.OtherKeys                             //
}
