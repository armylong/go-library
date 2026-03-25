package internal

import (
	"bytes"
	"github.com/gin-gonic/gin"
)

type CopyResponseWriter struct {
	gin.ResponseWriter
	Buffer *bytes.Buffer
}

func (c *CopyResponseWriter) Write(bytes []byte) (int, error) {
	_, _ = c.Buffer.Write(bytes)
	return c.ResponseWriter.Write(bytes)
}

func (c *CopyResponseWriter) Bytes() []byte {
	return c.Buffer.Bytes()
}
func (c *CopyResponseWriter) String() string {
	return c.Buffer.String()
}
