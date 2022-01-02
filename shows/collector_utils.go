package shows

import (
	"fengqi/kodi-metadata-tmdb-cli/tmdb"
	"fengqi/kodi-metadata-tmdb-cli/utils"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 通过一个绝对路径的文件分析剧集详情
func getFileInfoByFile(file string) *File {
	f, _ := os.Stat(file)
	return parseShowsFile(getDirInfoByFile(file), f)
}

// 通过一个绝对路径的文件分析剧集目录详情
func getDirInfoByFile(file string) *Dir {
	baseDir := filepath.Dir(file)
	fileInfo, _ := os.Stat(baseDir)
	return parseShowsDir(filepath.Dir(baseDir), fileInfo)
}

// 解析文件, 返回详情
func parseShowsFile(dir *Dir, file fs.FileInfo) *File {
	fileName := utils.FilterTmpSuffix(file.Name())

	// 判断是视频, 并获取后缀
	suffix := utils.IsVideo(fileName)
	if len(suffix) == 0 {
		utils.Logger.DebugF("pass : %s", fileName)
		return nil
	}

	// 提取季和集
	se, snum, enum := utils.MatchEpisode(fileName)
	if dir.Season > 0 {
		snum = dir.Season
	}
	utils.Logger.InfoF("find season: %d episode: %d %s", snum, enum, fileName)
	if len(se) == 0 || snum == 0 || enum == 0 {
		utils.Logger.WarningF("seaon or episode not find")
		return nil
	}

	return &File{
		Dir:           dir.Dir + "/" + dir.OriginTitle,
		OriginTitle:   fileName,
		Season:        snum,
		Episode:       enum,
		SeasonEpisode: se,
		Suffix:        suffix,
		TvId:          dir.TvId,
	}
}

// 解析目录, 返回详情
func parseShowsDir(baseDir string, file fs.FileInfo) *Dir {
	// 跳过合集
	if utils.IsCollection(file.Name()) {
		utils.Logger.WarningF("file pass, maybe not dir, or is collection: %s", file.Name())
		return nil
	}

	// 用点号.或者空格分割
	split := strings.Split(file.Name(), ".")
	if split == nil || len(split) < 3 {
		split = strings.Split(file.Name(), " ")
		if split == nil || len(split) < 5 {
			utils.Logger.WarningF("file name: %s syntax err, skipped", file.Name())
			return nil
		}
	}

	showsDir := &Dir{Dir: baseDir, OriginTitle: file.Name()}
	nameStart := false
	nameStop := false
	for _, item := range split {
		if year := utils.IsYear(item); year > 0 {
			showsDir.Year = year
			nameStop = true
			continue
		}

		if season := utils.IsSeason(item); len(season) > 0 {
			s := season[1:]
			i, err := strconv.Atoi(s)
			if err == nil {
				showsDir.Season = i
				nameStop = true
				continue
			}
		}

		if format := utils.IsFormat(item); len(format) > 0 {
			showsDir.Format = format
			nameStop = true
			continue
		}

		if utils.IsSubEpisodes(item) {
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

		if !nameStart {
			nameStart = true
			nameStop = false
		}

		if !nameStop {
			showsDir.Title += item + " "
		}
	}

	showsDir.Title = utils.CleanTitle(showsDir.Title)

	if len(showsDir.Title) == 0 {
		utils.Logger.WarningF("file: %s parse title empty: %v", file.Name(), showsDir)
		return nil
	}

	if showsDir.Season == 0 {
		showsDir.Season = 1
		seasonFile := baseDir + "/" + file.Name() + "/tmdb/season.txt"
		if _, err := os.Stat(seasonFile); err == nil {
			bytes, err := ioutil.ReadFile(seasonFile)
			if err == nil {
				showsDir.Season, _ = strconv.Atoi(strings.Trim(string(bytes), "\r\n "))
			} else {
				utils.Logger.WarningF("read season specially file: %s err: %v", seasonFile, err)
			}
		}
	}

	idFile := baseDir + "/" + file.Name() + "/tmdb/id.txt"
	if _, err := os.Stat(idFile); err == nil {
		bytes, err := ioutil.ReadFile(idFile)
		if err == nil {
			showsDir.TvId, _ = strconv.Atoi(strings.Trim(string(bytes), "\r\n "))
		} else {
			utils.Logger.WarningF("read tv id specially file: %s err: %v", idFile, err)
		}
	}

	return showsDir
}

func (f *File) getNfoFile() string {
	return f.Dir + "/" + f.getTitleWithoutSuffix() + ".nfo"
}

func (d *Dir) getNfoFile() string {
	return d.GetFullDir() + "/tvshow.nfo"
}

func (f *File) getTitleWithoutSuffix() string {
	return strings.Replace(f.OriginTitle, "."+f.Suffix, "", 1)
}

func (d *Dir) GetCacheDir() string {
	return d.GetFullDir() + "/tmdb"
}

func (d *Dir) GetFullDir() string {
	return d.Dir + "/" + d.OriginTitle
}

// CheckCacheDir tmdb 缓存目录
func (d *Dir) checkCacheDir() {
	dir := d.GetCacheDir()
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			utils.Logger.ErrorF("create cache: %s dir err: %v", dir, err)
		}
	}
}

func (f *File) getCacheDir() string {
	return f.Dir + "/tmdb"
}

// 下载电视剧的相关图片
func (d *Dir) downloadImage(detail *tmdb.TvDetail) {
	utils.Logger.DebugF("download %s images", d.Title)

	if len(detail.PosterPath) > 0 {
		_ = utils.DownloadFile(tmdb.ImageOriginal+detail.PosterPath, d.GetFullDir()+"/poster.jpg")
	}

	if len(detail.BackdropPath) > 0 {
		_ = utils.DownloadFile(tmdb.ImageOriginal+detail.BackdropPath, d.GetFullDir()+"/fanart.jpg")
	}
}

// 下载剧集的相关图片
func (f *File) downloadImage(d *tmdb.TvEpisodeDetail) {
	file := f.getTitleWithoutSuffix()
	if len(d.StillPath) > 0 {
		_ = utils.DownloadFile(tmdb.ImageOriginal+d.StillPath, f.Dir+"/"+file+"-thumb.jpg")
	}
}