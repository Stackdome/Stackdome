package handlers

import (
	"context"
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
})
