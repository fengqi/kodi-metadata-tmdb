package movies

import (
	"fengqi/kodi-metadata-tmdb-cli/tmdb"
	"fengqi/kodi-metadata-tmdb-cli/utils"

	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// 解析目录, 返回详情
// TODO 跳过电视剧，放错目录了
func parseMoviesDir(baseDir string, file fs.FileInfo) *Movie {
	movieName := utils.FilterTmpSuffix(file.Name())

	// 过滤无用文件
	if movieName[0:1] == "." || utils.InArray(collector.config.Collector.SkipFolders, movieName) {
		return nil
	}

	// 过滤可选字符
	movieName = utils.FilterOptionals(movieName)

	// 使用目录或者没有后缀的文件名
	suffix := utils.IsVideo(movieName)
	if !file.IsDir() {
		if suffix != "" {
			movieName = strings.Replace(movieName, "."+suffix, "", 1)
		} else {
			return nil
		}
	}

	// 使用自定义方法切割
	split := utils.Split(movieName)

	// 文件名识别
	nameStart := false
	nameStop := false
	movieDir := &Movie{Dir: baseDir, OriginTitle: file.Name(), IsFile: !file.IsDir(), Suffix: suffix}
	for _, item := range split {
		if item == "TLOTR" {
			continue
		}

		if resolution := utils.IsResolution(item); resolution != "" {
			nameStop = true
			continue
		}

		if year := utils.IsYear(item); year > 0 {
			movieDir.Year = year
			nameStop = true
			continue
		}

		if format := utils.IsFormat(item); len(format) > 0 {
			nameStop = true
			continue
		}

		if source := utils.IsSource(item); len(source) > 0 {
			nameStop = true
			continue
		}

		if studio := utils.IsStudio(item); len(studio) > 0 {
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
			movieDir.Title += item + " "
		}
	}

	movieDir.Title, movieDir.AliasTitle = utils.SplitTitleAlias(movieDir.Title)
	movieDir.ChsTitle, movieDir.EngTitle = utils.SplitChsEngTitle(movieDir.Title)
	if len(movieDir.Title) == 0 {
		utils.Logger.WarningF("file: %s parse title empty: %v", file.Name(), movieDir)
		return nil
	}

	// 通过文件指定id
	// todo all use baseDir + "/tmdb/"
	idFile := baseDir + "/" + file.Name() + "/tmdb/id.txt"
	if !file.IsDir() {
		idFile = baseDir + "/tmdb/" + movieName + ".id.txt"
	}
	if _, err := os.Stat(idFile); err == nil {
		bytes, err := ioutil.ReadFile(idFile)
		if err == nil {
			movieDir.MovieId, _ = strconv.Atoi(strings.Trim(string(bytes), "\r\n "))
		} else {
			utils.Logger.WarningF("read movies id specially file: %s err: %v", idFile, err)
		}
	}

	//识别是否时蓝光或dvd目录
	if file.IsDir() {
		fileInfo, err := ioutil.ReadDir(baseDir + "/" + file.Name())
		if err == nil {
			audioTs := false
			videoTs := false
			for _, item := range fileInfo {
				if item.IsDir() && item.Name() == "BDMV" || item.Name() == "CERTIFICATE" {
					movieDir.IsBluRay = true
					break
				}

				if item.IsDir() && item.Name() == "AUDIO_TS" {
					audioTs = true
				}
				if item.IsDir() && item.Name() == "VIDEO_TS" {
					videoTs = true
				}
				if videoTs && audioTs {
					movieDir.IsDvd = true
					break
				}

				if suffix := utils.IsVideo(item.Name()); suffix != "" {
					movieDir.IsSingleFile = true
					movieDir.VideoFileName = item.Name()
					break
				}
			}
		}
	}

	return movieDir
}

// tmdb 缓存目录
// TODO 统一使用一个目录
func (d *Movie) checkCacheDir() {
	dir := d.GetCacheDir()
	if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			utils.Logger.ErrorF("create cache: %s dir err: %v", dir, err)
		}
	}
}

func (d *Movie) GetCacheDir() string {
	if d.IsFile {
		return d.Dir + "/tmdb"
	}
	return d.GetFullDir() + "/tmdb"
}

