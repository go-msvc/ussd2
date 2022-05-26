package console

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	ussd "github.com/jansemmelink/ussd2"
	"github.com/jansemmelink/utils2/errors"
	"github.com/jansemmelink/utils2/logger"
	"github.com/jansemmelink/utils2/ms"
)

var log = logger.New()

type server struct {
	config  Config
	service ms.Service
}

func (s server) Serve(svc ms.Service) error {
	s.service = svc

	//todo: redirect to external file if necessary or to channel to show as part of console output

	//assuming this is a ussd service, it must have the following operations:
	startOper, ok := s.service.GetOper("start")
	if !ok {
		return errors.Errorf("service does not have a start operation")
	}
	continueOper, ok := s.service.GetOper("continue")
	if !ok {
		return errors.Errorf("service does not have a continue operation")
	}
	// abortOper, ok := s.service.GetOper("abort")
	// if !ok {
	// 	return errors.Errorf("service does not have an abort operation")
	// }

	//create a user input channel used for all console input
	//so we can constantly read the terminal
	userInputChan := make(chan string)
	go func(userInputChan chan string) {
		reader := bufio.NewReader(os.Stdin)
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "%+v", errors.Wrapf(err,
					"Error reading from stdin"))
				userInputChan <- "exit"
				return
			} // if err
			input = strings.Replace(input, "\n", "", -1)
			userInputChan <- input
		}
	}(userInputChan)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT) //<ctrl><C>
	go func() {
		<-signalChannel
		userInputChan <- "exit"
	}()

	sessionID := "console:" + s.config.Msisdn
	exited := false
	for !exited {
		startTime := time.Now()
		lastTime := startTime
		resTime := startTime
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "==========[ %-10.10s ]==========     (total=%7.3fs                  now=%s)\n",
			"START",
			float64((resTime.Sub(startTime))/time.Millisecond)/1000.0,
			resTime.Format("15:04:05.000"),
		)

		//start a new ussd session
		ctx := context.Background()
		startRequest := ussd.StartRequest{
			SessionID: sessionID,
			Data: map[string]interface{}{
				"msisdn": s.config.Msisdn,
			},
		}
		results := startOper.FncValue().Call([]reflect.Value{
			reflect.ValueOf(ctx),
			reflect.ValueOf(startRequest),
		})

		for {
			//get err from last result
			if err, ok := results[len(results)-1].Interface().(error); ok && err != nil {
				log.Errorf("service failed: %+v", err)
				break
			}
			//expect a ussd response
			if len(results) != 2 {
				log.Errorf("did not get service response")
				break
			}
			response, ok := results[0].Interface().(*ussd.Response)
			if !ok {
				log.Errorf("got %T instead of *ussd.Response", results[0].Interface())
				break
			}

			fmt.Printf("%s\n", response.Message)
			resTime = time.Now()
			fmt.Fprintf(os.Stdout, "----------[ %-10.10s ]----------     (total=%7.3fs content=%7.3fs now=%s)\n",
				response.Type,
				float64((resTime.Sub(startTime))/time.Millisecond)/1000.0,
				float64((resTime.Sub(lastTime))/time.Millisecond)/1000.0,
				resTime.Format("15:04:05.000"),
			)
			lastTime = resTime
			if response.Type != ussd.ResponseTypeResponse {
				break
			}

			//prompt for input
			fmt.Printf("reply > ")
			var input string
			for input == "" {
				input = <-userInputChan
			}
			if input == "exit" {
				exited = true
				break
			}
			resTime = time.Now()
			fmt.Fprintf(os.Stdout, "----------[ %-10.10s ]----------     (total=%7.3fs user   =%7.3fs now=%s)\n",
				"USER",
				float64((resTime.Sub(startTime))/time.Millisecond)/1000.0,
				float64((resTime.Sub(lastTime))/time.Millisecond)/1000.0,
				resTime.Format("15:04:05.000"),
			)
			lastTime = resTime

			//send input to service
			//reqTime = time.Now()
			ctx := context.Background()
			results = continueOper.FncValue().Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(ussd.ContinueRequest{
					SessionID: sessionID,
					Input:     input,
				}),
			})
		} //for main loop

		if exited {
			break
		}
		fmt.Printf("\n\nPress <enter> to start again or <ctrl><C> to exit ...\n")
		input := <-userInputChan
		if input == "exit" {
			break
		}
	} //for each USSD session

	fmt.Printf("\n\n")
	close(userInputChan)
	return nil
}
