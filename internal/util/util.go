package util

import "errors"

func ValidateEnvVars(envType, secret, privKey string) error {
	isEnvTypeValid := envType == "dev" || envType == "staging" || envType == "prod"

	if !isEnvTypeValid {
		return errors.New("ENV_TYPE is invalid")
	}

	if secret == "" {
		return errors.New("RECAPTCHA_SECRET is invalid")
	}

	if privKey == "" {
		return errors.New("PRIVATE_KEY is invalid")
	}
	return nil
}
