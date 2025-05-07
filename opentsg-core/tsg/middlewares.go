package tsg

import (
	"context"
	"fmt"
	"image/draw"
	"log/slog"
	"os"
	"path/filepath"
)

// os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0777)

// LogOptions provides the configuration options for
// the logging
type LogOptions struct {
	Folder string
	JobID  string
	// Make the slog that is used by the middleware
	// the default slog call as well
	MakeDefaultSlog bool
}

// LogToFile attaches a json slogging middleware to openTSG that writes to file of jobid.log
// If the file already exists it is appended to.
func LogToFile(otsg *OpenTSG, opts slog.HandlerOptions, options *LogOptions) {
	if options == nil {
		options = &LogOptions{
			JobID: "default",
		}
	}

	path := filepath.Join(options.Folder, fmt.Sprintf("%s.log", options.JobID))

	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0777)

	if err != nil {
		panic(err)
	}

	jSlog := slog.NewJSONHandler(f, &opts)
	slogging := slog.New(jSlog)
	if options.MakeDefaultSlog {
		slog.SetDefault(slogging)
	}
	otsg.Use(Logger(slogging))

}

// Logger initialises a slogger wrapper, that
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

// slogger writes the status code and message to the logger
// before forwarding the request to the wrapped wrtiers
func (s *slogger) Write(status StatusCode, message string, args ...any) {

	// search code here to find an appropriate error level
	level := getLogLevel(status)

	logFields := make([]any, len(args)+8)
	logFields[0] = "StatusCode"
	logFields[1] = status.String()
	logFields[2] = "RunID"
	logFields[3] = s.runID
	logFields[4] = "WidgetID"
	logFields[5] = s.alias
	logFields[6] = "FrameNumber"
	logFields[7] = s.frameNo

	for i := 8; i < 8+len(args); i++ {
		logFields[i] = args[i-8]
	}

	s.log.Log(s.c, level, message,
		logFields...,
	)

	s.r.Write(status, message, args...)
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
