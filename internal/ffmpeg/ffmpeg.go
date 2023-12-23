package ffmpeg

import "unsafe"

/*
#cgo pkg-config: libavcodec libavfilter libavformat libavutil
*/
import "C"

func arrayGet[T any](array *T, i uintptr) T {
	var inner T
	ptrPtr := (*T)(unsafe.Add(unsafe.Pointer(array), i*unsafe.Sizeof(inner)))
	return *ptrPtr
}
