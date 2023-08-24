package tmdb

import (
	"fengqi/kodi-metadata-tmdb-cli/config"
	"fengqi/kodi-metadata-tmdb-cli/utils"

	"io"
	"io/ioutil"
	"net/http"
)

var Api *tmdb
var HttpClient *http.Client

const (
	ApiSearchTv           = "/3/search/tv"
	ApiSearchMovie        = "/3/search/movie"
	ApiTvDetail           = "/3/tv/%d"
	ApiTvEpisode          = "/3/tv/%d/season/%d/episode/%d"
	ApiTvAggregateCredits = "/3/tv/%d/aggregate_credits"
	ApiTvContentRatings   = "/3/tv/%d/content_ratings"
	ApiTvEpisodeGroup     = "/3/tv/episode_group/%s"
	ApiMovieDetail        = "/3/movie/%d"
)

func InitTmdb(config *config.TmdbConfig) {
	HttpClient = utils.GetHttpClient(config.Proxy)
	Api = &tmdb{
		apiHost:   config.ApiHost,
		apiKey:    config.ApiKey,
		imageHost: config.ImageHost,
		language:  config.Language,
		rating:    config.Rating,
	}
}

// GetImageW500 压缩后的图片
func (t *tmdb) GetImageW500(path string) string {
	if path == "" {
		return ""
	}
	return Api.imageHost + "/t/p/w500" + path
}

// GetImageOriginal 原始图片
func (t *tmdb) GetImageOriginal(path string) string {
	if path == "" {
		return ""
	}
	return Api.imageHost + "/t/p/original" + path
}

func (t *tmdb) request(api string, args map[string]string) ([]byte, error) {
	if args == nil {
		args = make(map[string]string, 0)
	}

	args["api_key"] = t.apiKey
	args["language"] = t.language

	api = t.apiHost + api + "?" + utils.StringMapToQuery(args)
	resp, err := HttpClient.Get(api)
	if err != nil {
		utils.Logger.ErrorF("request tmdb: %s err: %v", api, err)
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			utils.Logger.WarningF("request tmdb close body err: %v", err)
		}
	}(resp.Body)

	return ioutil.ReadAll(resp.Body)
}
