package config

type Config struct {
	Log       *LogConfig           `json:"log"`           // 日志配置
	Ffmpeg    *FfmpegConfig        `json:"ffmpeg"`        // ffmpeg配置，给音乐视频使用的
	Tmdb      *TmdbConfig          `json:"tmdb"`          // TMDB 配置
	Subtitle  *OpensubtitlesConfig `json:"opensubtitles"` // Opensubtitles 配置
	Kodi      *KodiConfig          `json:"kodi"`          // kodi配置
	Collector *CollectorConfig     `json:"collector"`     // 刮削配置
}

type KodiConfig struct {
	Enable       bool   `json:"enable"`        // 是否开启 kodi 通知
	CleanLibrary bool   `json:"clean_library"` // 是否清理媒体库
	JsonRpc      string `json:"json_rpc"`      // kodi rpc 路径
	Timeout      int    `json:"timeout"`       // 连接kodi超时时间
	Username     string `json:"username"`      // rpc 用户名
	Password     string `json:"password"`      // rpc 密码
}

type LogConfig struct {
	Mode  int    `json:"mode"`  // 日志模式：1标准输出，2日志文件，3标准输出和日志文件
	Level int    `json:"level"` // 日志等级，0-4分别是：debug，info，warning，error，fatal
	File  string `json:"file"`  // 日志文件路径
}

type FfmpegConfig struct {
	MaxWorker   int    `json:"max_worker"`   // 最大进程数：建议为逻辑CPU个数
	FfmpegPath  string `json:"ffmpeg_path"`  // ffmpeg 可执行文件路径
	FfprobePath string `json:"ffprobe_path"` // ffprobe 可执行文件路径
}

type TmdbConfig struct {
	ApiHost   string `json:"api_host"`   // TMDB 接口地址
	ApiKey    string `json:"api_key"`    // api key
	ImageHost string `json:"image_host"` // 图片地址
	Language  string `json:"language"`   // 语言
	Rating    string `json:"rating"`     // 内容分级
	Proxy     string `json:"proxy"`      // 请求经过代理，支持 http、https、socks5、socks5h
}

type OpensubtitlesConfig struct {
	ApiHost   string   `json:"api_host"`  // Opensubtitles 接口地址
	ApiKey    string   `json:"api_key"`   // api key
	Languages []string `json:"languages"` // 语言
	Proxy     string   `json:"proxy"`     // 请求经过代理，支持 http、https、socks5、socks5h
}

type CollectorConfig struct {
	Watcher        bool     `json:"watcher"`          // 是否开启文件监听，比定时扫描及时
	CronSeconds    int      `json:"cron_seconds"`     // 定时扫描频率
	SkipFolders    []string `json:"skip_folders"`     // 跳过的目录，可多个
	MoviesNfoMode  int      `json:"movies_nfo_mode"`  // 电影NFO写入模式：1 movie.nfo，2 <VideoFileName>.nfo
	MoviesDir      []string `json:"movies_dir"`       // 电影文件根目录，可多个
	ShowsDir       []string `json:"shows_dir"`        // 电视剧文件根目录，可多个
	MusicVideosDir []string `json:"music_videos_dir"` // 音乐视频文件根目录，可多个
}
