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

type cohortInfo struct {
	Id           string   `json:"cohortId"`
	LastModified int64    `json:"lastModified"`
	Size         int      `json:"size"`
	MemberIds    []string `json:"memberIds"`
	GroupType    string   `json:"groupType"`
}

func (m *MockCohortDownloadApi) getCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	args := m.Called(cohortID, cohort)
	if args.Get(0) != nil {
		return args.Get(0).(*Cohort), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestCohortDownloadApi(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	api := newDirectCohortDownloadApi("api", "secret", 15000, 100, "https://server.amplitude.com", false)

	t.Run("test_cohort_download_success", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}
		response := cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_cohort_download_many_202s_success", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}
		response := &cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}

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

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_cohort_request_status_with_two_failures_succeeds", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}
		response := &cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}

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

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_cohort_request_status_429s_keep_retrying", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}
		response := &cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"user"}, GroupType: "userGroupType"}

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

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_group_cohort_download_success", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"group"}, GroupType: "org name"}
		response := &cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"group"}, GroupType: "org name"}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			func(req *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, response)
				if err != nil {
					return httpmock.NewStringResponse(500, ""), nil
				}
				return resp, nil
			},
		)

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_group_cohort_request_status_429s_keep_retrying", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"group"}, GroupType: "org name"}
		response := &cohortInfo{Id: "1234", LastModified: 0, Size: 1, MemberIds: []string{"group"}, GroupType: "org name"}

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

		resultCohort, err := api.getCohort("1234", cohort)
		assert.NoError(t, err)
		assert.Equal(t, cohort.Id, resultCohort.Id)
		assert.Equal(t, cohort.LastModified, resultCohort.LastModified)
		assert.Equal(t, cohort.Size, resultCohort.Size)
		assert.Equal(t, cohort.MemberIds, resultCohort.MemberIds)
		assert.Equal(t, cohort.GroupType, resultCohort.GroupType)
	})

	t.Run("test_cohort_size_too_large", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 0, Size: 16000, MemberIds: []string{}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(413, ""),
		)

		_, err := api.getCohort("1234", cohort)
		assert.Error(t, err)
		_, isCohortTooLargeException := err.(*CohortTooLargeException)
		assert.True(t, isCohortTooLargeException)
	})

	t.Run("test_cohort_not_modified_exception", func(t *testing.T) {
		cohort := &Cohort{Id: "1234", LastModified: 1000, Size: 1, MemberIds: []string{}}

		httpmock.RegisterResponder("GET", api.buildCohortURL("1234", cohort),
			httpmock.NewStringResponder(204, ""),
		)

		result, err := api.getCohort("1234", cohort)
		assert.Nil(t, result)
		assert.NoError(t, err)
	})
}
