package local

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	CdnCohortSyncUrl = "https://cohort-v2.lab.amplitude.com"
)

type HTTPErrorResponseException struct {
	StatusCode int
	Message    string
}

func (e *HTTPErrorResponseException) Error() string {
	return e.Message
}

type CohortTooLargeException struct {
	Message string
}

func (e *CohortTooLargeException) Error() string {
	return e.Message
}

type CohortNotModifiedException struct {
	Message string
}

func (e *CohortNotModifiedException) Error() string {
	return e.Message
}

type CohortDownloadApi interface {
	GetCohort(cohortID string, cohort *Cohort) (*Cohort, error)
}

type DirectCohortDownloadApi struct {
	ApiKey                   string
	SecretKey                string
	MaxCohortSize            int
	CohortRequestDelayMillis int
	Debug                    bool
	Logger                   *log.Logger
}

func NewDirectCohortDownloadApi(apiKey, secretKey string, maxCohortSize, cohortRequestDelayMillis int, debug bool) *DirectCohortDownloadApi {
	api := &DirectCohortDownloadApi{
		ApiKey:                   apiKey,
		SecretKey:                secretKey,
		MaxCohortSize:            maxCohortSize,
		CohortRequestDelayMillis: cohortRequestDelayMillis,
		Debug:                    debug,
		Logger:                   log.New(log.Writer(), "Amplitude: ", log.LstdFlags),
	}
	if debug {
		api.Logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}
	return api
}

func (api *DirectCohortDownloadApi) GetCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	api.Logger.Printf("getCohortMembers(%s): start", cohortID)
	errors := 0
	client := &http.Client{}

	for {
		response, err := api.getCohortMembersRequest(client, cohortID, cohort)
		if err != nil {
			api.Logger.Printf("getCohortMembers(%s): request-status error %d - %v", cohortID, errors, err)
			errors++
			if errors >= 3 || isSpecificError(err) {
				return nil, err
			}
			time.Sleep(time.Duration(api.CohortRequestDelayMillis) * time.Millisecond)
			continue
		}

		if response.StatusCode == http.StatusOK {
			var cohortInfo struct {
				CohortId     string   `json:"cohortId"`
				LastModified int64    `json:"lastModified"`
				Size         int      `json:"size"`
				MemberIds    []string `json:"memberIds"`
				GroupType    string   `json:"groupType"`
			}
			if err := json.NewDecoder(response.Body).Decode(&cohortInfo); err != nil {
				return nil, err
			}
			memberIDs := make(map[string]struct{}, len(cohortInfo.MemberIds))
			for _, id := range cohortInfo.MemberIds {
				memberIDs[id] = struct{}{}
			}
			api.Logger.Printf("getCohortMembers(%s): end - resultSize=%d", cohortID, cohortInfo.Size)
			return &Cohort{
				ID:           cohortInfo.CohortId,
				LastModified: cohortInfo.LastModified,
				Size:         cohortInfo.Size,
				MemberIDs:    memberIDs,
				GroupType:    cohortInfo.GroupType,
			}, nil
		} else if response.StatusCode == http.StatusNoContent {
			return nil, &CohortNotModifiedException{Message: "Cohort not modified"}
		} else if response.StatusCode == http.StatusRequestEntityTooLarge {
			return nil, &CohortTooLargeException{Message: "Cohort exceeds max cohort size"}
		} else {
			return nil, &HTTPErrorResponseException{StatusCode: response.StatusCode, Message: "Unexpected response code"}
		}
	}
}

func isSpecificError(err error) bool {
	switch err.(type) {
	case *CohortNotModifiedException, *CohortTooLargeException:
		return true
	default:
		return false
	}
}

func (api *DirectCohortDownloadApi) getCohortMembersRequest(client *http.Client, cohortID string, cohort *Cohort) (*http.Response, error) {
	req, err := http.NewRequest("GET", api.buildCohortURL(cohortID, cohort), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Basic "+api.getBasicAuth())
	return client.Do(req)
}

func (api *DirectCohortDownloadApi) getBasicAuth() string {
	auth := api.ApiKey + ":" + api.SecretKey
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (api *DirectCohortDownloadApi) buildCohortURL(cohortID string, cohort *Cohort) string {
	url := CdnCohortSyncUrl + "/sdk/v1/cohort/" + cohortID + "?maxCohortSize=" + strconv.Itoa(api.MaxCohortSize)
	if cohort != nil {
		url += "&lastModified=" + strconv.FormatInt(cohort.LastModified, 10)
	}
	return url
}
