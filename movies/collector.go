package movies

import (
	"fengqi/kodi-metadata-tmdb-cli/config"
	"fengqi/kodi-metadata-tmdb-cli/kodi"
	"fengqi/kodi-metadata-tmdb-cli/subtitle"
	"fengqi/kodi-metadata-tmdb-cli/utils"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

var collector *Collector

func RunCollector(config *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Logger.FatalF("new movies watcher err: %v", err)
	}

	collector = &Collector{
		config:  config,
		watcher: watcher,
		channel: make(chan *Movie, 100),
	}

	go collector.runWatcher()
	go collector.runMoviesProcess()
	collector.runCronScan()
}

// 开启文件夹监听
func (c *Collector) runWatcher() {
	if !c.config.Collector.Watcher {
		return
	}

	utils.Logger.Debug("run movies watcher")

	// 监听顶级目录
	for _, item := range c.config.Collector.MoviesDir {
		utils.Logger.DebugF("add movies dir: %s to watcher", item)

		err := c.watcher.Add(item)
		if err != nil {
			utils.Logger.FatalF("add movies dir: %s to watcher err :%v", item, err)
		}
	}

	for {
		select {
		// 接受事件，增删改查都会收到，需要过滤，部分情况下可能收不到create而是chmod
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}

			if !event.Has(fsnotify.Create) {
				continue
			}

			fileInfo, _ := os.Stat(event.Name)
			if fileInfo == nil || (!fileInfo.IsDir() && utils.IsVideo(event.Name) == "") {
				continue
			}

			utils.Logger.InfoF("created file: %s", event.Name)

			moviesDir := parseMoviesDir(filepath.Dir(event.Name), fileInfo)
			if moviesDir != nil {
				c.channel <- moviesDir
			}

		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}

			utils.Logger.ErrorF("movies watcher error: %v", err)
		}
	}
}

// 电影信息处理：来源包括cron和inotify监听的
func (c *Collector) runMoviesProcess() {
	utils.Logger.Debug("run movies process")

	for {
		select {
		case dir := <-c.channel:
			utils.Logger.DebugF("receive movies task: %v", dir)

			dir.checkCacheDir()
			detail, err := dir.getMovieDetail()
			if err != nil || detail == nil {
				continue
			}

			if !detail.FromCache || !dir.NfoExist(c.config.Collector.MoviesNfoMode) {
				_ = dir.saveToNfo(detail, c.config.Collector.MoviesNfoMode)
				kodi.Rpc.AddRefreshTask(kodi.TaskRefreshMovie, detail.OriginalTitle)
			}

			_ = dir.downloadImage(detail)

			cacheFile := dir.GetCacheDir() + "/" + subtitle.CacheFileSubfix
			if dir.IsFile {
				cacheFile = dir.GetCacheDir() + "/" + dir.OriginTitle + "." + subtitle.CacheFileSubfix
			}

			subtitles, err := subtitle.GetSubtitles(detail.Id, cacheFile)
			if err != nil {
				utils.Logger.ErrorF("GetSubtitles error: %v", err)
				return
			}
			_ = subtitle.Download(subtitles, dir.GetFullDir())
		}
	}
}

// 运行定时扫描
func (c *Collector) runCronScan() {
	utils.Logger.DebugF("run movies scan cron_seconds: %d", c.config.Collector.CronSeconds)

	task := func() {
		for _, item := range c.config.Collector.MoviesDir {
			utils.Logger.DebugF("movies scan ticker trigger")

			movieDirs, err := c.scanDir(item)
			if err != nil {
				utils.Logger.FatalF("scan movies dir: %s err :%v", item, err)
				continue
			}

			for _, movieDir := range movieDirs {
				c.channel <- movieDir
			}
		}

		if c.config.Kodi.CleanLibrary {
			kodi.Rpc.AddCleanTask("")
		}
	}

	task()
	ticker := time.NewTicker(time.Second * time.Duration(c.config.Collector.CronSeconds))
	for {
		select {
		case <-ticker.C:
			task()
		}
	}
}

// 扫描普通目录，返回其中的电影
func (c *Collector) scanDir(dir string) ([]*Movie, error) {
	movieDirs := make([]*Movie, 0)

	if f, err := os.Stat(dir); err != nil || !f.IsDir() {
		return movieDirs, nil
	}

	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		utils.Logger.ErrorF("scan dir: %s err: %v", dir, err)
		return nil, err
	}

	for _, file := range fileInfo {
		// 合集，以 Iron.Man.2008-2013.Blu-ray.x264.MiniBD1080P-CMCT 为例，暂定使用 2008-2013 做为判断特征
		if yearRange := utils.IsYearRangeLike(file.Name()); yearRange != "" {
			movieDir, err := c.scanDir(dir + "/" + file.Name())
			if err != nil {
				utils.Logger.ErrorF("scan collection dir: %s err: %v", dir+"/"+file.Name(), err)
				continue
			}
			movieDirs = append(movieDirs, movieDir...)
			continue
		}

		movieDir := parseMoviesDir(dir, file)
		if movieDir == nil {
			continue
		}

		movieDirs = append(movieDirs, movieDir)
	}

	return movieDirs, nil
}

func (c *Collector) skipFolders(path, filename string) bool {
	base := filepath.Base(path)
	for _, item := range c.config.Collector.SkipFolders {
		if item == base || item == filename {
			return true
		}
	}
	return false
}

func (c *Collector) listFilesAndFolders(path string) []os.FileInfo {
	list := make([]os.FileInfo, 0)
	pathInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return list
	}

	for _, file := range pathInfo {
		if c.skipFolders(path, file.Name()) {
			continue
		}

		list = append(list, file)
	}

	return list
}
