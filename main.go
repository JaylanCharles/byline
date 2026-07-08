package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JaylanCharles/byline/ioc"
	"github.com/fsnotify/fsnotify"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initLogger()
	initViper()

	closeFunc := ioc.InitOTEL()
	initPrometheus()
	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	app.cron.Start()

	server := app.web
	server.Run(":8080")

	// 一分钟内你要关完，要退出
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	closeFunc(ctx)

	ctx = app.cron.Stop()
	// 想办法 close ？？
	// 这边可以考虑超时强制退出，防止有些任务，执行特别长的时间
	tm := time.NewTimer(time.Minute * 10)
	select {
	case <-tm.C:
	case <-ctx.Done():
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/byline")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

func initViper() {
	cfile := pflag.String("config", "config/config.yaml", "指定配置文件路径")
	pflag.Parse()
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println(in.Name, in.Op)
		// 重新读新配置
	})

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8081", nil)
	}()
}
