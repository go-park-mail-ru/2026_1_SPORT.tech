package handler

import (
	"net/url"
	"strings"
)

func (handler *Handler) normalizePublicURL(rawURL *string) *string {
	if rawURL == nil || *rawURL == "" || handler.storagePublicBaseURL == "" {
		return rawURL
	}

	parsedURL, err := url.Parse(*rawURL)
	if err != nil {
		return rawURL
	}

	hostname := parsedURL.Hostname()
	if hostname != "localhost" && hostname != "127.0.0.1" {
		return rawURL
	}

	publicBaseURL := strings.TrimRight(handler.storagePublicBaseURL, "/")
	baseParsedURL, err := url.Parse(publicBaseURL)
	if err != nil {
		return rawURL
	}

	objectPath := strings.TrimPrefix(parsedURL.Path, "/")
	bucketPrefix := strings.TrimPrefix(strings.TrimRight(baseParsedURL.Path, "/"), "/")
	if bucketPrefix != "" {
		pathParts := strings.SplitN(objectPath, "/", 2)
		if len(pathParts) != 2 || pathParts[0] != bucketPrefix {
			return rawURL
		}

		objectPath = pathParts[1]
	}

	normalized := publicBaseURL + "/" + strings.TrimLeft(objectPath, "/")
	return &normalized
}
