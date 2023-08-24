package subtitle

import (
	"fengqi/kodi-metadata-tmdb-cli/utils"

	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func (o *opensubtitles) getSubtitleList(tmdbId int) (files []*File, err error) {

	url := fmt.Sprintf("/api/v1/subtitles?languages=%s&tmdb_id=%d", strings.Join(o.languages, ","), tmdbId)

	b, err := o.request(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	var resp subtitlesResp
	if err = json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("no subtitle found in opensubtitles")
	}

	for _, v := range resp.Data {
		for _, f := range v.Attributes.Files {

			f.Filename = f.Filename + ".srt"
			files = append(files, f)
		}
	}

	return
}

func (o *opensubtitles) getDownloadLink(fileID int) (info *downloadInfo, err error) {
	reader := strings.NewReader(fmt.Sprintf("{\n  \"file_id\": %d\n}", fileID))
	b, err := o.request(http.MethodPost, "/api/v1/download", reader)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &info); err != nil {
		return nil, err
	}
	return
}

func (o *opensubtitles) request(method, url string, body io.Reader) ([]byte, error) {
	url = o.apiHost + url
	utils.Logger.DebugF(url)

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Api-Key", o.apiKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
