package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/casept/configdir"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Sets up viper
func setupConfig() {

	// Set defaults
	homedir, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Unable to determine user's home directory: %v\n", err)
	}
	defaultGopherroot := filepath.Join(homedir, ".gopher")
	viper.SetDefault("gopherroot", defaultGopherroot)
	viper.SetDefault("port", 70)
	viper.SetDefault("address", "0.0.0.0")
	viper.SetDefault("testmode", false)

	// Set up CLI flags using pflag
	pflag.StringP("gopherroot", "g", defaultGopherroot, "Path to the directory to be served.")
	pflag.IntP("port", "p", 70, "The port to listen on. Default requires root/admin privileges.")
	pflag.StringP("address", "a", "0.0.0.0", "An IPv4/v6 address to listen on. Multiple addresses are currently unsupported.")
	pflag.StringP("config", "c", "", "Path to configuration file outside the standard config directories.")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	// Add the config file specified by the CLI flag if set
	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
		// Only add the other config files if the user doesn't specify one using the flag.
	} else {
		viper.SetConfigName("yagopherd")
		// Add PWD to config search path
		viper.AddConfigPath(".")
		// Get OS-specific config paths
		configDirs := configdir.New("yagopherd", "yagopherd")
		// Add per-user config directory
		viper.AddConfigPath(configDirs.GetFolder(configdir.Local))
		// Add system default systemwide config directory
		viper.AddConfigPath(configDirs.GetFolder(configdir.Global))
		// Set up automatic env var handling
		viper.SetEnvPrefix("yagopherd") // Becomes "YAGOPHERD_"
		viper.AutomaticEnv()
	}

	// Read config file
	err = viper.ReadInConfig()
	if err != nil {
		switch t := err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Printf("No config file found, relying on env vars/flags/defaults!")
		default:
			log.Fatalf("Error while reading config file: %s\n", err)
			_ = t
		}
	}

	// Ensure gopherroot is an absolute path
	absGopherroot, err := filepath.Abs(viper.GetString("gopherroot"))
	if err != nil {
		log.Fatalf("Failed to expand relative path: %v", err)
	} else {
		viper.Set("gopherroot", absGopherroot)
	}

	// Make sure gopherroot directory exists and is readable
	_, err = os.Stat(viper.GetString("gopherroot"))
	if err != nil {
		log.Fatalf("Cannot stat gopherroot %v: %v", viper.GetString("gopherroot"), err)
	}
}
