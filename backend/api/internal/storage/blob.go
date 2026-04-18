package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	logger := slog.With("component", "blob_storage", "operation", "upload_raw_video",
		"video_id", videoID, "blob_name", blobName, "container", s.rawContainer)

	logger.Info("upload starting", "content_type", contentType, "size_bytes", size)

	opts := &azblob.UploadStreamOptions{
		BlockSize:   4 * 1024 * 1024, // 4MB
		HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
	}

	_, err := s.client.UploadStream(ctx, s.rawContainer, blobName, reader, opts)
	if err != nil {
		logger.Error("upload failed", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		return "", fmt.Errorf("failed to upload raw video stream: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.rawContainer, blobName)
	logger.Info("upload succeeded", "duration_ms", time.Since(start).Milliseconds(), "bytes_uploaded", size)
	return url, nil
}

func (s *BlobStorage) GenerateVideoDirectorySASURL(ctx context.Context, videoID string) (masterURL string, sasToken string, err error) {
	logger := slog.With("component", "blob_storage", "operation", "generate_sas", "video_id", videoID)

	cred, err := azdatalake.NewSharedKeyCredential(s.accountName, s.accountKey)
	if err != nil {
		logger.Error("failed to create datalake credential", "error", err.Error())
		return "", "", fmt.Errorf("create datalake shared key credential: %w", err)
	}

	now := time.Now().UTC()
	expiryTime := now.Add(2 * time.Hour)

	perms := datalakesas.DirectoryPermissions{
		Read:    true,
		List:    true,
		Execute: true,
	}

	sv := datalakesas.DatalakeSignatureValues{
		Protocol:       datalakesas.ProtocolHTTPS,
		StartTime:      now.Add(-5 * time.Minute),
		ExpiryTime:     expiryTime,
		FileSystemName: s.hlsContainer,
		DirectoryPath:  videoID,
		Permissions:    perms.String(),
	}

	qp, err := sv.SignWithSharedKey(cred)
	if err != nil {
		logger.Error("failed to sign directory SAS", "error", err.Error())
		return "", "", fmt.Errorf("sign directory SAS: %w", err)
	}

	enc := qp.Encode()

	// Extract sr parameter to confirm directory-level SAS (sr=d)
	srParam := "sr=?"
	for _, p := range strings.Split(enc, "&") {
		if strings.HasPrefix(p, "sr=") {
			srParam = p
			break
		}
	}

	// blob endpoint for serving — dfs endpoint for signing
	master := fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s/%s/master.m3u8?%s",
		s.accountName, s.hlsContainer, videoID, enc,
	)

	logger.Info("SAS token generated",
		"expiry_time", expiryTime.Format(time.RFC3339),
		"permissions", perms.String(),
		"sr", srParam,
		"container", s.hlsContainer,
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
	logger := slog.With("component", "blob_storage", "operation", "delete_raw_video", "video_id", videoID)

	blobName := fmt.Sprintf("%s.mp4", videoID)
	_, err := s.client.DeleteBlob(ctx, s.rawContainer, blobName, nil)
	if err != nil {
		logger.Error("failed to delete raw video blob", "blob_name", blobName, "error", err.Error())
		return fmt.Errorf("failed to delete raw video: %w", err)
	}

	logger.Info("raw video blob deleted", "blob_name", blobName, "blobs_deleted", 1)
	return nil
}

func (s *BlobStorage) DeleteHLSFolder(ctx context.Context, videoID string) error {
	logger := slog.With("component", "blob_storage", "operation", "delete_hls_folder", "video_id", videoID)

	// With HNS enabled, use the datalake client to delete the entire
	// directory in a single API call instead of listing + deleting blobs individually
	dirClient := s.datalake.NewFileSystemClient(s.hlsContainer).NewDirectoryClient(videoID)

	_, err := dirClient.Delete(ctx, nil)
	if err != nil {
		// If directory doesn't exist, treat as success — nothing to delete
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) && respErr.StatusCode == http.StatusNotFound {
			logger.Info("HLS folder already deleted or never existed")
			return nil
		}
		logger.Error("failed to delete HLS directory", "error", err.Error())
		return fmt.Errorf("failed to delete HLS directory %s: %w", videoID, err)
	}

	logger.Info("HLS folder deleted", "container", s.hlsContainer, "blobs_deleted", 1)
	return nil
}
