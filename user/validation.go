package user

import (
	"ddrp-relayer/restmodels"
	"errors"
	"regexp"
)

var emailRegexp *regexp.Regexp

func ValidateUserParams(params *restmodels.CreateUserParams) error {
	if params.Email != "" && !emailRegexp.Match([]byte(params.Email)) {
		return errors.New("e-mail is invalid")
	}
	if len(params.Password) < 8 {
		return errors.New("password must be more than 8 characters")
	}
	if params.Username == "" {
		return errors.New("username must be defined")
	}
	if len(params.Username) > 15 {
		return errors.New("username must be less than 15 characters")
	}
	return nil
}

func init() {
	e, err := regexp.Compile("^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$")
	if err != nil {
		panic(err)
	}
	emailRegexp = e
}
