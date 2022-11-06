package common

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"github.com/tmnhs/common/utils"
	"path"
)

const (
	extensionJson = ".json"
	extensionYaml = ".yaml"
	extensionInI  = ".ini"

	nameSpace = "conf"
)

var (
	//Automatic loading sequence of local Config
	autoLoadLocalConfigs = []string{
		extensionJson,
		extensionYaml,
		extensionInI,
	}
)

type (
	Mysql struct {
		Path         string `mapstructure:"path" json:"path" yaml:"path" ini:"path"`                                       // 服务器地址
		Port         string `mapstructure:"port" json:"port" yaml:"port" ini:"port"`                                       // 端口
		Config       string `mapstructure:"config" json:"config" yaml:"config" ini:"config"`                               // 高级配置
		Dbname       string `mapstructure:"db-name" json:"dbname" yaml:"db-name" ini:"db-name"`                            // 数据库名
		Username     string `mapstructure:"username" json:"username" yaml:"username" ini:"username"`                       // 数据库用户名
		Password     string `mapstructure:"password" json:"password" yaml:"password" ini:"password"`                       // 数据库密码
		MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns" ini:"max-idle-conns"` // 空闲中的最大连接数
		MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns" ini:"max-open-conns"` // 打开到数据库的最大连接数
		LogMode      string `mapstructure:"log-mode" json:"logMode" yaml:"log-mode" ini:"log-mode"`                        // 是否开启Gorm全局日志
		LogZap       bool   `mapstructure:"log-zap" json:"logZap" yaml:"log-zap" ini:"log-zap"`                            // 是否通过zap写入日志文件
	}
	Etcd struct {
		Endpoints   []string `mapstructure:"endpoints" json:"endpoints" yaml:"endpoints" ini:"endpoints"`
		Username    string   `mapstructure:"username" json:"username" yaml:"username" ini:"username"`
		Password    string   `mapstructure:"password" json:"password" yaml:"password" ini:"password"`
		DialTimeout int64    `mapstructure:"dial-timeout" json:"dial-timeout" yaml:"dial-timeout" ini:"dial-timeout"`
		ReqTimeout  int64    `mapstructure:"req-timeout" json:"req-timeout" yaml:"req-timeout" ini:"req-timeout"`
	}
	Redis struct {
		DB       int    `mapstructure:"db" json:"db" yaml:"db" ini:"db"`                         // redis的哪个数据库
		Addr     string `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`                 // 服务器地址:端口
		Password string `mapstructure:"password" json:"password" yaml:"password" ini:"password"` // 密码
	}
	System struct {
		Env        string `mapstructure:"env" json:"env" yaml:"env" ini:"env"`
		Addr       int    `mapstructure:"addr" json:"addr" yaml:"addr" ini:"addr"`
		UploadType string `mapstructure:"upload-type" json:"upload-type" yaml:"upload-type" ini:"upload-type"` // Oss类型
		Version    string `mapstructure:"version" json:"version" yaml:"version" ini:"version"`
	}
	Log struct {
		Level         string `mapstructure:"level" json:"level" yaml:"level" ini:"level"`                                    // 级别
		Format        string `mapstructure:"format" json:"format" yaml:"format" ini:"level"`                                 // 输出
		Prefix        string `mapstructure:"prefix" json:"prefix" yaml:"prefix" ini:"level"`                                 // 日志前缀
		Director      string `mapstructure:"director" json:"director"  yaml:"director" ini:"level"`                          // 日志文件夹
		ShowLine      bool   `mapstructure:"show-line" json:"showLine" yaml:"showLine" ini:"showLine"`                       // 显示行
		EncodeLevel   string `mapstructure:"encode-level" json:"encodeLevel" yaml:"encode-level" ini:"encode-level"`         // 编码级
		StacktraceKey string `mapstructure:"stacktrace-key" json:"stacktraceKey" yaml:"stacktrace-key" ini:"stacktrace-key"` // 栈名
		LogInConsole  bool   `mapstructure:"log-in-console" json:"logInConsole" yaml:"log-in-console" ini:"log-in-console"`  // 输出控制台
	}
)

//notify
type (
	Email struct {
		Port     int      `mapstructure:"port" json:"port" yaml:"port" ini:"port"`                 // 端口
		From     string   `mapstructure:"from" json:"from" yaml:"from" ini:"from"`                 // 收件人
		Host     string   `mapstructure:"host" json:"host" yaml:"host" ini:"host"`                 // 服务器地址
		IsSSL    bool     `mapstructure:"is-ssl" json:"is-ssl" yaml:"is-ssl" ini:"is-ssl"`         // 是否SSL
		Secret   string   `mapstructure:"secret" json:"secret" yaml:"secret" ini:"secret"`         // 密钥
		Nickname string   `mapstructure:"nickname" json:"nickname" yaml:"nickname" ini:"nickname"` // 昵称
		To       []string `mapstructure:"to" json:"to" yaml:"to" ini:"to"`
	}
	WebHook struct {
		Kind string `mapstructure:"kind" json:"kind" yaml:"kind" ini:"kind"`
		Url  string `mapstructure:"url" json:"url" yaml:"url" ini:"kind"`
	}
	Notify struct {
		Email   Email   `mapstructure:"email" json:"email" yaml:"email" ini:"email"`
		WebHook WebHook `mapstructure:"webhook" json:"webhook" yaml:"webhook" ini:"webhook"`
	}
)

