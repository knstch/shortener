package batchrequest

import (
	"encoding/json"
	"strings"

	config "github.com/knstch/shortener/cmd/config"
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
)

type originalLink struct {
	OriginalURL string `json:"original_url"`
}

type shortLink struct {
	Result string `json:"short_url"`
}

func PostBatch(req []byte) []uint8 {
	var shortLinks []string
	var responseBatch []shortLink
	trimedBatch := strings.Trim(string(req), `[]`)
	splitedBatch := strings.Split(trimedBatch, ",")
	for i := range splitedBatch {
		var longLink originalLink
		json.Unmarshal([]byte(splitedBatch[i]), &longLink)
		shortLinks = append(shortLinks, postLongLink.PostLongLink(longLink.OriginalURL,
			&URLstorage.StorageURLs, config.ReadyConfig.BaseURL))
	}

	for i := range shortLinks {
		var responseToJSON = shortLink{
			Result: shortLinks[i],
		}
		responseBatch = append(responseBatch, responseToJSON)
	}
	resp, err := json.Marshal(responseBatch)
	if err != nil {
		logger.ErrorLogger("Fail during convertion to json: ", err)
	}

	return resp
}
