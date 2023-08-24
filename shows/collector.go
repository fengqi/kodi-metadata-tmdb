package shows

import (
	"encoding/json"
	"fengqi/kodi-metadata-tmdb-cli/config"
	"fengqi/kodi-metadata-tmdb-cli/kodi"
	"fengqi/kodi-metadata-tmdb-cli/subtitle"
	"fengqi/kodi-metadata-tmdb-cli/utils"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Collector struct {
	config  *config.Config
	watcher *fsnotify.Watcher
	dirChan chan *Dir
}

var collector *Collector

func RunCollector(config *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.Logger.FatalF("new shows watcher err: %v", err)
	}

	collector = &Collector{
		config:  config,
		watcher: watcher,
		dirChan: make(chan *Dir, 100),
	}

	go collector.runWatcher()
	go collector.showsDirProcess()
	collector.runCronScan()
}

// 目录处理队列消费
func (c *Collector) showsDirProcess() {
	utils.Logger.Debug("run shows dir process")

	for {
		select {
		case dir := <-c.dirChan: // todo dir处理挪到独立的方法
			utils.Logger.DebugF("shows dir process receive task: %v", dir.OriginTitle)

			detail, err := dir.getTvDetail()
			if err != nil || detail == nil {
				continue
			}

			if !detail.FromCache || !dir.NfoExist() {
				_ = dir.saveToNfo(detail)
				kodi.Rpc.AddRefreshTask(kodi.TaskRefreshTVShow, detail.OriginalName)
			}

			dir.downloadImage(detail)

			if dir.IsCollection { // 合集
				subDir, err := c.scanDir(dir.GetFullDir())
				if err != nil {
					utils.Logger.ErrorF("scan collection dir: %s err: %v", dir.OriginTitle, err)
					continue
				}

				for _, item := range subDir {
					err := c.watcher.Add(item.Dir + "/" + item.OriginTitle)
					utils.Logger.DebugF("runCronScan add shows dir: %s to watcher", item.Dir+"/"+item.OriginTitle)
					if err != nil {
						utils.Logger.FatalF("add shows dir: %s to watcher err: %v", item.Dir+"/"+item.OriginTitle, err)
					}

					item.TvId = dir.TvId
					c.dirChan <- item
				}

				continue
			}

			// 普通剧集
			subFiles, err := c.scanShowsFile(dir)
			if err != nil {
				utils.Logger.ErrorF("scan shows dir: %s err: %v", dir.OriginTitle, err)
				continue
			}

			files := make(map[int]map[string]*File, 0)
			if len(subFiles) > 0 {
				files[dir.Season] = subFiles
			}

			if len(files) == 0 {
				utils.Logger.WarningF("scan shows file empty: %s", dir.OriginTitle)
				continue
			}

			// 剧集组的分集信息写入缓存, 供后面处理分集信息使用
			if dir.GroupId != "" && detail.TvEpisodeGroupDetail != nil {
				for _, group := range detail.TvEpisodeGroupDetail.Groups {
					group.SortEpisode()
					for k, episode := range group.Episodes {
						se := fmt.Sprintf("s%02de%02d", group.Order, k+1)
						file, ok := files[group.Order][se]
						if !ok {
							continue
						}

						cacheFile := fmt.Sprintf("%s/tmdb/%s.json", file.Dir, se)
						f, err := os.OpenFile(cacheFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
						if err != nil {
							utils.Logger.ErrorF("save tv to cache, open_file err: %v", err)
							return
						}

						episode.EpisodeNumber = k + 1
						episode.SeasonNumber = group.Order
						bytes, err := json.MarshalIndent(episode, "", "    ")
						if err != nil {
							utils.Logger.ErrorF("save tv to cache, marshal struct errr: %v", err)
							return
						}

						_, err = f.Write(bytes)
						_ = f.Close()
					}
				}
			}

			for _, file := range files {
				for _, subFile := range file {
					c.showsFileProcess(detail.OriginalName, subFile)
				}
			}
		}
	}
}

// 单个剧集处理
func (c *Collector) showsFileProcess(originalName string, showsFile *File) bool {
	utils.Logger.DebugF("episode process: season: %d episode: %d %s", showsFile.Season, showsFile.Episode, showsFile.OriginTitle)

	episodeDetail, err := showsFile.getTvEpisodeDetail()
	if err != nil || episodeDetail == nil {
		utils.Logger.WarningF("get tv episode detail err: %v", err)
		return false
	}

	if !episodeDetail.FromCache || !showsFile.NfoExist() {
		_ = showsFile.saveToNfo(episodeDetail)
		taskVal := fmt.Sprintf("%s|-|%d|-|%d", originalName, episodeDetail.SeasonNumber, episodeDetail.EpisodeNumber)
		kodi.Rpc.AddRefreshTask(kodi.TaskRefreshEpisode, taskVal)
	}

	showsFile.downloadImage(episodeDetail)

	cacheFile := showsFile.getCacheDir() + "/" + showsFile.SeasonEpisode + "." + subtitle.CacheFileSubfix

	subtitles, err := subtitle.GetSubtitles(episodeDetail.Id, cacheFile)
	if err != nil {
		utils.Logger.ErrorF("GetSubtitles error: %v", err)
		return false
	}
	_ = subtitle.Download(subtitles, showsFile.Dir)

	return true
}

// 目录监听，新增的增加到队列，删除的移除监听
func (c *Collector) runWatcher() {
	if !c.config.Collector.Watcher {
		return
	}

	utils.Logger.Debug("run shows watcher")

	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				continue
			}

			fileInfo, err := os.Stat(event.Name)
			if fileInfo == nil || err != nil {
				utils.Logger.WarningF("get shows stat err: %v", err)
				continue
			}

			// 删除文件夹
			if event.Has(fsnotify.Remove) && fileInfo.IsDir() {
				utils.Logger.InfoF("removed dir: %s", event.Name)

				err := c.watcher.Remove(event.Name)
				if err != nil {
					utils.Logger.WarningF("remove shows watcher: %s error: %v", event.Name, err)
				}
				continue
			}

			// 新增文件夹
			if event.Has(fsnotify.Create) && fileInfo.IsDir() {
				utils.Logger.InfoF("created dir: %s", event.Name)

				showsDir := c.parseShowsDir(filepath.Dir(event.Name), fileInfo)
				if showsDir != nil {
					c.dirChan <- showsDir
				}

				err = c.watcher.Add(event.Name)
				if err != nil {
					utils.Logger.FatalF("add shows dir: %s to watcher err: %v", event.Name, err)
				}
				continue
			}

			// 新增剧集文件
			if event.Has(fsnotify.Create) && utils.IsVideo(event.Name) != "" {
				utils.Logger.InfoF("created file: %s", event.Name)

				filePath := filepath.Dir(event.Name)
				dirInfo, _ := os.Stat(filePath)
				dir := c.parseShowsDir(filepath.Dir(filePath), dirInfo)
				if dir != nil {
					c.dirChan <- dir
				}
			}

		case err, ok := <-c.watcher.Errors:
			utils.Logger.ErrorF("shows watcher error: %v", err)

			if !ok {
				return
			}
		}
	}
}

