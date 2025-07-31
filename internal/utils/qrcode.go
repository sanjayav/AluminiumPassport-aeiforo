
package utils

import (
    "fmt"
    "time"

    "github.com/skip2/go-qrcode"
)

// GenerateQRCode creates a PNG file with a QR code pointing to the given URL
func GenerateQRCode(data string) (string, error) {
    filename := fmt.Sprintf("qr_%d.png", time.Now().Unix())

    err := qrcode.WriteFile(data, qrcode.Medium, 256, filename)
    if err != nil {
        return "", err
    }

    return filename, nil
}
