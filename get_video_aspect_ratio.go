package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	type resolution struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}
	}

	excmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	var buf bytes.Buffer
	excmd.Stdout = &buf

	err := excmd.Run()
	if err != nil {
		return "", err
	}

	res := resolution{}
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		return "", err
	}

	if len(res.Streams) <= 0 {
		return "", fmt.Errorf("Do not get video")
	}

	div := fmt.Sprintf("%.3f", float32(res.Streams[0].Height)/float32(res.Streams[0].Width))

	switch div {
	case "0.562":
		return "16:9", nil
	case "1.776":
		return "9:16", nil
	default:
		return "other", nil
	}

}
