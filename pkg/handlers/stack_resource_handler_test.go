package handlers

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/logger"
)

var _ = Describe("stackResourceHandler.StreamLogs", func() {
	var (
		ctrl        *gomock.Controller
		mockLogging *MockLoggingService
		handler     *stackResourceHandler
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLogging = NewMockLoggingService(ctrl)
		handler = NewStackResourceHandler(StackResourceHandlerSpec{
			LoggingService: mockLogging,
			Logger:         logger.NewLogger(),
		})
	})

	AfterEach(func() { ctrl.Finish() })

	serve := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		req = mux.SetURLVars(req, map[string]string{"org_id": "org-1", "id": "stack-1", "resource_name": "web"})
		rec := httptest.NewRecorder()
		handler.StreamLogs(rec, req)
		return rec
	}

	It("returns 404 when the resource has no workload to read logs from", func() {
		mockLogging.EXPECT().
			StreamLogsForStackResource(gomock.Any(), "org-1", "stack-1", "web", gomock.Any()).
			Return(nil, errors.NotFound("no logs available for web"))

		Expect(serve().Code).To(Equal(http.StatusNotFound))
	})

	It("returns 500 for a genuine failure", func() {
		mockLogging.EXPECT().
			StreamLogsForStackResource(gomock.Any(), "org-1", "stack-1", "web", gomock.Any()).
			Return(nil, errors.GeneralError("cluster unreachable"))

		Expect(serve().Code).To(Equal(http.StatusInternalServerError))
	})

	It("streams log lines for a crash-looping resource", func() {
		streamable := fakeStreamable{objects: []interfaces.StreamObject{
			plainStreamObject{data: "panic: boom"},
		}}
		mockLogging.EXPECT().
			StreamLogsForStackResource(gomock.Any(), "org-1", "stack-1", "web", gomock.Any()).
			Return(streamable, nil)

		Expect(serve().Body.String()).To(ContainSubstring("data: panic: boom\n\n"))
	})
})
