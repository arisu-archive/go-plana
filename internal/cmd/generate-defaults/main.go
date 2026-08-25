package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	versionURL          = "https://ba.pokeguy.dev/com.YostarJP.BlueArchive/version.txt"
	configURLFormat     = "https://ba.pokeguy.dev/com.YostarJP.BlueArchive/decompiled/%s/GameMainConfig.json"
	generatedOutputPath = "plana_gen.go"
	generatorUserAgent  = "go-plana-defaults-generator/1.0"
	requestTimeout      = 30 * time.Second
	maxResponseSize     = 4 << 20
)

type defaults struct {
	Version       string
	BundleVersion string
	GatewayURL    string
	GameURL       string
}

type generator struct {
	client          *http.Client
	versionURL      string
	configURLFormat string
	outputPath      string
}

type gameMainConfig struct {
	ServerInfoDataURL string `json:"ServerInfoDataUrl"`
}

type serverInfo struct {
	ConnectionGroups []connectionGroup `json:"ConnectionGroups"`
}

type connectionGroup struct {
	BundleVersion   string `json:"BundleVersion"`
	GatewayURL      string `json:"GatewayUrl"`
	GameURL         string `json:"ApiUrl"`
	IsLivePublished bool   `json:"IsLivePublished"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	g := generator{
		client:          &http.Client{Timeout: requestTimeout},
		versionURL:      versionURL,
		configURLFormat: configURLFormat,
		outputPath:      generatedOutputPath,
	}
	generated, err := g.run(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "generated %s for version %s\n", g.outputPath, generated.Version)
	return nil
}

func (g generator) run(ctx context.Context) (defaults, error) {
	generated, err := g.fetchDefaults(ctx)
	if err != nil {
		return defaults{}, err
	}

	source, err := renderDefaults(generated)
	if err != nil {
		return defaults{}, err
	}
	if err := os.WriteFile(g.outputPath, source, 0o644); err != nil { //nolint:gosec // generated Go source is intentionally world-readable
		return defaults{}, fmt.Errorf("write generated defaults to %q: %w", g.outputPath, err)
	}

	return generated, nil
}

func (g generator) fetchDefaults(ctx context.Context) (defaults, error) {
	versionBody, err := g.fetch(ctx, g.versionURL)
	if err != nil {
		return defaults{}, fmt.Errorf("fetch version: %w", err)
	}
	version := strings.TrimSpace(string(versionBody))
	if version == "" {
		return defaults{}, errors.New("version response is empty")
	}

	configURL := fmt.Sprintf(g.configURLFormat, url.PathEscape(version))
	var config gameMainConfig
	if err := g.fetchJSON(ctx, configURL, &config); err != nil {
		return defaults{}, fmt.Errorf("fetch game main config: %w", err)
	}
	if config.ServerInfoDataURL == "" {
		return defaults{}, errors.New("game main config has an empty ServerInfoDataUrl")
	}

	var info serverInfo
	if err := g.fetchJSON(ctx, config.ServerInfoDataURL, &info); err != nil {
		return defaults{}, fmt.Errorf("fetch server info: %w", err)
	}
	for _, group := range info.ConnectionGroups {
		if !group.IsLivePublished {
			continue
		}

		generated := defaults{
			Version:       version,
			BundleVersion: group.BundleVersion,
			GatewayURL:    group.GatewayURL,
			GameURL:       group.GameURL,
		}
		if err := validateDefaults(generated); err != nil {
			return defaults{}, err
		}
		return generated, nil
	}

	return defaults{}, errors.New("server info has no live-published connection group")
}

func (g generator) fetchJSON(ctx context.Context, sourceURL string, destination any) error {
	body, err := g.fetch(ctx, sourceURL)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return fmt.Errorf("decode JSON from %q: %w", sourceURL, err)
	}
	return nil
}

func (g generator) fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %q: %w", sourceURL, err)
	}
	request.Header.Set("User-Agent", generatorUserAgent)

	response, err := g.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %q: %w", sourceURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request %q: unexpected HTTP status %s", sourceURL, response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("read response from %q: %w", sourceURL, err)
	}
	if len(body) > maxResponseSize {
		return nil, fmt.Errorf("response from %q exceeds %d bytes", sourceURL, maxResponseSize)
	}
	return body, nil
}

func validateDefaults(generated defaults) error {
	values := []struct {
		name  string
		value string
	}{
		{name: "Version", value: generated.Version},
		{name: "BundleVersion", value: generated.BundleVersion},
		{name: "GatewayURL", value: generated.GatewayURL},
		{name: "GameURL", value: generated.GameURL},
	}
	for _, value := range values {
		if strings.TrimSpace(value.value) == "" {
			return fmt.Errorf("generated value %s is empty", value.name)
		}
	}
	return nil
}

func renderDefaults(generated defaults) ([]byte, error) {
	var source bytes.Buffer
	fmt.Fprintf(
		&source,
		`// Code generated by go generate; DO NOT EDIT.

package plana

const (
	Version = %s
	defaultBundleVersion = %s
	defaultGatewayURL = %s
	defaultGameURL = %s
)
`,
		strconv.Quote(generated.Version),
		strconv.Quote(generated.BundleVersion),
		strconv.Quote(generated.GatewayURL),
		strconv.Quote(generated.GameURL),
	)

	formatted, err := format.Source(source.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated defaults: %w", err)
	}
	return formatted, nil
}
