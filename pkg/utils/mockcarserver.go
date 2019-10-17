/*
Copyright 2018-2019 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/Ridecell/ridecell-operator/pkg/errors"
	"net/http"
	"os"
	"time"
)

func GetHttpClient() http.Client {
	return http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func httpRequest(method string, resourcePath string, data *bytes.Buffer) (*http.Response, error) {
	URI := os.Getenv("MOCKCARSERVER_URI")
	AUTH := os.Getenv("MOCKCARSERVER_AUTH")
	AUTH_CLIENT := "ridecell-operator"
	client := GetHttpClient()
	request, err := func() (*http.Request, error) {
		if data != nil {
			return http.NewRequest(method, URI+resourcePath, data)
		} else {
			return http.NewRequest(method, URI+resourcePath, nil)
		}
	}()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create request.")
	}
	request.Header.Set("API-KEY", AUTH)
	request.Header.Set("API-CLIENT", AUTH_CLIENT)
	request.Header.Set("Content-type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return nil, errors.Wrapf(err, "Something bad happened while connecting to Mock car server.")
	}
	defer resp.Body.Close()
	return resp, nil
}

func checkResponseStatus(statusCode int) error {
	if statusCode == 201 || statusCode == 200 {
		return nil
	} else if statusCode == 404 {
		return errors.New("Resource not found")
	} else if statusCode == 401 {
		return errors.New("Request not authorized")
	} else if statusCode == 400 {
		return errors.New("Bad request to server")
	}
	return errors.Errorf("Unknown Server Response Status code: %d", statusCode)
}

// Get the mock tenant
// GET request
// query param: name
// response code: 200 success (present), 404 (not found), 401 (invalid auth)
func GetMockTenant(tenantName string) (bool, error) {
	response, err := httpRequest("GET", "/common/tenant?name="+tenantName, nil)
	if err != nil {
		return false, errors.Wrapf(err, "mockcarserver error")
	}
	err = checkResponseStatus(response.StatusCode)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to get mock car server tenant")
	}
	return true, nil
}

// Create the mock tenant
// POST request
// param: name, callbackUrl, tenantHardwareType, apiKey, secretKey, apiToken, pushApiKey, pushSecretKey, pushToken
// response code: 201 created, 400 (bad params), 401 (invalid auth)
func CreateOrUpdateMockTenant(postData map[string]string) (bool, error) {
	jsonData, err := json.Marshal(postData)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to convert data into json format")
	}

	response, err := httpRequest("POST", "/common/tenant", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, errors.Wrapf(err, "mockcarserver error")
	}
	err = checkResponseStatus(response.StatusCode)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to create/update mock car server tenant")
	}
	return true, nil
}

// Delete the mock tenant
// DELETE request
// query param: name
// response code: 200 success, 400 (bad params), 401 (invalid auth)
func DeleteMockTenant(tenantName string) (bool, error) {
	response, err := httpRequest("DELETE", "/common/tenant?name="+tenantName, nil)
	if err != nil {
		return false, errors.Wrapf(err, "mockcarserver error")
	}
	err = checkResponseStatus(response.StatusCode)
	if err != nil {
		return false, errors.Wrapf(err, "Unable to delete tenant on mock car server")
	}
	return true, nil
}