// 目录扫描，定时任务，扫描到的目录和文件增加到队列
func (c *Collector) runCronScan() {
	utils.Logger.DebugF("run shows scan cron_seconds: %d", c.config.Collector.CronSeconds)

	task := func() {
		for _, item := range c.config.Collector.ShowsDir {
			// 扫描到的每个目录都添加到watcher，因为还不能只监听根目录
			err := c.watcher.Add(item)
			utils.Logger.DebugF("runCronScan add shows dir: %s to watcher", item)
			if err != nil {
				utils.Logger.FatalF("add shows dir: %s to watcher err: %v", item, err)
			}

			showDirs, err := c.scanDir(item)
			if err != nil {
				utils.Logger.FatalF("scan shows dir: %s err :%v", item, err)
			}

			for _, showDir := range showDirs {
				err := c.watcher.Add(showDir.Dir + "/" + showDir.OriginTitle)
				utils.Logger.DebugF("runCronScan add shows dir: %s to watcher", showDir.Dir+"/"+showDir.OriginTitle)
				if err != nil {
					utils.Logger.FatalF("add shows dir: %s to watcher err: %v", showDir.Dir+"/"+showDir.OriginTitle, err)
				}

				// 预留50%空间给可能重新放回队列的任务
				for {
					if len(c.dirChan) < 100*0.5 {
						break
					}
					time.Sleep(time.Second * 2)
				}

				c.dirChan <- showDir
			}
		}

		if c.config.Kodi.CleanLibrary {
			kodi.Rpc.AddCleanTask("")
		}
	}

	task() // TODO 启动后立即运行可控
	ticker := time.NewTicker(time.Second * time.Duration(c.config.Collector.CronSeconds))
	for {
		select {
		case <-ticker.C:
			task()
		}
	}
}

