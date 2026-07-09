package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

func TestInternalStreamHandler_WritesIDFrameWhenAvailable(t *testing.T) {
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
	if !strings.Contains(body, "id: 3\ndata: {\"sequence\":3}\n\n") {
		t.Fatalf("expected id+data frame for the object with an ID, got:\n%q", body)
	}

	// The plain object gets a data frame with no id line.
	if !strings.Contains(body, "data: {\"sequence\":4}\n\n") {
		t.Fatalf("expected data frame for the plain object, got:\n%q", body)
	}
	if strings.Contains(body, "id: 4") {
		t.Fatalf("did not expect an id line for the plain object, got:\n%q", body)
	}
}
