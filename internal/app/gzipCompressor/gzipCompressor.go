package gzipcompressor

import (
	"compress/gzip"
	"io"
	"net/http"
)

type gzipWriter struct {
	res http.ResponseWriter
	zw  *gzip.Writer
}

func NewGzipWriter(res http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		res: res,
		zw:  gzip.NewWriter(res),
	}
}

func (gw *gzipWriter) Header() http.Header {
	return gw.res.Header()
}

func (gw *gzipWriter) Write(b []byte) (int, error) {
	return gw.res.Write(b)
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.Header().Set("Content-Encoding", "gzip")
	}
	gw.res.WriteHeader(statusCode)
}

func (gw *gzipWriter) Close() error {
	return gw.zw.Close()
}

type gzipReader struct {
	req io.ReadCloser
	zr  *gzip.Reader
}

func NewCompressReader(req io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(req)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		req: req,
		zr:  zr,
	}, err
}

func (gr *gzipReader) Read(b []byte) (n int, err error) {
	return gr.zr.Read(b)
}

func (gr *gzipReader) Close() error {
	if err := gr.req.Close(); err != nil {
		return err
	}
	return gr.req.Close()
}
