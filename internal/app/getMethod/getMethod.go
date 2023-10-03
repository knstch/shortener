package getmethod

import (
	postMethod "github.com/knstch/shortener_url/internal/app/postMethod"
)

func GetMethod(url string, URLstorage postMethod.Storage) string {
	for k, v := range URLstorage.Data {
		if v == url {
			return k
		}
	}
	return ""
}
