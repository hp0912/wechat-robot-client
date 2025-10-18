package controller

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

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
	targetURL string
	proxy     *httputil.ReverseProxy
}

func NewPprofProxyController(targetURL string) *PprofProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic("invalid pprof target URL: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1/pprof")
		req.Host = target.Host
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("pprof proxy error: " + err.Error()))
	}

	return &PprofProxy{
		targetURL: targetURL,
		proxy:     proxy,
	}
}

func (p *PprofProxy) ProxyPprof(c *gin.Context) {
	p.proxy.ServeHTTP(c.Writer, c.Request)
}
