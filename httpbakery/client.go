package httpbakery

import (
	"net/http"
	"fmt"
	"encoding/json"
)

// Do makes an http request to the given client.
// If the request fails with a discharge-required error,
// any required discharge macaroons will be acquired,
// and the request will be repeated with those attached.
//
// If c.Jar field is non-nil, the macaroons will be
// stored there and made available to subsequent requests.
func Do(c *http.Client, req *http.Request) (*http.Response, error) {
	httpResp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode != http.StatusProxyAuthRequired {
		return httpResp, nil
	}
	if httpResp.Header.Get("Content-Type") != "application/json" {
		return httpResp, nil
	}
	defer httpResp.Body.Close()

	var resp dischargeRequestedResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal discharge-required response: %v", err)
	}
	if resp.ErrorCode != codeDischargeRequired {
		return nil, fmt.Errorf("unexpected error code: %q", resp.ErrorCode)
	}
	if resp.Macaroon == nil {
		return nil, fmt.Errorf("no macaroon found in response")
	}
	return nil, fmt.Errorf("we *will* discharge the macaroon, id %q, some day", resp.Macaroon.Id())
}
