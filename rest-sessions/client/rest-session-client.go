package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	ussd "github.com/jansemmelink/ussd2"
	"github.com/jansemmelink/utils2/errors"
	"github.com/jansemmelink/utils2/logger"
)

func New(addr string) ussd.Sessions {
	return httpSessions{addr: addr}
}

var log = logger.New().WithLevel(logger.LevelDebug)

//implements ussd.Sessions
type httpSessions struct {
	addr string
}

func (c httpSessions) New(id string, initData map[string]interface{}) (ussd.Session, error) {
	hs := httpSession{
		ID:   id,
		Data: initData,
	}
	buf := bytes.NewBuffer(nil)
	json.NewEncoder(buf).Encode(hs)
	httpReq, _ := http.NewRequest(
		http.MethodPost,
		c.addr+"/session/"+id,
		buf)
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access HTTP session")
	}
	switch httpRes.StatusCode {
	case http.StatusOK:
		if err := json.NewDecoder(httpRes.Body).Decode(&hs); err != nil {
			return nil, errors.Wrapf(err, "failed to decode HTTP session")
		}
		if hs.StartTime == nil || hs.LastTime == nil {
			t0 := time.Now()
			hs.StartTime = &t0 //just for sanity
			hs.LastTime = &t0  //just for sanity
		}
		return ussd.NewSession(
			c,
			id,
			*hs.StartTime,
			*hs.LastTime,
			initData,
		), nil
	default:
		return nil, errors.Errorf("failed to create session: %+v", httpRes.Status)
	}
}

func (c httpSessions) Get(id string) (ussd.Session, error) {
	httpReq, _ := http.NewRequest(
		http.MethodGet,
		c.addr+"/session/"+id,
		nil)
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to access HTTP session")
	}
	switch httpRes.StatusCode {
	case http.StatusOK:
		var hs httpSession
		if err := json.NewDecoder(httpRes.Body).Decode(&hs); err != nil {
			return nil, errors.Wrapf(err, "failed to decode HTTP session")
		}
		if hs.StartTime == nil || hs.LastTime == nil {
			t0 := time.Now()
			hs.StartTime = &t0 //just for sanity
			hs.LastTime = &t0  //just for sanity
		}
		return ussd.NewSession(
			c,
			id,
			*hs.StartTime,
			*hs.LastTime,
			hs.Data,
		), nil
	case http.StatusNotFound:
		return nil, errors.Errorf("session not found")
	default:
		return nil, errors.Errorf("failed to get session: %+v", httpRes.Status)
	}
}

func (c httpSessions) Del(id string) error {
	httpReq, _ := http.NewRequest(
		http.MethodDelete,
		c.addr+"/session/"+id,
		nil)
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return errors.Wrapf(err, "failed to access HTTP session")
	}
	switch httpRes.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return errors.Errorf("failed to delete session: %+v", httpRes.Status)
	}
}

func (c httpSessions) Sync(id string, set map[string]interface{}, del map[string]bool) error {
	hs := httpSession{
		ID:   id,
		Data: set,
	}
	for n := range del {
		log.Debugf("  del[%s]=(%T)%v", n, hs.Data[n])
		hs.Data[n] = nil
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(hs); err != nil {
		return errors.Wrapf(err, "encoder failed")
	}

	httpReq, _ := http.NewRequest(
		http.MethodPut,
		c.addr+"/session/"+id,
		buf)
	httpRes, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return errors.Wrapf(err, "failed to access HTTP session")
	}
	switch httpRes.StatusCode {
	case http.StatusOK:
		hr := httpSession{}
		if err := json.NewDecoder(httpRes.Body).Decode(&hr); err != nil {
			return errors.Wrapf(err, "failed to decode HTTP session")
		}
		//log.Debugf("Synced: %+v", hr)
		return nil
	default:
		return errors.Errorf("failed to sync session: %+v", httpRes.Status)
	}
}

type httpSession struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data,omitempty"`
	StartTime *time.Time             `json:"start_time,omitempty"`
	LastTime  *time.Time             `json:"last_time,omitempty"`
}
