package transcoder

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/csnewman/ffmpeg-go"
)

type DstFile struct {
	logger    *slog.Logger
	ctx       *ffmpeg.AVFormatContext
	segmented bool

	encStreams []*EncStream
}

func OpenDstFile(logger *slog.Logger, path string, segmented bool) (*DstFile, error) {
	ofmt := ffmpeg.AVGuessFormat(ffmpeg.ToCStr("mp4"), nil, nil)

	var ctx *ffmpeg.AVFormatContext

	if _, err := ffmpeg.AVFormatAllocOutputContext2(&ctx, ofmt, nil, nil); err != nil {
		return nil, fmt.Errorf("failed to alloc output: %w", err)
	}

	namePtr := ffmpeg.ToCStr(path)
	defer namePtr.Free()

	var pb *ffmpeg.AVIOContext

	if _, err := ffmpeg.AVIOOpen(&pb, namePtr, ffmpeg.AVIOFlagWrite); err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	ctx.SetPb(pb)

	return &DstFile{
		logger:    logger,
		ctx:       ctx,
		segmented: segmented,
	}, nil
}

func (f *DstFile) WriteHeader() error {
	var opts *ffmpeg.AVDictionary

	if _, err := ffmpeg.AVDictSet(&opts, ffmpeg.GlobalCStr("fflags"), ffmpeg.GlobalCStr("-autobsf"), 0); err != nil {
		return fmt.Errorf("failed to set opt: %w", err)
	}

	if f.segmented {
		_, err := ffmpeg.AVDictSet(&opts, ffmpeg.GlobalCStr("movflags"), ffmpeg.GlobalCStr("+frag_custom+dash+delay_moov"), 0)
		if err != nil {
			return fmt.Errorf("failed to set opt: %w", err)
		}
	}

	if _, err := ffmpeg.AVFormatInitOutput(f.ctx, &opts); err != nil {
		return fmt.Errorf("failed to init output: %w", err)
	}

	val, _ := ffmpeg.AVDictCount(opts)
	if val > 0 {
		panic("opts left over")
	}

	if _, err := ffmpeg.AVFormatWriteHeader(f.ctx, nil); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, stream := range f.encStreams {
		stream.Timebase = stream.Stream.TimeBase()
	}

	return nil
}

func (f *DstFile) NewSegment(path string) error {
	if err := f.Write(nil); err != nil {
		return fmt.Errorf("failed to flush ctx: %w", err)
	}

	ffmpeg.AVIOFlush(f.ctx.Pb())

	if _, err := ffmpeg.AVIOClose(f.ctx.Pb()); err != nil {
		return fmt.Errorf("failed to close existing file: %w", err)
	}

	namePtr := ffmpeg.GlobalCStr(path)
	defer namePtr.Free()

	var pb *ffmpeg.AVIOContext

	if _, err := ffmpeg.AVIOOpen(&pb, namePtr, ffmpeg.AVIOFlagWrite); err != nil {
		return fmt.Errorf("failed to open new output: %w", err)
	}

	f.ctx.SetPb(pb)

	// Write intro
	ffmpeg.AVIOWb32(pb, 24)
	ffmpeg.FFIOWFourCC(pb, 's', 't', 'y', 'p')
	ffmpeg.FFIOWFourCC(pb, 'm', 's', 'd', 'h')
	ffmpeg.AVIOWb32(pb, 0)
	ffmpeg.FFIOWFourCC(pb, 'm', 's', 'd', 'h')
	ffmpeg.FFIOWFourCC(pb, 'm', 's', 'i', 'x')

	return nil
}

func (f *DstFile) Write(pkt *ffmpeg.AVPacket) error {
	// TODO: Verify if needed - flush should make this redundant?
	if f.segmented {
		_, err := ffmpeg.AVWriteFrame(f.ctx, pkt)

		return err
	}

	_, err := ffmpeg.AVInterleavedWriteFrame(f.ctx, pkt)

	return err
}

func (f *DstFile) WriteTrailer() error {
	if err := f.Write(nil); err != nil {
		return fmt.Errorf("failed to flush ctx: %w", err)
	}

	if _, err := ffmpeg.AVWriteTrailer(f.ctx); err != nil {
		return fmt.Errorf("failed write trailer: %w", err)
	}

	if _, err := ffmpeg.AVIOClose(f.ctx.Pb()); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	f.ctx.SetPb(nil)

	return nil
}

