package transcoder

import (
	"errors"
	"fmt"

	"github.com/csnewman/ffmpeg-go"
)

type Filter struct {
	FilterGraph   *ffmpeg.AVFilterGraph
	BufferSrcCtx  *ffmpeg.AVFilterContext
	BufferSinkCtx *ffmpeg.AVFilterContext
	FilteredFrame *ffmpeg.AVFrame
}

func newFilter(in *DecStream, out *EncStream, filterSpec string) (*Filter, error) {
	filterGraph := ffmpeg.AVFilterGraphAlloc()

	var (
		bufferSrcCtx  *ffmpeg.AVFilterContext
		bufferSinkCtx *ffmpeg.AVFilterContext
	)

	if in.Type == ffmpeg.AVMediaTypeVideo {
		bufferSrc := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("buffer"))
		bufferSink := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("buffersink"))

		if bufferSrc == nil || bufferSink == nil {
			panic("failed to alloc src/sink")
		}

		pktTimebase := in.DecCtx.PktTimebase()
		args := fmt.Sprintf(
			"video_size=%vx%v:pix_fmt=%v:time_base=%v/%v:pixel_aspect=%v/%v",
			in.DecCtx.Width(), in.DecCtx.Height(),
			in.DecCtx.PixFmt(),
			pktTimebase.Num(), pktTimebase.Den(),
			in.DecCtx.SampleAspectRatio().Num(),
			in.DecCtx.SampleAspectRatio().Den(),
		)

		argsC := ffmpeg.ToCStr(args)
		defer argsC.Free()

		_, err := ffmpeg.AVFilterGraphCreateFilter(
			&bufferSrcCtx,
			bufferSrc,
			ffmpeg.GlobalCStr("in"),
			argsC,
			nil,
			filterGraph,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create in filter: %w", err)
		}

		_, err = ffmpeg.AVFilterGraphCreateFilter(
			&bufferSinkCtx,
			bufferSink,
			ffmpeg.GlobalCStr("out"),
			nil,
			nil,
			filterGraph,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create out filter: %w", err)
		}

		pixFmts := []ffmpeg.AVPixelFormat{
			out.EncCtx.PixFmt(),
		}

		_, err = ffmpeg.AVOptSetSlice(
			bufferSinkCtx.RawPtr(),
			ffmpeg.GlobalCStr("pix_fmts"),
			pixFmts,
			ffmpeg.AVOptSearchChildren,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set pix_fmts: %w", err)
		}
	} else {
		bufferSrc := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("abuffer"))
		bufferSink := ffmpeg.AVFilterGetByName(ffmpeg.GlobalCStr("abuffersink"))

		if bufferSrc == nil || bufferSink == nil {
			panic("failed to alloc src/sink")
		}

		if in.DecCtx.ChLayout().Order() == ffmpeg.AVChannelOrderUnspec {
			ffmpeg.AVChannelLayoutDefault(in.DecCtx.ChLayout(), in.DecCtx.ChLayout().NbChannels())
		}

		layoutPtr := ffmpeg.AllocCStr(64)
		defer layoutPtr.Free()

		if _, err := ffmpeg.AVChannelLayoutDescribe(in.DecCtx.ChLayout(), layoutPtr, 64); err != nil {
			return nil, fmt.Errorf("failed to describe dec channel layout: %w", err)
		}

		layout := layoutPtr.String()

		pktTimebase := in.DecCtx.PktTimebase()
		args := fmt.Sprintf(
			"time_base=%v/%v:sample_rate=%v:sample_fmt=%v:channel_layout=%v",
			pktTimebase.Num(), pktTimebase.Den(),
			in.DecCtx.SampleRate(),
			ffmpeg.AVGetSampleFmtName(in.DecCtx.SampleFmt()),
			layout,
		)

		argsC := ffmpeg.ToCStr(args)
		defer argsC.Free()

		_, err := ffmpeg.AVFilterGraphCreateFilter(
			&bufferSrcCtx,
			bufferSrc,
			ffmpeg.GlobalCStr("in"),
			argsC,
			nil,
			filterGraph,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create in filter: %w", err)
		}

		_, err = ffmpeg.AVFilterGraphCreateFilter(
			&bufferSinkCtx,
			bufferSink,
			ffmpeg.GlobalCStr("out"),
			nil,
			nil,
			filterGraph,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create out filter: %w", err)
		}

		sampleFmts := []ffmpeg.AVSampleFormat{
			out.EncCtx.SampleFmt(),
		}

		_, err = ffmpeg.AVOptSetSlice(
			bufferSinkCtx.RawPtr(),
			ffmpeg.GlobalCStr("sample_fmts"),
			sampleFmts,
			ffmpeg.AVOptSearchChildren,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set sample_fmts: %w", err)
		}

		layoutPtr = ffmpeg.AllocCStr(64)
		defer layoutPtr.Free()

		if _, err := ffmpeg.AVChannelLayoutDescribe(out.EncCtx.ChLayout(), layoutPtr, 64); err != nil {
			return nil, fmt.Errorf("failed to describe enc channel layout: %w", err)
		}

		_, err = ffmpeg.AVOptSet(
			bufferSinkCtx.RawPtr(),
			ffmpeg.GlobalCStr("ch_layouts"),
			layoutPtr,
			ffmpeg.AVOptSearchChildren,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set ch_layouts: %w", err)
		}

		sampleRates := []int{
			out.EncCtx.SampleRate(),
		}

		_, err = ffmpeg.AVOptSetSlice(
			bufferSinkCtx.RawPtr(),
			ffmpeg.GlobalCStr("sample_rates"),
			sampleRates,
			ffmpeg.AVOptSearchChildren,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to set sample_rates: %w", err)
		}
	}

	outputs := ffmpeg.AVFilterInoutAlloc()
	inputs := ffmpeg.AVFilterInoutAlloc()

	defer ffmpeg.AVFilterInoutFree(&outputs)
	defer ffmpeg.AVFilterInoutFree(&inputs)

	outputs.SetName(ffmpeg.ToCStr("in"))
	outputs.SetFilterCtx(bufferSrcCtx)
	outputs.SetPadIdx(0)
	outputs.SetNext(nil)

	inputs.SetName(ffmpeg.ToCStr("out"))
	inputs.SetFilterCtx(bufferSinkCtx)
	inputs.SetPadIdx(0)
	inputs.SetNext(nil)

	filterSpecC := ffmpeg.ToCStr(filterSpec)
	defer filterSpecC.Free()

	if _, err := ffmpeg.AVFilterGraphParsePtr(filterGraph, filterSpecC, &inputs, &outputs, nil); err != nil {
		return nil, fmt.Errorf("failed to parse filter graph: %w", err)
	}

	if _, err := ffmpeg.AVFilterGraphConfig(filterGraph, nil); err != nil {
		return nil, fmt.Errorf("failed to configure filter graph: %w", err)
	}

	filteredFrame := ffmpeg.AVFrameAlloc()

	return &Filter{
		FilterGraph:   filterGraph,
		BufferSrcCtx:  bufferSrcCtx,
		BufferSinkCtx: bufferSinkCtx,
		FilteredFrame: filteredFrame,
	}, nil
}

func (f *Filter) Write(frame *ffmpeg.AVFrame) error {
	if _, err := ffmpeg.AVBuffersrcAddFrameFlags(f.BufferSrcCtx, frame, 0); err != nil {
		return err
	}

	return nil
}

func (f *Filter) Read() (*ffmpeg.AVFrame, error) {
	_, err := ffmpeg.AVBuffersinkGetFrame(f.BufferSinkCtx, f.FilteredFrame)
	if errors.Is(err, ffmpeg.AVErrorEOF) || errors.Is(err, ffmpeg.EAgain) {
		return nil, ErrNoMore
	} else if err != nil {
		return nil, err
	}

	f.FilteredFrame.SetTimeBase(ffmpeg.AVBuffersinkGetTimeBase(f.BufferSinkCtx))
	f.FilteredFrame.SetPictType(ffmpeg.AVPictureTypeNone)

	return f.FilteredFrame, nil
}
