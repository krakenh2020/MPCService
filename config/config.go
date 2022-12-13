package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// init loads the default config file
func init() {
	// set reasonable defaults
	setDefaults()

	//override defaults with configuration read from configuration file
	viper.AddConfigPath("./config")
	err := loadConfig("defaults", "yml")
	if err != nil {
		fmt.Println(err)
	}
}

// loadConfig reads in the config file with configName being the name of the file (without suffix)
// and configType being "yml" or "json".
func loadConfig(configName string, configType string) error {
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("cannot read configuration file: %s\n", err)
	}

	return nil
}

// setDefaults sets default values that can be specified in the command line.
func setDefaults() {
	viper.SetDefault("name", "Ljubljana_node")
	viper.SetDefault("nodeAddr", "localhost")
	viper.SetDefault("managerPort", 5001)
	viper.SetDefault("guiPort", 5000)
	viper.SetDefault("scalePort", 5010)

	viper.SetDefault("sm", "/SCALE-MAMBA")
	viper.SetDefault("certLocation", "key_management/keys_certificates")

	viper.SetDefault("logLevel", "info")
	viper.SetDefault("logFile", "logging/log.log")

	viper.SetDefault("manAddr", "localhost/5001")

	viper.SetDefault("dataLoc", "data_provider/datasets")
	viper.SetDefault("assets", "manager/assets")
	viper.SetDefault("shareWith", "all")
	viper.SetDefault("description", "")
}

// LoadServerName returns the name of the server.
func LoadServerName() string {
	return viper.GetString("name")
}

// LoadNodeAddr returns the address where the node server will be listening.
func LoadNodeAddr() string {
	return viper.GetString("nodeAddr")
}

// LoadManagerPort returns the port where the manager will be listening.
func LoadManagerPort() int {
	return viper.GetInt("managerPort")
}

// LoadGuiPort returns the port where the manager will be listening.
func LoadGuiPort() int {
	return viper.GetInt("guiPort")
}

// LoadScalePort returns the port where scale will be listening.
func LoadScalePort() int {
	return viper.GetInt("scalePort")
}

// LoadScaleMamba returns location of SCALE-MAMBA.
func LoadScaleMamba() string {
	return viper.GetString("sm")
}

// LoadCertLocation returns the folder of the certificates.
func LoadCertLocation() string {
	return viper.GetString("certLocation")
}

// LoadLogFile returns the destination of log file.
func LoadLogFile() string {
	return viper.GetString("logFile")
}

// LoadLogLevel returns the level of logging.
func LoadLogLevel() string {
	return viper.GetString("logLevel")
}

// LoadManAddr returns the address of the manager.
func LoadManAddr() string {
	return viper.GetString("manAddr")
}

// LoadDataLoc returns the location of the datasets.
func LoadDataLoc() string {
	return viper.GetString("dataLoc")
}

// LoadAssets returns the location of assets needed to serve a web page py the manager.
func LoadAssets() string {
	return viper.GetString("assets")
}

func LoadShareWith() string {
	return viper.GetString("shareWith")
}

func LoadDescription() string {
	return viper.GetString("description")
}
