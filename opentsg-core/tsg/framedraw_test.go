package tsg

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

/*
func TestXxx(t *testing.T) {

		otsg := openTSG{widgets: map[string]func([]byte) (Generator, error){}} // map[string]func([]byte) (Generator, error){}}
		otsg.AddWidget("test", []byte("{}"), testWidget{})
		otsg.AddWidget("test2", []byte("{}"), testWidget2{})

		bs := []byte("{\"input\":\"yes\"}")
		//var x any
		x, err := otsg.widgets["test"](bs)

		fmt.Println(err)
		x.Generate(nil)

		x2, _ := otsg.widgets["test2"](bs)
		x2.Generate(nil)
	}
*/

func TestLegacy(t *testing.T) {
	otsg, err := BuildOpenTSG("./testdata/testloader.json", "", true)

	first := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			fmt.Println("fitrst")

			next.Handle(r1, r2)
		})
	}

	mid := func(next Handler) Handler {
		return HandlerFunc(func(r1 Response, r2 *Request) {
			fmt.Println("SOME THING FROM CUSTOM MIDDLEWARE")

			next.Handle(r1, r2)
		})
	}
	otsg.Use(first, mid)
	otsg.Handle("builtin.legacy", []byte("{}"), Legacy{})
	otsg.HandleFunc("builtin.canvasoptions", func(r1 Response, r2 *Request) { fmt.Println("ring a ding") })
	fmt.Println(err)
	otsg.Run(true, "", "stdout")

}

func TestMarshallHandler(t *testing.T) {

	testHandlers := map[string]Handler{
		"dummyHandler":       dummyHandler{},
		"secondDummyHandler": secondDummyHandler{},
	}

	expected := map[string]Handler{
		"dummyHandler":       &dummyHandler{"testInput"},
		"secondDummyHandler": &secondDummyHandler{"testInput"},
	}

	tragets := []string{"dummyHandler", "secondDummyHandler"}

	for _, target := range tragets {

		gotHandler, err := Unmarshal(testHandlers[target])([]byte(`{"input":"testInput"}`))

		Convey("Checking the unmarshaling of bytes to method structs", t, func() {
			Convey(fmt.Sprintf("Unmarshaling bytes to a struct of %v ", reflect.TypeOf(testHandlers[target])), func() {
				Convey("No error is returned and the struct is populated as expected", func() {
					So(err, ShouldBeNil)
					So(gotHandler, ShouldResemble, expected[target])
				})
			})
		})

	}
}

type dummyHandler struct {
	Input string `json:"input"`
}

func (d dummyHandler) Handle(resp Response, req *Request) {
}

type secondDummyHandler struct {
	Input string `json:"input"`
}

func (d secondDummyHandler) Handle(resp Response, req *Request) {
}
