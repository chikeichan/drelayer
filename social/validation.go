package social

import (
	"ddrp-relayer/restmodels"
	"errors"
	"regexp"
)

var hashRegexp *regexp.Regexp

func ValidatePostParams(params *restmodels.PostParams) error {
	if params.Body == "" {
		return errors.New("body must be defined")
	}
	if params.Reference != "" && !hashRegexp.Match([]byte(params.Reference)) {
		return errors.New("reference must be a 32-byte hex-encoded hash")
	}
	if len(params.Title) > 255 {
		return errors.New("title must be less than 255 characters")
	}
	if len(params.Topic) > 255 {
		return errors.New("topic must be less than 255 characters")
	}
	if len(params.Tags) > 255 {
		return errors.New("must have fewer than 255 tags")
	}
	for _, tag := range params.Tags {
		if len(tag) == 0 {
			return errors.New("tags must not be empty")
		}
		if len(tag) > 255 {
			return errors.New("tags must be less than 255 characters")
		}
	}
	return nil
}

func ValidateConnectionParams(params *restmodels.ConnectionParams) error {
	if err := ValidateTLD(params.ConnecteeTld); err != nil {
		return errors.New("connectee tld must be defined")
	}
	if params.ConnecteeSubdomain != "" && len(params.ConnecteeSubdomain) > 15 {
		return errors.New("connectee subdomain must be less than 15 characters")
	}
	if params.Type != "FOLLOW" &&params.Type != "BLOCK" {
		return errors.New("type must be either FOLLOW or BLOCK")
	}
	return nil
}

func ValidateModerationParams(params *restmodels.ModerationParams) error {
	if !hashRegexp.Match([]byte(params.Reference)) {
		return errors.New("reference must be a 32-byte hex-encoded hash")
	}
	if params.Type != "LIKE" && params.Type != "PIN" {
		return errors.New("type must be either LIKE or PIN")
	}
	return nil
}

var validCharset = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0,
	0, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0, 0, 0, 0, 4,
	0, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 0, 0, 0, 0, 0,
}

func ValidateTLD(name string) error {
	if len(name) == 0 {
		return errors.New("name must have nonzero length")
	}

	if len(name) > 63 {
		return errors.New("name over maximum length")
	}

	for i := 0; i < len(name); i++ {
		ch := name[i]

		if int(ch) > len(validCharset) {
			return errors.New("invalid character")
		}

		charType := validCharset[ch]
		switch charType {
		case 0:
			return errors.New("invalid character")
		case 1:
			continue
		case 2:
			return errors.New("name cannot contain capital letters")
		case 3:
			continue
		case 4:
			if i == 0 {
				return errors.New("name cannot start with a hyphen")
			}
			if i == len(name)-1 {
				return errors.New("name cannot end with a hyphen")
			}
		}
	}

	return nil
}

func init() {
	hr, err := regexp.Compile("^[a-f0-9]{64}$")
	if err != nil {
		panic(err)
	}
	hashRegexp = hr
}
