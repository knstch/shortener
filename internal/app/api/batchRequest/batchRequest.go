package batchrequest

import (
	"encoding/json"
	"net/http"

	config "github.com/knstch/shortener/cmd/config"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
	storage "github.com/knstch/shortener/internal/app/storage"
)

type originalLink struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type shortLink struct {
	Result        string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

func PostBatch(req *http.Request) []uint8 {
	var originalRequest []originalLink
	var shortenResponse []shortLink

	err := json.NewDecoder(req.Body).Decode(&originalRequest)
	if err != nil {
		logger.ErrorLogger("Failed to read json: ", err)
	}

	for i := range originalRequest {
		returnedShortLink, _ := postLongLink.PostLongLink(originalRequest[i].OriginalURL,
			&storage.StorageURLs, config.ReadyConfig.BaseURL)
		shortenResponse = append(shortenResponse,
			shortLink{
				Result:        returnedShortLink,
				CorrelationID: originalRequest[i].CorrelationID,
			})
	}
	response, err := json.Marshal(shortenResponse)
	if err != nil {
		logger.ErrorLogger("Failed to marshal json: ", err)
	}
	return response
}
