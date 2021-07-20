package abstractions

import (
	"flag"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yoyofx/yoyogo/abstractions/xlog"
	"github.com/yoyofx/yoyogo/utils"
	"path"
)

type Configuration struct {
	context *ConfigurationContext
	config  *viper.Viper
	log     xlog.ILogger
}

func NewConfiguration(configContext *ConfigurationContext) *Configuration {
	log := xlog.GetXLogger("Configuration")
	log.SetCustomLogFormat(nil)

	defaultConfig := viper.New()
	if configContext.enableEnv {
		defaultConfig.AutomaticEnv()
		defaultConfig.SetEnvPrefix("YYG")
	}
	if configContext.enableFlag {
		flag.String("app", "", "application name")
		flag.String("port", "", "application port")
		flag.String("profile", configContext.profile, "application profile")
		flag.String("f", "", "config file path")
		flag.String("conf", ".", "config dir")
		pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
		pflag.Parse()
		_ = defaultConfig.BindPFlags(pflag.CommandLine)
	}

	if pf := defaultConfig.GetString("profile"); pf != "" {
		configContext.profile = pf
	}

	if cf := defaultConfig.GetString("conf"); cf != "" {
		configContext.configDir = cf
	}

	if configFile := defaultConfig.GetString("f"); configFile != "" {
		configContext.configFile = configFile
	}
	configFilePath := configContext.configFile
	if configFilePath == "" {
		configName := configContext.configName + "_" + configContext.profile
		configFilePath = path.Join(configContext.configDir, configName+"."+configContext.ConfigType)
		exists, _ := utils.PathExists(configFilePath)
		if !exists {
			configName = configContext.configName
		}
		defaultConfig.AddConfigPath(configContext.configDir)
		defaultConfig.SetConfigName(configName)
		defaultConfig.SetConfigType(configContext.ConfigType)
	} else {
		defaultConfig.SetConfigFile(configFilePath)
	}

	if err := defaultConfig.ReadInConfig(); err != nil {
		panic(err)
		return nil
	}
	log.Debug(configFilePath)

	if configContext.EnableRemote {
		defaultConfig = configContext.RemoteProvider.GetProvider(defaultConfig)
		_ = defaultConfig.BindPFlags(pflag.CommandLine)
	}

	return &Configuration{
		context: configContext,
		config:  defaultConfig,
		log:     log,
	}
}

func (c *Configuration) Get(name string) interface{} {
	return c.config.Get(name)
}

func (c *Configuration) GetString(name string) string {
	return c.config.GetString(name)
}

func (c *Configuration) GetBool(name string) bool {
	return c.config.GetBool(name)
}

func (c *Configuration) GetInt(name string) int {
	return c.config.GetInt(name)
}

func (c *Configuration) GetSection(name string) IConfiguration {
	section := c.config.Sub(name)

	if section != nil {
		return &Configuration{config: section}
	}
	return nil
}

func (c *Configuration) Unmarshal(obj interface{}) {
	err := c.config.Unmarshal(obj)
	if err != nil {
		c.log.Error("unmarshal config is failed, err:", err)
	}
}

func (c *Configuration) GetProfile() string {
	return c.context.profile
}

func (c *Configuration) GetConfDir() string {
	return c.context.configDir
}
