//  Copyright (C) 2021 Zilliqa
//
//  This file is part of faucet-service.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package recaptcha

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// https://developers.google.com/recaptcha/docs/verify#api_request

type Response struct {
	Success bool `json:"success"`

	// timestamp of the challenge load (ISO format yyyy-MM-dd'T'HH:mm:ssZZ)
	ChallengeTs string `json:"challenge_ts"`

	// the hostname of the site where the reCAPTCHA was solved
	Hostname string `json:"hostname"`

	// Error code reference
	ErrorCodes []string `json:"error-codes"`
	//
	// missing-input-secret 	The secret parameter is missing.
	// invalid-input-secret 	The secret parameter is invalid or malformed.
	// missing-input-response 	The response parameter is missing.
	// invalid-input-response 	The response parameter is invalid or malformed.
	// bad-request 	The request is invalid or malformed.
	// timeout-or-duplicate 	The response is no longer valid: either is too old or has been used previously.
	//
	// https://developers.google.com/recaptcha/docs/verify#error_code_reference
}

// Verifier returns verify function
func Verifier(secret string) func(*log.Entry, string, string) error {
	return func(logger *log.Entry, token string, remoteIP string) error {
		url := "https://www.google.com/recaptcha/api/siteverify" +
			"?secret=" + secret +
			"&response=" + token +
			"&remoteip=" + remoteIP

		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		logger.Info("recaptcha status code: ", resp.StatusCode)

		body, _ := ioutil.ReadAll(resp.Body)
		res := Response{}
		err = json.Unmarshal(body, &res)
		if err != nil {
			return err
		}
		logger.Info("recaptcha body: ", res)

		if !res.Success {
			errorCodes := strings.Join(res.ErrorCodes, ",")
			return errors.New(errorCodes)
		}

		return nil
	}
}
