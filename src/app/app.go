package app

import (
	"flag"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/pprof"
	"time"

	"gopnik"
	"loghelper"
	"plugins"
	"program_version"
	"xmlstatus"

	"github.com/op/go-logging"
	"github.com/orofarne/hmetrics2"
	"github.com/orofarne/hmetrics2/expvarexport"
	json "github.com/orofarne/strict-json"
)

var log = logging.MustGetLogger("global")

type app struct {
	metatiler    *gopnik.Metatiler // Metatile converter
	appName      string            // Application name (in config), e.g. Render
	config       interface{}       // Common config
	appConfig    interface{}       // Application config
	commonConfig CommonConfig      // Common config
}

var App app

func init() {
	App.configureDefault()
}

func (self *app) configureDefault() {
	self.metatiler = gopnik.NewMetatiler(8, 256)
}

func (self *app) Configure(appName string, defaultCofig interface{}) {
	self.appName = appName

	// Read config
	self.config = defaultCofig
	if err := self.parseFlags(); err != nil {
		stdlog.Fatal(err)
	}

	// Application section
	if self.appConfig != nil {
		// Set up logger
		if err := self.setupLogger(); err != nil {
			stdlog.Fatal(err)
		}

		// Set up GOMAXPROCS
		if err := self.setupGOMAXPROCS(); err != nil {
			stdlog.Fatal(err)
		}

		// Set up monitoring
		if err := self.setupMonitoring(); err != nil {
			stdlog.Fatal(err)
		}
	}

	// Common section
	// Set up metatiler
	if err := self.setupMetatiler(); err != nil {
		stdlog.Fatal(err)
	}
}

func (self *app) Metatiler() *gopnik.Metatiler {
	return self.metatiler
}

// Private methods

// Parse command line args
func (self *app) parseFlags() error {
	configFile := flag.String("config", "", "Config file")
	showVersion := flag.Bool("version", false, "Show version")
	// Profiling stuff ... from http://blog.golang.org/profiling-go-programs
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	// -version
	if *showVersion {
		fmt.Printf("Version: %v\n", program_version.GetVersion())
		os.Exit(1)
	}

	// -config "cfg.json"
	if *configFile != "" {
		f, err := os.Open(*configFile)
		if err != nil {
			return fmt.Errorf("Failed to open file '%s': %v", *configFile, err)
		}
		defer f.Close()
		if err := self.parseConfig(f); err != nil {
			return err
		}
	}

	// -cpuprofile "filename.prof"
	if *cpuprofile != "" {
		if err := self.startCpuProfile(*cpuprofile); err != nil {
			return err
		}
	}

	return nil
}

// Parse configuration
func (self *app) parseConfig(input io.Reader) error {
	// Decode JSON
	dec := json.NewDecoder(input)
	if err := dec.Decode(self.config); err != nil {
		return fmt.Errorf("JSON error: %v", err)
	}

	// Validate config
	cfgStPtr := reflect.ValueOf(self.config)
	if cfgStPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("Invalid config struct ref")
	}
	cfgSt := cfgStPtr.Elem()
	if cfgSt.Kind() != reflect.Struct {
		return fmt.Errorf("Invalid config struct")
	}

	// Parse application config
	if self.appName != "" {
		appCfg := cfgSt.FieldByName(self.appName)
		if !appCfg.IsValid() {
			return fmt.Errorf(`Field "%s" not found in config`, self.appName)
		}
		if appCfg.Kind() != reflect.Struct {
			return fmt.Errorf(`Invalid field "%s" struct`, self.appName)
		}
		self.appConfig = appCfg.Interface()
	}

	// Parse common config
	cCfg := cfgSt.FieldByName("CommonConfig")
	if !cCfg.IsValid() {
		return fmt.Errorf(`Field "CommonConfig" not found`)
	}
	cCfgConv, cCfgOk := cCfg.Interface().(CommonConfig)
	if !cCfgOk {
		return fmt.Errorf("Config is invalid")
	}
	self.commonConfig = cCfgConv

	return nil
}

