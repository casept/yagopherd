// configdir provides access to configuration folder in each platforms.
//
// System wide configuration folders:
//
//   - Windows: %PROGRAMDATA% (C:\ProgramData)
//   - Linux/BSDs: ${XDG_CONFIG_DIRS} (/etc/xdg)
//   - MacOSX: "/Library/Application Support"
//
// User wide configuration folders:
//
//   - Windows: %APPDATA% (C:\Users\<User>\AppData\Roaming)
//   - Linux/BSDs: ${XDG_CONFIG_HOME} (${HOME}/.config)
//   - MacOSX: "${HOME}/Library/Application Support"
//
// User wide cache folders:
//
//   - Windows: %LOCALAPPDATA% (C:\Users\<User>\AppData\Local)
//   - Linux/BSDs: ${XDG_CACHE_HOME} (${HOME}/.cache)
//   - MacOSX: "${HOME}/Library/Caches"
//
// configdir returns paths inside the above folders.

package configdir

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

//ConfigType is the numerical identifier of the type of configuration directory.
type ConfigType int

// Various types of directories, such as the per-user config directory or cache directory.
const (
	System ConfigType = iota
	Global
	All
	Existing
	Local
	Cache
)

// Config represents each folder
type Config struct {
	Path string
	// Type is the numerical identifier for the type of folder.
	// You shouldn't set this manually.
	Type ConfigType
}

// Open opens the specified file inside of Config.Path.
func (c Config) Open(fileName string) (*os.File, error) {
	return os.Open(filepath.Join(c.Path, fileName))
}

// Create creates the specified file inside of Config.Path (and it's parent directory, if needed).
func (c Config) Create(fileName string) (*os.File, error) {
	err := c.CreateParentDir(fileName)
	if err != nil {
		return nil, err
	}
	return os.Create(filepath.Join(c.Path, fileName))
}

// ReadFile reads the specified file inside of Config.Path
func (c Config) ReadFile(fileName string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(c.Path, fileName))
}

// CreateParentDir creates the parent directory of fileName inside c. fileName
// is a relative path inside c, containing zero or more path separators.
func (c Config) CreateParentDir(fileName string) error {
	return os.MkdirAll(filepath.Dir(filepath.Join(c.Path, fileName)), 0755)
}

// WriteFile writes data to the specified file inside of Config.Path.
// The parent directory of the file is created if not present.
func (c Config) WriteFile(fileName string, data []byte) error {
	err := c.CreateParentDir(fileName)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(c.Path, fileName), data, 0644)
}

// MkdirAll creates Config.Path and all needed parent directories.
func (c Config) MkdirAll() error {
	return os.MkdirAll(c.Path, 0755)
}

// Exists checks whether a file exists inside of Config.Path.
func (c Config) Exists(fileName string) bool {
	_, err := os.Stat(filepath.Join(c.Path, fileName))
	return !os.IsNotExist(err)
}

// ConfigDir keeps setting for querying folders.
type ConfigDir struct {
	VendorName      string
	ApplicationName string
	LocalPath       string
}

// New creates a new ConfigDir struct.
func New(vendorName, applicationName string) ConfigDir {
	return ConfigDir{
		VendorName:      vendorName,
		ApplicationName: applicationName,
	}
}

func (c ConfigDir) joinPath(root string) string {
	if c.VendorName != "" && hasVendorName {
		return filepath.Join(root, c.VendorName, c.ApplicationName)
	}
	return filepath.Join(root, c.ApplicationName)
}

// QueryFolders returns the location of the specified type of config folder.
func (c ConfigDir) QueryFolders(configType ConfigType) []*Config {
	if configType == Cache {
		return []*Config{c.QueryCacheFolder()}
	}
	var result []*Config
	if c.LocalPath != "" && configType != System && configType != Global {
		result = append(result, &Config{
			Path: c.LocalPath,
			Type: Local,
		})
	}
	if configType != System && configType != Local {
		result = append(result, &Config{
			Path: c.joinPath(globalSettingFolder),
			Type: Global,
		})
	}
	if configType != Global && configType != Local {
		for _, root := range systemSettingFolders {
			result = append(result, &Config{
				Path: c.joinPath(root),
				Type: System,
			})
		}
	}
	if configType != Existing {
		return result
	}
	var existing []*Config
	for _, entry := range result {
		if _, err := os.Stat(entry.Path); !os.IsNotExist(err) {
			existing = append(existing, entry)
		}
	}
	return existing
}

// QueryFolderContainsFile checks whether ConfigDir contains the specified file.
func (c ConfigDir) QueryFolderContainsFile(fileName string) *Config {
	configs := c.QueryFolders(Existing)
	for _, config := range configs {
		if _, err := os.Stat(filepath.Join(config.Path, fileName)); !os.IsNotExist(err) {
			return config
		}
	}
	return nil
}

// QueryCacheFolder returns a Config struct containing the path to the cache directory.
func (c ConfigDir) QueryCacheFolder() *Config {
	return &Config{
		Path: c.joinPath(cacheFolder),
		Type: Cache,
	}
}

// GetFolder returns the location of the specified type of config folder.
func (c ConfigDir) GetFolder(configType ConfigType) string {
	var path string
	if configType == Cache {
		return c.getCacheFolder()
	}
	if c.LocalPath != "" && configType != System && configType != Global {
		path = c.LocalPath
	}
	if configType != System && configType != Local {
		path = c.joinPath(globalSettingFolder)
	}
	if configType != Global && configType != Local {
		path = c.joinPath(systemSettingFolders[0])
	}
	return path
}

func (c ConfigDir) getCacheFolder() string {
	return c.joinPath(cacheFolder)
}
