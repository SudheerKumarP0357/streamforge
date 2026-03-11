package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"streamforge/transcoder/internal/config"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type BlobClient struct {
	client       *azblob.Client
	accountName  string
	rawContainer string
	hlsContainer string
	endpoint     string
}

func NewBlobClient(cfg config.Config) *BlobClient {
	if cfg.AzureStorageName == "" || cfg.AzureStorageKey == "" {
		log.Println("Warning: Azure Storage credentials not provided, transcoder cannot run properly")
		return nil
	}

	cred, err := azblob.NewSharedKeyCredential(cfg.AzureStorageName, cfg.AzureStorageKey)
	if err != nil {
		log.Fatalf("Invalid blob credentials: %v", err)
	}

	endpoint := cfg.AzureStorageEndpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", cfg.AzureStorageName)
	}

	client, err := azblob.NewClientWithSharedKeyCredential(endpoint, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create blob client: %v", err)
	}

	return &BlobClient{
		client:       client,
		accountName:  cfg.AzureStorageName,
		rawContainer: cfg.AzureBlobRawContainer,
		hlsContainer: cfg.AzureBlobHLSContainer,
		endpoint:     endpoint,
	}
}

// DownloadBlob downloads a blob from raw container to a local file
func (b *BlobClient) DownloadBlob(ctx context.Context, blobName, destPath string) error {
	if b == nil || b.client == nil {
		return fmt.Errorf("blob client not initialized")
	}

	file, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer file.Close()

	_, err = b.client.DownloadFile(ctx, b.rawContainer, blobName, file, nil)
	if err != nil {
		return fmt.Errorf("failed to download blob: %w", err)
	}

	return nil
}

// UploadBlob uploads a local file to the hls container under videoID/ prefix
func (b *BlobClient) UploadBlob(ctx context.Context, videoID string, sourcePath string, contentType string) (string, error) {
	if b == nil || b.client == nil {
		return "", fmt.Errorf("blob client not initialized")
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer file.Close()

	fileName := filepath.Base(sourcePath)
	blobName := fmt.Sprintf("%s/%s", videoID, fileName)

	opts := &azblob.UploadStreamOptions{}
	if contentType != "" {
		opts.HTTPHeaders = &blob.HTTPHeaders{BlobContentType: &contentType}
	}

	_, err = b.client.UploadStream(ctx, b.hlsContainer, blobName, file, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", b.endpoint, b.hlsContainer, blobName)
	return url, nil
}
