// Package v1 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package v1

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/runtime"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Transcode Manifest M3U8
	// (GET /transcode/{transcodeId}/manifest.m3u8)
	GetTranscodeManifestM3u8(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID)
	// Transcode Segment
	// (GET /transcode/{transcodeId}/{segment})
	GetTranscodeSegment(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID, segment string)
}

// Unimplemented server implementation that returns http.StatusNotImplemented for each endpoint.

type Unimplemented struct{}

// Transcode Manifest M3U8
// (GET /transcode/{transcodeId}/manifest.m3u8)
func (_ Unimplemented) GetTranscodeManifestM3u8(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID) {
	w.WriteHeader(http.StatusNotImplemented)
}

// Transcode Segment
// (GET /transcode/{transcodeId}/{segment})
func (_ Unimplemented) GetTranscodeSegment(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID, segment string) {
	w.WriteHeader(http.StatusNotImplemented)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetTranscodeManifestM3u8 operation middleware
func (siw *ServerInterfaceWrapper) GetTranscodeManifestM3u8(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "transcodeId" -------------
	var transcodeId openapi_types.UUID

	err = runtime.BindStyledParameterWithOptions("simple", "transcodeId", chi.URLParam(r, "transcodeId"), &transcodeId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "transcodeId", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetTranscodeManifestM3u8(w, r, transcodeId)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// GetTranscodeSegment operation middleware
func (siw *ServerInterfaceWrapper) GetTranscodeSegment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "transcodeId" -------------
	var transcodeId openapi_types.UUID

	err = runtime.BindStyledParameterWithOptions("simple", "transcodeId", chi.URLParam(r, "transcodeId"), &transcodeId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "transcodeId", Err: err})
		return
	}

	// ------------- Path parameter "segment" -------------
	var segment string

	err = runtime.BindStyledParameterWithOptions("simple", "segment", chi.URLParam(r, "segment"), &segment, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "segment", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetTranscodeSegment(w, r, transcodeId, segment)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/transcode/{transcodeId}/manifest.m3u8", wrapper.GetTranscodeManifestM3u8)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/transcode/{transcodeId}/{segment}", wrapper.GetTranscodeSegment)
	})

	return r
}

type GetTranscodeManifestM3u8RequestObject struct {
	TranscodeId openapi_types.UUID `json:"transcodeId"`
}

type GetTranscodeManifestM3u8ResponseObject interface {
	VisitGetTranscodeManifestM3u8Response(w http.ResponseWriter) error
}

type GetTranscodeManifestM3u8200ApplicationvndAppleMpegurlResponse struct {
	Body          io.Reader
	ContentLength int64
}

func (response GetTranscodeManifestM3u8200ApplicationvndAppleMpegurlResponse) VisitGetTranscodeManifestM3u8Response(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	if response.ContentLength != 0 {
		w.Header().Set("Content-Length", fmt.Sprint(response.ContentLength))
	}
	w.WriteHeader(200)

	if closer, ok := response.Body.(io.ReadCloser); ok {
		defer closer.Close()
	}
	_, err := io.Copy(w, response.Body)
	return err
}

type GetTranscodeSegmentRequestObject struct {
	TranscodeId openapi_types.UUID `json:"transcodeId"`
	Segment     string             `json:"segment"`
}

type GetTranscodeSegmentResponseObject interface {
	VisitGetTranscodeSegmentResponse(w http.ResponseWriter) error
}

type GetTranscodeSegment200VideoResponse struct {
	Body          io.Reader
	ContentType   string
	ContentLength int64
}

func (response GetTranscodeSegment200VideoResponse) VisitGetTranscodeSegmentResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", response.ContentType)
	if response.ContentLength != 0 {
		w.Header().Set("Content-Length", fmt.Sprint(response.ContentLength))
	}
	w.WriteHeader(200)

	if closer, ok := response.Body.(io.ReadCloser); ok {
		defer closer.Close()
	}
	_, err := io.Copy(w, response.Body)
	return err
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Transcode Manifest M3U8
	// (GET /transcode/{transcodeId}/manifest.m3u8)
	GetTranscodeManifestM3u8(ctx context.Context, request GetTranscodeManifestM3u8RequestObject) (GetTranscodeManifestM3u8ResponseObject, error)
	// Transcode Segment
	// (GET /transcode/{transcodeId}/{segment})
	GetTranscodeSegment(ctx context.Context, request GetTranscodeSegmentRequestObject) (GetTranscodeSegmentResponseObject, error)
}

type StrictHandlerFunc = strictnethttp.StrictHTTPHandlerFunc
type StrictMiddlewareFunc = strictnethttp.StrictHTTPMiddlewareFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// GetTranscodeManifestM3u8 operation middleware
func (sh *strictHandler) GetTranscodeManifestM3u8(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID) {
	var request GetTranscodeManifestM3u8RequestObject

	request.TranscodeId = transcodeId

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetTranscodeManifestM3u8(ctx, request.(GetTranscodeManifestM3u8RequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetTranscodeManifestM3u8")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetTranscodeManifestM3u8ResponseObject); ok {
		if err := validResponse.VisitGetTranscodeManifestM3u8Response(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// GetTranscodeSegment operation middleware
func (sh *strictHandler) GetTranscodeSegment(w http.ResponseWriter, r *http.Request, transcodeId openapi_types.UUID, segment string) {
	var request GetTranscodeSegmentRequestObject

	request.TranscodeId = transcodeId
	request.Segment = segment

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.GetTranscodeSegment(ctx, request.(GetTranscodeSegmentRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "GetTranscodeSegment")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(GetTranscodeSegmentResponseObject); ok {
		if err := validResponse.VisitGetTranscodeSegmentResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+RSPW/bPBD+K8SNL2RZifImttamKDx4SdopyMCQJ5uF+VHyZDQw9N+Lo2XZcV0gGTp1",
	"Em3y+bjnnh0ob4N36ChBs4Ok1mhlPn6O0ccHTMG7hPxHiD5gJIP5GvmaD/QaEBpIFI1bQV+AxZTkCi/c",
	"9QVE/NGZiBqap4HiCHgugAxtGPFWvDgQ+ZfvqAh6ZjKu9VljgHyStPYaxRK1keIR4xaZfIsxGe+ggauy",
	"Yns+oJPBQAN1WZXXUECQtM4zTSlKl5TXON2Nx4Xup1Y602Ki0tbdjF+ukPijMaloAu0FHpC66JKgNYpl",
	"/W0m0DGBFge4aH0UtDZJjOwlZEdRMsdCQwNfkL4ebpcDcMmybDRKi4QxQfN0rr64F77N2iO5SJh4eBbB",
	"n9KGHFStru+w+r+azOrb28nNfDabzOf11WTeqrm+k/pKvtwC5wtNjgYKcNIy8iQTOF0lxQ6LoTu/xzJO",
	"Ixb37KT10UqCBrrO6ONux5I8M/V+83kr11XFH+UdocuxyxA2RuXIplunS/6NpQ246uLmWGM+jWIvxsn4",
	"ekGuL878PnZKYUpl7mvqrGXc6RiHreQl51d/LM4u4cqio/5dpUkBlWkNajHAPlyYxz3un+lKcU4wBCDY",
	"xdtRjDNU2nBz2W4ag3uv1XOlj9d4azT66X9/v7CHVvR93/8KAAD//4tQuq3yBQAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}