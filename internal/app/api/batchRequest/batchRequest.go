package batchrequest

import (
	"encoding/json"
	"net/http"

	config "github.com/knstch/shortener/cmd/config"
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
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
		shortenResponse = append(shortenResponse,
			shortLink{
				Result: postLongLink.PostLongLink(originalRequest[i].OriginalURL,
					&URLstorage.StorageURLs, config.ReadyConfig.BaseURL),
				CorrelationId: originalRequest[i].CorrelationId,
			})
	}
	response, err := json.Marshal(shortenResponse)
	if err != nil {
		logger.ErrorLogger("Failed to marshal json: ", err)
	}
	return response
}
