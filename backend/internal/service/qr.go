package service

import (
	"fmt"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

type QR struct {
	BaseURL string
}

func (q *QR) ScanURL(token string) string {
	base := strings.TrimRight(strings.TrimSpace(q.BaseURL), "/")
	return fmt.Sprintf("%s/q/%s", base, token)
}

func (q *QR) PNG(token string) ([]byte, error) {
	url := q.ScanURL(token)
	return qrcode.Encode(url, qrcode.Medium, 256)
}
