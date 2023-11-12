package postlonglinkjson

import (
	"encoding/json"

	config "github.com/knstch/shortener/cmd/config"
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
)

// Структура для приема URL
type link struct {
	URL string `json:"url"`
}

// Структура для записи в json
type result struct {
	Result string `json:"result"`
}

// Функция принимает ссылку в json и отдает короткую в json
func PostLongLinkJSON(req []byte) ([]uint8, int) {
	var longLink link
	json.Unmarshal(req, &longLink)
	shortenURL, statusCode := postLongLink.PostLongLink(string(longLink.URL), &URLstorage.StorageURLs, config.ReadyConfig.BaseURL)
	var resultJSON = result{
		Result: shortenURL,
	}
	resp, err := json.Marshal(resultJSON)
	if err != nil {
		logger.ErrorLogger("Fail during convertion to json: ", err)
	}
	return resp, statusCode
}
