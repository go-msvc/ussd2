package ussd

import (
	"context"

	"bitbucket.org/vservices/utils/errors"
	"bitbucket.org/vservices/utils/ms"
	"github.com/google/uuid"
)

func NewService(initItem Item) ms.Service {
	if initItem == nil {
		panic(errors.Errorf("cannot create NewService(nil)"))
	}
	svc := service{
		initItem: initItem,
	}
	msService := ms.NewService().
		Handle("ussd", svc.handleUSSD).
		Handle("start", svc.handleStart).
		Handle("continue", svc.handleContinue).
		Handle("abort", svc.handleAbort)
	return msService
}

type service struct {
	initItem Item
}

type StartRequest struct {
	ID           string                 `json:"id" doc:"Unique session ID, typically made up of source and user e.g. 'sigtran:27821234567'. It could be UUID as well, but using same string always for a user ensures the user can only have one session at any time and starting new session will delete any old session."`
	Data         map[string]interface{} `json:"data" doc:"Initial data values to set in the new session"`
	ItemID       string                 `json:"item_id" doc:"ID of USSD item to start the session. It must be a server side item to return next item, typically a ussd router to process the dialed USSD string."`
	Input        string                 `json:"input" doc:"User input is initially dialed USSD string for start, and prompt/menu input for continuation."`
	ResponderID  string                 `json:"responder_id" doc:"Identifies the responder to use"`
	ResponderKey string                 `json:"responder_key" doc:"Key given to the responder to send to the correct user"`
}

type StartResponse struct {
	ID      string       `json:"id" doc:"Session ID to repeat in subsequent requests for this session."`
	Type    ResponseType `json:"type" doc:"Type of response is either RESPONSE(=prompt/menu)|RELEASE(=final)|REDIRECT"`
	Message string       `json:"message" doc:"Content for RESPONSE|RELEASE, or USSD string for REDIRECT"`
}

func (svc service) handleStart(ctx context.Context, req StartRequest) (StartResponse, error) {
	//session ID on this provider will only be specified if consumer is continuing
	//on an existing session
	// sid := m.Header.Provider.Sid //"ussd:" + m.Request.Msisdn
	// ...sid.... ussd start open must expect "" while continue/abort expects an id
	// or ussd still create/get session because other services does not need sesison...
	// so ignore sid in consumer and provider? I think so...

	if req.ID == "" {
		req.ID = uuid.New().String()
	}
	s, err := sessions.New(req.ID, req.Data)
	if err != nil {
		return StartResponse{}, errors.Wrapf(err, "failed to create session(%s)", req.ID)
	}
	s.Set("start_request", req)
	//ctx = context.WithValue(ctx, CtxSession{}, s)

	//process until reached next user item
	return StartResponse{
		Message: "NYI",
		Type:    ResponseTypeRelease, //todo: must be a Prompt or Final
	}, nil
}

type ContinueRequest StartRequest

type ContinueResponse StartResponse

func (svc service) handleContinue(ctx context.Context, req ContinueRequest) (ContinueResponse, error) {
	return ContinueResponse{}, errors.Errorf("NYI")
}

type AbortRequest struct {
	ID string `json:"id" doc:"Unique session ID also used in start/continue Request."`
}

type AbortResponse struct {
}

func (svc service) handleAbort(ctx context.Context, req AbortRequest) (AbortResponse, error) {
	return AbortResponse{}, errors.Errorf("NYI")
}

func (s service) handleUSSD(ctx context.Context, req Request) (Response, error) {
	switch req.Type {
	case RequestTypeRequest:
		res, err := s.handleStart(
			ctx,
			StartRequest{})
		if err != nil {
			return Response{
				Type:    ResponseTypeRelease,
				Message: err.Error(),
			}, nil
		} else {
			return Response{
				Type:    res.Type,
				Message: res.Message,
			}, nil
		}
	case RequestTypeResponse:
		res, err := s.handleContinue(
			ctx,
			ContinueRequest{})
		if err != nil {
			return Response{
				Message: res.Message,
			}, nil
		} else {
			return Response{}, nil
		}
	case RequestTypeRelease:
		_, err := s.handleAbort(
			ctx,
			AbortRequest{})
		if err != nil {
			return Response{}, nil
		} else {
			return Response{}, nil
		}
	default:
		return Response{
			Type:    ResponseTypeRelease,
			Message: "Unexpected request type",
		}, nil
	}
} //handleUSSD()
