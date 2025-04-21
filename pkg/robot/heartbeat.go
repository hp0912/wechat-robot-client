package robot

type AutoHeartbeatStatusResponse struct {
	Success bool   `json:"Success"`
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    any    `json:"Data"`
	Running bool   `json:"Running"`
}
