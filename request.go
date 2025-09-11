package runcomfyserverless

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURLRaw   = "https://api.runcomfy.net"
	authHeader   = "Authorization"
	ctHeader     = "Content-Type"
	acceptHeader = "Accept"
	ctJSON       = "application/json; charset=UTF-8"
)

var (
	baseURL *url.URL
)

func init() {
	var err error
	if baseURL, err = url.Parse(baseURLRaw); err != nil {
		panic(err)
	}
}

func (d *Deployment) request(ctx context.Context, method, endpointURI string, reqPayload, respPayload any) (err error) {
	// Prepare body if necessary
	var bodyReader io.Reader
	if reqPayload != nil {
		var bodyData []byte
		if bodyData, err = json.Marshal(reqPayload); err != nil {
			err = fmt.Errorf("failed to marshal body data to JSON: %w", err)
			return
		}
		bodyReader = bytes.NewBuffer(bodyData)
	}
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, d.baseURL.JoinPath(endpointURI).String(), bodyReader)
	if err != nil {
		err = fmt.Errorf("failed to create HTTP request: %w", err)
		return
	}
	req.Header.Set(authHeader, d.auth)
	if reqPayload != nil {
		req.Header.Set(ctHeader, ctJSON)
	}
	if respPayload != nil {
		req.Header.Add(acceptHeader, ctJSON)
	}
	// Execute request
	resp, err := d.http.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to execute HTTP request: %w", err)
		return
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		if respPayload == nil {
			return
		}
		decoder := json.NewDecoder(resp.Body)
		decoder.DisallowUnknownFields()
		if err = decoder.Decode(respPayload); err != nil {
			err = fmt.Errorf("failed to decode response payload: %w", err)
			return
		}
		return
	case http.StatusForbidden: // may be some HTTP codes used by the API should be added too
		var he HTTPError
		if err = json.NewDecoder(resp.Body).Decode(&he); err != nil {
			err = fmt.Errorf("failed to decode error payload after HTTP status code %s: %w", resp.Status, err)
		} else {
			he.Code = resp.StatusCode
			err = he
		}
		return
	default:
		err = fmt.Errorf("unexpected HTTP status code: %s", resp.Status)
		return
	}
}

// APIError represents an error received thru the API (request was successfull but API returns an inner error)
type APIError struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_message"`
}

func (ae APIError) Error() string {
	return fmt.Sprintf("%d: %s", ae.Code, ae.Message)
}

// HTTPEror represents an error on the HTTP level
type HTTPError struct {
	Code    int    `json:"http_code"` // not received as payload, fed from header
	Message string `json:"message"`
}

func (he HTTPError) Error() string {
	return fmt.Sprintf("%d %s: %s", he.Code, http.StatusText(he.Code), he.Message)
}
