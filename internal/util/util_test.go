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

package util

import (
	"errors"
	"testing"
)

func TestValidateEnvVars(t *testing.T) {
	testCases := []struct {
		envType    string
		secret     string
		privateKey string
		want       error
	}{
		{
			"",
			"secretX",
			"privateKeyX",
			errors.New("ENV_TYPE is invalid"),
		},
		{
			"dev",
			"",
			"privateKeyX",
			errors.New("RECAPTCHA_SECRET is invalid"),
		},
		{
			"dev",
			"secretX",
			"",
			errors.New("PRIVATE_KEY is invalid"),
		},
		{
			"dev",
			"secretX",
			"privateKeyX",
			nil,
		},
	}

	for _, testCase := range testCases {
		err := ValidateEnvVars(
			testCase.envType,
			testCase.secret,
			testCase.privateKey,
		)
		if testCase.want == nil {
			if err != testCase.want {
				t.Errorf("%#v; want %#v", err, testCase.want)
			}
		} else {
			if err.Error() != testCase.want.Error() {
				t.Errorf("%#v; want %#v", err, testCase.want)
			}
		}
	}
}
