package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"wechat-robot-client/vars"

	"github.com/gin-gonic/gin"
)

/*
*
#### 1. 查看pprof首页
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/
```
或在浏览器中访问：
```
http://your-domain/api/v1/pprof/debug/pprof/
```

#### 2. 查看堆内存信息
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/heap
```

#### 3. 查看goroutine信息
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/goroutine?debug=1
```

#### 4. 采集CPU Profile（30秒）
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/profile?seconds=30 -o cpu.prof
```

#### 5. 查看内存分配信息
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/allocs
```

#### 6. 查看阻塞信息
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/block
```

#### 7. 查看互斥锁信息
```bash
curl http://your-domain/api/v1/pprof/debug/pprof/mutex
```

### 使用go tool pprof分析

下载profile文件后，可以使用go工具进行分析：

```bash
# 下载CPU profile
curl http://your-domain/api/v1/pprof/debug/pprof/profile?seconds=30 -o cpu.prof

# 使用go tool分析
go tool pprof cpu.prof

# 或者直接在线分析
go tool pprof http://your-domain/api/v1/pprof/debug/pprof/heap
```

### 生成可视化图表

如果安装了graphviz，可以生成可视化图表：

```bash
# 生成CPU profile的PDF图表
go tool pprof -pdf http://your-domain/api/v1/pprof/debug/pprof/profile > cpu.pdf

# 生成内存profile的PDF图表
go tool pprof -pdf http://your-domain/api/v1/pprof/debug/pprof/heap > heap.pdf
```
*/
type PprofProxy struct {
	proxy *httputil.ReverseProxy
	err   error
}

var hrefRegexp = regexp.MustCompile(`href=(["'])([^"']+)(["'])`)

func NewPprofProxyController() *PprofProxy {
	target, err := url.Parse(vars.PprofProxyURL)
	if err != nil || target.Scheme == "" || target.Host == "" {
		if err == nil {
			err = fmt.Errorf("missing scheme or host")
		}
		return &PprofProxy{err: fmt.Errorf("invalid pprof target URL %q: %w", vars.PprofProxyURL, err)}
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1/robot/pprof")
		req.Host = target.Host
	}
	proxy.ModifyResponse = func(res *http.Response) error {
		if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
			return nil
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		body = rewritePprofHTMLLinks(body, "/api/v1/robot/pprof/debug/pprof/")
		res.Body = io.NopCloser(bytes.NewReader(body))
		res.ContentLength = int64(len(body))
		res.Header.Set("Content-Length", strconv.Itoa(len(body)))
		return nil
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("pprof proxy error: " + err.Error()))
	}

	return &PprofProxy{
		proxy: proxy,
	}
}

func (p *PprofProxy) ProxyPprof(c *gin.Context) {
	if p.err != nil {
		c.String(http.StatusBadGateway, "pprof proxy error: %s", p.err.Error())
		return
	}
	p.proxy.ServeHTTP(c.Writer, c.Request)
}

func rewritePprofHTMLLinks(body []byte, basePath string) []byte {
	return hrefRegexp.ReplaceAllFunc(body, func(match []byte) []byte {
		parts := hrefRegexp.FindSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		rewritten, ok := rewritePprofHref(string(parts[2]), basePath)
		if !ok {
			return match
		}
		return []byte("href=" + string(parts[1]) + rewritten + string(parts[3]))
	})
}

func rewritePprofHref(rawHref string, basePath string) (string, bool) {
	parsed, err := url.Parse(rawHref)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "", false
	}
	profilePath := parsed.Path
	if index := strings.Index(profilePath, "/debug/pprof"); index >= 0 {
		profilePath = profilePath[index+len("/debug/pprof"):]
	} else if strings.HasPrefix(profilePath, "/") {
		return "", false
	}
	profilePath = strings.TrimLeft(profilePath, "/")

	rewritten := basePath + profilePath
	if parsed.RawQuery != "" {
		rewritten += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		rewritten += "#" + parsed.Fragment
	}
	return rewritten, true
}
