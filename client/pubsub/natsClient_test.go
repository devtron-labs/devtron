package pubsub

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	util1 "github.com/devtron-labs/devtron/util"
	"github.com/nats-io/nats.go"
	"testing"
	"time"
)

func TestNewPubSubClient(t *testing.T) {

	t.SkipNow()
	const payload = "stop-msg"
	var globalVal = 0

	t.Run("subscriber", func(t *testing.T) {
		sugaredLogger, _ := util.NewSugardLogger()
		pubSubClient, _ := NewPubSubClient(sugaredLogger)
		globalVar := false

		_ = util1.AddStream(pubSubClient.JetStrCtxt, util1.CI_RUNNER_STREAM)
		subs, err := pubSubClient.JetStrCtxt.QueueSubscribe("CI-COMPLETE", "CI-COMPLETE_GROUP-1", func(msg *nats.Msg) {
			println("msg received")
			defer msg.Ack()
			println(string(msg.Data))
			if string(msg.Data) == payload {
				globalVar = true
			}
		}, nats.Durable(util1.WORKFLOW_STATUS_UPDATE_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util1.CI_RUNNER_STREAM))
		if err != nil {
			fmt.Println("error is ", err)
			return
		}
		for true {
			if globalVar {
				break
			}
			fmt.Println("looping & checking subs status: ", subs.IsValid())
			time.Sleep(5 * time.Second)
		}
	})

	t.Run("pullSubscriber", func(t *testing.T) {
		sugaredLogger, _ := util.NewSugardLogger()
		pubSubClient, _ := NewPubSubClient(sugaredLogger)

		_ = util1.AddStream(pubSubClient.JetStrCtxt, util1.ORCHESTRATOR_STREAM)
		subs, err := pubSubClient.JetStrCtxt.PullSubscribe("CD.TRIGGER", util1.WORKFLOW_STATUS_UPDATE_DURABLE, nats.BindStream(util1.ORCHESTRATOR_STREAM))
		if err != nil {
			fmt.Println("error occurred while subscribing pull reason: ", err)
			return
		}
		for subs.IsValid() {
			msgs, err := subs.Fetch(10)
			if err != nil && err == nats.ErrTimeout {
				fmt.Println(" timeout occurred but we have to try again")
				time.Sleep(5 * time.Second)
				continue
			} else if err != nil {
				fmt.Println("error occurred while extracting msg", err)
				return
			}
			for _, nxtMsg := range msgs {
				fmt.Println("Received a JetStream message: ", string(nxtMsg.Data))
				if string(nxtMsg.Data) == payload {
					return
				}
				defer nxtMsg.Ack()
			}
		}
	})

	t.Run("publisher", func(t *testing.T) {
		sugaredLogger, _ := util.NewSugardLogger()
		pubSubClient, _ := NewPubSubClient(sugaredLogger)
		//topic := "CD.TRIGGER"
		//topic := "CI-COMPLETE"
		topic := "hello.world"
		//streamName := util1.ORCHESTRATOR_STREAM
		//streamName := util1.CI_RUNNER_STREAM
		streamName := "New_Stream_2"

		globalVal++
		helloWorld := "Hello World " + string(rune(globalVal))
		WriteNatsEvent(pubSubClient, topic, helloWorld, streamName)
	})

	t.Run("stopPublisher", func(t *testing.T) {
		sugaredLogger, _ := util.NewSugardLogger()
		pubSubClient, _ := NewPubSubClient(sugaredLogger)
		topic := "CD.TRIGGER" // for pull subs
		//topic := "CI-COMPLETE"
		streamName := util1.ORCHESTRATOR_STREAM
		//streamName := util1.CI_RUNNER_STREAM

		WriteNatsEvent(pubSubClient, topic, payload, streamName)
	})
}

func WriteNatsEvent(psc *PubSubClient, topic string, payload string, streamName string) {
	_ = util1.AddStream(psc.JetStrCtxt, streamName)
	//Generate random string for passing as Header Id in message
	randString := "MsgHeaderId-" + util1.Generate(10)
	_, err := psc.JetStrCtxt.Publish(topic, []byte(payload), nats.MsgId(randString))
	if err != nil {
		fmt.Println("error occurred while publishing event reason: ", err)
	}
}
