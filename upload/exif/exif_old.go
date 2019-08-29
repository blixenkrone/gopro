package exif

import (
	"bytes"
	"errors"
	"fmt"
	"image"

	// Keep this import so the compiler knows the format
	_ "image/jpeg"
	_ "image/png"
	"io"
	"sync"

	"github.com/rwcarlsen/goexif/exif"
)

// ImageReader contains image info
type ImageReader struct {
	Image  image.Image
	Name   string
	Format string
	Buffer *bytes.Buffer
}

// ImgService contains methods for imgs
type ImgService interface {
	TagExif(*sync.WaitGroup, chan<- *exif.Exif, chan<- error)
	TagExifSync() (*exif.Exif, error) // For tests no goroutines
}

// NewExifReq request exif data for image
func NewExifReq(r io.Reader) (ImgService, error) {
	var buf = new(bytes.Buffer)
	teeRead := io.TeeReader(r, buf)
	src, format, err := image.Decode(teeRead)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Image format is: %s\n", format)

	return &ImageReader{
		Image:  src,
		Format: format,
		Buffer: buf,
	}, nil
}

// TagExif returns the bytes of the image/tiff in ch
func (img *ImageReader) TagExif(wg *sync.WaitGroup, ch chan<- *exif.Exif, cherr chan<- error) {
	defer wg.Done()
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Fatalf("exif.Decode, critical error: %v", err)
		}
		log.Printf("exif.Decode, warning: %v", err)
	}
	log.Printf("Tagged exif: %s", img.Name)

	if err := img.requiredExifData(out); err != nil {
		cherr <- err
	} else {
		ch <- out
	}
}

// // RequiredExifData - testing with ch errors
func (img *ImageReader) requiredExifData(out *exif.Exif) error {
	minExifData := [3]exif.FieldName{"DateTime", "GPSLatitude", "GPSLongitude"}
	for _, r := range minExifData {
		_, err := out.Get(r)
		if err != nil {
			return errors.New("EXIF Tag was not present: " + string(r))
		}
	}
	return nil
}

// TagExifSync returns the bytes of the image/tiff in ch - dont use in production
func (img *ImageReader) TagExifSync() (*exif.Exif, error) {
	out, err := exif.Decode(img.Buffer)
	if err != nil {
		if exif.IsCriticalError(err) {
			log.Errorf("exif.Decode, critical error: %v", err)
			return nil, errors.New("exif.Decode, critical error: " + err.Error())
		}
		log.Printf("exif.Decode, warning: " + err.Error())
		return nil, err
	}
	log.Printf("Tagged exif: %s", img.Name)
	return out, nil
}