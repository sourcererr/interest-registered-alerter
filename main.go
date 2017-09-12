package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/nats-io/go-nats-streaming"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Interest register alerter"
	app.Usage = "Sends an alert to slack when someone registers an interest"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "nats-endpoint",
			Value:  "nats://0.0.0.0:4222",
			EnvVar: "NATS_ENDPOINT",
		},
		cli.StringFlag{
			Name:   "nats-cluster",
			Value:  "nats-cluster",
			EnvVar: "NATS_CLUSTER",
		},
		cli.StringFlag{
			Name:   "nats-subscription-id",
			Value:  "interest-registered-alerter",
			EnvVar: "NATS_SUBSCRIPTION_ID",
		},
		cli.StringFlag{
			Name:   "nats-topic",
			Value:  "interest.registered",
			EnvVar: "NATS_TOPIC",
		},
		cli.StringFlag{
			Name:   "slack-url, su",
			Value:  "https://slackurl.com",
			Usage:  "slack alert url",
			EnvVar: "SLACK_URL",
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))

	app.Action = startAlerter

	app.Run(os.Args)
}

func startAlerter(c *cli.Context) error {
	log.Println("[INFO]: Slack Alerter started......")

	//create a notification channel to shutdown
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	topic := c.String("nats-topic")
	clusterID := c.String("nats-cluster")
	natsSubscriptionId := c.String("nats-subscription-id")
	natsUrl := c.String("nats-endpoint")

	sc, err := stan.Connect(clusterID, natsSubscriptionId, stan.NatsURL(natsUrl))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, natsUrl)
	}

	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", natsUrl, clusterID, natsSubscriptionId)

	startOpt := stan.DeliverAllAvailable()

	alerter := NewSlackAlerter(c.String("nats-endpoint"))

	sub, err := sc.QueueSubscribe(topic, natsSubscriptionId+uuid.NewV4().String(), handler(alerter), startOpt, stan.DurableName("slack-alerter"))
	if err != nil {
		sc.Close()
		log.Fatal(err)
	}

	log.Printf("Listening on [%s], clientID=[%s], qgroup=[%s] durable=[%s]\n", topic, natsSubscriptionId, "slack-alerter", "slack-alerter")

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	go func() {
		for _ = range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			// Do not unsubscribe a durable on exit, except if asked to.
			sub.Unsubscribe()

			sc.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone

	log.Println("[INFO]: Slack alerter stopped.")

	return nil
}

func handler(alerter Alerter) func(msg *stan.Msg) {
	return func(msg *stan.Msg) {
		emailAddress := string(msg.Data)
		err := alerter.InterestRegistered(emailAddress)

		if err != nil {
			fmt.Printf("Error on consuming: %v", err)
		}
	}
}
