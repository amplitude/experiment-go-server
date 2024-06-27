package local

import (
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCohortDownloadApi struct {
	mock.Mock
}

func (m *MockCohortDownloadApi) GetCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	args := m.Called(cohortID, cohort)
	if args.Get(0) != nil {
		return args.Get(0).(*Cohort), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCohortDownloadApi(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	api := NewDirectCohortDownloadApi("api", "secret", 15000, 100, "https://server.amplitude.com", false)

	t.Run("test_cohort_download_success", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_cohort_download_many_202s_success", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}

		for i := 0; i < 9; i++ {
			httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
				httpmock.NewStringResponder(202, ""),
			)
		}
		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_cohort_request_status_with_two_failures_succeeds", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(503, ""),
		)
		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(503, ""),
		)
		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_cohort_request_status_429s_keep_retrying", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"user"}}

		for i := 0; i < 9; i++ {
			httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
				httpmock.NewStringResponder(429, ""),
			)
		}
		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_group_cohort_download_success", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"group"}, GroupType: "org name"}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"group"}, GroupType: "org name"}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_group_cohort_request_status_429s_keep_retrying", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"group"}, GroupType: "org name"}
		response := &Cohort{ID: "1234", LastModified: 0, Size: 1, MemberIDs: []string{"group"}, GroupType: "org name"}

		for i := 0; i < 9; i++ {
			httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
				httpmock.NewStringResponder(429, ""),
			)
		}
		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.GetCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort, resultCohort)
	})

	t.Run("test_cohort_size_too_large", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 0, Size: 16000, MemberIDs: []string{}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(413, ""),
		)

		_, err := api.GetCohort("1234", cohort)
		assert.Error(t, err)
		_, isCohortTooLargeException := err.(*CohortTooLargeException)
		assert.True(t, isCohortTooLargeException)
	})

	t.Run("test_cohort_not_modified_exception", func(t *testing.T) {
		cohort := &Cohort{ID: "1234", LastModified: 1000, Size: 1, MemberIDs: []string{}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(204, ""),
		)

		_, err := api.GetCohort("1234", cohort)
		assert.Error(t, err)
		_, isCohortNotModifiedException := err.(*CohortNotModifiedException)
		assert.True(t, isCohortNotModifiedException)
	})
}
