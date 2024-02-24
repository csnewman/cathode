package transcoder

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/csnewman/ffmpeg-go"
)

var (
	errNoDecoder         = errors.New("decoder not found")
	errUnsupportedStream = errors.New("unsupported stream type")
)

type SrcFile struct {
	logger *slog.Logger
	ctx    *ffmpeg.AVFormatContext
	pkt    *ffmpeg.AVPacket
}

func OpenSrcFile(logger *slog.Logger, path string) (*SrcFile, error) {
	logger.Info("Opening src file", "path", path)

	urlPtr := ffmpeg.ToCStr(path)
	defer urlPtr.Free()

	var ctx *ffmpeg.AVFormatContext

	if _, err := ffmpeg.AVFormatOpenInput(&ctx, urlPtr, nil, nil); err != nil {
		return nil, fmt.Errorf("failed to open input: %w", err)
	}

	if _, err := ffmpeg.AVFormatFindStreamInfo(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to find streams: %w", err)
	}

	ffmpeg.AVDumpFormat(ctx, 0, urlPtr, 0)

	pkt := ffmpeg.AVPacketAlloc()

	return &SrcFile{
		logger: logger,
		ctx:    ctx,
		pkt:    pkt,
	}, nil
}

func (f *SrcFile) BestVideoStreamID() (int, error) {
	vsi, err := ffmpeg.AVFindBestStream(f.ctx, ffmpeg.AVMediaTypeVideo, -1, -1, nil, 0)
	if err != nil {
		return -1, fmt.Errorf("failed to find video stream: %w", err)
	}

	return vsi, nil
}

func (f *SrcFile) Read() (*ffmpeg.AVPacket, error) {
	_, err := ffmpeg.AVReadFrame(f.ctx, f.pkt)
	if errors.Is(err, ffmpeg.AVErrorEOF) {
		return nil, ErrNoMore
	} else if err != nil {
		return nil, err
	}

	return f.pkt, nil
}

type DecStream struct {
	Index    int
	Type     ffmpeg.AVMediaType
	DecCtx   *ffmpeg.AVCodecContext
	DecFrame *ffmpeg.AVFrame
}

func (f *SrcFile) OpenAVStream(id int) (*DecStream, error) {
	streams := f.ctx.Streams()
	stream := streams.Get(uintptr(id))

	cid := stream.Codecpar().CodecId()
	codecName := ffmpeg.AVCodecGetName(cid).String()

	cType := stream.Codecpar().CodecType()
	typeNme := ffmpeg.AVGetMediaTypeString(cType).String()

	f.logger.Debug("Opening stream", "type", typeNme, "codec", codecName)

	if cType != ffmpeg.AVMediaTypeVideo && cType != ffmpeg.AVMediaTypeAudio {
		return nil, fmt.Errorf("%w: %v", errUnsupportedStream, typeNme)
	}

	dec := ffmpeg.AVCodecFindDecoder(cid)
	if dec == nil {
		return nil, fmt.Errorf("%w: %v", errNoDecoder, codecName)
	}

	codecCtx := ffmpeg.AVCodecAllocContext3(dec)

	if _, err := ffmpeg.AVCodecParametersToContext(codecCtx, stream.Codecpar()); err != nil {
		return nil, fmt.Errorf("failed to copy codec params: %w", err)
	}

	codecCtx.SetPktTimebase(stream.TimeBase())

	if cType == ffmpeg.AVMediaTypeVideo {
		fr := ffmpeg.AVGuessFrameRate(f.ctx, stream, nil)
		codecCtx.SetFramerate(fr)
	}

	if _, err := ffmpeg.AVCodecOpen2(codecCtx, dec, nil); err != nil {
		return nil, fmt.Errorf("failed to open decoder: %w", err)
	}

	frame := ffmpeg.AVFrameAlloc()

	return &DecStream{
		Index:    id,
		Type:     cType,
		DecCtx:   codecCtx,
		DecFrame: frame,
	}, nil
}

func (s *DecStream) Send(packet *ffmpeg.AVPacket) error {
	if _, err := ffmpeg.AVCodecSendPacket(s.DecCtx, packet); err != nil {
		return err
	}

	return nil
}

func (s *DecStream) Receive() (*ffmpeg.AVFrame, error) {
	_, err := ffmpeg.AVCodecReceiveFrame(s.DecCtx, s.DecFrame)
	if errors.Is(err, ffmpeg.AVErrorEOF) || errors.Is(err, ffmpeg.EAgain) {
		return nil, ErrNoMore
	} else if err != nil {
		return nil, err
	}

	s.DecFrame.SetPts(s.DecFrame.BestEffortTimestamp())

	return s.DecFrame, nil
}
