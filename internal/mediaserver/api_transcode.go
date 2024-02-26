package mediaserver

import (
	"context"

	v1 "github.com/csnewman/cathode/internal/v1"
)

func (a *v1API) GetTranscodeManifestM3u8(
	ctx context.Context,
	request v1.GetTranscodeManifestM3u8RequestObject,
) (v1.GetTranscodeManifestM3u8ResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (a *v1API) GetTranscodeSegment(
	ctx context.Context,
	request v1.GetTranscodeSegmentRequestObject,
) (v1.GetTranscodeSegmentResponseObject, error) {
	//TODO implement me
	panic("implement me")
}
