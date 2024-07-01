package local

import (
	"encoding/base64"
	"encoding/json"
	"github.com/amplitude/experiment-go-server/internal/logger"
	"net/http"
	"strconv"
	"time"
)

type DirectCohortDownloadApi struct {
	ApiKey                   string
	SecretKey                string
	MaxCohortSize            int
	CohortRequestDelayMillis int
	ServerUrl                string
	Debug                    bool
	log                      *logger.Log
}

func NewDirectCohortDownloadApi(apiKey, secretKey string, maxCohortSize, cohortRequestDelayMillis int, serverUrl string, debug bool) *DirectCohortDownloadApi {
	api := &DirectCohortDownloadApi{
		ApiKey:                   apiKey,
		SecretKey:                secretKey,
		MaxCohortSize:            maxCohortSize,
		CohortRequestDelayMillis: cohortRequestDelayMillis,
		ServerUrl:                serverUrl,
		Debug:                    debug,
		log:                      logger.New(debug),
	}
	return api
}

func (api *DirectCohortDownloadApi) GetCohort(cohortID string, cohort *Cohort) (*Cohort, error) {
	api.log.Debug("getCohortMembers(%s): start", cohortID)
	errors := 0
	client := &http.Client{}

	for {
		response, err := api.getCohortMembersRequest(client, cohortID, cohort)
		if err != nil {
			api.log.Error("getCohortMembers(%s): request-status error %d - %v", cohortID, errors, err)
			errors++
			if errors >= 3 || func(err error) bool {
				switch err.(type) {
				case *CohortNotModifiedException, *CohortTooLargeException:
					return true
				default:
					return false
				}
			}(err) {
				return nil, err
			}
			time.Sleep(time.Duration(api.CohortRequestDelayMillis) * time.Millisecond)
			continue
		}

		if response.StatusCode == http.StatusOK {
			var cohortInfo struct {
				Id           string   `json:"cohortId"`
				LastModified int64    `json:"lastModified"`
				Size         int      `json:"size"`
				MemberIds    []string `json:"memberIds"`
				GroupType    string   `json:"groupType"`
			}
			if err := json.NewDecoder(response.Body).Decode(&cohortInfo); err != nil {
				return nil, err
			}
			api.log.Debug("getCohortMembers(%s): end - resultSize=%d", cohortID, cohortInfo.Size)
			return &Cohort{
				ID:           cohortInfo.Id,
				LastModified: cohortInfo.LastModified,
				Size:         cohortInfo.Size,
				MemberIDs:    cohortInfo.MemberIds,
				GroupType: func() string {
					if cohortInfo.GroupType == "" {
						return userGroupType
					}
					return cohortInfo.GroupType
				}(),
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
	url := api.ServerUrl + "/sdk/v1/cohort/" + cohortID + "?maxCohortSize=" + strconv.Itoa(api.MaxCohortSize)
	if cohort != nil {
		url += "&lastModified=" + strconv.FormatInt(cohort.LastModified, 10)
	}
	return url
}