// 扫描普通目录，返回其中的电视剧
func (c *Collector) scanDir(dir string) ([]*Dir, error) {
	movieDirs := make([]*Dir, 0)

	if f, err := os.Stat(dir); err != nil || !f.IsDir() {
		return movieDirs, nil
	}

	dirEntry, err := os.ReadDir(dir)
	if err != nil {
		utils.Logger.ErrorF("scan dir: %s err: %v", dir, err)
		return nil, err
	}

	for _, entry := range dirEntry {
		if !entry.IsDir() {
			continue
		}

		fi, err := entry.Info()
		if fi == nil || err != nil {
			continue
		}

		movieDir := c.parseShowsDir(dir, fi)
		if movieDir == nil {
			continue
		}

		movieDirs = append(movieDirs, movieDir)
	}

	return movieDirs, nil
}

// ScanMovieFile 扫描可以确定的单个电影、电视机目录，返回其中的视频文件信息
func (c *Collector) scanShowsFile(d *Dir) (map[string]*File, error) {
	fileInfo, err := ioutil.ReadDir(d.Dir + "/" + d.OriginTitle)
	if err != nil {
		return nil, err
	}

	movieFiles := make([]*File, 0)
	for _, file := range fileInfo {
		movieFile := c.parseShowsFile(d, file)
		if movieFile != nil {
			if d.PartMode > 0 {
				movieFile.Part = utils.MatchPart(file.Name())
			}
			movieFiles = append(movieFiles, movieFile)
		}
	}

	// 处理分卷
	// part=1会根据part出现的次数累加, 适合没有规律的, 比如E01只有上下, E03有上中下, 但是如果中间部分剧集缺失会导致算错
	// part=2或者更大的数字会使用当前集数*2, 比如part=2的时候, E05.Part1会映射成E09, 可以缺失中间部分剧集
	if d.PartMode == 1 {
		// 使用season episode part多重排序
		sort.Slice(movieFiles, func(i, j int) bool {
			if movieFiles[i].Season == movieFiles[j].Season {
				if movieFiles[i].Episode == movieFiles[j].Episode {
					return movieFiles[i].Part < movieFiles[j].Part
				}
				return movieFiles[i].Episode < movieFiles[j].Episode
			}
			return movieFiles[i].Season < movieFiles[j].Season
		})

		// 重新计算episode
		for i, item := range movieFiles {
			item.Episode = i + 1
			item.SeasonEpisode = fmt.Sprintf("s%02de%02d", item.Season, item.Episode)
			utils.Logger.DebugF("scanShowsFile partMode=%d, correct episode to %d", d.PartMode, item.Episode, item.OriginTitle)
		}
	} else if d.PartMode > 1 {
		for _, item := range movieFiles {
			item.Episode = (item.Episode-1)*d.PartMode + item.Part
			item.SeasonEpisode = fmt.Sprintf("s%02de%02d", item.Season, item.Episode)
			utils.Logger.DebugF("scanShowsFile partMode=%d, correct episode to %d", d.PartMode, item.Episode, item.OriginTitle)
		}
	}

	// TODO 忘记这里为啥返回map，而不是slice了，先临时转成map，后续看看能不能改回来
	movieFilesMap := make(map[string]*File)
	for _, item := range movieFiles {
		movieFilesMap[item.SeasonEpisode] = item
	}

	return movieFilesMap, nil
}

