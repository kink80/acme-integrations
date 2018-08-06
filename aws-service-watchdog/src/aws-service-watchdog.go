package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkArgs "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/integration"
)

type serviceWatchdog struct {
	name string
	tag  string
	port string
}

type serverWatchDog struct {
	service string
	name    string
	url     string
}

type argumentList struct {
	sdkArgs.DefaultArgumentList
	Services sdkArgs.JSON
	Timeout  string `default:5 help: "Client http timeout in seconds"`
	Profile  string `default:"" help: "AWS profile, can be blank"`
	Region   string `default:"" help: "AWS region, mandatory"`
	TagName  string `default:"tag:Service" help: "Tag to inspect for resource selection"`
}

const (
	integrationName    = "com.acme.aws-watchdog"
	integrationVersion = "0.1.0"
)

var (
	args argumentList
)

func main() {
	// Create Integration
	i, err := integration.New(integrationName, integrationVersion, integration.Args(&args))
	panicOnErr(err)

	// Add Event
	if args.All() || args.Events {
		for _, server := range fetchServerList(args) {
			var timeout, err = strconv.Atoi(args.Timeout)
			panicOnErr(err)
			var state = status(server, timeout)
			if 0 == state {
				entity, err := i.Entity("app-status"+server.name, "acme.safeassign.health")
				panicOnErr(err)
				err = entity.AddEvent(event.New("Service on "+server.name+" not responding", "app-status-"+server.service))
				panicOnErr(err)
			}
		}

		panicOnErr(err)
	}

	panicOnErr(i.Publish())
}

func status(serverDef serverWatchDog, timeout int) int {
	conn, err := net.DialTimeout("tcp", serverDef.url, time.Duration(timeout)*time.Second)
	if err != nil {
		log.Println("Connection error:", err)
		return 0
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	return 1
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func fetchServerList(args argumentList) []serverWatchDog {
	sess, err := session.NewSession(aws.NewConfig())
	panicOnErr(err)

	ec2metadataSvc := ec2metadata.New(sess)

	var providers []credentials.Provider
	if len(args.Profile) > 0 {
		providers = append(providers, &credentials.SharedCredentialsProvider{
			Profile: args.Profile,
		})
	}
	providers = append(providers, &credentials.EnvProvider{})
	providers = append(providers, &ec2rolecreds.EC2RoleProvider{
		Client: ec2metadataSvc,
	})

	creds := credentials.NewChainCredentials(providers)

	config := aws.NewConfig().
		WithCredentialsChainVerboseErrors(true).
		WithCredentials(creds).
		WithRegion(args.Region)

	sess, err = session.NewSession(aws.NewConfig().WithCredentials(creds))
	ec2Service := ec2.New(sess, config)

	var s []serverWatchDog
	var m = args.Services.Get().(map[string]interface{})
	for k, v := range m {
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {
				var serviceDef = parseService(u.(map[string]interface{}))

				params := &ec2.DescribeInstancesInput{
					Filters: []*ec2.Filter{
						{
							Name: aws.String(args.TagName),
							Values: []*string{
								aws.String(serviceDef.tag),
							},
						},
					},
				}

				res, _ := ec2Service.DescribeInstances(params)
				if len(res.Reservations) > 0 {
					for _, i := range res.Reservations[0].Instances {
						var nt string
						for _, t := range i.Tags {
							if *t.Key == "Name" {
								nt = *t.Value
								break
							}
						}
						s = append(s, serverWatchDog{
							service: serviceDef.name,
							name:    nt,
							url:     *i.PrivateIpAddress + ":" + serviceDef.port,
						})
					}
				}

			}
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
		}
	}
	return s
}

func parseService(serverDef map[string]interface{}) serviceWatchdog {
	return serviceWatchdog{
		name: serverDef["name"].(string),
		tag:  serverDef["tag"].(string),
		port: serverDef["port"].(string),
	}
}

func (wd serviceWatchdog) ListServers() []serverWatchDog {
	var s []serverWatchDog
	return s
}
