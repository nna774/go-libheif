package heif

/*
#cgo LDFLAGS: -lheif
#include <stdlib.h>
#include <libheif/heif.h>
*/
import "C"

import (
	"errors"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"unsafe"
)

const heifErrorOK C.enum_heif_error_code = C.heif_error_Ok

// Decode is
func Decode(r io.Reader) (image.Image, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var herr C.struct_heif_error
	ctx := C.heif_context_alloc()
	herr = C.heif_context_read_from_memory_without_copy(ctx, unsafe.Pointer(&bytes[0]), C.size_t(len(bytes)), nil)
	if herr.code != heifErrorOK {
		return nil, errors.New(C.GoString(herr.message))
	}

	var handle *C.struct_heif_image_handle
	herr = C.heif_context_get_primary_image_handle(ctx, &handle)
	defer C.heif_context_free(ctx) // "Once you obtained an heif_image_handle, you can already release the heif_context"
	if herr.code != heifErrorOK {
		return nil, errors.New(C.GoString(herr.message))
	}
	defer C.heif_image_handle_release(handle)

	var img *C.struct_heif_image
	herr = C.heif_decode_image(handle, &img, C.heif_colorspace_RGB, C.heif_chroma_interleaved_RGB, nil)
	if herr.code != heifErrorOK {
		return nil, errors.New(C.GoString(herr.message))
	}
	defer C.heif_image_release(img)

	var channel C.enum_heif_channel = C.heif_channel_interleaved
	var stride C.int
	var data *C.uint8_t
	data = C.heif_image_get_plane_readonly(img, channel, &stride)
	width := C.heif_image_get_width(img, channel)
	height := C.heif_image_get_height(img, channel)
	buf := C.GoBytes(unsafe.Pointer(data), stride*height)
	// `bytes` should be alive from here(is this correct?).

	return &rgbImage{
		Pix:    buf,
		Stride: int(stride),
		Rect: image.Rectangle{
			Min: image.Point{0, 0},
			Max: image.Point{int(width), int(height)},
		}}, nil
}

type rgbImage struct {
	Pix    []byte
	Stride int
	Rect   image.Rectangle
}

func (i *rgbImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (i *rgbImage) Bounds() image.Rectangle {
	return i.Rect
}

func (i *rgbImage) At(x, y int) color.Color {
	if !(image.Point{x, y}).In(i.Rect) {
		return color.RGBA{}
	}
	offset := (y-i.Rect.Min.Y)*i.Stride + (x-i.Rect.Min.X)*3
	p := i.Pix
	return color.RGBA{p[offset], p[offset+1], p[offset+2], 0xFF}
}

// Make sure rgbI	mage implements image.Image.
// See https://golang.org/doc/effective_go.html#blank_implements
var _ image.Image = new(rgbImage)
