package controller

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/proxy"
)

func GenDefaultIpadUA() string {
	code := 8059
	major := 0x0f & (code >> 24)
	minor := 0xff & (code >> 16)
	patch := 0xff & (code >> 8)
	//build := 0xff & (code >> 0)
	wxVersion := strconv.Itoa(int(major)) + "." + strconv.Itoa(int(minor)) + "." + strconv.Itoa(int(patch))
	iPadOsVersionS := strings.Replace("18.8.1", ".", "_", -1)
	wechatUserAgent := fmt.Sprintf("Mozilla/5.0 (iPad; CPU iPad iPhone OS %s like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/%s(%s) NetType/WIFI Language/zh_CN", iPadOsVersionS, wxVersion, strconv.Itoa(int(code)))
	return wechatUserAgent
}

func Socks5Transport(proxyAddr string, proxyUser string, proxyPass string) (client *http.Transport, err error) {

	//设定代理
	var proxyAuth *proxy.Auth
	var transport *http.Transport
	if proxyAddr != "" && proxyAddr != "string" {
		//设定账号和用户名
		if proxyUser != "" && proxyUser != "string" && proxyPass != "" && proxyPass != "string" {
			proxyAuth = &proxy.Auth{
				User:     proxyUser,
				Password: proxyPass,
			}
		} else {
			proxyAuth = nil
		}
		dialer, err := proxy.SOCKS5("tcp", proxyAddr,
			proxyAuth,
			&net.Dialer{
				Timeout:  15 * time.Second,
				Deadline: time.Now().Add(time.Second * 15),
			},
		)
		if err != nil {
			return nil, err
		}
		transport = &http.Transport{
			Proxy:               nil,
			Dial:                dialer.Dial,
			TLSHandshakeTimeout: 15 * time.Second,
			MaxIdleConnsPerHost: -1,   //连接池禁用缓存
			DisableKeepAlives:   true, //禁用客户端连接缓存到连接池
		}
	} else {
		transport = &http.Transport{
			Proxy:               nil,
			TLSHandshakeTimeout: 15 * time.Second,
			MaxIdleConnsPerHost: -1,   //连接池禁用缓存
			DisableKeepAlives:   true, //禁用客户端连接缓存到连接池
		}
	}

	return transport, nil
}

func WxHttpRequesthb(Url string, action string, headers *map[string]string, body io.Reader, ua string, proxyAddr string, proxyUser string, proxyPass string) (*http.Response, error) {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(action, Url, body)
	} else {
		req, err = http.NewRequest(action, Url, nil)
	}
	if err != nil {
		return nil, err
	}
	if ua == "" {
		ua = GenDefaultIpadUA()
	}
	postUri, err := url.Parse(Url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,zh-TW;q=0.8,en-US;q=0.7,en;q=0.6")
	if action == "POST" {
		req.Header.Set("Host", postUri.Host)
		req.Header.Set("Origin", postUri.Scheme+"://"+postUri.Host)
		req.Header.Set("Content-type", "application/json")
		req.Header.Set("X-Requested-With", "com.tencent.mm")
	}
	if headers != nil {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}
	transport, err := Socks5Transport(proxyAddr, proxyUser, proxyPass)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	return resp, err
}

func TestLogoutCallback(t *testing.T) {
	notiUrl := "http://127.0.0.1:9001/api/v1/wechat-client/x/logout"
	notiData := fmt.Sprintf(`{"wxid":"%s","type":"heartbeat","status":"failed","retry_count":%d}`, "x", 1)
	resp, err := WxHttpRequesthb(notiUrl, "POST", nil, bytes.NewBuffer([]byte(notiData)), "", "", "", "")
	if err != nil {
		t.Errorf("Failed to send logout callback: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code: %d", resp.StatusCode)
	}
}
