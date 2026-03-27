package handler

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "path"
    "strings"

    "ClientToR2/internal/s3"

    "github.com/jelius-sama/logger"
)

// itemDetails holds only the fields we care about from /Items/{itemId} response.
type itemDetails struct {
    Path string `json:"Path"`
}

// getItemPath queries Jellyfin for the file path of a given itemId.
// It returns the raw filesystem path that Jellyfin has on record, e.g:
//
//	/media/akane edit __ capsize.mp4
func getItemPath(itemId string) (string, error) {
    endpoint := fmt.Sprintf("%s/Items/%s?UserId=%s", os.Getenv("JELLYFIN_HOST"), itemId, os.Getenv("JELLYFIN_USER_ID"))

    req, err := http.NewRequest("GET", endpoint, nil)
    if err != nil {
        return "", fmt.Errorf("failed to build request: %w", err)
    }

    req.Header.Set("X-Emby-Token", os.Getenv("JELLYFIN_API_KEY"))

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("failed to contact Jellyfin: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("Jellyfin returned status %d for itemId %s", resp.StatusCode, itemId)
    }

    var details itemDetails
    if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
        return "", fmt.Errorf("failed to decode Jellyfin response: %w", err)
    }

    if details.Path == "" {
        return "", fmt.Errorf("Jellyfin returned empty path for itemId %s", itemId)
    }

    return details.Path, nil
}

// extractItemId pulls the itemId out of a path like /Videos/{itemId}/stream
func extractItemId(urlPath string) (string, error) {
    parts := strings.Split(path.Clean(urlPath), "/")

    // Expected: ["", "Videos", "{itemId}", "stream"]
    if len(parts) < 3 {
        return "", fmt.Errorf("unexpected path format: %s", urlPath)
    }

    return parts[2], nil
}

func ApplyVideosPatch(w http.ResponseWriter, r *http.Request, s3Client *s3.S3Client) {
    logger.Info("Applying videos patch, original path:", r.URL.Path)

    itemId, err := extractItemId(r.URL.Path)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        logger.Fatal("Failed to extract itemId:", err)
    }
    logger.Info("Extracted itemId:", itemId)

    filePath, err := getItemPath(itemId)
    if err != nil {
        http.Error(w, "Failed to resolve media path", http.StatusInternalServerError)
        logger.Fatal("Failed to get item path from Jellyfin:", err)
    }
    logger.Info("Jellyfin returned file path:", filePath)

    // Strip the Jellyfin media mount prefix to get the S3 object key.
    // e.g. "/media/akane edit __ capsize.mp4" -> "akane edit __ capsize.mp4"
    // TODO: When on R2, update this prefix to match your rclone mount path,
    //       e.g. strings.TrimPrefix(filePath, "/mnt/r2/")
    objectKey := strings.TrimPrefix(filePath, "/media/") // INFO: Jellyfin sees as /media/ (because it uses docker, I can do stuff like this)
    // INFO: In actuality the object is stored in /media-tmp/ directory (In S3 terms it is not a directory BTW but who really cares).
    objectKey = "media-tmp/" + objectKey // FIX: Because in my current setup jellyfin directory and S3 directory doesn't match
    logger.Info("Resolved S3 object key:", objectKey)

    presignedURL, err := s3Client.CreateSignedURL(context.TODO(), objectKey, nil)
    if err != nil {
        http.Error(w, "Failed to generate media URL", http.StatusInternalServerError)
        // TODO: Instead of crashing handle 404 like jellyfin server would
        logger.Fatal("Failed to create presigned URL:", err)
    }
    logger.Okay("S3 URL:", presignedURL)

    // Redirect the client directly to S3.
    // From this point the client fetches the video bytes straight from S3,
    // our EC2 server is no longer in the data path.
    http.Redirect(w, r, presignedURL, http.StatusTemporaryRedirect)
    logger.Okay("Redirected client to S3 for object:", objectKey)
}

