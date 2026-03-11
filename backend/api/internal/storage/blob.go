package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake"
	datalakesas "github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/sas"
	datalakeservice "github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/service"

	"streamforge/api/internal/metrics"
)

type BlobStorage struct {
	client       *azblob.Client
	datalake     *datalakeservice.Client
	accountName  string
	accountKey   string
	rawContainer string
	hlsContainer string
	endpoint     string
}

func NewBlobStorage(accountName, accountKey, endpoint, rawContainer, hlsContainer string) (*BlobStorage, error) {
	if accountName == "" || accountKey == "" {
		return nil, fmt.Errorf("azure storage credentials missing")
	}

	blobCred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("invalid azure storage credentials: %w", err)
	}

	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
	}

	blobClient, err := azblob.NewClientWithSharedKeyCredential(endpoint, blobCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure blob client: %w", err)
	}

	// --- NEW azdatalake client for directory SAS ---
	// HNS accounts expose a separate DFS endpoint — this is required for sr=d SAS tokens
	dfsEndpoint := fmt.Sprintf("https://%s.dfs.core.windows.net", accountName)

	dlCred, err := azdatalake.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("invalid datalake credentials: %w", err)
	}

	dlClient, err := datalakeservice.NewClientWithSharedKeyCredential(dfsEndpoint, dlCred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create datalake client: %w", err)
	}

	return &BlobStorage{
		client:       blobClient,
		datalake:     dlClient,
		accountName:  accountName,
		accountKey:   accountKey,
		rawContainer: rawContainer,
		hlsContainer: hlsContainer,
		endpoint:     endpoint,
	}, nil
}

func (s *BlobStorage) UploadRawVideo(ctx context.Context, videoID string, reader io.Reader, size int64, contentType string) (string, error) {
	start := time.Now()
	defer func() {
		metrics.BlobUploadDuration.WithLabelValues(s.rawContainer).Observe(time.Since(start).Seconds())
	}()

	blobName := fmt.Sprintf("%s.mp4", videoID)

	opts := &azblob.UploadStreamOptions{
		BlockSize:   4 * 1024 * 1024, // 4MB
		HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
	}

	_, err := s.client.UploadStream(ctx, s.rawContainer, blobName, reader, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload raw video stream: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.rawContainer, blobName)
	return url, nil
}

func (s *BlobStorage) GenerateVideoDirectorySASURL(ctx context.Context, videoID string) (masterURL string, sasToken string, err error) {
	cred, err := azdatalake.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		return "", "", fmt.Errorf("create datalake shared key credential: %w", err)
	}

	now := time.Now().UTC()

	perms := datalakesas.DirectoryPermissions{
		Read:    true,
		List:    true,
		Execute: true,
	}

	sv := datalakesas.DatalakeSignatureValues{
		Protocol:       datalakesas.ProtocolHTTPS,
		StartTime:      now.Add(-5 * time.Minute),
		ExpiryTime:     now.Add(2 * time.Hour),
		FileSystemName: s.hlsContainer,
		DirectoryPath:  videoID,
		Permissions:    perms.String(),
	}

	qp, err := sv.SignWithSharedKey(cred)
	if err != nil {
		return "", "", fmt.Errorf("sign directory SAS: %w", err)
	}

	enc := qp.Encode()

	// blob endpoint for serving — dfs endpoint for signing
	master := fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s/%s/master.m3u8?%s",
		s.accountName, s.hlsContainer, videoID, enc,
	)

	log.Printf("[SAS] video=%s sp=%s sr=%s expires=%s",
		videoID, perms.String(),
		func() string {
			for _, p := range strings.Split(enc, "&") {
				if strings.HasPrefix(p, "sr=") {
					return p
				}
			}
			return "sr=?"
		}(),
		now.Add(2*time.Hour).Format(time.RFC3339),
	)

	// Return master URL and raw token separately
	return master, enc, nil
}

func (s *BlobStorage) UploadHLSFile(ctx context.Context, videoID, filename string, data []byte, contentType string) error {
	blobName := fmt.Sprintf("%s/%s", videoID, filename)

	opts := &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
	}

	_, err := s.client.UploadBuffer(ctx, s.hlsContainer, blobName, data, opts)
	if err != nil {
		return fmt.Errorf("failed to upload hls file %s: %w", filename, err)
	}

	return nil
}

func (s *BlobStorage) DeleteRawVideo(ctx context.Context, videoID string) error {
	blobName := fmt.Sprintf("%s.mp4", videoID)
	_, err := s.client.DeleteBlob(ctx, s.rawContainer, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete raw video: %w", err)
	}
	return nil
}

func (s *BlobStorage) DeleteHLSFolder(ctx context.Context, videoID string) error {
	// With HNS enabled, use the datalake client to delete the entire
	// directory in a single API call instead of listing + deleting blobs individually
	dirClient := s.datalake.NewFileSystemClient(s.hlsContainer).NewDirectoryClient(videoID)

	_, err := dirClient.Delete(ctx, nil)
	if err != nil {
		// If directory doesn't exist, treat as success — nothing to delete
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
			log.Printf("[storage] HLS folder %s already deleted or never existed", videoID)
			return nil
		}
		return fmt.Errorf("failed to delete HLS directory %s: %w", videoID, err)
	}

	log.Printf("[storage] deleted HLS folder for video=%s", videoID)
	return nil
}
