package util

import (
    "net/http/httputil"
    "net/url"
    "regexp"
)

// TODO: Handle these URLs:
// Streaming / Playback:
// /Videos/{itemId}/stream
// /Videos/{itemId}/stream.{container}
// /Audio/{itemId}/stream
// /Audio/{itemId}/stream.{container}
// /Items/{itemId}/Download
//
// HLS (adaptive streaming):
// /Videos/{itemId}/master.m3u8
// /Videos/{itemId}/main.m3u8
// /Audio/{itemId}/hls/...
//
// Images (thumbnails, posters):
// /Items/{itemId}/Images/{imageType}

var videoPaths = []*regexp.Regexp{
    regexp.MustCompile(`^/Videos/[^/]+/stream$`),
    regexp.MustCompile(`^/Videos/[^/]+/stream\.[a-zA-Z0-9]+$`),
}

var hlsPaths = []*regexp.Regexp{}

var imagePaths = []*regexp.Regexp{}

var streamPaths = []*regexp.Regexp{}

var audioPaths = []*regexp.Regexp{}

type PathKindT = int8

const (
    PathKindVideos PathKindT = iota
    PathKindStreams
    PathKindAudios
    PathKindImage
    PathKindHLS
)

func ShouldForward(path string) (bool, PathKindT) {
    for _, pattern := range videoPaths {
        if pattern.MatchString(path) {
            return true, PathKindVideos
        }
    }

    for _, pattern := range streamPaths {
        if pattern.MatchString(path) {
            return true, PathKindStreams
        }
    }

    for _, pattern := range audioPaths {
        if pattern.MatchString(path) {
            return true, PathKindAudios
        }
    }

    for _, pattern := range hlsPaths {
        if pattern.MatchString(path) {
            return true, PathKindHLS
        }
    }

    for _, pattern := range imagePaths {
        if pattern.MatchString(path) {
            return true, PathKindImage
        }
    }

    return false, -1
}

func MakeReverseProxy(target string) (*httputil.ReverseProxy, error) {
    parsed, err := url.Parse(target)
    if err != nil {
        return nil, err
    }
    return httputil.NewSingleHostReverseProxy(parsed), nil
}

