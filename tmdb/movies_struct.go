package tmdb

// MovieDetail 电影详情
type MovieDetail struct {
	Adult               bool                `json:"adult"`
	BackdropPath        string              `json:"backdrop_path"`
	BelongsToCollection BelongsToCollection `json:"belongs_to_collection"`
	Budget              int                 `json:"budget"`
	Genres              []Genre             `json:"genres"`
	Homepage            string              `json:"homepage"`
	Id                  int                 `json:"id"`
	ImdbId              string              `json:"imdb_id"`
	OriginalLanguage    string              `json:"original_language"`
	OriginalTitle       string              `json:"original_title"`
	Overview            string              `json:"overview"`
	Popularity          float32             `json:"popularity"`
	PosterPath          string              `json:"poster_path"`
	ProductionCompanies []ProductionCompany `json:"production_companies"`
	ProductionCountries []ProductionCountry `json:"production_countries"`
	ReleaseDate         string              `json:"release_date"`
	Revenue             int                 `json:"revenue"`
	Runtime             int                 `json:"runtime"`
	SpokenLanguages     []SpokenLanguage    `json:"spoken_languages"`
	Status              string              `json:"status"`
	Tagline             string              `json:"tagline"`
	Title               string              `json:"title"`
	Video               bool                `json:"video"`
	VoteAverage         float32             `json:"vote_average"`
	VoteCount           int                 `json:"vote_count"`
	Credits             *Credit             `json:"credits"`
	FromCache           bool                `json:"from_cache"`
	Releases            MovieRelease        `json:"releases"`
}

type BelongsToCollection struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type Credit struct {
	Id   string      `json:"id"`
	Cast []MovieCast `json:"cast"`
	Crew []MovieCrew `json:"crew"`
}

type MovieCast struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	CastId             int     `json:"cast_id"`
	Character          string  `json:"character"`
	CreditId           string  `json:"credit_id"`
	Order              int     `json:"order"`
}

// MovieCrew 电影工作人员
type MovieCrew struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float32 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	CreditId           string  `json:"credit_id"`
	Department         string  `json:"department"`
	Job                string  `json:"job"`
}

// SearchMoviesResponse 搜索电影的结果
type SearchMoviesResponse struct {
	Page         int                    `json:"page"`
	TotalResults int                    `json:"total_results"`
	TotalPages   int                    `json:"total_pages"`
	Results      []*SearchMoviesResults `json:"results"`
}

// SearchMoviesResults 搜索电影的结果
type SearchMoviesResults struct {
	PosterPath       string  `json:"poster_path"`
	Adult            bool    `json:"adult"`
	Overview         string  `json:"overview"`
	ReleaseDate      string  `json:"release_date"`
	GenreIds         []int   `json:"genre_ids"`
	Id               int     `json:"id"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Title            string  `json:"title"`
	BackdropPath     string  `json:"backdrop_path"`
	Popularity       float32 `json:"popularity"`
	VoteCount        int     `json:"vote_count"`
	Video            bool    `json:"video"`
	VoteAverage      float32 `json:"vote_average"`
}

// MovieRelease 电影各国家上映时间和分级
type MovieRelease struct {
	Countries []ReleaseCountry `json:"countries"`
}

type ReleaseCountry struct {
	Certification string `json:"certification"`
	ISO31661      string `json:"iso_3166_1"`
	Primary       bool   `json:"primary"`
	ReleaseDate   string `json:"release_date"`
}
