package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/preedep/go-nixcopy/internal/domain/entity"
	"github.com/preedep/go-nixcopy/internal/domain/repository"
	appconfig "github.com/preedep/go-nixcopy/internal/infrastructure/config"
)

// GCSStorage implements repository.Storage for Google Cloud Storage.
type GCSStorage struct {
	config *appconfig.GCSConfig
	client *storage.Client
}

func NewGCSStorage(cfg *appconfig.GCSConfig) repository.Storage {
	return &GCSStorage{config: cfg}
}

func (g *GCSStorage) Connect(ctx context.Context) error {
	var opts []option.ClientOption

	switch g.config.AuthType {
	case appconfig.GCSAuthServiceAccount:
		switch {
		case g.config.CredentialsFile != "":
			opts = append(opts, option.WithAuthCredentialsFile(option.ServiceAccount, g.config.CredentialsFile))
		case g.config.CredentialsJSON != "":
			opts = append(opts, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(g.config.CredentialsJSON)))
		default:
			return fmt.Errorf("service_account auth requires credentials_file or credentials_json")
		}

	case appconfig.GCSAuthImpersonate:
		if g.config.ImpersonateServiceAccount == "" {
			return fmt.Errorf("impersonate auth requires impersonate_service_account")
		}
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: g.config.ImpersonateServiceAccount,
			Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
			Delegates:       g.config.Delegates,
		})
		if err != nil {
			return fmt.Errorf("failed to create impersonation credentials: %w", err)
		}
		opts = append(opts, option.WithTokenSource(ts))

	case appconfig.GCSAuthAccessToken:
		if g.config.AccessToken == "" {
			return fmt.Errorf("access_token auth requires access_token")
		}
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: g.config.AccessToken})
		opts = append(opts, option.WithTokenSource(ts))

	case appconfig.GCSAuthApplicationDefault, "workload_identity", "":
		// ADC covers: GCE SA, GKE Workload Identity, Cloud Run, App Engine, local gcloud.
		// "workload_identity" is accepted as a user-friendly alias — GKE Workload Identity
		// is surfaced through ADC automatically without extra configuration.

	default:
		return fmt.Errorf("unsupported GCS auth type: %s", g.config.AuthType)
	}

	if g.config.Endpoint != "" {
		opts = append(opts, option.WithEndpoint(g.config.Endpoint))
		// Emulator mode (application_default or empty auth with a custom endpoint):
		// redirect all HTTP traffic — including the XML API used by NewReader — to
		// the emulator host, and suppress credential requirements.
		if g.config.AuthType == appconfig.GCSAuthApplicationDefault || g.config.AuthType == "" {
			emulatorURL, err := url.Parse(g.config.Endpoint)
			if err != nil {
				return fmt.Errorf("invalid GCS endpoint URL: %w", err)
			}
			opts = append(opts,
				option.WithoutAuthentication(),
				option.WithHTTPClient(&http.Client{
					Transport: &gcsEmulatorTransport{
						base:   http.DefaultTransport,
						host:   emulatorURL.Host,
						scheme: emulatorURL.Scheme,
					},
				}),
			)
		}
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}

	g.client = client
	return nil
}

func (g *GCSStorage) Disconnect(ctx context.Context) error {
	if g.client != nil {
		return g.client.Close()
	}
	return nil
}

func (g *GCSStorage) List(ctx context.Context, path string) ([]entity.FileInfo, error) {
	if g.client == nil {
		return nil, fmt.Errorf("GCS client not connected")
	}

	prefix := strings.TrimPrefix(path, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	it := g.client.Bucket(g.config.Bucket).Objects(ctx, &storage.Query{Prefix: prefix})

	var files []entity.FileInfo
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list GCS objects: %w", err)
		}
		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}
		files = append(files, entity.FileInfo{
			Path:         attrs.Name,
			Name:         filepath.Base(attrs.Name),
			Size:         attrs.Size,
			ModifiedTime: attrs.Updated,
			IsDirectory:  false,
		})
	}

	return files, nil
}

func (g *GCSStorage) Read(ctx context.Context, path string) (io.ReadCloser, int64, error) {
	if g.client == nil {
		return nil, 0, fmt.Errorf("GCS client not connected")
	}

	key := strings.TrimPrefix(path, "/")
	obj := g.client.Bucket(g.config.Bucket).Object(key)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get GCS object attributes: %w", err)
	}

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open GCS object: %w", err)
	}

	return rc, attrs.Size, nil
}

func (g *GCSStorage) Stat(ctx context.Context, path string) (*entity.FileInfo, error) {
	if g.client == nil {
		return nil, fmt.Errorf("GCS client not connected")
	}

	key := strings.TrimPrefix(path, "/")
	attrs, err := g.client.Bucket(g.config.Bucket).Object(key).Attrs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get GCS object attributes: %w", err)
	}

	return &entity.FileInfo{
		Path:         path,
		Name:         filepath.Base(path),
		Size:         attrs.Size,
		ModifiedTime: attrs.Updated,
		IsDirectory:  false,
	}, nil
}

func (g *GCSStorage) Write(ctx context.Context, path string, reader io.Reader, size int64) error {
	if g.client == nil {
		return fmt.Errorf("GCS client not connected")
	}

	key := strings.TrimPrefix(path, "/")
	wc := g.client.Bucket(g.config.Bucket).Object(key).NewWriter(ctx)

	if _, err := io.Copy(wc, reader); err != nil {
		_ = wc.Close()
		return fmt.Errorf("failed to write GCS object: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to finalize GCS object: %w", err)
	}

	return nil
}

func (g *GCSStorage) CreateDirectory(ctx context.Context, path string) error {
	return nil
}

func (g *GCSStorage) Delete(ctx context.Context, path string) error {
	if g.client == nil {
		return fmt.Errorf("GCS client not connected")
	}

	key := strings.TrimPrefix(path, "/")
	if err := g.client.Bucket(g.config.Bucket).Object(key).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete GCS object: %w", err)
	}

	return nil
}

// gcsEmulatorTransport rewrites every outbound request to target fake-gcs-server.
// The Go GCS SDK uses two HTTP APIs:
//   - JSON API: /storage/v1/... — routed correctly by option.WithEndpoint
//   - XML API:  /<bucket>/<object> — used by NewReader; fake-gcs serves this at
//     /download/storage/v1/b/<bucket>/o/<object>?alt=media
//
// This transport fixes both: it redirects all host:port to the emulator and
// rewrites XML-style paths to the fake-gcs download path.
type gcsEmulatorTransport struct {
	base   http.RoundTripper
	host   string
	scheme string
}

func (t *gcsEmulatorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Host = t.host
	clone.URL.Scheme = t.scheme
	clone.Host = t.host

	// Rewrite XML API path "/<bucket>/<object>" → fake-gcs download path.
	// XML paths have exactly two non-empty segments and no leading /storage/ prefix.
	p := clone.URL.Path
	if !strings.HasPrefix(p, "/storage/") && !strings.HasPrefix(p, "/upload/") && !strings.HasPrefix(p, "/download/") {
		// strip leading slash, split into bucket + object
		parts := strings.SplitN(strings.TrimPrefix(p, "/"), "/", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			clone.URL.Path = "/download/storage/v1/b/" + parts[0] + "/o/" + parts[1]
			q := clone.URL.Query()
			q.Set("alt", "media")
			clone.URL.RawQuery = q.Encode()
		}
	}

	return t.base.RoundTrip(clone)
}
