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

package faucet

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func TestControler(t *testing.T) {
	verifyTokenNoop := func(l *log.Entry, token string, remoteIP string) error {
		return nil
	}
	verifyTokenErr := func(l *log.Entry, token string, remoteIP string) error {
		return errors.New("Negative Testing")
	}
	insertNoop := func(*FundRequest) error {
		return nil
	}
	insertNoopErr := func(*FundRequest) error {
		return errors.New("Negative Testing")
	}

	testCases := []struct {
		body           string
		mockVerify     func(*log.Entry, string, string) error
		mockInsert     func(*FundRequest) error
		wantStatusCode int
	}{
		// 400 for invalid body
		{
			"{\"address\": \"\"}",
			verifyTokenNoop,
			insertNoop,
			400,
		},
		{
			"{\"address\": \"\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoop,
			400,
		},
		{
			"{\"address\": \"A09e79\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoop,
			400,
		},

		// 400 for invalid token
		{
			"{\"address\": \"0x0334995e2CFc53CF785C554839F6e845A3A09e79\",\"token\": \"test\"}",
			verifyTokenErr,
			insertNoop,
			http.StatusBadRequest,
		},

		// 500 for insertion failure
		{
			"{\"address\": \"0x0334995e2CFc53CF785C554839F6e845A3A09e79\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoopErr,
			http.StatusInternalServerError,
		},

		// 200
		{
			"{\"address\": \"zil1kkgy7ph9cfzalpg3ygwryk5prqd432jc48yz5k\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoop,
			http.StatusOK,
		},
		{
			"{\"address\": \"0334995e2CFc53CF785C554839F6e845A3A09e79\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoop,
			http.StatusOK,
		},
		{
			"{\"address\": \"0x0334995e2CFc53CF785C554839F6e845A3A09e79\",\"token\": \"test\"}",
			verifyTokenNoop,
			insertNoop,
			http.StatusOK,
		},
	}
	for i, testCase := range testCases {
		setupServer := func() *gin.Engine {
			r := gin.Default()
			r.POST("/test", Controller(
				testCase.mockVerify,
				testCase.mockInsert,
			))
			return r
		}
		testServer := httptest.NewServer(setupServer())
		defer testServer.Close()

		resp, err := http.Post(
			testServer.URL+"/test",
			"application/json",
			bytes.NewBufferString(testCase.body),
		)
		if err != nil {
			t.Errorf("%v", err)
		}
		if resp.StatusCode != testCase.wantStatusCode {
			t.Errorf("index:%d, %#v; want %#v", i, resp.StatusCode, testCase.wantStatusCode)
		}
	}
}
