package ffmpeg

/*
#include <stdlib.h>
#include <libavutil/dict.h>
*/
import "C"
import "unsafe"

type AVDictFlag int

const (
	AVDictMatchCase    AVDictFlag = C.AV_DICT_MATCH_CASE
	AVDictIgnoreSuffix AVDictFlag = C.AV_DICT_IGNORE_SUFFIX
)

type AVDictionaryEntry struct {
	ptr *C.AVDictionaryEntry
}

func (e *AVDictionaryEntry) Key() string {
	return C.GoString(e.ptr.key)
}

func (e *AVDictionaryEntry) Value() string {
	return C.GoString(e.ptr.value)
}

type AVDictionary struct {
	ptr *C.AVDictionary
}

func (d *AVDictionary) Get(key string, prev *AVDictionaryEntry, flags AVDictFlag) *AVDictionaryEntry {
	keyPtr := C.CString(key)
	defer C.free(unsafe.Pointer(keyPtr))

	var prevPtr *C.AVDictionaryEntry
	if prev != nil {
		prevPtr = prev.ptr
	}

	ret := C.av_dict_get(d.ptr, keyPtr, prevPtr, C.int(flags))

	if ret != nil {
		return &AVDictionaryEntry{
			ptr: ret,
		}
	}

	return nil
}

func (d *AVDictionary) Count() int {
	return int(C.av_dict_count(d.ptr))
}
