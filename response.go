package ussd

import (
	"fmt"
	"strings"

	"github.com/jansemmelink/utils2/errors"
)

//Response is compatible with ms-vservices-ussd-router
type Response struct {
	SessionID string       `json:"session_id" doc:"If not final, this will be populated and must be repeated in next Continue or Abort request."`
	Type      ResponseType `json:"type" doc:"Type of response will be RESPONSE|RELEASE|REDIRECT"`
	Message   string       `json:"message" doc:"Message is content for user on RESPONSE|RELEASE or new id for REDIRECT."`
}

func (res Response) Validate() error {
	if res.Message == "" {
		return errors.Errorf("missing text")
	}
	if _, ok := resTypeString[res.Type]; !ok {
		return errors.Errorf("invalid type:%d", res.Type)
	}
	return nil
}

type ResponseType int

const (
	ResponseTypeRedirect ResponseType = iota
	ResponseTypeResponse
	ResponseTypeRelease
)

var (
	resTypeString = map[ResponseType]string{
		ResponseTypeRedirect: "REDIRECT",
		ResponseTypeResponse: "RESPONSE",
		ResponseTypeRelease:  "RELEASE",
	}
	resTypeValue = map[string]ResponseType{}
)

func init() {
	for t, s := range resTypeString {
		resTypeValue[s] = t
	}
}

func (t ResponseType) String() string {
	if s, ok := resTypeString[t]; ok {
		return s
	}
	return fmt.Sprintf("unknown ussd.ResponseType(%d)", t)
}

func (t *ResponseType) Parse(s string) error {
	s = strings.ToUpper(s)
	if v, ok := resTypeValue[s]; ok {
		*t = v
		return nil
	}
	return errors.Errorf("unknown ussd.ResponseType(%d)", t)
}

func (t *ResponseType) UnmarshalJSON(v []byte) error {
	s := string(v)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return errors.Errorf("ResponseType(%s) expected quoted value", s)
	}
	s = s[1 : len(s)-1]
	if err := t.Parse(s); err != nil {
		return errors.Wrapf(err, "unable to unmarshal ResponseType(%s)", s)
	}
	return nil
}

func (t ResponseType) MarshalJSON() ([]byte, error) {
	if s, ok := resTypeString[t]; ok {
		return []byte("\"" + s + "\""), nil
	}
	return nil, errors.Errorf("unknown ussd.ResponseType(%d)", t)
}
