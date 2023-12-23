package ffmpeg

import (
	"unsafe"
)

/*
#include <libavformat/avformat.h>
*/
import "C"

type AVFormatContext struct {
	ptr *C.AVFormatContext
}

func (c *AVFormatContext) NBStreams() uint {
	return uint(c.ptr.nb_streams)
}

func (c *AVFormatContext) NBPrograms() uint {
	return uint(c.ptr.nb_programs)
}

func (c *AVFormatContext) Program(id uint) *AVProgram {
	ptr := arrayGet[*C.AVProgram](c.ptr.programs, uintptr(id))

	return &AVProgram{
		ptr: ptr,
	}
}

func AVFormatOpenInput(url string) (*AVFormatContext, error) {
	var ptr *C.AVFormatContext

	urlPtr := C.CString(url)
	defer C.free(unsafe.Pointer(urlPtr))

	ret := C.avformat_open_input(&ptr, urlPtr, nil, nil)
	if ret != 0 {
		return nil, wrapError(ret)
	}

	return &AVFormatContext{
		ptr: ptr,
	}, nil
}

func (c *AVFormatContext) FindStreamInfo() error {
	return wrapError(C.avformat_find_stream_info(c.ptr, nil))
}

func (c *AVFormatContext) ReadFrame(pkt *AVPacket) error {
	ret := C.av_read_frame(c.ptr, pkt.ptr)

	return wrapError(ret)
}

type AVProgram struct {
	ptr *C.AVProgram
}

func (p *AVProgram) Id() int {
	return int(p.ptr.id)
}

func (p *AVProgram) Flags() int {
	return int(p.ptr.flags)
}

func (p *AVProgram) NBStreamIndexes() uint {
	return uint(p.ptr.nb_stream_indexes)
}

func (p *AVProgram) StreamIndex(id uint) uint {
	return uint(arrayGet[C.uint](p.ptr.stream_index, uintptr(id)))
}

func (p *AVProgram) Metadata() *AVDictionary {
	return &AVDictionary{ptr: p.ptr.metadata}
}

func (p *AVProgram) ProgramNum() int {
	return int(p.ptr.program_num)
}

func (p *AVProgram) PmtPid() int {
	return int(p.ptr.pmt_pid)
}

func (p *AVProgram) PcrPid() int {
	return int(p.ptr.pcr_pid)
}

func (p *AVProgram) PmtVersion() int {
	return int(p.ptr.pmt_version)
}
