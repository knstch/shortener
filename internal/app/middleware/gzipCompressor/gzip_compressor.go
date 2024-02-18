// Модуль gzipcompressor имеет middleware для сжатия данных
// и модифицированные методы интерфейса http.ResponseWriter.
package gzipcompressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	errorLogger "github.com/knstch/shortener/internal/app/logger"
)

type gzipWriter struct {
	res http.ResponseWriter
	zw  *gzip.Writer
}

func newGzipWriter(res http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		res: res,
		zw:  gzip.NewWriter(res),
	}
}

// Header - это модифицированный метод интерфейса http.ResponseWriter.
func (gw *gzipWriter) Header() http.Header {
	return gw.res.Header()
}

// Header - это модифицированный метод интерфейса http.ResponseWriter.
func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.zw.Write(b)
}

// WriteHeader - это модифицированный метод интерфейса http.ResponseWriter.
func (gw *gzipWriter) WriteHeader(statusCode int) {
	gw.res.Header().Set("Content-Encoding", "gzip")
	gw.res.WriteHeader(statusCode)
}

// Close - это модифицированный метод интерфейса http.ResponseWriter.
func (gw *gzipWriter) Close() error {
	return gw.zw.Close()
}

type gzipReader struct {
	req io.ReadCloser
	zr  *gzip.Reader
}

func newCompressReader(req io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(req)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		req: req,
		zr:  zr,
	}, nil
}

// Read - это модифицированный метод интерфейса http.ResponseWriter.
func (gr *gzipReader) Read(b []byte) (n int, err error) {
	return gr.zr.Read(b)
}

// Close - это модифицированный метод интерфейса http.ResponseWriter.
func (gr *gzipReader) Close() error {
	if err := gr.req.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}

// GzipMiddleware сжимает данные в формате gzip, если клиент
// может распаковать эти данные.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		originalRes := res
		supportsGzip := strings.Contains(req.Header.Get("Accept-Encoding"), "gzip")
		contentEncodingGzip := strings.Contains(req.Header.Get("Content-Encoding"), "gzip")
		if supportsGzip {
			compressedRes := newGzipWriter(res)
			originalRes = compressedRes
			err := compressedRes.Close()
			if err != nil {
				errorLogger.ErrorLogger("Can't close compressed response body: ", err)
			}
		}
		if contentEncodingGzip {
			decompressedReq, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				errorLogger.ErrorLogger("Error during decompression: ", err)
				return
			}
			req.Body = decompressedReq
			err = decompressedReq.Close()
			if err != nil {
				errorLogger.ErrorLogger("Can't close decompressed response body: ", err)
			}
		}
		h.ServeHTTP(originalRes, req)
	})
}