// 解析文件, 返回详情
func (c *Collector) parseShowsFile(dir *Dir, file fs.FileInfo) *File {
	fileName := utils.FilterTmpSuffix(file.Name())

	// 判断是视频, 并获取后缀
	suffix := utils.IsVideo(fileName)
	if len(suffix) == 0 {
		utils.Logger.DebugF("pass : %s", file.Name())
		return nil
	}

	fileName = utils.ReplaceChsNumber(fileName)

	// 提取季和集
	se, snum, enum := utils.MatchEpisode(fileName)
	if dir.Season > 0 {
		snum = dir.Season
	}
	utils.Logger.InfoF("find season: %d episode: %d %s", snum, enum, file.Name())
	if len(se) == 0 || snum == 0 || enum == 0 {
		utils.Logger.WarningF("seaon or episode not find: %s", file.Name())
		return nil
	}

	return &File{
		Dir:           dir.Dir + "/" + dir.OriginTitle,
		OriginTitle:   utils.FilterTmpSuffix(file.Name()),
		Season:        snum,
		Episode:       enum,
		SeasonEpisode: se,
		Suffix:        suffix,
		TvId:          dir.TvId,
	}
}

// 解析目录, 返回详情
// TODO 参数合并，只需要传完整的路径
func (c *Collector) parseShowsDir(baseDir string, file fs.FileInfo) *Dir {
	showName := file.Name()

	// 过滤无用文件
	if showName[0:1] == "." || utils.InArray(collector.config.Collector.SkipFolders, showName) {
		utils.Logger.DebugF("pass file: %s", showName)
		return nil
	}

	// 过滤可选字符
	showName = utils.FilterOptionals(showName)

	// 过滤掉或替换歧义的内容
	showName = utils.SeasonCorrecting(showName)

	// 过滤掉分段的干扰
	if subEpisodes := utils.IsSubEpisodes(showName); subEpisodes != "" {
		showName = strings.Replace(showName, subEpisodes, "", 1)
	}

	showsDir := &Dir{
		Dir:          baseDir,
		OriginTitle:  file.Name(),
		IsCollection: utils.IsCollection(file.Name()),
	}

	// 年份范围
	if yearRange := utils.IsYearRange(showName); len(yearRange) > 0 {
		showsDir.YearRange = yearRange
		showName = strings.Replace(showName, yearRange, "", 1)
	}

	// 使用自定义方法切割
	split := utils.Split(showName)

	nameStart := false
	nameStop := false
	for _, item := range split {
		if year := utils.IsYear(item); year > 0 {
			// 名字带年的，比如 reply 1994
			if showsDir.Year == 0 {
				showsDir.Year = year
			} else {
				showsDir.Title += strconv.Itoa(showsDir.Year)
				showsDir.Year = year
			}
			nameStop = true
			continue
		}

		if season := utils.IsSeason(item); len(season) > 0 {
			if !showsDir.IsCollection {
				if season != item { // TODO 这里假定只有名字和season在一起，没有其他特殊字符的情况，如：黄石S01，否则可能不适合这样处理
					showsDir.Title += strings.TrimRight(item, season) + " "
				}
				s := season[1:]
				i, err := strconv.Atoi(s)
				if err == nil {
					showsDir.Season = i
					nameStop = true
				}
			}
			continue
		}

		if format := utils.IsFormat(item); len(format) > 0 {
			showsDir.Format = format
			nameStop = true
			continue
		}

		if source := utils.IsSource(item); len(source) > 0 {
			showsDir.Source = source
			nameStop = true
			continue
		}

		if studio := utils.IsStudio(item); len(studio) > 0 {
			showsDir.Studio = studio
			nameStop = true
			continue
		}

		if channel := utils.IsChannel(item); len(channel) > 0 {
			nameStop = true
			continue
		}

		if !nameStart {
			nameStart = true
			nameStop = false
		}

		if !nameStop {
			showsDir.Title += item + " "
		}
	}

	// 文件名清理
	showsDir.Title, showsDir.AliasTitle = utils.SplitTitleAlias(showsDir.Title)
	showsDir.ChsTitle, showsDir.EngTitle = utils.SplitChsEngTitle(showsDir.Title)
	if len(showsDir.Title) == 0 {
		utils.Logger.WarningF("file: %s parse title empty: %v", file.Name(), showsDir)
		return nil
	}

	// 读特殊指定的值
	showsDir.ReadSeason()
	showsDir.ReadTvId()
	showsDir.ReadGroupId()
	showsDir.checkCacheDir()
	showsDir.ReadPart()

	return showsDir
}