func (d *Movie) GetFullDir() string {
	return d.Dir + "/" + d.OriginTitle
}

func (m *Movie) VideoFileNameWithoutSuffix() string {
	if !m.IsSingleFile {
		return ""
	}

	suffix := utils.IsVideo(m.VideoFileName)
	return m.GetFullDir() + "/" + strings.Replace(m.VideoFileName, "."+suffix, "", 1)
}

func (d *Movie) downloadImage(detail *tmdb.MovieDetail) error {
	utils.Logger.DebugF("download %s images", d.Title)

	var err error
	if len(detail.PosterPath) > 0 {
		posterFile := ""
		if d.IsFile {
			suffix := utils.IsVideo(d.OriginTitle)
			posterFile = d.Dir + "/" + strings.Replace(d.OriginTitle, "."+suffix, "", 1) + "-poster.jpg"
		} else if name := d.VideoFileNameWithoutSuffix(); name != "" {
			posterFile = name + "-poster.jpg"
		}
		err = utils.DownloadFile(tmdb.Api.GetImageOriginal(detail.PosterPath), posterFile)
	}

	if len(detail.BackdropPath) > 0 {
		fanArtFile := ""
		if d.IsFile {
			suffix := utils.IsVideo(d.OriginTitle)
			fanArtFile = d.Dir + "/" + strings.Replace(d.OriginTitle, "."+suffix, "", 1) + "-fanart.jpg"
		} else if name := d.VideoFileNameWithoutSuffix(); name != "" {
			fanArtFile = name + "-fanart.jpg"
		}
		err = utils.DownloadFile(tmdb.Api.GetImageOriginal(detail.BackdropPath), fanArtFile)
	}

	if len(detail.ProductionCompanies) > 0 {
		for _, item := range detail.ProductionCompanies {
			if item.LogoPath == "" {
				continue
			}

			logoFile := ""
			if d.IsFile {
				suffix := utils.IsVideo(d.OriginTitle)
				logoFile = d.Dir + "/" + strings.Replace(d.OriginTitle, "."+suffix, "", 1) + "-clearlogo.png"
			} else if name := d.VideoFileNameWithoutSuffix(); name != "" {
				logoFile = name + "-clearlogo.png"
			}
			if err = utils.DownloadFile(tmdb.Api.GetImageOriginal(item.LogoPath), logoFile); err == nil {
				break
			}
		}
	}

	return err
}

// maybe <VideoFileName>.nfo
// Kodi比较推荐 <VideoFileName>.nfo 但是存在一种情况就是，使用inotify监听文件变动，可能电影目录先创建
// 里面的视频文件会迟一点，这个时候 VideoFileName 就会为空，导致NFO写入失败
// 部分资源可能存在 <VideoFileName>.nfo 且里面写入了一些无用的信息，会产生冲突
// 如果使用 movie.nfo 就不需要考虑这个情况，但是需要打开媒体源的 "电影在以片名命名的单独目录中"
func (m *Movie) getNfoFile(mode int) string {
	if m.IsFile {
		suffix := utils.IsVideo(m.OriginTitle)
		return m.Dir + "/" + strings.Replace(m.OriginTitle, "."+suffix, "", 1) + ".nfo"
	}

	if m.IsBluRay {
		if utils.FileExist(m.GetFullDir() + "/BDMV/MovieObject.bdmv") {
			return m.GetFullDir() + "/BDMV/MovieObject.nfo"
		}
		return m.GetFullDir() + "/BDMV/index.nfo"
	}

	if m.IsDvd {
		return m.GetFullDir() + "/VIDEO_TS/VIDEO_TS.nfo"
	}

	if mode == 2 {
		return m.GetFullDir() + "/movie.nfo"
	}

	if mode == 1 && m.VideoFileName == "" {
		return ""
	}

	return m.VideoFileNameWithoutSuffix() + ".nfo"
}

func (m *Movie) NfoExist(mode int) bool {
	nfo := m.getNfoFile(mode)

	if info, err := os.Stat(nfo); err == nil && info.Size() > 0 {
		return true
	}

	return false
}
