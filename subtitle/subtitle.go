package subtitle

import (
	"fengqi/kodi-metadata-tmdb-cli/config"
	"fengqi/kodi-metadata-tmdb-cli/utils"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

type opensubtitles struct {
	apiHost   string
	apiKey    string
	languages []string
	client    *http.Client
}

const CacheFileSubfix = "opensubtitles.json"

var osb *opensubtitles

func InitSubtitles(config *config.OpensubtitlesConfig) {
	osb = &opensubtitles{
		apiHost:   config.ApiHost,
		apiKey:    config.ApiKey,
		languages: config.Languages,
		client:    utils.GetHttpClient(config.Proxy),
	}
}

func GetSubtitles(tmdbId int, cacheFile string) (subtitles []*File, err error) {
	// 从缓存读取
	if cf, err := os.Stat(cacheFile); err == nil && cf.Size() > 0 {
		utils.Logger.DebugF("get subtitles from cache: %s", cacheFile)

		bytes, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			utils.Logger.WarningF("read opensubtitles.json cache: %s err: %v", cacheFile, err)
		}

		err = json.Unmarshal(bytes, &subtitles)
		if err != nil {
			utils.Logger.WarningF("parse opensubtitles: %s file err: %v", cacheFile, err)
		}
	}

	// 缓存失效，重新搜索
	if subtitles == nil {

		// 获取字幕
		subtitles, err = osb.getSubtitleList(tmdbId)
		if err != nil {
			utils.Logger.ErrorF("get opensubtitles err: %v", err)
			return
		}

		// 保存到缓存
		file, _ := json.MarshalIndent(subtitles, "", " ")
		_ = ioutil.WriteFile(cacheFile, file, 0644)

	}

	return
}

func Download(files []*File, dir string) error {
	for _, f := range files {

		filename := filepath.Join(dir, f.Filename)

		if info, err := os.Stat(filename); err == nil && info.Size() > 0 {
			continue
		}

		info, err := osb.getDownloadLink(f.FileID)
		if err != nil {
			return err
		}
		utils.DownloadFile(info.Link, filename)
	}

	return nil
}
