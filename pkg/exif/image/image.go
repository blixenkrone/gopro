package image

import (
	"bytes"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/blixenkrone/gopro/pkg/conversion"
	"github.com/blixenkrone/gopro/pkg/exif"
	"github.com/blixenkrone/gopro/pkg/logger"

	goexif "github.com/rwcarlsen/goexif/exif"
)

var (
	log = logger.NewLogger()
)

// tiff.Tag struct return values as number(i.e. 0 == int)
const (
	exifIntVal = iota
	EOFError   = "error reading exif from file: EOF"
)

type ExifMetadata struct {
	x *goexif.Exif
}

type ExifReader struct {
	io.Reader
}

func NewExifReader(r io.Reader) *ExifReader {
	return &ExifReader{
		r,
	}
}

func (r *ExifReader) Read(b []byte) error {
	// reader := bytes.NewReader(b)
	return nil
}

// DecodeImageMetadata returns the struct *Output containing img data.
// This will include the errors from missing/broken exif will follow.
// If an error is != nil, its a panic
func DecodeImageMetadata(data []byte) (*exif.Output, error) {
	r := bytes.NewReader(data)
	xErr := &exif.Output{MissingExif: make(map[string]string)}

	x, err := loadExifData(r)
	if err != nil {
		err = errors.Cause(err)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			return nil, errors.New(EOFError)
		}
		// Missing exif should probably not happen
		xErr.AddMissingExif("decode", err)
		return nil, err
	}
	lat, err := x.calcGeoCoordinate(goexif.GPSLatitude)
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("lat", err)
	}
	lng, err := x.calcGeoCoordinate(goexif.GPSLongitude)
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("lng", err)
	}
	date, err := x.getDateTime()
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("date", err)
	}
	author, err := x.getCopyright()
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("copyright", err)
	}
	model, err := x.getCameraModel()
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("model", err)
	}
	dimensions, err := x.getImageDimensions()
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("dimension", err)
	}
	size, err := x.getFileSize(r)
	if err != nil {
		err = errors.Cause(err)
		xErr.AddMissingExif("fileSize", err)
	}

	return &exif.Output{
		Lat:             lat,
		Lng:             lng,
		Date:            date,
		Model:           model,
		PixelXDimension: dimensions[goexif.PixelXDimension],
		PixelYDimension: dimensions[goexif.PixelYDimension],
		Copyright:       author,
		MediaSize:       size,
		MissingExif:     xErr.MissingExif,
		// ? do this MediaFormat:     mediaFmt,
	}, nil
}

// loadExifData request exif data for image
func loadExifData(r io.Reader) (*ExifMetadata, error) {
	x, err := goexif.Decode(r)
	if err != nil {
		err := errors.Wrap(err, "loading exif error")
		return nil, err
	}
	return &ExifMetadata{x}, nil
}

func (e *ExifMetadata) calcGeoCoordinate(fieldName goexif.FieldName) (float64, error) {
	tag, err := e.x.Get(fieldName)
	if err != nil {
		return 0.0, errors.WithMessagef(err, "error getting location coordinates from %s", fieldName)
	}
	ratValues := map[string]int{"deg": 0, "min": 1, "sec": 2}
	fValues := make(map[string]float64, len(ratValues))

	for key, val := range ratValues {
		v, err := tag.Rat(val)
		if err != nil {
			return 0.0, err
		}
		f, _ := v.Float64()
		fValues[key] = f
	}

	res := fValues["deg"] + (fValues["min"] / 60) + (fValues["sec"] / 3600)
	return res, nil
}

func (e *ExifMetadata) getDateTime() (d int64, err error) {
	t, err := e.x.DateTime()
	if err != nil {
		return d, err
	}
	d = conversion.UnixNanoToMillis(t)
	return d, nil
}

func (e *ExifMetadata) getCopyright() (author string, err error) {
	tag, err := e.x.Get(goexif.Copyright)
	if err != nil {
		return author, err
	}
	return tag.StringVal()
}

func (e *ExifMetadata) getCameraModel() (model string, err error) {
	n := goexif.FieldName(goexif.Model)
	tag, err := e.x.Get(n)
	if err != nil {
		return model, err
	}
	return tag.StringVal()
}

func (e *ExifMetadata) getImageDimensions() (map[goexif.FieldName]int, error) {
	var fNames = []goexif.FieldName{goexif.PixelXDimension, goexif.PixelYDimension}
	var fNameVal = make(map[goexif.FieldName]int, len(fNames))
	for _, n := range fNames {
		tag, err := e.x.Get(n)
		if err != nil {
			return nil, err
		}
		i, err := tag.Int(exifIntVal)
		if err != nil {
			return nil, err
		}
		fNameVal[n] = i
	}
	return fNameVal, nil
}

// get file size
func (e *ExifMetadata) getFileSize(r io.Reader) (float64, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	var buf bytes.Buffer
	n, err := buf.Write(b)
	if err != nil {
		return 0, err
	}
	size := conversion.FileSizeBytesToFloat(n)
	return size, nil
}

// get image fmt
// ! switch between image and video - evt create struct input
func (e *ExifMetadata) getMediaFmt(r io.Reader) (fmt string, err error) {
	// _, fmt, err = image.DecodeConfig(r)
	// if err != nil {
	// 	log.Errorln(err)
	// 	return "", err
	// }
	fmt = ".jpeg"
	return fmt, err
}
