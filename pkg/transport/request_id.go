package transport

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

var prefix string
var reqId atomic.Uint64

func GetPrefix() string {
	return prefix
}

func NextReqID() uint64 {
	return reqId.Add(1)
}

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}
	var buf [12]byte
	var b64 string
	for len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	prefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}
