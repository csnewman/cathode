package mediaserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	v1 "github.com/csnewman/cathode/internal/v1"
	oapi "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type v1API struct {
	logger *slog.Logger
	router *chi.Mux
}

func newV1API(logger *slog.Logger) (*v1API, error) {
	r := chi.NewRouter()

	api := &v1API{
		logger: logger,
		router: r,
	}

	r.Use(middleware.RequestID)
	r.Use(api.httpLogger)
	//r.Use(middleware.NoCache)
	r.Use(api.httpRecoverer)

	spec, err := v1.GetSwagger()
	if err != nil {
		return nil, err
	}

	// Change mount point
	newPaths := make(openapi3.Paths)
	for s, item := range spec.Paths {
		newPaths[fmt.Sprintf("/api/v1%v", s)] = item
	}

	spec.Paths = newPaths

	for s, item := range spec.Servers {
		logger.Info("s", "s", s, "it", item)
	}

	r.Use(oapi.OapiRequestValidatorWithOptions(
		spec,
		&oapi.Options{
			Options: openapi3filter.Options{},
			ErrorHandler: func(w http.ResponseWriter, message string, _ int) {
				api.logger.Error(
					"API bad request",
					"message", message,
				)

				v := &v1.ErrorResponse{
					Error:   "bad-request",
					Message: message,
				}

				buf := &bytes.Buffer{}
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(true)
				if err := enc.Encode(v); err != nil {
					panic(err)
				}

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(buf.Bytes())
			},
			MultiErrorHandler: nil,
		},
	))

	v1.HandlerWithOptions(
		v1.NewStrictHandlerWithOptions(
			api,
			nil,
			v1.StrictHTTPServerOptions{
				RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					api.logger.Error(
						"API request error",
						"path", r.URL.Path,
						"request_id", middleware.GetReqID(r.Context()),
						"err", err,
					)

					writeResponse(
						w, r, http.StatusBadRequest,
						"bad-request",
						"An error was found in the request",
					)
				},
				ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					api.logger.Error(
						"API response error",
						"path", r.URL.Path,
						"request_id", middleware.GetReqID(r.Context()),
						"err", err,
					)

					writeResponse(
						w, r, http.StatusInternalServerError,
						"internal-error",
						"An internal server error has occurred",
					)
				},
			},
		),
		v1.ChiServerOptions{
			BaseRouter: r,
			ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				api.logger.Error(
					"API request error",
					"path", r.URL.Path,
					"request_id", middleware.GetReqID(r.Context()),
					"err", err,
				)

				writeResponse(
					w, r, http.StatusInternalServerError,
					"internal-error",
					"An internal server error has occurred",
				)
			},
		},
	)

	return api, nil
}

func (a *v1API) httpLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			ua := ww.Header().Get("User-Agent")
			if ua == "" {
				ua = r.Header.Get("User-Agent")
			}

			a.logger.Debug(
				"Served API Request",
				"path", r.URL.Path,
				"request_id", middleware.GetReqID(r.Context()),
				"took", time.Since(t1),
				"status", ww.Status(),
				"size", ww.BytesWritten(),
				"ua", ua,
			)
		}()
		next.ServeHTTP(ww, r)
	})
}

func (a *v1API) httpRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				//nolint:errorlint,goerr113
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				a.logger.Error(
					"API Request Error",
					"path", r.URL.Path,
					"request_id", middleware.GetReqID(r.Context()),
					"err", rvr,
				)

				writeResponse(
					w, r, http.StatusInternalServerError,
					"internal-error",
					"An internal server error has occurred",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func writeResponse(w http.ResponseWriter, r *http.Request, status int, code string, msg string) {
	w.WriteHeader(status)
	render.JSON(w, r, &v1.ErrorResponse{
		Error:   code,
		Message: msg,
	})
}
