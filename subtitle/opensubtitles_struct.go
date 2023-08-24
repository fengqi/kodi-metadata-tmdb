package subtitle

type File struct {
	FileID   int    `json:"file_id"`
	Filename string `json:"file_name"`
}

type attributes struct {
	// SubtitleID        string    `json:"subtitle_id"`
	// Language          string    `json:"language"`
	// DownloadCount     int       `json:"download_count"`
	// NewDownloadCount  int       `json:"new_download_count"`
	// HearingImpaired   bool      `json:"hearing_impaired"`
	// Hd                bool      `json:"hd"`
	// Fps               float64   `json:"fps"`
	// Votes             int       `json:"votes"`
	// Ratings           float64   `json:"ratings"`
	// FromTrusted       bool      `json:"from_trusted"`
	// ForeignPartsOnly  bool      `json:"foreign_parts_only"`
	// UploadDate        time.Time `json:"upload_date"`
	// AiTranslated      bool      `json:"ai_translated"`
	// MachineTranslated bool      `json:"machine_translated"`
	// Release           string    `json:"release"`
	// Comments          string    `json:"comments"`
	// LegacySubtitleID  int       `json:"legacy_subtitle_id"`
	// Uploader          struct {
	// 	UploaderID int    `json:"uploader_id"`
	// 	Name       string `json:"name"`
	// 	Rank       string `json:"rank"`
	// } `json:"uploader"`
	// FeatureDetails struct {
	// 	FeatureID       int    `json:"feature_id"`
	// 	FeatureType     string `json:"feature_type"`
	// 	Year            int    `json:"year"`
	// 	Title           string `json:"title"`
	// 	MovieName       string `json:"movie_name"`
	// 	ImdbID          int    `json:"imdb_id"`
	// 	TmdbID          int    `json:"tmdb_id"`
	// 	SeasonNumber    int    `json:"season_number"`
	// 	EpisodeNumber   int    `json:"episode_number"`
	// 	ParentImdbID    int    `json:"parent_imdb_id"`
	// 	ParentTitle     string `json:"parent_title"`
	// 	ParentTmdbID    int    `json:"parent_tmdb_id"`
	// 	ParentFeatureID int    `json:"parent_feature_id"`
	// } `json:"feature_details"`
	// URL          string `json:"url"`
	// RelatedLinks []struct {
	// 	Label  string `json:"label"`
	// 	URL    string `json:"url"`
	// 	ImgURL string `json:"img_url,omitempty"`
	// } `json:"related_links"`
	// MoviehashMatch bool `json:"moviehash_match"`
	Files []*File `json:"files"`
}

type data struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Attributes attributes `json:"attributes"`
}

type subtitlesResp struct {
	TotalPages int    `json:"total_pages"`
	TotalCount int    `json:"total_count"`
	PerPage    int    `json:"per_page"`
	Page       int    `json:"page"`
	Data       []data `json:"data"`
}

type downloadInfo struct {
	Link string `json:"link"`
	// FileID   int    `json:"file_id"`
	// Filename string `json:"file_name"`
	// Requests     int       `json:"requests"`
	// Remaining    int       `json:"remaining"`
	// Message      string    `json:"message"`
	// ResetTime    string    `json:"reset_time"`
	// ResetTimeUtc time.Time `json:"reset_time_utc"`
}
