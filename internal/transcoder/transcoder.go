package transcoder

import (
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/csnewman/ffmpeg-go"
)

var ErrNoMore = errors.New("no more data available")

type Session struct {
	logger *slog.Logger
	path   string

	input  *SrcFile
	output *DstFile

	videoIn     *DecStream
	videoOut    *EncStream
	videoFilter *Filter
}

func Open(logger *slog.Logger, path string) *Session {
	return &Session{
		logger: logger,
		path:   path,
	}
}

func (s *Session) Start() error {
	var err error

	s.input, err = OpenSrcFile(s.logger, s.path)
	if err != nil {
		return fmt.Errorf("failed to open src: %w", err)
	}

	vsi, err := s.input.BestVideoStreamID()
	if err != nil {
		return fmt.Errorf("failed to find best src video stream: %w", err)
	}

	s.logger.Debug("Video stream", "idx", vsi)

	s.videoIn, err = s.input.OpenAVStream(vsi)
	if err != nil {
		return fmt.Errorf("failed to open src video stream: %w", err)
	}

	s.output, err = OpenDstFile(s.logger, "init.mp4", true)
	if err != nil {
		return fmt.Errorf("failed to open dst file ctx: %w", err)
	}

	s.videoOut, err = s.output.OpenAVStream(EncOptions{
		Type:        ffmpeg.AVMediaTypeVideo,
		Width:       s.videoIn.DecCtx.Width(),
		Height:      s.videoIn.DecCtx.Height(),
		AspectRatio: s.videoIn.DecCtx.SampleAspectRatio(),
		FrameRate:   s.videoIn.DecCtx.Framerate(),
	})
	if err != nil {
		return fmt.Errorf("failed to open dst video stream: %w", err)
	}

	if err := s.output.WriteHeader(); err != nil {
		return fmt.Errorf("failed to write dst header: %w", err)
	}

	if err := s.output.NewSegment("seg1.m4s"); err != nil {
		return fmt.Errorf("failed to open segement: %w", err)
	}

	//"scale=w=256:h=256"
	filt, err := newFilter(s.videoIn, s.videoOut, "null")
	if err != nil {
		return err
	}

	s.videoFilter = filt

	return nil
}

var justFlushed = false //nolint:gochecknoglobals

func (s *Session) Run() error {
	i := 0

	for {
		i++

		if i == 2000 {
			if err := s.output.Write(nil); err != nil {
				log.Panicln(err)
			}

			justFlushed = true

			if err := s.output.NewSegment("seg2.m4s"); err != nil {
				return fmt.Errorf("failed to open segement: %w", err)
			}
		}

		if i > 3000 {
			break
		}

		pkt, err := s.input.Read()
		if errors.Is(err, ffmpeg.AVErrorEOF) {
			log.Println("End of file")

			break
		} else if err != nil {
			return err
		}

		idx := pkt.StreamIndex()

		if idx == s.videoIn.Index {
			s.processAVPacket(pkt, s.videoIn, s.videoFilter, s.videoOut)
		}

		ffmpeg.AVPacketUnref(pkt)
	}

	s.flushAV(s.videoFilter, s.videoOut)

	if err := s.output.WriteTrailer(); err != nil {
		return err
	}

	return nil
}

func (s *Session) processAVPacket(packet *ffmpeg.AVPacket, in *DecStream, filter *Filter, out *EncStream) {
	err := in.Send(packet)
	if err != nil {
		log.Panicln(err)
	}

	for {
		frame, err := in.Receive()
		if errors.Is(err, ErrNoMore) {
			break
		} else if err != nil {
			log.Panicln(err)
		}

		if err := filter.Write(frame); err != nil {
			log.Panicln(err)
		}

		for {
			filtered, err := filter.Read()
			if errors.Is(err, ErrNoMore) {
				break
			} else if err != nil {
				log.Panicln(err)
			}

			if justFlushed {
				filtered.SetPictType(ffmpeg.AVPictureTypeI)

				justFlushed = false
			}

			if err := out.Write(filtered); err != nil {
				log.Panicln(err)
			}

			ffmpeg.AVFrameUnref(filtered)
		}
	}
}

func (s *Session) flushAV(filter *Filter, out *EncStream) {
	if err := filter.Write(nil); err != nil {
		log.Panicln(err)
	}

	for {
		filtered, err := filter.Read()
		if errors.Is(err, ErrNoMore) {
			break
		} else if err != nil {
			log.Panicln(err)
		}

		if err := out.Write(filtered); err != nil {
			log.Panicln(err)
		}

		ffmpeg.AVFrameUnref(filtered)
	}

	if err := out.Write(nil); err != nil {
		log.Panicln(err)
	}
}
