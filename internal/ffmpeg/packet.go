package ffmpeg

/*
#include <libavcodec/packet.h>
*/
import "C"

type AVPacket struct {
	ptr *C.AVPacket
}

func AVPacketAlloc() *AVPacket {
	ptr := C.av_packet_alloc()

	return &AVPacket{
		ptr: ptr,
	}
}

func (p *AVPacket) Free() {
	C.av_packet_free(&p.ptr)
}

func (p *AVPacket) Unref() {
	C.av_packet_unref(p.ptr)
}

func (p *AVPacket) PTS() int64 {
	return int64(p.ptr.pts)
}

func (p *AVPacket) DTS() int64 {
	return int64(p.ptr.dts)
}

func (p *AVPacket) StreamIndex() int {
	return int(p.ptr.stream_index)
}

func (p *AVPacket) Duration() int64 {
	return int64(p.ptr.duration)
}
