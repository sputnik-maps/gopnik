package gopnik

type ClusterPluginInterface interface {
	GetRenders() ([]string, error)
}

type KVStore interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
}

type CachePluginInterface interface {
	Get(TileCoord) ([]byte, error)
	Set(TileCoord, []Tile) error
}

type MonitoringPluginInterface interface {
	Exporter() (func(map[string]float64), error)
}

type FilterPluginInterface interface {
	Filter(TileCoord) (TileCoord, error)
}
