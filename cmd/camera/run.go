package main

import (
	"camera"
	"camera/config"
	"camera/echo"
	"camera/goonvif"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"os"
	"os/signal"
	"path"
	"syscall"
	"time"
)

func run(cmd *cobra.Command, args []string) error {
	tasks := []func() error{
		setLogLevel,
		setMQTT,
		setIntervalCheck,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	sigChan := make(chan os.Signal)
	exitChan := make(chan struct{})
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")

	select {
	case <-exitChan:
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received, stopping immediately")
	}

	return nil
}

/*设置日志的输出方式*/
func setLogLevel() error {
	formatter := echo.NewTextFormat(true)
	log.SetFormatter(formatter)

	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	// 开启日志写入文件
	if config.C.General.LogPath != "" {
		if config.C.General.Name == "" {
			config.C.General.Name = "camera-gateway-bridge"
		}
		configLocalFilesystemLogger(config.C.General.LogPath, config.C.General.Name, time.Hour*24*7, time.Hour*12)
	}
	return nil
}

// 配置日志文件输出
func configLocalFilesystemLogger(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) {
	var exist bool
	if _, err := os.Stat(logPath); err == nil {
		exist = true
	}
	if !exist {
		if err := os.Mkdir(logPath, os.ModePerm); err != nil {
			log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
		}
	}
	baseLogPath := path.Join(logPath, logFileName)
	writerDebug, err := rotatelogs.New(
		baseLogPath+".debug.%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}

	writerInfo, err := rotatelogs.New(
		baseLogPath+".info.%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}

	writerWarn, err := rotatelogs.New(
		baseLogPath+".warn.%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	writerError, err := rotatelogs.New(
		baseLogPath+".error.%Y%m%d.log",
		rotatelogs.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(maxAge),             // 文件最大保存时间
		rotatelogs.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		log.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}

	lfHook := echo.NewLfsHook(echo.WriterMap{
		log.DebugLevel: writerDebug, // 为不同级别设置不同的输出目的
		log.InfoLevel:  writerInfo,
		log.WarnLevel:  writerWarn,
		log.ErrorLevel: writerError,
		log.FatalLevel: writerError,
		log.PanicLevel: writerError,
	}, nil)

	formatter := echo.NewTextFormat(true)
	lfHook.SetFormatter(formatter)
	log.AddHook(lfHook)
}

func setMQTT() error {
	log.WithFields(log.Fields{
		"server":   config.C.Camera.MQTTServer,
		"username": config.C.Camera.MQTTUsername,
		"password": config.C.Camera.MQTTPassword,
	}).Info("connecting to mqtt")
	cfg := camera.Config{
		Server:       config.C.Camera.MQTTServer,
		Username:     config.C.Camera.MQTTUsername,
		Password:     config.C.Camera.MQTTPassword,
		QOS:          2,
		CleanSession: true,
	}
	if err := camera.NewBackend(cfg); err != nil {
		return err
	}
	return nil
}

func setIntervalCheck() error {
	go func() {
		for {
			t := time.NewTicker(time.Second * 60)
			select {
			case <-t.C:
				// check
				dev, _ := goonvif.NewDevice(config.C.General.Addr)
				if dev == nil {
					go camera.HandleIntervalCheck("offline(离线)", camera.HandleInterval)
					//fmt.Println("offline")
				} else {
					go camera.HandleIntervalCheck("online(在线)", camera.HandleInterval)
					//fmt.Println("online")
				}
			}
		}

	}()
	return nil
}
