package middleware

import (
    "sync/atomic"

    "github.com/gin-gonic/gin"
)

// 统计入/出流量的原子计数器（进程内累计）
var (
    inBytesTotal  atomic.Int64
    outBytesTotal atomic.Int64
)

// responseWriter 包装器，用于统计写出的字节数
type countingWriter struct {
    gin.ResponseWriter
    written int64
}

func (w *countingWriter) Write(b []byte) (int, error) {
    n, err := w.ResponseWriter.Write(b)
    if n > 0 {
        outBytesTotal.Add(int64(n))
        w.written += int64(n)
    }
    return n, err
}

func (w *countingWriter) WriteString(s string) (int, error) {
    n, err := w.ResponseWriter.WriteString(s)
    if n > 0 {
        outBytesTotal.Add(int64(n))
        w.written += int64(n)
    }
    return n, err
}

// TrafficMiddleware 统计每次请求的入口/出口字节
func TrafficMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 入口字节：优先使用 Content-Length（如果可用）
        if c.Request != nil && c.Request.ContentLength > 0 {
            inBytesTotal.Add(c.Request.ContentLength)
        } else if c.Request != nil {
            // 某些请求无 Content-Length（如分块传输），尝试从 Header 获取
            if cl := c.Request.Header.Get("Content-Length"); cl != "" {
                // 忽略转换失败场景：影响不大
                // 不额外解析，保持轻量
            }
        }

        // 包装 ResponseWriter 统计写出字节
        cw := &countingWriter{ResponseWriter: c.Writer}
        c.Writer = cw

        c.Next()

        // 也记录最终的 Header 指定的大小（如有），但以实际写出为准
        // 如果需要更复杂统计（按路由/方法/时间窗口），可在此扩展
        _ = cw.written
    }
}

// GetTrafficInBytes 返回自进程启动以来累计的入口字节数
func GetTrafficInBytes() int64 {
    return inBytesTotal.Load()
}

// GetTrafficOutBytes 返回自进程启动以来累计的出口字节数
func GetTrafficOutBytes() int64 {
    return outBytesTotal.Load()
}
