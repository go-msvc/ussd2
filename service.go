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
	SessionID string                 `json:"session_id" doc:"Unique session ID, typically made up of source and user e.g. 'sigtran:27821234567'. It could be UUID as well, but using same string always for a user ensures the user can only have one session at any time and starting new session will delete any old session."`
	Data      map[string]interface{} `json:"data" doc:"Initial data values to set in the new session, e.g. msisdn, email or account_id of the user that requested the service, dialled ussd string if there is a router step, etc..."`
}

func (req StartRequest) Validate() error {
	if req.SessionID == "" {
		return errors.Errorf("missing session_id")
	}
	return nil
}

func (svc service) handleStart(ctx context.Context, req StartRequest) (*Response, error) {
	//set started true so that any attempts to define static items at runtime will panic
	started = true
	log.Debugf("START: %+v", req)

	//session ID on this provider will only be specified if consumer is continuing
	//on an existing session
	// sid := m.Header.Provider.Sid //"ussd:" + m.Request.Msisdn
	// ...sid.... ussd start open must expect "" while continue/abort expects an id
	// or ussd still create/get session because other services does not need sesison...
	// so ignore sid in consumer and provider? I think so...
	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}
	sessions.Del(req.SessionID)
	s, err := sessions.New(req.SessionID, req.Data)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create session(%s)", req.SessionID)
	}
	return svc.Run(s, svc.initItem, "")
}

type ContinueRequest struct {
	SessionID string `json:"session_id" doc:"Copied from previous response"`
	Input     string `json:"input" doc:"User's answer to the last prompt"`
}

func (req ContinueRequest) Validate() error {
	if req.SessionID == "" {
		return errors.Errorf("missing session_id")
	}
	return nil
}

func (svc service) handleContinue(ctx context.Context, req ContinueRequest) (*Response, error) {
	log.Debugf("CONTINUE: %+v", req)
	//get existing session
	s, err := sessions.Get(req.SessionID)
	if err != nil {
		return nil, errors.Wrapf(err, "not an existing session(%s)", req.SessionID)
	}
	if s == nil {
		return nil, errors.Errorf("not an existing session(%s)", req.SessionID)
	}

	log.Debugf("Got existing session(%s)", s.ID())
	var item Item
	if id, ok := s.Get("item_id").(string); !ok {
		return nil, errors.Errorf("item_id not defined for existing session(%s)", req.SessionID)
	} else {
		var ok bool
		if item, ok = ItemByID(id, s); !ok {
			return nil, errors.Errorf("session(%s).item_id(%s) not found to continue", req.SessionID, id)
		}
	}

	return svc.Run(s, item, req.Input)
}

