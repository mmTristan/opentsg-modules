package tsg

import (
	"context"
	"fmt"
	"image/draw"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mrmxf/opentsg-modules/opentsg-core/config/validator"
)

// os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)

func LogToFile(otsg *OpenTSG, opts slog.HandlerOptions, folder, jobID string) {

	path := filepath.Join(folder, fmt.Sprintf("%s.log", jobID))

	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)

	if err != nil {
		panic(err)
	}

	jSlog := slog.NewJSONHandler(f, &opts)
	otsg.Use(Logger(slog.New(jSlog)))

}

// jsonValidator validates the input json request, against a schema.
// It is designed to be the last middleware put on the handler stack.
func jSONValidator(loggedJson validator.JSONLines, schema []byte, id string) func(Handler) Handler {

	return func(h Handler) Handler {

		return HandlerFunc(func(resp Response, req *Request) {

			err := validator.SchemaValidator(schema, req.RawWidgetYAML, id, loggedJson)

			if err != nil {
				// write an error and return
				// skip the rest of the process
				eMess := ""
				for _, e := range err {
					eMess += fmt.Sprintf("%s,", e)
				}
				resp.Write(400, eMess)
				return
			}

			h.Handle(resp, req)
		})
	}
}

// Logger initialises a slogger wrapper,
// records every write() call during the tsg run.
func Logger(logger *slog.Logger) func(Handler) Handler {

	return func(h Handler) Handler {
		return HandlerFunc(func(resp Response, req *Request) {
			// wrap the writer in the slogger body
			slg := slogger{log: logger, r: resp, runID: req.JobID, frameNo: req.FrameProperties.FrameNumber, alias: req.PatchProperties.WidgetFullID}
			h.Handle(&slg, req)
		})
	}
}

// the slogger body for each request
type slogger struct {
	log     *slog.Logger
	r       Response
	c       context.Context
	runID   any
	frameNo int
	alias   string
}

func (s *slogger) Write(status StatusCode, message string) {

	// search code here to find an appropriate error level
	level := getLogLevel(status)

	s.log.Log(s.c, level, message,
		"StatusCode", status.String(),
		"RunID", s.runID,
		"WidgetID", s.alias,
		"FrameNumber", s.frameNo,
	)

	s.r.Write(status, message)
}

// getLogLevel converts the status code into and error level for slog.
func getLogLevel(status StatusCode) slog.Level {
	switch status {
	case WidgetSuccess, 200, Profiler:
		return slog.LevelDebug
	case FrameSuccess:
		return slog.LevelInfo
	case WidgetNotFound, WidgetWarning:
		return slog.LevelWarn
	case 400, 500, EncoderNotFound:
		return slog.LevelError

	default:
		return slog.LevelError

	}
}

func (s *slogger) BaseImage() draw.Image {
	return s.r.BaseImage()
}

/*
func (s *slogger) At(x int, y int) color.Color {
	return s.r.At(x, y)
}
func (s *slogger) Bounds() image.Rectangle {
	return s.r.Bounds()
}
func (s *slogger) ColorModel() color.Model {
	return s.r.ColorModel()
}
func (s *slogger) Set(x int, y int, c color.Color) {
	s.r.Set(x, y, c)
}

func (s *slogger) Space() colour.ColorSpace {
	return s.r.Space()
}
*/
