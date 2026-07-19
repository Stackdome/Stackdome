package githubapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PullRequestCommenter", func() {
	var (
		srv       *httptest.Server
		commenter PullRequestCommenter
		gotPath   string
		gotBody   map[string]any
		status    int
		reply     any
	)

	BeforeEach(func() {
		status = http.StatusOK
		gotPath = ""
		gotBody = nil
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.Method + " " + r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(reply)
		}))
		DeferCleanup(srv.Close)
		commenter = NewPullRequestCommenter(PullRequestCommenterSpec{BaseURL: srv.URL})
	})

	It("creates a comment and returns its id", func() {
		status = http.StatusCreated
		reply = map[string]any{"id": 4242}

		id, err := commenter.CreateComment(context.Background(), "tok", "acme", "app", 7, "hello")
		Expect(err).ToNot(HaveOccurred())
		Expect(id).To(Equal(int64(4242)))
		Expect(gotPath).To(Equal("POST /repos/acme/app/issues/7/comments"))
		Expect(gotBody["body"]).To(Equal("hello"))
	})

	It("edits a comment by id", func() {
		reply = map[string]any{"id": 4242}

		err := commenter.EditComment(context.Background(), "tok", "acme", "app", 4242, "updated")
		Expect(err).ToNot(HaveOccurred())
		Expect(gotPath).To(Equal("PATCH /repos/acme/app/issues/comments/4242"))
		Expect(gotBody["body"]).To(Equal("updated"))
	})

	It("maps a 404 on edit to ErrCommentNotFound", func() {
		status = http.StatusNotFound
		reply = map[string]any{"message": "Not Found"}

		err := commenter.EditComment(context.Background(), "tok", "acme", "app", 4242, "updated")
		Expect(err).To(MatchError(ErrCommentNotFound))
	})
})
