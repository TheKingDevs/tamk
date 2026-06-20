package qrcode

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func PrintTerminal(url string) error {
	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	art := qr.ToString(false)
	fmt.Println()
	fmt.Print(art)
	fmt.Println()
	return nil
}
