package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-service-broker/contrib/pkg/client"
	"github.com/Azure/azure-service-broker/pkg/api"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

func provision(c *cli.Context) error {
	host := c.GlobalString(flagHost)
	port := c.GlobalInt(flagPort)
	username := c.GlobalString(flagUsername)
	password := c.GlobalString(flagPassword)
	serviceID := c.String(flagServiceID)
	if serviceID == "" {
		return fmt.Errorf("--%s is a required flag", flagServiceID)
	}
	planID := c.String(flagPlanID)
	if planID == "" {
		return fmt.Errorf("--%s is a required flag", flagPlanID)
	}
	params := make(map[string]string)
	rawParamStrs := c.StringSlice(flagParameter)
	for _, rawParamStr := range rawParamStrs {
		rawParamStr = strings.TrimSpace(rawParamStr)
		tokens := strings.Split(rawParamStr, "=")
		if len(tokens) != 2 {
			return errors.New("parameter string is incorrectly formatted")
		}
		key := strings.TrimSpace(tokens[0])
		value := strings.TrimSpace(tokens[1])
		params[key] = value
	}
	instanceID, err := client.Provision(
		host,
		port,
		username,
		password,
		serviceID,
		planID,
		params,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nProvisioning service instance %s\n\n", instanceID)
	if c.Bool(flagPoll) {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for range ticker.C {
			result, err := client.Poll(
				host,
				port,
				username,
				password,
				instanceID,
				api.OperationProvisioning,
			)
			if err != nil {
				return fmt.Errorf("error polling for provisioning status: %s", err)
			}
			switch result {
			case api.OperationStateInProgress:
				fmt.Print(".")
			case api.OperationStateSucceeded:
				fmt.Printf(
					"\n\nService instance %s has been successfully provisioned\n\n",
					instanceID,
				)
				return nil
			case api.OperationStateFailed:
				return fmt.Errorf(
					"Provisioning service instance %s has failed",
					instanceID,
				)
			default:
				return fmt.Errorf("Unrecognized operation status: %s", result)
			}
		}
	}
	return nil
}