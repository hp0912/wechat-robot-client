package robot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientResponseLogoutError(test *testing.T) {
	response := ClientResponse[struct{}]{Code: -7}
	if err := response.CheckError(nil); !errors.Is(err, ErrNotLoggedIn) {
		test.Fatalf("CheckError() = %v, want ErrNotLoggedIn", err)
	}
	if err := response.CheckError(context.DeadlineExceeded); !errors.Is(err, context.DeadlineExceeded) {
		test.Fatalf("CheckError() = %v, want the transport error", err)
	}
}

func TestRobotCheckLoggedIn(test *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
		wantLogin  bool
		wantError  bool
	}{
		{
			name:       "logged in",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_test"}}}}`,
			wantLogin:  true,
		},
		{
			name:       "confirmed logout",
			statusCode: http.StatusOK,
			body:       `{"Success":false,"Code":-7,"Message":"已退出登录"}`,
		},
		{
			name:       "logout code overrides success flag",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":-7,"Message":"已退出登录"}`,
		},
		{
			name:       "transient protocol error",
			statusCode: http.StatusOK,
			body:       `{"Success":false,"Code":-1,"Message":"请求超时"}`,
			wantError:  true,
		},
		{
			name:       "unknown protocol error",
			statusCode: http.StatusOK,
			body:       `{"Success":false,"Code":-99,"Message":"unknown error"}`,
			wantError:  true,
		},
		{
			name:       "gateway error with profile",
			statusCode: http.StatusBadGateway,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_test"}}}}`,
			wantError:  true,
		},
		{
			name:       "gateway error with logout code",
			statusCode: http.StatusServiceUnavailable,
			body:       `{"Success":false,"Code":-7}`,
			wantError:  true,
		},
		{
			name:       "malformed JSON",
			statusCode: http.StatusOK,
			body:       `{"Code":`,
			wantError:  true,
		},
		{
			name:       "empty response",
			statusCode: http.StatusOK,
			wantError:  true,
		},
		{
			name:       "missing user info",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{}}`,
			wantError:  true,
		},
		{
			name:       "missing user name",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{}}}`,
			wantError:  true,
		},
		{
			name:       "missing user name string",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{"UserName":{}}}}`,
			wantError:  true,
		},
		{
			name:       "empty user name",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{"UserName":{"String":""}}}}`,
			wantError:  true,
		},
		{
			name:       "different account",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_other"}}}}`,
			wantError:  true,
		},
		{
			name:       "nested protocol error",
			statusCode: http.StatusOK,
			body:       `{"Success":true,"Code":0,"Data":{"baseResponse":{"ret":-1},"userInfo":{"UserName":{"String":"wxid_test"}}}}`,
			wantError:  true,
		},
	}

	for _, testcase := range cases {
		test.Run(testcase.name, func(test *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost || request.URL.Path != "/api"+UserGetContactProfile || request.URL.Query().Get("wxid") != "wxid_test" {
					test.Errorf("unexpected profile request: %s %s", request.Method, request.URL)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testcase.statusCode)
				fmt.Fprint(writer, testcase.body)
			}))
			defer server.Close()

			runtime := &Robot{
				WxID:   "wxid_test",
				Client: NewClient(WechatDomain(strings.TrimPrefix(server.URL, "http://")), ProxyInfo{}),
			}
			loggedIn, err := runtime.CheckLoggedIn()
			if loggedIn != testcase.wantLogin || (err != nil) != testcase.wantError {
				test.Fatalf("CheckLoggedIn() = (%v, %v), want (%v, error=%v)", loggedIn, err, testcase.wantLogin, testcase.wantError)
			}
		})
	}
}

func TestRobotCheckLoggedInWithoutAccount(test *testing.T) {
	runtime := &Robot{}
	if loggedIn, err := runtime.CheckLoggedIn(); loggedIn || err != nil {
		test.Fatalf("CheckLoggedIn() = (%v, %v), want (false, nil)", loggedIn, err)
	}
}

func TestRobotCheckLoggedInTransportFailure(test *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	runtime := &Robot{
		WxID:   "wxid_test",
		Client: NewClient(WechatDomain(strings.TrimPrefix(server.URL, "http://")), ProxyInfo{}),
	}
	if loggedIn, err := runtime.CheckLoggedIn(); loggedIn || err == nil || errors.Is(err, ErrNotLoggedIn) {
		test.Fatalf("CheckLoggedIn() = (%v, %v), want an unknown login state with a transport error", loggedIn, err)
	}
}

func TestRobotCheckLoggedInTimeout(test *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	client := NewClient(WechatDomain(strings.TrimPrefix(server.URL, "http://")), ProxyInfo{})
	client.client.SetTimeout(20 * time.Millisecond)
	runtime := &Robot{WxID: "wxid_test", Client: client}
	if loggedIn, err := runtime.CheckLoggedIn(); loggedIn || !errors.Is(err, context.DeadlineExceeded) {
		test.Fatalf("CheckLoggedIn() = (%v, %v), want a deadline error", loggedIn, err)
	}
}
