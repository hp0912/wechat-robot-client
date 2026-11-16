package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"wechat-robot-client/model"
	"wechat-robot-client/pkg/robot"
	"wechat-robot-client/vars"
)

func TestLoginServiceIsRunningReadOnly(test *testing.T) {
	originalRuntime := vars.RobotRuntime
	test.Cleanup(func() { vars.RobotRuntime = originalRuntime })

	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	vars.RobotRuntime = &robot.Robot{
		WxID:   "wxid_test",
		Status: model.RobotStatusOnline,
		Client: robot.NewClient(robot.WechatDomain(strings.TrimPrefix(server.URL, "http://")), robot.ProxyInfo{}),
	}
	service := NewLoginService(context.Background())
	if !service.IsRunning() {
		test.Fatal("IsRunning() = false for a running server")
	}
	server.Close()
	if service.IsRunning() {
		test.Fatal("IsRunning() = true for a stopped server")
	}
	if vars.RobotRuntime.Status != model.RobotStatusOnline {
		test.Fatalf("IsRunning() changed runtime status to %q", vars.RobotRuntime.Status)
	}
}

func TestLoginServiceIsLoggedInReadOnly(test *testing.T) {
	cases := []struct {
		name       string
		status     model.RobotStatus
		statusCode int
		body       string
		wantLogin  bool
	}{
		{
			name:       "online profile",
			status:     model.RobotStatusOnline,
			statusCode: http.StatusOK,
			body:       `{"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_test"}}}}`,
			wantLogin:  true,
		},
		{
			name:       "offline profile",
			status:     model.RobotStatusOffline,
			statusCode: http.StatusOK,
			body:       `{"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_test"}}}}`,
			wantLogin:  true,
		},
		{
			name:       "transient error while online",
			status:     model.RobotStatusOnline,
			statusCode: http.StatusBadGateway,
			body:       `{"Code":-7}`,
			wantLogin:  true,
		},
		{
			name:       "transient error while offline",
			status:     model.RobotStatusOffline,
			statusCode: http.StatusBadGateway,
			body:       `{"Code":-7}`,
		},
		{
			name:       "transient error with unknown status",
			status:     model.RobotStatusError,
			statusCode: http.StatusServiceUnavailable,
		},
		{
			name:       "malformed profile while online",
			status:     model.RobotStatusOnline,
			statusCode: http.StatusOK,
			body:       `{}`,
			wantLogin:  true,
		},
		{
			name:       "confirmed logout while online",
			status:     model.RobotStatusOnline,
			statusCode: http.StatusOK,
			body:       `{"Code":-7}`,
		},
		{
			name:       "confirmed logout while offline",
			status:     model.RobotStatusOffline,
			statusCode: http.StatusOK,
			body:       `{"Code":-7}`,
		},
	}
	originalRuntime := vars.RobotRuntime
	test.Cleanup(func() { vars.RobotRuntime = originalRuntime })

	for _, testcase := range cases {
		test.Run(testcase.name, func(test *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/api"+robot.UserGetContactProfile {
					test.Errorf("status query triggered a lifecycle request: %s", request.URL)
				}
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(testcase.statusCode)
				fmt.Fprint(writer, testcase.body)
			}))
			defer server.Close()
			vars.RobotRuntime = &robot.Robot{
				WxID:   "wxid_test",
				Status: testcase.status,
				Client: robot.NewClient(robot.WechatDomain(strings.TrimPrefix(server.URL, "http://")), robot.ProxyInfo{}),
			}
			service := NewLoginService(context.Background())
			if loggedIn := service.IsLoggedIn(); loggedIn != testcase.wantLogin {
				test.Fatalf("IsLoggedIn() = %v, want %v", loggedIn, testcase.wantLogin)
			}
			if vars.RobotRuntime.Status != testcase.status {
				test.Fatalf("IsLoggedIn() changed runtime status to %q", vars.RobotRuntime.Status)
			}
		})
	}
}

func TestLoginServiceStatusRecoversAfterTransientFailures(test *testing.T) {
	originalRuntime := vars.RobotRuntime
	test.Cleanup(func() { vars.RobotRuntime = originalRuntime })
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api"+robot.UserGetContactProfile {
			test.Errorf("status query triggered a lifecycle request: %s", request.URL)
		}
		writer.Header().Set("Content-Type", "application/json")
		if requests.Add(1) <= 3 {
			fmt.Fprint(writer, `{"Code":-1,"Message":"temporary timeout"}`)
			return
		}
		fmt.Fprint(writer, `{"Code":0,"Data":{"userInfo":{"UserName":{"String":"wxid_test"}}}}`)
	}))
	defer server.Close()
	vars.RobotRuntime = &robot.Robot{
		WxID:   "wxid_test",
		Status: model.RobotStatusOnline,
		Client: robot.NewClient(robot.WechatDomain(strings.TrimPrefix(server.URL, "http://")), robot.ProxyInfo{}),
	}
	service := NewLoginService(context.Background())
	for attempt := 0; attempt < 4; attempt++ {
		if !service.IsLoggedIn() {
			test.Fatalf("IsLoggedIn() = false on attempt %d", attempt+1)
		}
		if vars.RobotRuntime.Status != model.RobotStatusOnline {
			test.Fatalf("runtime status changed to %q on attempt %d", vars.RobotRuntime.Status, attempt+1)
		}
	}
	if requests.Load() != 4 {
		test.Fatalf("profile requests = %d, want 4", requests.Load())
	}
}
