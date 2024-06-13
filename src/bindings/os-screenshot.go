//go:build !darwin || CI

package bindings

import (
	"fmt"

	"github.com/kbinani/screenshot"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/multi/qrcode"
)

func (b *Bindings) ScreenshotQR() ([]string, error) {
	codes := []string{}

	screens := screenshot.NumActiveDisplays()
	reader := qrcode.NewQRCodeMultiReader()
	for i := 0; i < screens; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		if b.options.Verbose {
			fmt.Println("Capturing screen", i, bounds)
		}
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			return codes, err
		}

		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			return codes, err
		}

		result, err := reader.DecodeMultiple(bmp, nil)
		if err == nil {
			for _, result := range result {
				if b.options.Verbose {
					fmt.Println("QR Code found on screen", i, result.GetResultMetadata(), result.GetResultPoints())
				}
				codes = append(codes, result.GetText())
			}
		}
	}

	return codes, nil
}
