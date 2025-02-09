package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func (cfg apiConfig) getObjectURL(key string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
}

func getAssetPath(mediaType string) string {
	ext := mediaTypeToExt(mediaType)
	base := make([]byte, 32)
	_, err := rand.Read(base)
	if err != nil {
		panic("failed to generate random bytes")
	}
	fileName := base64.RawURLEncoding.EncodeToString(base)
	return fmt.Sprintf("%s%s", fileName, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}

func getVideoAspectRatio(filePath string) (string, error) {
	type AutoGenerated struct {
		Streams []struct {
			// Index              int    `json:"index"`
			// Width              int    `json:"width,omitempty"`
			// Height             int    `json:"height,omitempty"`
			// CodedWidth         int    `json:"coded_width,omitempty"`
			// CodedHeight        int    `json:"coded_height,omitempty"`
			// SampleAspectRatio  string `json:"sample_aspect_ratio,omitempty"`
			DisplayAspectRatio string `json:"display_aspect_ratio,omitempty"`
		} `json:"streams"`
	}

	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-print_format",
		"json",
		"-show_streams",
		filePath,
	)
	var jsonBuffer bytes.Buffer
	cmd.Stdout = &jsonBuffer
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	videoMetadata := AutoGenerated{}
	err := json.Unmarshal(jsonBuffer.Bytes(), &videoMetadata)
	if err != nil {
		return "", fmt.Errorf("couldn't unmarshal json data: %v", err)
	}

	// fmt.Printf("Height: %d\n", videoMetadata.Streams[0].Height)
	// fmt.Printf("Width: %d\n\n", videoMetadata.Streams[0].Width)

	ratio := videoMetadata.Streams[0].DisplayAspectRatio
	if ratio == "16:9" || ratio == "9:16" {
		return ratio, nil
	}

	return "other", nil
}
