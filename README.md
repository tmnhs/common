# common
<div align=center>

<img src="https://img.shields.io/badge/golang-1.16.5-blue"/>
<img src="https://img.shields.io/badge/gin-1.8.1-lightBlue"/>
<img src="https://img.shields.io/badge/gorm-1.23.10-red"/>
<img src="https://img.shields.io/badge/etcd-3.5-red"/>
</div>

> 基于gorm、gin、zap、viper实现的对一些常用库的二次封装以及连接mysql、redis、etcd、日志的公共包，简化开发过程，能快速的搭建一个web后端服务器

##  1. 技术选型
- 数据库:         [Gorm](http://gorm.cn)
- web框架:      [Gin](https://gin-gonic.com/)
- 日志:              [Zap](https://github.com/uber-go/zap)
- 读取配置文件:[Viper](https://github.com/spf13/viper)

## 2.功能特性

- 支持数据库mysql、redis、etcd连接
- 提供五种文件上传(存储)的方式，包括本地上传、阿里云存储对象、七牛云存储对象、腾讯云存储对象、华为云存储对象
- 提供通知功能，提供email和webhook两种方式
- 提供http请求Get方法和Post方法
- 提供日志封装
- 通过viper加载配置文件，支持testing和production两种环境，支持json、yaml、ini三种文件格式
- 提供一些有用的工具包
    - event.go: 监听程序的退出信号
    - file.go:  一些对文件目录处理的函数
    - ip.go:    获取本机ip
    - map.go:   对map的一些封装，提高安全性
    - md5.go:  对数据的普通加密
    - parse.go: 对cmd命令的解析
    - platform.go: 一些关于各平台的常量
    - scrypt.go: 对数据的高级加密，不可逆
    - strings.go: 字符串转化的一些处理函数
    - system.go:  获取服务器的cpu、硬盘、内存等信息
    - task.go:   对定时任务的简单封装
    - time.go:   对时间的一些处理函数
    - uuid.go:   获取uuid（唯一）


| 目录         | 说明                               |
| ---------- | -------------------------------- |
| dbclient   | 基于gorm实现的数据库mysql连接              |
| etcdclient | 提供etcd连接                         |
| httpclient | 提供http(get/post)请求方法             |
| logger     | 基于zap实现的日志管理，实现日志分类以及分割          |
| notify     | 提供email和webhook两种通知方式            |
| server     | 基于gin实现对web服务的启动                 |
| utils      | 一些工具类 |
| config.go  | 配置信息的结构体并基于viper加载配置文件           |
| upload.go  | 文件上传功能的方法封装           |
| env.go     | 一些环境变量                           |
| request.go | 常见的绑定请求的的结构体                     |
| response.go | 常见的请求返回的结构体   

## 3. 使用方法

```shell
go get -u github.com/tmnhs/common
```

**详细的使用示例可见:**  [common-test](https://github.com/tmnhs/common-test)

### 3.1.配置文件

> 配置文件支持多种环境(testing/prodution)和多种格式(json/yaml/ini)
>
> **注意事项**:配置文件的目录必须是下面这个样子


```shell
├── cmd
├── conf
|     ├── production                #生产环境，支持json、yaml、ini三种配置文件
|     |          └── main.json
|     └── testing
|                └── main.json      #测试环境，支持json、yaml、ini三种配置文件
└── internal
```
**配置文件示例(json格式)**

```json
{
  "mysql": {
    "path": "127.0.0.1",
    "port": "3306",
    "config": "charset=utf8mb4&parseTime=True&loc=Local",
    "db-name": "common-test",
    "username": "",
    "password": "",
    "max-idle-conns": 100,
    "max-open-conns": 100,
    "log-mode": "info",
    "log-zap": false
  },
  "redis": {
    "addr": "127.0.0.1:6379",
    "password": "",
    "db": 0
  },
  "system": {
    "env": "testing",
    "addr": 8089,
    "upload-type": "qiniu",
    "version": "v1.0.2"
  },
  "etcd": {
    "endpoints": [
      "http://127.0.0.1:2379"
    ],
    "username": "",
    "password": "",
    "dial-timeout": 2,
    "req-timeout": 5
  },
  "notify": {
    "email": {
      "port": 465,
      "from": "test@qq.com",
      "host": "smtp.qq.com",
      "is-ssl": true,
      "secret": "test",
      "nickname": "common-test",
      "to": [
        "test@test.mobi"
      ]
    },
    "webhook": {
      "url": "url",
      "kind": "feishu"
    }
  },
  "upload": {
    "local": {
      "path": "upload"
    },
    "aliyun-oss": {
      "endpoint": "yourEndpoint",
      "access-key-id": "yourAccessKeyId",
      "access-key-secret": "yourAccessKeySecret",
      "bucket-name": "yourBucketName",
      "bucket-url": "yourBucketUrl",
      "base-path": "yourBasePath"
    },
    "hua-wei-obs": {
      "path": "you-path",
      "bucket": "you-bucket",
      "endpoint": "you-endpoint",
      "access-key": "you-access-key",
      "secret-key": "you-secret-key"
    },
    "qiniu": {
      "zone": "ZoneHuanan",
      "bucket": "",
      "img-path": "http://qny.tmnhs.top",
      "use-https": false,
      "access-key": "",
      "secret-key": "",
      "use-cdn-domains": false
    },
    "tencent-cos": {
      "bucket": "xxxxx-10005608",
      "region": "ap-shanghai",
      "secret-id": "xxxxxxxx",
      "secret-key": "xxxxxxxx",
      "base-url": "xxxx",
      "path-prefix": "your path"
    }
  },
  "log": {
    "level": "debug",
    "format": "console",
    "prefix": "[common-test]",
    "director": "logs",
    "showLine": false,
    "encode-level": "LowercaseLevelEncoder",
    "stacktrace-key": "stacktrace",
    "log-in-console": true
  }
}
```

### 3.2 开启一个web应用

```go
func main() {
  	//参数为需要启动的服务(etcd/mysql/redis)
    //连接成功后可以通过dbclient.GetMysqlDD(),etcdClient.GetEtcd(),redisclient.GetRedis()获取对应的client
    //通过logger.GetLogger()获取日志处理器
    //通过common.GetConfigModels()获取配置文件的信息
	srv, err := server.NewApiServer(server.WithEtcd(),server.WithMysql(),server.WithRedis())
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("new api server error:%s", err.Error()))
		os.Exit(1)
	}
	// 注册路由
	srv.RegisterRouters(handler.RegisterRouters)

	// 建表，当然，如果不需要可以直接注释掉
	err = service.RegisterTables(dbclient.GetMysqlDB())
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("init db table error:%#v", err))
	}
	err = srv.ListenAndServe()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("startup api server error:%v", err.Error()))
		os.Exit(1)
	}
	os.Exit(0)
}

```

### 3.3 注册路由

```go

func RegisterRouters(r *gin.Engine) {

	configRoute(r)

	configNoRoute(r)
}

func configRoute(r *gin.Engine) {

	hello := r.Group("/ping")
	{
		hello.GET("", func(c *gin.Context) {
			c.JSON(200, "pong")
		})
	}

	base := r.Group("")
	{
		base.POST("register", defaultUserRouter.Register)
		base.POST("login", defaultUserRouter.Login)
	}

	user := r.Group("/user")
	user.Use(middlerware.JWTAuth())
	{
		user.POST("del", defaultUserRouter.Delete)
		user.POST("update", defaultUserRouter.Update)
		user.POST("change_pw", defaultUserRouter.ChangePassword)
		user.GET("find", defaultUserRouter.FindById)
		user.POST("search", defaultUserRouter.Search)
	}
}

func configNoRoute(r *gin.Engine) {
	/*	r.LoadHTMLGlob("./dist/*.html") // npm打包成dist的路径
		r.StaticFile("favicon.ico", "./dist/favicon.ico")
		r.Static("/css", "./dist/css")
		r.Static("/fonts", "./dist/fonts")
		r.Static("/js", "./dist/js")
		r.Static("/img", "./dist/img")
		r.StaticFile("/", "./dist/index.html") // 前端网页入口页面*/
}
```

## 4. 可能出现的问题

如果引入包并且go mod tidy 出现以下错误时

```go
go: finding module for package google.golang.org/grpc/naming
github.com/tmnhs/common-test/cmd imports
        github.com/tmnhs/common/server imports
        github.com/tmnhs/common/etcdclient imports
        github.com/coreos/etcd/clientv3 tested by
        github.com/coreos/etcd/clientv3.test imports
        github.com/coreos/etcd/integration imports
        github.com/coreos/etcd/proxy/grpcproxy imports
        google.golang.org/grpc/naming: module google.golang.org/grpc@latest found (v1.50.1), but does not contain package google.golang.org/grpc/naming
```

可以在go.mod中添加以下一行（这个报错和etcd连接的第三方库有版本冲突）

```
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
```
## 5. 其他功能
如果你需要添加其他的功能，建议将common克隆到你的项目里自行修改，然后
```
replace github.com/tmnhs/common => ../common
```

## 6. 交流讨论

如有问题欢迎加qq:1685290935一起交流讨论