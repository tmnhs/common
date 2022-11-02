package common

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/tmnhs/common/utils"
	"path"
)

const (
	ExtensionJson = ".json"
	ExtensionYaml = ".yaml"
	ExtensionInI  = ".ini"

	NameSpace = "conf"
)

var (
	//Automatic loading sequence of local Config
	autoLoadLocalConfigs = []string{
		ExtensionJson,
		ExtensionYaml,
		ExtensionInI,
	}
)

type (
	Mysql struct {
		Path         string `mapstructure:"path" json:"path" yaml:"path"`                             // 服务器地址
		Port         string `mapstructure:"port" json:"port" yaml:"port"`                             // 端口
		Config       string `mapstructure:"config" json:"config" yaml:"config"`                       // 高级配置
		Dbname       string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`                     // 数据库名
		Username     string `mapstructure:"username" json:"username" yaml:"username"`                 // 数据库用户名
		Password     string `mapstructure:"password" json:"password" yaml:"password"`                 // 数据库密码
		MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns"` // 空闲中的最大连接数
		MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns"` // 打开到数据库的最大连接数
		LogMode      string `mapstructure:"log-mode" json:"logMode" yaml:"log-mode"`                  // 是否开启Gorm全局日志
		LogZap       bool   `mapstructure:"log-zap" json:"logZap" yaml:"log-zap"`                     // 是否通过zap写入日志文件
	}
	Email struct {
		Port     int      `mapstructure:"port" json:"port" yaml:"port"`             // 端口
		From     string   `mapstructure:"from" json:"from" yaml:"from"`             // 收件人
		Host     string   `mapstructure:"host" json:"host" yaml:"host"`             // 服务器地址
		IsSSL    bool     `mapstructure:"is-ssl" json:"isSSL" yaml:"is-ssl"`        // 是否SSL
		Secret   string   `mapstructure:"secret" json:"secret" yaml:"secret"`       // 密钥
		Nickname string   `mapstructure:"nickname" json:"nickname" yaml:"nickname"` // 昵称
		To       []string `mapstructure:"to" json:"to" yaml:"to" ini:"to"`
	}
	WebHook struct {
		Kind string `mapstructure:"kind" json:"kind" yaml:"kind" ini:"kind"`
		Url  string `mapstructure:"url" json:"url" yaml:"url" ini:"kind"`
	}
	Etcd struct {
		Endpoints   []string `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints" ini:"endpoints"`
		Username    string   `mapstructure:"username" json:"username" yaml:"username" ini:"username"`
		Password    string   `mapstructure:"password" json:"password" yaml:"password" ini:"password"`
		DialTimeout int64    `mapstructure:"dial-timeout" json:"dial-timeout" yaml:"dial-timeout" ini:"dial-timeout"`
		ReqTimeout  int64    `mapstructure:"req-timeout" json:"req-timeout" yaml:"req-timeout" ini:"req-timeout"`
	}
	Redis struct {
		DB       int    `mapstructure:"db" json:"db" yaml:"db"`                   // redis的哪个数据库
		Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`             // 服务器地址:端口
		Password string `mapstructure:"password" json:"password" yaml:"password"` // 密码
	}
	System struct {
		Env     string `mapstructure:"env" json:"env" yaml:"env" ini:"env"`
		Addr    int    `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`
		Version string `mapstructure:"version" json:"version" yaml:"version" ini:"version"`
	}
	Log struct {
		Level         string `mapstructure:"level" json:"level" yaml:"level"`                           // 级别
		Format        string `mapstructure:"format" json:"format" yaml:"format"`                        // 输出
		Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`                        // 日志前缀
		Director      string `mapstructure:"director" json:"director"  yaml:"director"`                 // 日志文件夹
		ShowLine      bool   `mapstructure:"show-line" json:"showLine" yaml:"showLine"`                 // 显示行
		EncodeLevel   string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level"`       // 编码级
		StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key"` // 栈名
		LogInConsole  bool   `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console"`  // 输出控制台
	}
	Config struct {
		WebHook WebHook `mapstructure:"webhook" json:"webhook" yaml:"webhook" ini:"webhook"`
		Log     Log     `mapstructure:"log" json:"log" yaml:"log" ini:"log"`
		Email   Email   `mapstructure:"email" json:"email" yaml:"email" ini:"email"`
		System  System  `mapstructure:"system" json:"system" yaml:"system" ini:"system"`
		Mysql   Mysql   `mapstructure:"mysql" json:"mysql" yaml:"mysql" ini:"mysql"`
		Redis   Redis   `mapstructure:"redis" json:"redis" yaml:"redis" ini:"redis"`
		Etcd    Etcd    `mapstructure:"etcd" json:"etcd" yaml:"etcd" ini:"etcd"`
	}
)

func (m *Mysql) Dsn() string {
	return m.Username + ":" + m.Password + "@tcp(" + m.Path + ":" + m.Port + ")/" + m.Dbname + "?" + m.Config
}

func (m *Mysql) EmptyDsn() string {
	if m.Path == "" {
		m.Path = "127.0.0.1"
	}
	if m.Port == "" {
		m.Port = "3306"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/", m.Username, m.Password, m.Path, m.Port)
}

var _defaultConfig *Config

func LoadConfig(env, configFileName string) (*Config, error) {
	var c Config
	var confPath string
	dir := fmt.Sprintf("%s/%s", NameSpace, env)
	for _, registerExt := range autoLoadLocalConfigs {
		confPath = path.Join(dir, configFileName+registerExt)
		if utils.Exists(confPath) {
			break
		}
	}
	fmt.Println("the path to the configuration file you are using is :", confPath)
	v := viper.New()
	v.SetConfigFile(confPath)
	ext := utils.Ext(confPath)
	v.SetConfigType(ext)
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	v.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("config file changed:", e.Name)
		if err := v.Unmarshal(&c); err != nil {
			fmt.Println(err)
		}
	})
	if err := v.Unmarshal(&c); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("load config is :%#v\n", c)
	_defaultConfig = &c
	return &c, nil
}

func GetConfigModels() *Config {
	return _defaultConfig
}
