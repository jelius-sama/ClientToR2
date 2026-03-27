package handler

import (
    "net/http"

    "github.com/jelius-sama/logger"
)

// TODO:
// 1. Implement getting video file details from ID
// 2. Integrate with S3 compatible storage like R2, S3
// 3. Generating HLS on the fly would incur egress because processesing must be done in EC2 instance
// 4. Test if we can just send regular video instead of HLS segment and if the jellyfin client would accept it or reject it.

// INFO: Why would generating HLS incur egress cost?
//	Because R2 will store raw videos only, jellyfin uses ffmpeg to generate HLS on the fly.
//	Our EC2 can fetch the media data required from R2 as much as it wants for free BUT it has to write the HLS
//	data back to R2 for transimission to the client or the EC2 can directly transmit the data to the client, either
//	way both operations (sending to R2 then R2 to client & sending directly to client) will incur egress cost as
//	the data leaves from EC2 instance to the internet whether the client is R2 or the actual user.

// SOLUTION:
//  Test if we can just compute the segment required using byte ranges and only send that byte range instead of the HLS segment.
//  If it doesn't work, then maybe use CF workers, at least it will not incur those expensive egress costs or maybe we can use Cloudfront
//  for this sole purpose at least it will give us 1000GB of egress that we can work with in a month and perhaps we can cache the HLS
//  segments instead of deleting them after use, though it will increase our storage bill.

func ApplyHLSPatch(r *http.Request) {
    logger.Debug("TODO: Implement patch to the caught video path")
}

