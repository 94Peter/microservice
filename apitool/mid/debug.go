package mid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func DebugHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		fmt.Println("-------Request-------")
		fmt.Println()
		fmt.Println("remote IP: " + c.RemoteIP())
		fmt.Println("client IP: " + c.ClientIP())
		path := c.FullPath()
		fmt.Println("full path: " + path)
		path = fmt.Sprintf("%s,%s?%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
		fmt.Println("path: " + path)
		header, _ := json.Marshal(c.Request.Header)
		fmt.Println("header: " + string(header))
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println("read body fail: ", err)
		}
		c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(b))
		fmt.Println("body: " + string(b))
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw
		start := time.Now()
		c.Next()
		delta := time.Since(start)
		if delta.Seconds() > 3 {
			fmt.Println("!!!! too slow !!!")
		}
		fmt.Println("-------End Request-------")
		fmt.Println("-------Response-------")
		fmt.Println(c.Writer.Status())
		header, _ = json.Marshal(c.Writer.Header())
		fmt.Println("header: " + string(header))
		fmt.Println("Response body: " + blw.body.String())
		fmt.Println("-------End Response-------")
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
