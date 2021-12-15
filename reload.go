package reload

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/xiusin/logger"
	"github.com/xiusin/reload/util"
)

const winExt = ".exe"

var (
	rebuildNotifier    = make(chan struct{})
	watcher            *fsnotify.Watcher
	counter            int32
	globalCancel       func()
	printRegisterFile  bool
	subProcessCancelFn func()
)

func init() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
}

func SetPrintRegisterInfo(val bool) {
	printRegisterFile = val
}

func Loop(fn func() error, cnf *Conf) error {
	if util.IsChildMode() {
		return fn()
	}
	var ctx context.Context
	ctx, subProcessCancelFn = context.WithCancel(context.Background())
	if cnf != nil {
		if len(cnf.File) > 0 {
			defaultConf.File = cnf.File
		}
		if cnf.Cmd != nil {
			defaultConf.Cmd = cnf.Cmd
		}
	}
	parseConf()
	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, os.Interrupt, syscall.SIGTERM)
	if util.IsWindows() {
		defaultConf.conf.BuildName += winExt
	}
	_ = os.MkdirAll(filepath.Dir(defaultConf.conf.BuildName), os.ModePerm)
	_ = os.Remove(defaultConf.conf.BuildName)
	defer func() { _ = watcher.Close() }()
	if err := build(); err != nil {
		return err
	}
	if err := registerFile(); err != nil {
		panic(err)
	}
	go eventNotify()
	go serve()
	if len(defaultConf.Cmd.SubProcessCb) > 0 {
		for _, f := range defaultConf.Cmd.SubProcessCb {
			go f(ctx)
		}
	}
	<-closeCh
	if subProcessCancelFn != nil {
		subProcessCancelFn()
	}
	if globalCancel != nil {
		globalCancel()
	}
	return nil
}

func serve() {
	var nextEventCh = make(chan struct{})
	if defaultConf.Cmd.Base == nil {
		defaultConf.Cmd.Base = func(s string) string { return fmt.Sprintf("./%s", s) }
	}
	for {
		ctx, cancel := context.WithCancel(context.Background())
		globalCancel = cancel

		process := exec.CommandContext(ctx, defaultConf.Cmd.Base(defaultConf.conf.BuildName), defaultConf.Cmd.Params...)
		process.Dir = util.AppPath()
		process.Stdout = os.Stdout
		process.Stderr = os.Stdout
		process.Env = append(process.Env, defaultConf.Cmd.buildEnv()...)

		go func() {
			<-rebuildNotifier
			if process.Process != nil {
				_ = process.Process.Kill()
			}
			cancel()
			nextEventCh <- struct{}{}
		}()
		if err := process.Start(); err != nil {
			logger.Error(err)
		}
		_ = process.Wait()
		process = nil
		<-nextEventCh
	}
}

func build() error {
	start := time.Now()
	cmd := exec.Command("go", "build", "-o", defaultConf.conf.BuildName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Env = os.Environ()
	cmd.Dir = util.AppPath()
	if err := cmd.Run(); err != nil {
		return err
	}
	logger.Printf("构建耗时: %.2fs", time.Since(start).Seconds())
	return nil
}

func registerFile() error {
	files, err := util.ScanDir(defaultConf.conf.RootDir, defaultConf.conf.IgnoreDirs)
	if err != nil {
		return err
	}
	for _, file := range files {
		if counter > int32(defaultConf.conf.Limit) {
			logger.Warning("监听文件已达上限")
			break
		}
		if len(defaultConf.conf.FileExts) > 0 && !util.InSlice(".*", defaultConf.conf.FileExts) && !file.IsDir {
			suffixPartial := strings.Split(file.Path, ".")
			if !util.InSlice("."+suffixPartial[len(suffixPartial)-1], defaultConf.conf.FileExts) {
				continue
			}
		}
		if !file.IsDir && strings.HasSuffix(file.Path, defaultConf.conf.BuildName) {
			continue
		}
		if err := watcher.Add(file.Path); err != nil {
			return err
		} else if !file.IsDir {
			atomic.AddInt32(&counter, 1)
			if printRegisterFile {
				logger.Print("监听文件:", file.Path)
			}
		}
	}
	return nil
}

func eventNotify() {
	var lockerTimestamp time.Time
	var building = false
	for {
		select {
		case event := <-watcher.Events:
			if util.IsIgnoreAction(&event) {
				continue
			}
			if time.Since(lockerTimestamp) > time.Duration(defaultConf.conf.DelayMS)*time.Millisecond && !building {
				name := util.Replace(event.Name, util.AppPath(), "")
				fileInfo := strings.Split(name, ".")
				if !util.InSlice(".*", defaultConf.conf.FileExts) && !util.InSlice("."+strings.TrimRight(fileInfo[len(fileInfo)-1], "~"), defaultConf.conf.FileExts) {
					continue
				}
				lockerTimestamp, building = time.Now(), true
				if event.Op&fsnotify.Create == fsnotify.Create {
					_ = watcher.Add(event.Name)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					_ = watcher.Remove(event.Name)
				}
				logger.Warningf("%s event %s", name, strings.ToLower(event.Op.String()))
				go func() {
					if err := build(); err != nil {
						logger.Warning("构建错误", err)
						building = false
					}
					rebuildNotifier <- struct{}{}
					building = false
				}()
			}
		case err, ok := <-watcher.Errors:
			if ok {
				logger.Warning("watcher error: %s", err)
			}
		}
	}
}
