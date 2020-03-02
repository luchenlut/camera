package main

import (
	"bytes"
	"camera/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	_ "net/http/pprof"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	Execute()
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var cfgFile string // config file
var version = "v2.0.8"

var rootCmd = &cobra.Command{
	Use:     "camera",
	Short:   "camera gateway bridge",
	Long:    ``,
	RunE:    run,
	Version: version,
}

func init() {
	cobra.OnInitialize(initConfig)
	//// for backwards compatibility
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
}

func initConfig() {
	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		//fmt.Printf("配置文件打印 \n %s \n",string(b))
		viper.SetConfigType("toml")
		if err = viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("iot-hub")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/iot-hub")
		viper.AddConfigPath("/etc/")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("Deprecation warning! no configuration file found, falling back on environment variables. Set your configuration")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}
	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
