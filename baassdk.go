package baassdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

type Playable interface {
	Play([3]string, string) string
}

type SDK struct {
	sessionID      string
	host           string
	feed           string
	subscriptionID string
	playable       Playable
}

func NewGame(sessionID string, host string, playable Playable) *SDK {
	sdk := SDK{
		sessionID: sessionID,
		host:      host,
		playable:  playable,
	}
	return &sdk
}

func (sdk *SDK) Play() (int, error) {

	// 1. Connect to the server and receive the feed and the connection token
	startResponse, err := sdk.startRequest()
	if err != nil {
		log.Fatalln(err)
	}

	show(startResponse)

	if startResponse.Error == true {
		return 0, errors.New(startResponse.Message)
	}

	if startResponse.SubscriptionID == "" {
		return 0, errors.New("No subscriptionID was returned from the server")
	}

	sdk.subscriptionID = startResponse.SubscriptionID

	// 2. Listen to the returned feed
	playedCard := "0"
	playResponse, err := sdk.playRequest(playedCard)
	if err != nil {
		log.Fatalln(err)
	}

	for true {
		show(playResponse)
		if playResponse.ErrorCode == 408 && playResponse.Message == "Request timeout" {
			playResponse, err = sdk.playRequest(playedCard)
			if err != nil {
				return 0, err
			}
		} else {
			// End of game
			// ... todo

			// Play a card
			// .. todo
			/*
				playedCard := sdk.playable.Play([3]string{"A", "B", "C"}, "D")
				playResponse, err = sdk.playRequest(playedCard)
				if err != nil {
					return 0, err
				}
			*/
		}
	}

	return 10, nil
}

func (sdk *SDK) startRequest() (genericResponse, error) {
	var gr genericResponse
	startRequest := sdk.host + "/start?sessionID=" + sdk.sessionID + "&type=TEST"

	httpResponse, err := http.Get(startRequest)
	if err != nil {
		return gr, err
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return gr, err
	}

	gr, err = ParseResponse(body)
	if err != nil {
		return gr, err
	}
	return gr, nil
}

func (sdk *SDK) playRequest(card string) (genericResponse, error) {
	var gr genericResponse
	playRequest := sdk.host + "/play?sessionID=" + sdk.sessionID + "&subscriptionID=" + sdk.subscriptionID + "&card=" + card

	httpResponse, err := http.Get(playRequest)
	if err != nil {
		return gr, err
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return gr, err
	}

	gr, err = ParseResponse(body)
	if err != nil {
		return gr, err
	}

	return gr, nil
}

type event struct {
	Data      interface{} `json:"Data"`
	Timestamp int32       `json:"Timestamp"`
}
type genericResponse struct {
	SubscriptionID string   `json:"SubscriptionID"`
	Feeds          []string `json:"Feeds"`
	Error          bool     `json:"Error"`
	ErrorCode      int      `json:"ErrorCode"`
	Message        string   `json:"Message"`
	Events         []event  `json:"Events"`
}

func ParseResponse(stream []byte) (genericResponse, error) {
	var response genericResponse
	err := json.Unmarshal(stream, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func show(i interface{}) {
	fmt.Printf("*** %+v\n", i)
}
