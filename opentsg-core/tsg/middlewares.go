package tsg

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"log/slog"

	gonanoid "github.com/matoous/go-nanoid"
	"github.com/mrmxf/opentsg-modules/opentsg-core/config/validator"
)

/*


need
- logger
- validator
- metadata substituions
*/

// jsonValidator validates the input json request, against a schema
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

// Logger initialises a slogger wrapper of any writes that
// occur during the TSG run
func Logger(logger *slog.Logger, runID any) func(Handler) Handler {

	if runID == nil {
		runID = gonanoid.MustID(18)
	}

	return func(h Handler) Handler {
		return HandlerFunc(func(resp Response, req *Request) {
			// wrap the writer in the slogger body
			slg := slogger{log: logger, r: resp, runID: runID, frameNo: req.FrameProperties.FrameNumber, alias: req.PatchProperties.WidgetFullID}
			h.Handle(&slg, req)
		})
	}
}

type slogger struct {
	log     *slog.Logger
	r       Response
	c       context.Context
	runID   any
	frameNo int
	alias   string
}

func (s *slogger) Write(status int, message string) {
	// switch the code here to find an appropriate error level
	s.log.Log(s.c, slog.LevelError, fmt.Sprintf("%v:%s", status, message),
		"RunID", s.runID,
		"WidgetID", s.alias,
		"FrameNumber", s.frameNo,
	)

	s.r.Write(status, message)
}

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
