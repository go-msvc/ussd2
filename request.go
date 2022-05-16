package ussd

import (
	"fmt"
	"strings"

	"bitbucket.org/vservices/utils/errors"
)

//Request is compatible with ms-vservices-ussd-router and -menu
//but we do without prompt request, which is how menu is used from router
//
// type Request struct {
// 	RequestHeader                        //embedded
// 	Data          map[string]interface{} `json:"data,omitempty" doc:"additional data are JSON encoded inside \"data\":{...} object"`
// }
// type RequestHeader struct {
// 	Type          MessageType      `json:"type" doc:"Indicates REQUEST (initial request), RESPONSE (for user input) or RELEASE (caller aborts the session)."`
// 	Source        string           `json:"source" doc:"Name of interface/service where the USSD session originated from."`
// 	MSISDN        string           `json:"msisdn" doc:"Subscriber MSISDN in international format, e.g. \"27821234567\""`
// 	Message       string           `json:"message" doc:"USSD string dialed by user or input provided by the user after being prompted."`
// 	SessionID     string           `json:"session_id,omitempty" doc:"Session ID to be echoed in the ussd.Response."`
// 	PromptService *ServiceProvider `json:"prompt_service,omitempty" doc:"Optional: Service to call to prompt the user."`
// }

type Request struct {
	Type      RequestType            `json:"type"`
	Msisdn    string                 `json:"msisdn"`
	Message   string                 `json:"message"`
	SessionID string                 `json:"session_id,omitempty" doc:"Session ID to be echoed in the ussd.Response."`
	Data      map[string]interface{} `json:"data" doc:"Data to store in the session"`
}

func (req Request) Validate() error {
	if req.Msisdn == "" {
		return errors.Errorf("missing msisdn")
	}
	if req.Message == "" {
		return errors.Errorf("missing message")
	}
	if _, ok := reqTypeString[req.Type]; !ok {
		return errors.Errorf("invalid type:%d", req.Type)
	}
	return nil
}

type RequestType int

const (
	RequestTypeRequest RequestType = iota
	RequestTypeResponse
	RequestTypeRelease
)

var (
	reqTypeString = map[RequestType]string{
		RequestTypeRequest:  "REQUEST",  //user request to begin a new session
		RequestTypeResponse: "RESPONSE", //user input that continues a session
		RequestTypeRelease:  "RELEASE",  //user request to abort the session
	}
	reqTypeValue = map[string]RequestType{}
)

func init() {
	for t, s := range reqTypeString {
		reqTypeValue[s] = t
	}
}

func (t RequestType) String() string {
	if s, ok := reqTypeString[t]; ok {
		return s
	}
	return fmt.Sprintf("unknown ussd.RequestType(%d)", t)
}

func (t *RequestType) Parse(s string) error {
	s = strings.ToUpper(s)
	if v, ok := reqTypeValue[s]; ok {
		*t = v
		return nil
	}
	return errors.Errorf("unknown ussd.RequestType(%s)", s)
}

func (t *RequestType) UnmarshalJSON(v []byte) error {
	s := string(v)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return errors.Errorf("RequestType(%s) expected quoted value", s)
	}
	s = s[1 : len(s)-1]
	if err := t.Parse(s); err != nil {
		return errors.Wrapf(err, "unable to unmarshal RequestType(%s)", s)
	}
	return nil
}

func (t RequestType) MarshalJSON() ([]byte, error) {
	if s, ok := reqTypeString[t]; ok {
		return []byte("\"" + s + "\""), nil
	}
	return nil, errors.Errorf("unknown ussd.RequestType(%d)", t)
}