func (svc service) Run(s Session, item Item, input string) (*Response, error) {
	//rebuild list of next items from session data
	nextItems := []Item{}
	if nextItemIds, ok := s.Get("next_item_ids").([]string); ok {
		for _, id := range nextItemIds {
			if nextItem, ok := itemByID[id]; !ok {
				return nil, errors.Errorf("unknown next item_id(%s)", id)
			} else {
				nextItems = append(nextItems, nextItem)
			}
		}
	}
	log.Debugf("Run: item(%s) and %d next items", item.ID(), len(nextItems))
	for _, i := range nextItems {
		log.Debugf("   item(%s): %T", i.ID(), i)
	}

	defer func() {
		if s != nil {
			if item != nil {
				s.Set("item_id", item.ID())
				nextItemIDs := []string{}
				for _, i := range nextItems {
					nextItemIDs = append(nextItemIDs, i.ID())
				}
				s.Set("next_item_ids", nextItemIDs)
				s.Sync()
			} else {
				s.Del("item_id")
				s.Del("next_item_ids")
			}
		} //if session still exists
	}()

	ctx := context.WithValue(context.Background(), CtxSession{}, s)

	//see if current item is a user item waiting for input
	//(even when user input is "")
	if usrPromptItem, ok := item.(ItemUsrPrompt); ok {
		log.Debugf("Item(%s):%T is processing input(%s)...", item.ID(), item, input)
		moreNextItems, err := usrPromptItem.Process(ctx, input)
		if err != nil {
			return nil, errors.Wrapf(err, "prompt(%s).Process(%s) failed", item.ID(), input)
		}
		if len(moreNextItems) > 0 {
			nextItems = append(moreNextItems, nextItems...)
		}
		//step to next
		//note: prompt/menu that want's to repeat the same prompt/menu, must return
		//themselves in the moreNextItems list
		if len(nextItems) < 1 {
			return nil, errors.Errorf("no next item after processing input")
		}
		item = nextItems[0]
		nextItems = nextItems[1:]
	}

	for item != nil {
		log.Infof("LOOP Item(%s):%T ...", item.ID(), item)
		if svcItem, ok := item.(ItemSvc); ok {
			log.Debugf("Service item...")
			moreNextItems, err := svcItem.Exec(ctx)
			if err != nil {
				return nil, errors.Wrapf(err, "service item(%s:%T).Exec() failed", item, item)
			}
			if len(moreNextItems) > 0 {
				nextItems = append(moreNextItems, nextItems...)
			}
			if len(nextItems) == 0 {
				item = nil
			} else {
				item = nextItems[0]
				nextItems = nextItems[1:]
			}
			continue
		} //if ItemSvc

		if usrItem, ok := item.(ItemUsr); ok {
			res := Response{
				SessionID: s.ID(),
				Type:      ResponseTypeResponse,
				Message:   usrItem.Render(ctx),
			}
			if _, ok := item.(ItemUsrPrompt); !ok {
				res.SessionID = ""
				res.Type = ResponseTypeRelease
				sessions.Del(s.ID())
				s = nil
			}
			log.Debugf("User item(%s):%T -> Response: %+v", item.ID(), item, res)
			return &res, nil
		} //if ItemUsr
		return nil, errors.Errorf("item(%s) unknown type %T", item.ID(), item)
	} //for loop
	return nil, errors.Errorf("not expected to get here")
} //service.Run()

type AbortRequest struct {
	SessionID string `json:"session_id" doc:"Copied from previous response"`
}

func (req *AbortRequest) Validate() error {
	if req.SessionID == "" {
		return errors.Errorf("missing session_id")
	}
	return nil
}

func (svc service) handleAbort(ctx context.Context, req AbortRequest) error {
	log.Debugf("ABORT: %+v", req)
	//get existing session
	s, err := sessions.Get(req.SessionID)
	if err != nil {
		return errors.Wrapf(err, "not an existing session(%s)", req.SessionID)
	}
	if s == nil {
		return errors.Errorf("not an existing session(%s)", req.SessionID)
	}
	sessions.Del(req.SessionID)
	return nil
}

func (s service) handleUSSD(ctx context.Context, req Request) (*Response, error) {
	switch req.Type {
	case RequestTypeRequest:
		startRequest := StartRequest{
			SessionID: "ussd:" + req.Msisdn,
			Data: map[string]interface{}{
				"msisdn": req.Msisdn,
				"ussd":   req.Message,
			},
		}
		res, err := s.handleStart(ctx, startRequest)
		if err != nil {
			return &Response{
				Type:    ResponseTypeRelease,
				Message: err.Error(),
			}, nil
		}
		return res, nil

	case RequestTypeResponse:
		contRequest := ContinueRequest{
			SessionID: req.SessionID,
			Input:     req.Message,
		}
		res, err := s.handleContinue(ctx, contRequest)
		if err != nil {
			return &Response{
				Type:    ResponseTypeRelease,
				Message: err.Error(),
			}, nil
		}
		return res, nil

	case RequestTypeRelease:
		err := s.handleAbort(
			ctx,
			AbortRequest{})
		if err != nil {
			log.Errorf("abort failed: %+v", err)
		}
		return &Response{
			Type:    ResponseTypeRelease,
			Message: "", //nothing will be displayed to the user
		}, nil

	default:
		return nil, errors.Errorf("unexpected type %d", req.Type)
	}
} //handleUSSD()
