package pkg

type JimengRequest struct {
	Model          string  `json:"model"`
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	SampleStrength float64 `json:"sample_strength"`
}

type JimengConfig struct {
	SessionID []string `json:"sessionid"`
	JimengRequest
}

type JimengResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func Jimeng(config *JimengConfig) (string, error) {
	// TODO: Implement Jimeng function
	return "", nil
}
