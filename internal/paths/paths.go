package paths

import "github.com/adrg/xdg"

const appName = "bab"

func CacheFile(name string) (string, error) {
	return xdg.CacheFile(appName + "/" + name)
}

func SearchCacheFile(name string) (string, error) {
	return xdg.SearchCacheFile(appName + "/" + name)
}

func ConfigFile(name string) (string, error) {
	return xdg.ConfigFile(appName + "/" + name)
}

func SearchConfigFile(name string) (string, error) {
	return xdg.SearchConfigFile(appName + "/" + name)
}
