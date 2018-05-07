package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

type assetInfo struct {
	Owner string
	Repo  string
	Re    *regexp.Regexp
}

// Perform a single attempt to download the asset.
func downloadAsset(info assetInfo, dir string, extract bool, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, info.Owner, info.Repo)
	if err != nil {
		return errors.Wrap(err, "failed to get latest release")
	}
	log.Printf("Latest release %d %s", *release.ID, *release.Name)
	for _, asset := range release.Assets {
		if !info.Re.MatchString(*asset.Name) {
			log.Printf("Skip non-matching asset %d %s", *asset.ID, *asset.Name)
			continue
		}
		log.Printf("Found matching asset %d %s (%d bytes)", *asset.ID, *asset.Name, *asset.Size)
		rc, url, err := client.Repositories.DownloadReleaseAsset(ctx, info.Owner, info.Repo, *asset.ID)
		if err != nil {
			return errors.Wrap(err, "failed to download release asset")
		}

		if url != "" {
			res, err := http.Get(url)
			if err != nil {
				return errors.Wrap(err, "failed to download release asset redirect")
			}
			rc = res.Body
		}

		defer rc.Close()

		if extract && strings.HasSuffix(*asset.Name, ".tar.gz") {
			if err := untar(dir, rc); err != nil {
				return errors.Wrap(err, "failed to read or extract download stream")
			}
		} else {
			data, err := ioutil.ReadAll(rc)
			if err != nil {
				return errors.Wrap(err, "failed to read download stream")
			}

			if err := ioutil.WriteFile(filepath.Join(dir, *asset.Name), data, 0644); err != nil {
				log.Fatal("Failed to write downloaded asset", err)
			}
		}

		break
	}

	return nil
}
