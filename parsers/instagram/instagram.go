package instagram

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/Gasoid/regular-go-bot/parsers"
)

type InstagramParser struct{}

func (p *InstagramParser) Name() string {
	return "instagram"
}

func (p *InstagramParser) Handler(text string, callback parsers.Callback) error {
	instagramURLs := extractInstagramURLs(text)
	if len(instagramURLs) == 0 {
		return nil // No Instagram URLs found, do nothing
	}

	for _, url := range instagramURLs {
		if err := processInstagramURL(url, callback); err != nil {
			slog.Error("failed to process Instagram URL", "url", url, "error", err)
			callback.ReplyMessage(fmt.Sprintf("Failed to process Instagram content: %v", err))
		}
	}

	return nil
}

func extractInstagramURLs(text string) []string {
	// Regex pattern to match Instagram URLs
	pattern := `https?://(?:www\.)?instagram\.com/(?:p|reel)/[A-Za-z0-9_-]+/?`
	re := regexp.MustCompile(pattern)
	return re.FindAllString(text, -1)
}

func processInstagramURL(url string, callback parsers.Callback) error {
	slog.Info("Processing Instagram URL", "url", url)

	// Try to extract media using a simple approach
	// Note: Instagram has strict anti-scraping measures, so this is a simplified approach
	// In production, you might need to use specialized services or APIs

	mediaInfo, err := extractMediaInfo(url)
	if err != nil {
		return fmt.Errorf("failed to extract media info: %w", err)
	}

	if mediaInfo.IsVideo {
		// Download and send video
		filePath, err := downloadMedia(mediaInfo.URL, "video")
		if err != nil {
			return fmt.Errorf("failed to download video: %w", err)
		}
		defer os.Remove(filePath)

		callback.SendVideo(filePath)
	} else {
		// Download and send photo
		filePath, err := downloadMedia(mediaInfo.URL, "photo")
		if err != nil {
			return fmt.Errorf("failed to download photo: %w", err)
		}
		defer os.Remove(filePath)

		callback.SendPhoto(filePath, "📸 Instagram photo")
	}

	return nil
}

type MediaInfo struct {
	URL     string
	IsVideo bool
	Caption string
}

func extractMediaInfo(instagramURL string) (*MediaInfo, error) {
	// Instagram has complex anti-scraping measures. This is a simplified approach.
	// For production use, consider:
	// 1. Instagram Basic Display API (requires user authentication)
	// 2. Third-party services like RapidAPI Instagram scrapers
	// 3. Browser automation tools

	// Try to get the page content with proper headers
	client := &http.Client{}
	req, err := http.NewRequest("GET", instagramURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Instagram page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instagram returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	bodyStr := string(body)

	// Try to extract media from various possible JSON structures
	// Look for video URLs first
	videoPatterns := []string{
		`"video_url":"([^"]+)"`,
		`"src":"([^"]+\.mp4[^"]*)"`,
		`videoUrl":"([^"]+)"`,
	}

	for _, pattern := range videoPatterns {
		videoRe := regexp.MustCompile(pattern)
		if match := videoRe.FindStringSubmatch(bodyStr); len(match) > 1 {
			videoURL := strings.ReplaceAll(match[1], "\\u0026", "&")
			videoURL = strings.ReplaceAll(videoURL, "\\/", "/")
			return &MediaInfo{
				URL:     videoURL,
				IsVideo: true,
			}, nil
		}
	}

	// Look for image URLs
	imagePatterns := []string{
		`"display_url":"([^"]+)"`,
		`"src":"([^"]+\.jpg[^"]*)"`,
		`"src":"([^"]+\.jpeg[^"]*)"`,
		`"thumbnail_src":"([^"]+)"`,
	}

	for _, pattern := range imagePatterns {
		imageRe := regexp.MustCompile(pattern)
		if match := imageRe.FindStringSubmatch(bodyStr); len(match) > 1 {
			imageURL := strings.ReplaceAll(match[1], "\\u0026", "&")
			imageURL = strings.ReplaceAll(imageURL, "\\/", "/")
			return &MediaInfo{
				URL:     imageURL,
				IsVideo: false,
			}, nil
		}
	}

	// If direct extraction fails, try a fallback approach
	// Look for any high-resolution image URLs in the page
	fallbackPattern := `https://[^"]*\.(?:jpg|jpeg|png|mp4)[^"]*`
	fallbackRe := regexp.MustCompile(fallbackPattern)
	matches := fallbackRe.FindAllString(bodyStr, -1)

	for _, match := range matches {
		if strings.Contains(match, "instagram") && (strings.Contains(match, "jpg") || strings.Contains(match, "jpeg") || strings.Contains(match, "png")) {
			return &MediaInfo{
				URL:     match,
				IsVideo: false,
			}, nil
		}
		if strings.Contains(match, "instagram") && strings.Contains(match, "mp4") {
			return &MediaInfo{
				URL:     match,
				IsVideo: true,
			}, nil
		}
	}

	return nil, fmt.Errorf("no media found in Instagram post - Instagram may have blocked access")
}

func downloadMedia(url, mediaType string) (string, error) {
	// Create HTTP client with proper headers
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://www.instagram.com/")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Determine file extension based on content type or URL
	var ext string
	contentType := resp.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "video/mp4") || strings.Contains(url, ".mp4"):
		ext = ".mp4"
	case strings.Contains(contentType, "image/jpeg") || strings.Contains(url, ".jpg") || strings.Contains(url, ".jpeg"):
		ext = ".jpg"
	case strings.Contains(contentType, "image/png") || strings.Contains(url, ".png"):
		ext = ".png"
	case mediaType == "video":
		ext = ".mp4"
	default:
		ext = ".jpg"
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("instagram_%s_*%s", mediaType, ext))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Copy media data to file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save media: %w", err)
	}

	slog.Info("Downloaded Instagram media", "file", tmpFile.Name(), "type", mediaType, "size", tmpFile.Name())
	return tmpFile.Name(), nil
}

func init() {
	parsers.RegisterTextParser(&InstagramParser{})
}
