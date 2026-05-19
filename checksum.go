package cinc

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

// md5Hex returns the lower-case hex MD5 digest of data.
func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// md5Base64 returns the base64-encoded MD5 digest of data, as Chef sandboxes
// expect in the Content-MD5 header.
func md5Base64(data []byte) string {
	sum := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}
