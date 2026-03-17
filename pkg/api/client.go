package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client handles WebUntis API communication
type Client struct {
	BaseURL      string
	School       string
	SessionID    string // Classic JSESSIONID, needed for JWT exchange
	JWTToken     string // Modern authentication token
	TenantID     string // Required header
	SchoolYearID string // Required header
	CSRFToken    string // Required for POST
	Students     []StudentInfo
	HTTPClient   *http.Client
}

// NewClient creates a new WebUntis API client
func NewClient(server, school string) *Client {
	protocol := "https://"
	if strings.HasPrefix(server, "http://") || strings.HasPrefix(server, "https://") {
		protocol = ""
	}

	u := fmt.Sprintf("%s%s/WebUntis/jsonrpc.do?school=%s", protocol, server, url.QueryEscape(school))

	return &Client{
		BaseURL: u,
		School:  school,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Main Authentication Orchestrator ---

// Authenticate performs the full multi-step authentication flow.
func (c *Client) Authenticate(user, password string) (*AuthResult, error) {
	// Step 1: Legacy authentication to get session
	authResult, err := c.authenticateLegacy(user, password)
	if err != nil {
		return nil, fmt.Errorf("step 1 (legacy auth) failed: %w", err)
	}

	// Step 2: Get JWT token using the classic session
	if err := c.fetchJWTToken(); err != nil {
		return nil, fmt.Errorf("step 2 (fetch jwt) failed: %w", err)
	}

	// Step 3: Get app data (tenantId, schoolYearId) using the JWT
	if err := c.fetchAppData(); err != nil {
		return nil, fmt.Errorf("step 3 (fetch app data) failed: %w", err)
	}

	return authResult, nil
}

// --- Authentication Sub-steps ---

// authenticateLegacy performs the classic login to get a JSESSIONID.
func (c *Client) authenticateLegacy(user, password string) (*AuthResult, error) {
	params := AuthParams{
		User:     user,
		Password: password,
		Client:   "go-cli",
	}

	var result AuthResult

	// The old endpoint is still used for the initial authentication
	oldAuthURL, err := c.getEndpointURL("/WebUntis/jsonrpc.do")
	if err != nil {
		return nil, err
	}

	// We use a temporary client for this call to not interfere with the main one
	tempClient := *c
	tempClient.BaseURL = oldAuthURL

	if err := tempClient.call("authenticate", params, &result); err != nil {
		return nil, err
	}

	// Store classic session ID, it's needed for the next step
	c.SessionID = result.SessionID
	return &result, nil
}

// fetchJWTToken calls the /api/token/new endpoint to get the JWT.
func (c *Client) fetchJWTToken() error {
	tokenURL, err := c.getEndpointURL("/WebUntis/api/token/new")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return err
	}
	req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.SessionID})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// The body is just the JWT token as a string
	c.JWTToken = strings.TrimSpace(string(body))

	return nil
}

// fetchAppData calls the /api/rest/view/v1/app/data endpoint.
func (c *Client) fetchAppData() error {
	appDataURL, err := c.getEndpointURL("/WebUntis/api/rest/view/v1/app/data")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", appDataURL, nil)
	if err != nil {
		return err
	}

	// This call requires the JWT token
	req.Header.Set("Authorization", "Bearer "+c.JWTToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	var appData AppDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&appData); err != nil {
		return err
	}

	c.TenantID = appData.Tenant.ID
	c.SchoolYearID = strconv.Itoa(appData.CurrentSchoolYear.ID)
	c.Students = appData.User.Students

	return nil
}

// --- Core JSON-RPC and Helpers ---