//upload
type (
	//阿里云存储对象
	AliyunOSS struct {
		Endpoint        string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint" ini:"endpoint"`
		AccessKeyId     string `mapstructure:"access-key-id" json:"access-key-id" yaml:"access-key-id" ini:"access-key-id"`
		AccessKeySecret string `mapstructure:"access-key-secret" json:"access-key-secret" yaml:"access-key-secret" ini:"access-key-secret"`
		BucketName      string `mapstructure:"bucket-name" json:"bucket-name" yaml:"bucket-name" ini:"bucket-name"`
		BucketUrl       string `mapstructure:"bucket-url" json:"bucket-url" yaml:"bucket-url" ini:"bucket-url"`
		BasePath        string `mapstructure:"base-path" json:"base-path" yaml:"base-path" ini:"base-path"`
	}
	//华为云存储对象
	HuaWeiObs struct {
		Path      string `mapstructure:"path" json:"path" yaml:"path" ini:"path"`
		Bucket    string `mapstructure:"bucket" json:"bucket" yaml:"bucket" ini:"bucket"`
		Endpoint  string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint" ini:"endpoint"`
		AccessKey string `mapstructure:"access-key" json:"access-key" yaml:"access-key" ini:"access-key"`
		SecretKey string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key" ini:"secret-key"`
	}
	Local struct {
		Path string `mapstructure:"path" json:"path" yaml:"path" ini:"path"` // 本地文件路径
	}
	//七牛云存对象
	Qiniu struct {
		Zone          string `mapstructure:"zone" json:"zone" yaml:"zone" ini:"zone"`                                             // 存储区域
		Bucket        string `mapstructure:"bucket" json:"bucket" yaml:"bucket" ini:"bucket"`                                     // 空间名称
		ImgPath       string `mapstructure:"img-path" json:"img-path" yaml:"img-path" ini:"img-path"`                             // CDN加速域名
		UseHTTPS      bool   `mapstructure:"use-https" json:"use-https" yaml:"use-https" ini:"use-https"`                         // 是否使用https
		AccessKey     string `mapstructure:"access-key" json:"access-key" yaml:"access-key" ini:"access-key"`                     // 秘钥AK
		SecretKey     string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key" ini:"secret-key"`                     // 秘钥SK
		UseCdnDomains bool   `mapstructure:"use-cdn-domains" json:"use-cdn-domains" yaml:"use-cdn-domains" ini:"use-cdn-domains"` // 上传是否使用CDN上传加速
	}
	//腾讯云存储对象
	TencentCOS struct {
		Bucket     string `mapstructure:"bucket" json:"bucket" yaml:"bucket" ini:"bucket"`
		Region     string `mapstructure:"region" json:"region" yaml:"region" ini:"region"`
		SecretID   string `mapstructure:"secret-id" json:"secret-id" yaml:"secret-id" ini:"secret-id"`
		SecretKey  string `mapstructure:"secret-key" json:"secret-key" yaml:"secret-key" ini:"secret-key"`
		BaseURL    string `mapstructure:"base-url" json:"base-url" yaml:"base-url" ini:"base-url"`
		PathPrefix string `mapstructure:"path-prefix" json:"path-prefix" yaml:"path-prefix" ini:"path-prefix"`
	}
	Upload struct {
		// oss
		Local      Local      `mapstructure:"local" json:"local" yaml:"local" ini:"local"`
		Qiniu      Qiniu      `mapstructure:"qiniu" json:"qiniu" yaml:"qiniu" ini:"qiniu"`
		AliyunOSS  AliyunOSS  `mapstructure:"aliyun-oss" json:"aliyun-oss" yaml:"aliyun-oss" ini:"aliyun-oss"`
		HuaWeiObs  HuaWeiObs  `mapstructure:"hua-wei-obs" json:"hua-wei-obs" yaml:"hua-wei-obs" ini:"hua-wei-obs"`
		TencentCOS TencentCOS `mapstructure:"tencent-cos" json:"tencent-cos" yaml:"tencent-cos" ini:"tencent-cos"`
	}
)

type Config struct {
	Log    Log    `mapstructure:"log" json:"log" yaml:"log" ini:"log"`
	System System `mapstructure:"system" json:"system" yaml:"system" ini:"system"`
	Mysql  Mysql  `mapstructure:"mysql" json:"mysql" yaml:"mysql" ini:"mysql"`
	Redis  Redis  `mapstructure:"redis" json:"redis" yaml:"redis" ini:"redis"`
	Etcd   Etcd   `mapstructure:"etcd" json:"etcd" yaml:"etcd" ini:"etcd"`
	Notify Notify `mapstructure:"notify" json:"notify" yaml:"notify" ini:"notify"`
	Upload Upload `mapstructure:"upload" json:"upload" yaml:"upload" ini:"upload"`
}

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
	dir := fmt.Sprintf("%s/%s", nameSpace, env)
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
