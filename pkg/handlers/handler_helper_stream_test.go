package handlers

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/interfaces"
)

// idStreamObject implements interfaces.StreamObjectWithID.
type idStreamObject struct {
	data string
	id   string
}

func (o idStreamObject) Data() string { return o.data }
func (o idStreamObject) Error() error { return nil }
func (o idStreamObject) ID() string   { return o.id }

// plainStreamObject implements only interfaces.StreamObject (no ID).
type plainStreamObject struct{ data string }

func (o plainStreamObject) Data() string { return o.data }
func (o plainStreamObject) Error() error { return nil }

// errStreamObject implements interfaces.StreamObject and reports a stream error.
type errStreamObject struct{ err error }

func (o errStreamObject) Data() string { return "" }
func (o errStreamObject) Error() error { return o.err }

// fakeStreamable emits a fixed set of objects then closes.
type fakeStreamable struct {
	objects []interfaces.StreamObject
}

func (f fakeStreamable) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	ch := make(chan interfaces.StreamObject)
	go func() {
		defer close(ch)
		for _, o := range f.objects {
			select {
			case ch <- o:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// streamBody runs internalStreamHandler against a streamable and returns the
// recorded SSE body.
func streamBody(streamable interfaces.ServerSideStreamable) string {
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	rec := httptest.NewRecorder()
	internalStreamHandler(rec, req, streamable, &handlerConfig{})
	return rec.Body.String()
}

var _ = Describe("internalStreamHandler", func() {
	It("writes an id frame when the object provides one", func() {
		streamable := fakeStreamable{objects: []interfaces.StreamObject{
			idStreamObject{data: `{"sequence":3}`, id: "3"},
			plainStreamObject{data: `{"sequence":4}`},
		}}

		req := httptest.NewRequest(http.MethodGet, "/stream", nil)
		rec := httptest.NewRecorder()

		cfg := &handlerConfig{}
		internalStreamHandler(rec, req, streamable, cfg)

		body := rec.Body.String()

		// The object implementing StreamObjectWithID gets an id line before its data.
		Expect(body).To(ContainSubstring("id: 3\ndata: {\"sequence\":3}\n\n"))

		// The plain object gets a data frame with no id line.
		Expect(body).To(ContainSubstring("data: {\"sequence\":4}\n\n"))
		Expect(body).NotTo(ContainSubstring("id: 4"))
	})

	It("emits a terminal end event when the source completes", func() {
		body := streamBody(fakeStreamable{objects: []interfaces.StreamObject{
			plainStreamObject{data: "line-1"},
		}})
		Expect(body).To(HaveSuffix("event: end\ndata: {}\n\n"))
	})

	It("does not emit end after a stream error", func() {
		body := streamBody(fakeStreamable{objects: []interfaces.StreamObject{
			errStreamObject{err: stderrors.New("boom")},
		}})
		Expect(body).To(ContainSubstring("event: error"))
		Expect(body).NotTo(ContainSubstring("event: end"))
	})
})
