package registry

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yannh/kubeconform/pkg/cache"
)

type httpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

// SchemaRegistry is a file repository (local or remote) that contains JSON schemas for Kubernetes resources
type SchemaRegistry struct {
	c                  httpGetter
	schemaPathTemplate string
	cache              cache.Cache
	strict             bool
	debug              bool
}

func newHTTPRegistry(schemaPathTemplate string, cacheFolder string, strict bool, skipTLS bool, debug bool) (*SchemaRegistry, error) {
	reghttp := &http.Transport{
		MaxIdleConns:       100,
		IdleConnTimeout:    3 * time.Second,
		DisableCompression: true,
		Proxy:              http.ProxyFromEnvironment,
	}

	if skipTLS {
		reghttp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	var filecache cache.Cache = nil
	if cacheFolder != "" {
		fi, err := os.Stat(cacheFolder)
		if err != nil {
			return nil, fmt.Errorf("failed opening cache folder %s: %s", cacheFolder, err)
		}
		if !fi.IsDir() {
			return nil, fmt.Errorf("cache folder %s is not a directory", err)
		}

		filecache = cache.NewOnDiskCache(cacheFolder)
	}

	return &SchemaRegistry{
		c:                  &http.Client{Transport: reghttp},
		schemaPathTemplate: schemaPathTemplate,
		cache:              filecache,
		strict:             strict,
		debug:              debug,
	}, nil
}

// DownloadSchema downloads the schema for a particular resource from an HTTP server
func (r SchemaRegistry) DownloadSchema(resourceKind, resourceAPIVersion, k8sVersion string) ([]byte, error) {
	url, err := schemaPath(r.schemaPathTemplate, resourceKind, resourceAPIVersion, k8sVersion, r.strict)
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if b, err := r.cache.Get(resourceKind, resourceAPIVersion, k8sVersion); err == nil {
			return b.([]byte), nil
		}
	}

	resp, err := r.c.Get(url)
	if err != nil {
		msg := fmt.Sprintf("failed downloading schema at %s: %s", url, err)
		if r.debug {
			log.Println(msg)
		}
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		msg := fmt.Sprintf("could not find schema at %s", url)
		if r.debug {
			log.Print(msg)
		}
		return nil, newNotFoundError(errors.New(msg))
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("error while downloading schema at %s - received HTTP status %d", url, resp.StatusCode)
		if r.debug {
			log.Print(msg)
		}
		return nil, fmt.Errorf(msg)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := fmt.Sprintf("failed parsing schema from %s: %s", url, err)
		if r.debug {
			log.Print(msg)
		}
		return nil, errors.New(msg)
	}

	if r.debug {
		log.Printf("using schema found at %s", url)
	}

	if r.cache != nil {
		if err := r.cache.Set(resourceKind, resourceAPIVersion, k8sVersion, body); err != nil {
			return nil, fmt.Errorf("failed writing schema to cache: %s", err)
		}
	}

	return body, nil
}
