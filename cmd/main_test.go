package main

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
		err := validateEnvVars(
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
