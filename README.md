# common

> 基于gorm、gin、zap、viper实现的对一些常用库的二次封装以及连接mysql、redis、etcd、日志的公共包，简化开发过程，能快速的搭建一个web后端服务器

##  技术选型
- 数据库:         [Gorm](http://gorm.cn)
- web框架:      [Gin](https://gin-gonic.com/)
- 日志:              [Zap](https://github.com/uber-go/zap)
- 读取配置文件:[Viper](https://github.com/spf13/viper)

## 目录结构


| 目录          | 说明                               |
| ----------- | -------------------------------- |
| dbclient    | 基于gorm实现的数据库mysql连接              |
| etcdclient  | 提供etcd连接                         |
| httpclient  | 提供http(get/post)请求方法             |
| logger      | 基于zap实现的日志管理，实现日志分类以及分割          |
| notify      | 提供email和webhook两种通知方式            |
| server      | 基于gin实现对web服务的启动                 |
| utils       | 一些工具类，比如system.go用来获取服务器cpu和内存信息 |
| config.go   | 配置信息的结构体并加载配置文件                  |
| env.go      | 一些环境变量                           |
| request.go  | 常见的绑定请求的的结构体                     |
| response.go | 常见的请求返回的结构体   

## 使用方法

```shell
go get -u github.com/tmnhs/common
```

使用示例可见:  [common-test](https://github.com/tmnhs/common-test)

### 1.配置文件

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
    "username": "root",
    "password": "root",
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
  },
  "log": {
    "level": "debug",
    "format": "console",
    "prefix": "[crony-admin]",
    "director": "logs",
    "showLine": false,
    "encode-level": "LowercaseLevelEncoder",
    "stacktrace-key": "stacktrace",
    "log-in-console": true
  }
}
```

### 2.开启一个web应用

```go
func main() {
  	//参数为需要启动的服务(etcd/mysql/redis)
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

### 3.注册路由

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

                   |

