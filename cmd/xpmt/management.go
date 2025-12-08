package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type managementClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
	timeout time.Duration
}

type managementConfig struct {
	ServerURL string
	Timeout   time.Duration
}

func newManagementClient(apiKey string, config *managementConfig) *managementClient {
	if config == nil {
		config = &managementConfig{}
	}

	baseURL := config.ServerURL
	if baseURL == "" {
		baseURL = os.Getenv("MANAGEMENT_API_URL")
		if baseURL == "" {
			baseURL = "https://experiment.amplitude.com/api"
		}
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &managementClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

func (c *managementClient) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return c.doRequestWithQuery(ctx, method, path, url.Values{}, body)
}

func (c *managementClient) doRequestWithQuery(ctx context.Context, method, path string, query url.Values, body interface{}) ([]byte, error) {
	endpoint, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, err
	}
	endpoint.Path = endpoint.Path + path
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &managementAPIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

type managementAPIError struct {
	StatusCode int
	Message    string
}

func (e *managementAPIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

type managementFlag struct {
	Id                *string                 `json:"id,omitempty"`
	Key               *string                 `json:"key,omitempty"`
	Name              *string                 `json:"name,omitempty"`
	Description       *string                 `json:"description,omitempty"`
	ProjectId         *string                 `json:"projectId,omitempty"`
	Deployments       *[]string               `json:"deployments,omitempty"`
	Variants          *[]interface{}          `json:"variants,omitempty"`
	BucketingKey      *string                 `json:"bucketingKey,omitempty"`
	BucketingSalt     *string                 `json:"bucketingSalt,omitempty"`
	BucketingUnit     *string                 `json:"bucketingUnit,omitempty"`
	RolloutWeights    *map[string]int         `json:"rolloutWeights,omitempty"`
	TargetSegments    *interface{}            `json:"targetSegments,omitempty"`
	EvaluationMode    *string                 `json:"evaluationMode,omitempty"`
	RolloutPercentage *int                    `json:"rolloutPercentage,omitempty"`
	Enabled           *bool                   `json:"enabled,omitempty"`
	Archive           *bool                   `json:"archive,omitempty"`
	Metadata          *map[string]interface{} `json:"metadata,omitempty"`
	Tags              *[]string               `json:"tags,omitempty"`
}

type managementFlagListResponse struct {
	Flags      []managementFlag `json:"flags"`
	NextCursor string           `json:"nextCursor,omitempty"`
}

func (c *managementClient) listFlags(ctx context.Context, flagKey *string, projectId *string, cursor string) (*managementFlagListResponse, error) {
	path := "/1/flags"
	query := url.Values{}
	if flagKey != nil && *flagKey != "" {
		query.Add("key", *flagKey)
	}
	if projectId != nil && *projectId != "" {
		query.Add("projectId", *projectId)
	}
	if cursor != "" {
		query.Add("cursor", cursor)
	}
	data, err := c.doRequestWithQuery(ctx, "GET", path, query, nil)
	if err != nil {
		return nil, err
	}

	var response managementFlagListResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *managementClient) getFlag(ctx context.Context, flagId string) (*managementFlag, error) {
	path := fmt.Sprintf("/1/flags/%s", flagId)

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var flag managementFlag
	err = json.Unmarshal(data, &flag)
	if err != nil {
		return nil, err
	}

	return &flag, nil
}

type managementFlagCreateResponse struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}

func (c *managementClient) createFlag(ctx context.Context, flag *managementFlag) (*managementFlagCreateResponse, error) {
	data, err := c.doRequest(ctx, "POST", "/1/flags", flag)
	if err != nil {
		return nil, err
	}

	var createdFlag managementFlagCreateResponse
	err = json.Unmarshal(data, &createdFlag)
	if err != nil {
		return nil, err
	}

	return &createdFlag, nil
}

func (c *managementClient) updateFlag(ctx context.Context, flagId string, updates *managementFlag) error {
	path := fmt.Sprintf("/1/flags/%s", flagId)

	_, err := c.doRequest(ctx, "PATCH", path, updates)
	if err != nil {
		return err
	}

	return nil
}

func (c *managementClient) listVariantUserIds(ctx context.Context, flagId string, variantKey string) ([]string, error) {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/users", flagId, variantKey)

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response []string
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *managementClient) addVariantUserIds(ctx context.Context, flagKey, variantKey string, userIds []string) error {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/users", flagKey, variantKey)

	body := map[string]interface{}{
		"inclusions": userIds,
	}

	_, err := c.doRequest(ctx, "POST", path, body)
	return err
}

func (c *managementClient) deleteVariantUserIds(ctx context.Context, flagKey, variantKey string, userIds []string) error {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/bulk-delete-users", flagKey, variantKey)

	body := map[string]interface{}{
		"users": userIds,
	}

	_, err := c.doRequest(ctx, "DELETE", path, body)
	return err
}

func (c *managementClient) listVariantCohortIds(ctx context.Context, flagKey, variantKey string) ([]string, error) {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/cohorts", flagKey, variantKey)

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response []string
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *managementClient) addVariantCohortIds(ctx context.Context, flagKey, variantKey string, cohortIds []string) error {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/cohorts", flagKey, variantKey)

	body := map[string]interface{}{
		"inclusions": cohortIds,
	}

	_, err := c.doRequest(ctx, "POST", path, body)
	return err
}

func (c *managementClient) deleteVariantCohortIds(ctx context.Context, flagKey, variantKey string, cohortIds []string) error {
	path := fmt.Sprintf("/1/flags/%s/variants/%s/bulk-delete-cohorts", flagKey, variantKey)

	body := map[string]interface{}{
		"users": cohortIds,
	}

	_, err := c.doRequest(ctx, "DELETE", path, body)
	return err
}

func (c *managementClient) addFlagDeployments(ctx context.Context, flagKey string, deploymentIds []string) error {
	path := fmt.Sprintf("/1/flags/%s/deployments", flagKey)

	body := map[string]interface{}{
		"deployments": deploymentIds,
	}

	_, err := c.doRequest(ctx, "POST", path, body)
	return err
}

func (c *managementClient) deleteFlagDeployments(ctx context.Context, flagKey string, deploymentId string) error {
	path := fmt.Sprintf("/1/flags/%s/deployments/%s", flagKey, deploymentId)

	_, err := c.doRequest(ctx, "DELETE", path, nil)
	return err
}

type managementDeployment struct {
	Id        *string `json:"id,omitempty"`
	ProjectId *string `json:"projectId,omitempty"`
	Key       *string `json:"key,omitempty"`
	Label     *string `json:"label,omitempty"`
	Deleted   *bool   `json:"deleted,omitempty"`
	Type      *string `json:"type,omitempty"`
	Archive   *bool   `json:"archive,omitempty"`
}

type managementDeploymentListResponse struct {
	Deployments []*managementDeployment `json:"deployments,omitempty"`
	NextCursor  string                  `json:"nextCursor,omitempty"`
}

func (c *managementClient) listDeployments(ctx context.Context, cursor string) (*managementDeploymentListResponse, error) {
	path := "/1/deployments"
	if cursor != "" {
		path = fmt.Sprintf("%s?cursor=%s", path, cursor)
	}

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var response managementDeploymentListResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *managementClient) getDeployment(ctx context.Context, deploymentId string) (*managementDeployment, error) {
	path := fmt.Sprintf("/1/deployments/%s", deploymentId)

	data, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var deployment managementDeployment
	err = json.Unmarshal(data, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func (c *managementClient) getDeploymentByProjectIdLabel(ctx context.Context, projectId *string, label *string) ([]*managementDeployment, error) {
	cursor := ""
	result := make([]*managementDeployment, 0)
	for {
		response, err := c.listDeployments(ctx, cursor)
		if err != nil {
			return nil, err
		}

		for _, dep := range response.Deployments {
			if (projectId == nil || *dep.ProjectId == *projectId) && (label == nil || *dep.Label == *label) {
				result = append(result, dep)
			}
		}
		cursor = response.NextCursor
		if cursor == "" {
			break
		}
	}

	return result, nil
}

func (c *managementClient) createDeployment(ctx context.Context, deployment *managementDeployment) (*managementDeployment, error) {
	data, err := c.doRequest(ctx, "POST", "/1/deployments", deployment)
	if err != nil {
		return nil, err
	}

	var createdDeployment managementDeployment
	err = json.Unmarshal(data, &createdDeployment)
	if err != nil {
		return nil, err
	}

	return &createdDeployment, nil
}

func (c *managementClient) updateDeployment(ctx context.Context, deploymentId string, deployment *managementDeployment) error {
	path := fmt.Sprintf("/1/deployments/%s", deploymentId)
	_, err := c.doRequest(ctx, "PATCH", path, deployment)
	return err
}
