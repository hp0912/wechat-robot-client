package podcast

import "testing"

func TestPodcast(t *testing.T) {
	secrets := PodcastSecrets{
		AppID:       "",
		AccessToken: "",
	}
	config := PodcastConfig{
		Action: 0,
		InputInfo: InputInfo{
			InputURL:       "https://mp.weixin.qq.com/s/hdYs186AcxeDFBgrrQTUVg",
			ReturnAudioURL: true,
		},
	}
	audioURL, err := Podcast(secrets, config)
	if err != nil {
		t.Fatalf("Podcast failed: %v", err)
	}
	t.Logf("播客音频链接: %s", audioURL)
}
