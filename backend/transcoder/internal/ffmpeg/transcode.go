package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Transcoder struct {
	FfmpegPath string
	TempDir    string
}

type RenditionResult struct {
	Resolution   string   `json:"resolution"`
	PlaylistFile string   `json:"playlist_file"`
	SegmentFiles []string `json:"segment_files"`
	Bandwidth    int      `json:"bandwidth"`
}

type TranscodeResult struct {
	VideoID    string            `json:"video_id"`
	OutputDir  string            `json:"output_dir"`
	Renditions []RenditionResult `json:"renditions"`
}

func NewTranscoder(ffmpegPath, tempDir string) *Transcoder {
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	return &Transcoder{
		FfmpegPath: ffmpegPath,
		TempDir:    tempDir,
	}
}

func hasAudioStream(input string) bool {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		input,
	)

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(out) > 0
}

func (t *Transcoder) Transcode(ctx context.Context, videoID, inputPath string) (*TranscodeResult, error) {

	outputDir := filepath.Join(t.TempDir, videoID, "output")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	audioExists := hasAudioStream(inputPath)

	args := []string{
		"-y",
		"-threads", "2",
		"-i", inputPath,
		"-preset", "veryfast",
		"-g", "48",
		"-sc_threshold", "0",
	}

	// video maps
	for i := 0; i < 4; i++ {
		args = append(args, "-map", "0:v:0")
	}

	// audio maps if exists
	if audioExists {
		for i := 0; i < 4; i++ {
			args = append(args, "-map", "0:a:0")
		}
	}

	args = append(args,
		"-c:v", "libx264",
	)

	if audioExists {
		args = append(args,
			"-c:a", "aac",
			"-ar", "48000",
		)
	}

	// 1080p
	args = append(args,
		"-filter:v:0", "scale=1920:-2",
		"-b:v:0", "5000k",
		"-maxrate:v:0", "5300k",
		"-bufsize:v:0", "7500k",
	)

	// 720p
	args = append(args,
		"-filter:v:1", "scale=1280:-2",
		"-b:v:1", "2800k",
		"-maxrate:v:1", "2996k",
		"-bufsize:v:1", "4200k",
	)

	// 480p
	args = append(args,
		"-filter:v:2", "scale=854:-2",
		"-b:v:2", "1400k",
		"-maxrate:v:2", "1498k",
		"-bufsize:v:2", "2100k",
	)

	// 360p
	args = append(args,
		"-filter:v:3", "scale=640:-2",
		"-b:v:3", "700k",
		"-maxrate:v:3", "756k",
		"-bufsize:v:3", "1050k",
	)

	if audioExists {
		args = append(args,
			"-b:a:0", "128k",
			"-b:a:1", "128k",
			"-b:a:2", "96k",
			"-b:a:3", "64k",
		)
	}

	varStreamMap := ""

	if audioExists {
		varStreamMap = "v:0,a:0,name:1080p v:1,a:1,name:720p v:2,a:2,name:480p v:3,a:3,name:360p"
	} else {
		varStreamMap = "v:0,name:1080p v:1,name:720p v:2,name:480p v:3,name:360p"
	}

	args = append(args,
		"-var_stream_map", varStreamMap,
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, "%v_%03d.ts"),
		"-master_pl_name", "master_ffmpeg.m3u8",
		filepath.Join(outputDir, "%v.m3u8"),
	)

	cmd := exec.CommandContext(ctx, t.FfmpegPath, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout

	log.Printf("Executing FFmpeg for video %s", videoID)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg execution failed: %w\nstderr: %s", err, stderr.String())
	}

	log.Printf("FFmpeg finished successfully for video: %s", videoID)

	result := &TranscodeResult{
		VideoID:   videoID,
		OutputDir: outputDir,
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	resolutions := []struct {
		name string
		bw   int
	}{
		{"1080p", 5128000},
		{"720p", 2928000},
		{"480p", 1496000},
		{"360p", 764000},
	}

	for _, res := range resolutions {

		r := RenditionResult{
			Resolution: res.name,
			Bandwidth:  res.bw,
		}

		for _, f := range files {

			if f.IsDir() {
				continue
			}

			name := f.Name()

			if strings.HasPrefix(name, res.name) {

				if strings.HasSuffix(name, ".m3u8") {
					r.PlaylistFile = filepath.Join(outputDir, name)
				}

				if strings.HasSuffix(name, ".ts") {
					r.SegmentFiles = append(r.SegmentFiles, filepath.Join(outputDir, name))
				}
			}
		}

		if r.PlaylistFile != "" && len(r.SegmentFiles) > 0 {
			result.Renditions = append(result.Renditions, r)
		}
	}

	return result, nil
}

func (t *Transcoder) GenerateMasterPlaylist(result *TranscodeResult, outputDir string) error {
	masterPath := filepath.Join(outputDir, "master.m3u8")

	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:3\n")

	for _, r := range result.Renditions {
		var resolutionStr string
		switch r.Resolution {
		case "1080p":
			resolutionStr = "1920x1080"
		case "720p":
			resolutionStr = "1280x720"
		case "480p":
			resolutionStr = "854x480"
		case "360p":
			resolutionStr = "640x360"
		}

		sb.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n", r.Bandwidth, resolutionStr))
		sb.WriteString(fmt.Sprintf("%s.m3u8\n", r.Resolution))
	}

	if err := os.WriteFile(masterPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write master playlist: %w", err)
	}

	return nil
}

func (t *Transcoder) Cleanup(videoID string) error {
	dir := filepath.Join(t.TempDir, videoID)

	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed cleanup: %w", err)
	}

	return nil
}
