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
	rebuildNotifier  = make(chan struct{})
	watcher          *fsnotify.Watcher
	counter          int32
	globalCancel     func()
	execCmdConf      *CmdConf
	printReisterFile bool
)

func init() {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
}

func SetPrintRegisterInfo(val bool) {
	printReisterFile = val
}

func Loop(fn func() error, cnf *CmdConf) error {
	if util.IsChildMode() {
		return fn()
	}
	execCmdConf = cnf
	if cnf == nil {
		execCmdConf = &cmdConf
	}
	closeCh := make(chan os.Signal, 1)
	signal.Notify(closeCh, os.Interrupt, syscall.SIGTERM)
	if util.IsWindows() {
		conf.BuildName += winExt
	}
	_ = os.MkdirAll(filepath.Dir(conf.BuildName), os.ModePerm)
	_ = os.Remove(conf.BuildName)
	defer func() { _ = watcher.Close() }()
	if err := build(); err != nil {
		return err
	}
	if err := registerFile(); err != nil {
		panic(err)
	}
	go eventNotify()
	go serve()
	<-closeCh
	if globalCancel != nil {
		globalCancel()
	}
	return nil
}

func serve() {
	var nextEventCh = make(chan struct{})
	if execCmdConf.Base == nil {
		execCmdConf.Base = func(s string) string { return fmt.Sprintf("./%s", s) }
	}
	for {
		ctx, cancel := context.WithCancel(context.Background())
		globalCancel = cancel

		process := exec.CommandContext(ctx, execCmdConf.Base(conf.BuildName), execCmdConf.Params...)
		process.Dir = util.AppPath()
		process.Stdout = os.Stdout
		process.Stderr = os.Stdout
		process.Env = append(process.Env, execCmdConf.buildEnv()...)

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
	cmd := exec.Command("go", "build", "-o", conf.BuildName)
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
	files, err := util.ScanDir(conf.RootDir, conf.IgnoreDirs)
	if err != nil {
		return err
	}
	for _, file := range files {
		if counter > int32(conf.Limit) {
			logger.Warning("监听文件已达上限")
			break
		}
		if len(conf.FileExts) > 0 && !util.InSlice(".*", conf.FileExts) && !file.IsDir {
			suffixPartial := strings.Split(file.Path, ".")
			if !util.InSlice("."+suffixPartial[len(suffixPartial)-1], conf.FileExts) {
				continue
			}
		}
		if !file.IsDir && strings.HasSuffix(file.Path, conf.BuildName) {
			continue
		}
		if err := watcher.Add(file.Path); err != nil {
			return err
		} else if !file.IsDir {
			atomic.AddInt32(&counter, 1)
			if printReisterFile {
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
			if time.Since(lockerTimestamp) > time.Duration(conf.DelayMS)*time.Millisecond && !building {
				name := util.Replace(event.Name, util.AppPath(), "")
				fileInfo := strings.Split(name, ".")
				if !util.InSlice(".*", conf.FileExts) && !util.InSlice("."+strings.TrimRight(fileInfo[len(fileInfo)-1], "~"), conf.FileExts) {
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
