package handlers

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/gorilla/mux"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
)

var _ = Describe("imageBuildHandler.StreamLogs", func() {
	It("streams build log lines as SSE", func() {
		ctrl := gomock.NewController(GinkgoT())
		defer ctrl.Finish()

		mockLogging := mocks.NewMockLoggingService(ctrl)
		streamable := fakeStreamable{objects: []interfaces.StreamObject{
			plainStreamObject{data: "step 1/5 : FROM golang"},
		}}
		mockLogging.EXPECT().
			StreamLogsForBuild(gomock.Any(), "org-1", "build-1", gomock.Any()).
			Return(streamable, nil)

		h := NewImageBuildHandler(ImageBuildHandlerSpec{
			LoggingService: mockLogging,
			Logger:         logger.NewLogger(),
		})

		req := httptest.NewRequest(http.MethodGet, "/logs", nil)
		req = mux.SetURLVars(req, map[string]string{"org_id": "org-1", "build_id": "build-1"})
		rec := httptest.NewRecorder()

		h.StreamLogs(rec, req)

		Expect(rec.Body.String()).To(ContainSubstring("data: step 1/5 : FROM golang\n\n"))
	})
})
