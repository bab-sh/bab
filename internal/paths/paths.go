package paths

import "github.com/adrg/xdg"

const appName = "bab"

func CacheFile(name string) (string, error) {
	return xdg.CacheFile(appName + "/" + name)
}