// Set up logger
func (self *app) setupLogger() error {
	appCfg := reflect.ValueOf(self.appConfig)
	logCfg := appCfg.FieldByName("Logging")
	if !logCfg.IsValid() {
		return fmt.Errorf(`Field "Logging" not found in "%s"`, self.appName)
	}
	logCfgRaw, logCfgOk := logCfg.Interface().(json.RawMessage)
	if !logCfgOk {
		return fmt.Errorf(`Invalid field "Logging" struct`)
	}
	if err := loghelper.IntiLog(logCfgRaw); err != nil {
		return fmt.Errorf("Failed to initialize logging: %v", err)
	}
	return nil
}

// Set up metatiler
func (self *app) setupMetatiler() error {
	if self.commonConfig.MetaSize < 1 {
		return fmt.Errorf("Invalid MetaSize %v", self.commonConfig.MetaSize)
	}
	if self.commonConfig.TileSize < 1 {
		return fmt.Errorf("Invalid TileSize %v", self.commonConfig.TileSize)
	}

	self.metatiler = gopnik.NewMetatiler(
		uint64(self.commonConfig.MetaSize),
		uint64(self.commonConfig.TileSize))
	return nil
}

// Set up GOMAXPROCS
func (self *app) setupGOMAXPROCS() error {
	threadsCfg := reflect.ValueOf(self.appConfig).FieldByName("Threads")
	if threadsCfg.IsValid() {
		threads, ok := threadsCfg.Interface().(int)
		if !ok {
			return fmt.Errorf(`Field "Threads" in "%s" is invalid`,
				self.appName)
		}
		if threads < 0 {
			threads = runtime.NumCPU()
		}
		runtime.GOMAXPROCS(threads)
	}
	return nil
}

// Set up monitoring
func (self *app) setupMonitoring() error {
	// Set up debug HTTP interface
	debugAddrCfg := reflect.ValueOf(self.appConfig).FieldByName("DebugAddr")
	if debugAddrCfg.IsValid() {
		debugAddr, ok := debugAddrCfg.Interface().(string)
		if !ok {
			return fmt.Errorf(`Field "DebugAddr" in "%s" is invalid`,
				self.appName)
		}
		if debugAddr != "" {
			if err := self.startMonitoring(debugAddr); err != nil {
				return err
			}
		}
	}

	// Set up exporter
	for _, pCfg := range self.commonConfig.MonitoringPlugins {
		plug, err :=
			plugins.DefaultPluginStore.Create(pCfg.Plugin, pCfg.PluginConfig)
		if err != nil {
			return fmt.Errorf("Filed to create new %s: %v", pCfg.Plugin, err)
		}
		mon, ok := plug.(gopnik.MonitoringPluginInterface)
		if !ok {
			fmt.Errorf(`Invalid monitoring plugin "%s"`, pCfg.Plugin)
		}
		exporter, err := mon.Exporter()
		if err != nil {
			return fmt.Errorf("%s exporter error: %v", pCfg.Plugin, err)
		}
		hmetrics2.AddHook(exporter)
	}

	return nil
}

// Start monitoring
func (self *app) startMonitoring(addr string) error {
	go func() {
		log.Info("Serving debug data (/debug/vars) on %s...", addr)
		log.Info("Serving monitoring xml data on %s...", addr)
		http.Handle("/", xmlstatus.CreateXMLStatusHandler(self.config))
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
	// Hmetrics2
	hmetrics2.SetPeriod(time.Minute)
	hmetrics2.AddHook(expvarexport.Exporter("metrics"))
	// Export metrics
	// TODO: monitoring plugins
	return nil
}

// Set up CPU profiler
func (self *app) startCpuProfile(cpuprofile string) error {
	f, err := os.Create(cpuprofile)
	if err != nil {
		return fmt.Errorf("Failed to create CPU profile: %v", err)
	}
	pprof.StartCPUProfile(f)
	// Handle
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.Info("captured %v, stopping profiler and exiting...", sig)
			pprof.StopCPUProfile()
			os.Exit(2)
		}
	}()

	return nil
}
