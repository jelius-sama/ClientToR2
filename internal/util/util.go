package util

import (
    "errors"
    "net/http/httputil"
    "net/url"
    "os"
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

func EnsureENV() error {
    var errs string = "The following environment variables are not set:\n"
    var errCount int = 0

    if val := os.Getenv("JELLYFIN_HOST"); len(val) == 0 {
        errCount++
        errs = errs + "\tJELLYFIN_HOST is not set\n"
    }

    if val := os.Getenv("JELLYFIN_API_KEY"); len(val) == 0 {
        errCount++
        errs = errs + "\tJELLYFIN_API_KEY is not set\n"
    }

    if val := os.Getenv("JELLYFIN_USER_ID"); len(val) == 0 {
        errCount++
        errs = errs + "\tJELLYFIN_USER_ID is not set\n"
    }

    if val := os.Getenv("ACCESS_KEY_ID"); len(val) == 0 {
        errCount++
        errs = errs + "\tACCESS_KEY_ID is not set\n"
    }

    if val := os.Getenv("SECRET_ACCESS_KEY"); len(val) == 0 {
        errCount++
        errs = errs + "\tSECRET_ACCESS_KEY is not set\n"
    }

    if val := os.Getenv("ACCOUNT_ID"); len(val) == 0 {
        errCount++
        errs = errs + "\tACCOUNT_ID is not set\n"
    }

    if val := os.Getenv("BUCKET_NAME"); len(val) == 0 {
        errCount++
        errs = errs + "\tBUCKET_NAME is not set\n"
    }

    if errCount > 0 {
        return errors.New(errs)
    }

    return nil
}