type EncStream struct {
	File     *DstFile
	Index    int
	Enc      *ffmpeg.AVCodec
	EncCtx   *ffmpeg.AVCodecContext
	Stream   *ffmpeg.AVStream
	EncPkt   *ffmpeg.AVPacket
	Timebase *ffmpeg.AVRational
}

type EncOptions struct {
	Type ffmpeg.AVMediaType

	// Video
	Width       int
	Height      int
	AspectRatio *ffmpeg.AVRational
	FrameRate   *ffmpeg.AVRational
}

func (f *DstFile) OpenAVStream(opts EncOptions) (*EncStream, error) {
	out := &EncStream{
		File:   f,
		EncPkt: ffmpeg.AVPacketAlloc(),
	}

	out.Stream = ffmpeg.AVFormatNewStream(f.ctx, nil)
	out.Index = out.Stream.Index()

	if opts.Type == ffmpeg.AVMediaTypeVideo {
		out.Enc = ffmpeg.AVCodecFindEncoder(ffmpeg.AVCodecIdH264)
		if out.Enc == nil {
			panic("h264 not found")
		}

		out.EncCtx = ffmpeg.AVCodecAllocContext3(out.Enc)
		out.EncCtx.SetHeight(opts.Height)
		out.EncCtx.SetWidth(opts.Width)
		out.EncCtx.SetSampleAspectRatio(opts.AspectRatio)

		fmts := out.Enc.PixFmts()
		if fmts != nil {
			out.EncCtx.SetPixFmt(fmts.Get(0))
		} else {
			// out.EncCtx.SetPixFmt(in.DecCtx.PixFmt())

			panic("not implemented: pixfmt fallback")
		}

		out.EncCtx.SetTimeBase(ffmpeg.AVInvQ(opts.FrameRate))
	} else {
		panic("not implemented")
	}

	if f.ctx.Oformat().Flags()&ffmpeg.AVFmtGlobalheader != 0 {
		out.EncCtx.SetFlags(out.EncCtx.Flags() | ffmpeg.AVCodecFlagGlobalHeader)
	}

	// Third parameter can be used to pass settings to encoder
	if _, err := ffmpeg.AVCodecOpen2(out.EncCtx, out.Enc, nil); err != nil {
		return nil, fmt.Errorf("failed to open encoder codec: %w", err)
	}

	if _, err := ffmpeg.AVCodecParametersFromContext(out.Stream.Codecpar(), out.EncCtx); err != nil {
		return nil, fmt.Errorf("failed to set codec parameters: %w", err)
	}

	out.Stream.SetTimeBase(out.EncCtx.TimeBase())

	f.encStreams = append(f.encStreams, out)

	return out, nil
}

func (s *EncStream) Write(frame *ffmpeg.AVFrame) error {
	ffmpeg.AVPacketUnref(s.EncPkt)

	if frame != nil && frame.Pts() != ffmpeg.AVNoptsValue {
		frame.SetPts(ffmpeg.AVRescaleQ(frame.Pts(), frame.TimeBase(), s.EncCtx.TimeBase()))
	}

	if _, err := ffmpeg.AVCodecSendFrame(s.EncCtx, frame); err != nil {
		return fmt.Errorf("failed to send frame to encoder: %w", err)
	}

	for {
		if _, err := ffmpeg.AVCodecReceivePacket(s.EncCtx, s.EncPkt); err != nil {
			if errors.Is(err, ffmpeg.EAgain) || errors.Is(err, ffmpeg.AVErrorEOF) {
				break
			}

			return fmt.Errorf("failed to receive frame from encoder: %w", err)
		}

		s.EncPkt.SetStreamIndex(s.Index)
		ffmpeg.AVPacketRescaleTs(s.EncPkt, s.EncCtx.TimeBase(), s.Timebase)

		if err := s.File.Write(s.EncPkt); err != nil {
			return fmt.Errorf("failed to write frame: %w", err)
		}
	}

	return nil
}