// call executes a JSON-RPC method. Renamed to be private.
func (c *Client) call(method string, params interface{}, result interface{}) error {
	reqBody := JsonRpcRequest{
		ID:      "go-cli-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Method:  method,
		Params:  params,
		JSONRPC: "2.0",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.SessionID != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.SessionID})
	}
	if c.JWTToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.JWTToken)
	}
	if c.TenantID != "" {
		req.Header.Set("tenant-id", c.TenantID)
	}
	if c.SchoolYearID != "" {
		req.Header.Set("x-webuntis-api-school-year-id", c.SchoolYearID)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(body))
		return fmt.Errorf("HTTP error: %d - Body: %s", resp.StatusCode, string(body))
	}

	var rpcResp rawRpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if rpcResp.Error != nil {
		return fmt.Errorf("API Error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	if result != nil && rpcResp.Result != nil {
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

func (c *Client) getEndpointURL(path string) (string, error) {
	parsedURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	// Construct URL relative to the server root, not the json-rpc path
	return fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, path), nil
}

// --- Public API Methods ---

// Logout invalidates the current session
func (c *Client) Logout() error {
	oldAuthURL, err := c.getEndpointURL("/WebUntis/jsonrpc.do")
	if err != nil {
		return err
	}
	tempClient := *c
	tempClient.BaseURL = oldAuthURL

	if err := tempClient.call("logout", map[string]string{}, nil); err != nil {
		return err
	}
	c.SessionID = ""
	c.JWTToken = ""
	return nil
}

// GetKlassen retrieves all classes
func (c *Client) GetKlassen() ([]Klasse, error) {
	var result []Klasse
	err := c.call("getKlassen", map[string]string{}, &result)
	return result, err
}

// GetTeachers retrieves all teachers
func (c *Client) GetTeachers() ([]Teacher, error) {
	var result []Teacher
	err := c.call("getTeachers", map[string]string{}, &result)
	return result, err
}

// GetSubjects retrieves all subjects
func (c *Client) GetSubjects() ([]Subject, error) {
	var result []Subject
	err := c.call("getSubjects", map[string]string{}, &result)
	return result, err
}

// GetRooms retrieves all rooms
func (c *Client) GetRooms() ([]Room, error) {
	var result []Room
	err := c.call("getRooms", map[string]string{}, &result)
	return result, err
}

// GetTimetable retrieves timetable entries
func (c *Client) GetTimetable(options TimetableOptions) ([]TimetableEntry, error) {
	params := TimetableParams{
		Options: options,
	}
	var result []TimetableEntry
	err := c.call("getTimetable", params, &result)
	return result, err
}

// GetStudents returns the list of students accessible to the logged-in user.
// It calls the app/data endpoint which requires JWT authentication.
func (c *Client) GetStudents() ([]StudentInfo, error) {
	if err := c.fetchAppData(); err != nil {
		return nil, fmt.Errorf("failed to fetch app data: %w", err)
	}
	return c.Students, nil
}

// doRESTGet performs an authenticated GET request to a REST API path.
// It sets JWT bearer auth, tenant-id, JSESSIONID cookie, and optionally
// the school-year-id header. Returns the raw response body bytes.
func (c *Client) doRESTGet(path string, withSchoolYearID bool) ([]byte, error) {
	reqURL, err := c.getEndpointURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.JWTToken)
	req.Header.Set("tenant-id", c.TenantID)
	if withSchoolYearID {
		req.Header.Set("x-webuntis-api-school-year-id", c.SchoolYearID)
	}
	if c.SessionID != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.SessionID})
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return body, nil
}

// GetTimetableREST fetches timetable entries from the REST API with resolved names.
// resourceType: "CLASS", "TEACHER", "SUBJECT", "ROOM", or "STUDENT"
// startDate/endDate: "YYYY-MM-DD" format
func (c *Client) GetTimetableREST(resourceType string, resourceID int, startDate, endDate string) (*RestTimetableResponse, error) {
	path := fmt.Sprintf("/WebUntis/api/rest/view/v1/timetable/entries?start=%s&end=%s&format=1&resourceType=%s&resources=%d&periodTypes=&timetableType=MY_TIMETABLE&layout=START_TIME",
		startDate, endDate, resourceType, resourceID)

	body, err := c.doRESTGet(path, false)
	if err != nil {
		return nil, err
	}

	var result RestTimetableResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode timetable response: %w", err)
	}
	return &result, nil
}

// GetAbsences fetches student absences from the REST API.
// startDate/endDate in YYYYMMDD format (e.g. "20260301").
func (c *Client) GetAbsences(studentID int, startDate, endDate string) (*AbsencesResponse, error) {
	path := fmt.Sprintf("/WebUntis/api/classreg/absences/students?startDate=%s&endDate=%s&studentId=%d&excuseStatusId=-1",
		startDate, endDate, studentID)

	body, err := c.doRESTGet(path, false)
	if err != nil {
		return nil, err
	}

	var result AbsencesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode absences response: %w", err)
	}
	return &result, nil
}

// GetHomework fetches homework assignments from the REST API.
// startDate/endDate in YYYYMMDD format.
func (c *Client) GetHomework(startDate, endDate string) (*HomeworkResponse, error) {
	path := fmt.Sprintf("/WebUntis/api/homeworks/lessons?startDate=%s&endDate=%s", startDate, endDate)

	body, err := c.doRESTGet(path, false)
	if err != nil {
		return nil, err
	}

	var result HomeworkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode homework response: %w", err)
	}
	return &result, nil
}

// GetMessages fetches inbox messages from the REST API.
func (c *Client) GetMessages() (*MessagesResponse, error) {
	body, err := c.doRESTGet("/WebUntis/api/rest/view/v1/messages", true)
	if err != nil {
		return nil, err
	}

	var result MessagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode messages response: %w", err)
	}
	return &result, nil
}

// --- Helper Structs ---

// rawRpcResponse is used internally to decode Result as RawMessage
type rawRpcResponse struct {
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JsonRpcError   `json:"error,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
}
