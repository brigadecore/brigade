package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/v2/internal/file"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/apimachinery/pkg/util/duration"
)

var eventCommand = &cli.Command{
	Name:  "event",
	Usage: "Manage events",
	Subcommands: []*cli.Command{
		{
			Name:  "cancel",
			Usage: "Cancel a single event without deleting it",
			Description: "Unconditionally cancels (and aborts if applicable) a " +
				"single event whose worker is in a non-terminal phase",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Cancel (and abort if applicable) the specified event " +
						"(required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    flagYes,
					Aliases: []string{"y"},
					Usage:   "Non-interactively confirm cancellation",
				},
			},
			Action: eventCancel,
		},
		{
			Name:    "cancel-many",
			Aliases: []string{"cm"},
			Usage:   "Cancel multiple events without deleting them",
			Description: "By default, only cancels events for the specified " +
				"project with their worker in a PENDING phase",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Cancel events for the specified project only (required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    flagRunning,
					Aliases: []string{"r"},
					Usage: "If set, will additionally abort and cancel events with " +
						"their worker in a RUNNING phase",
				},
				&cli.BoolFlag{
					Name:    flagYes,
					Aliases: []string{"y"},
					Usage:   "Non-interactively confirm cancellation",
				},
			},
			Action: eventCancelMany,
		},
		{
			Name:        "create",
			Usage:       "Create a new event",
			Description: "Creates a new event for the specified project",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  flagPayload,
					Usage: "The event payload",
				},
				&cli.StringFlag{
					Name:  flagPayloadFile,
					Usage: "The location of a file containing the event payload",
				},
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Create an event for the specified project (required)",
					Required: true,
				},
				&cli.StringFlag{
					Name:    flagSource,
					Aliases: []string{"s"},
					Usage:   "Override the default event source",
					Value:   "github.com/brigadecore/brigade/cli",
				},
				&cli.StringFlag{
					Name:    flagType,
					Aliases: []string{"t"},
					Usage:   "Override the default event type",
					Value:   "exec",
				},
			},
			Action: eventCreate,
		},
		{
			Name:  "delete",
			Usage: "Delete a single event",
			Description: "Unconditionally deletes (and aborts if applicable) a " +
				"single event whose worker is in any phase",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagID,
					Aliases: []string{"i"},
					Usage: "Delete (and abort if applicable) the specified event " +
						"(required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name:    flagYes,
					Aliases: []string{"y"},
					Usage:   "Non-interactively confirm deletion",
				},
			},
			Action: eventDelete,
		},
		{
			Name:    "delete-many",
			Aliases: []string{"dm"},
			Usage:   "Delete multiple events",
			Description: "Deletes (and aborts if applicable) events for the " +
				"specified project with their workers in the specified phase(s)",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: flagAborted,
					Usage: "If set, will delete events with their worker in an ABORTED " +
						"phase; mutually exclusive with --any-phase and --terminal",
				},
				&cli.BoolFlag{
					Name: flagAnyPhase,
					Usage: "If set, will delete events with their worker in any phase; " +
						"mutually exclusive with all other phase flags",
				},
				&cli.BoolFlag{
					Name: flagCanceled,
					Usage: "If set, will delete events with their worker in a CANCELED " +
						"phase; mutually exclusive with --any-phase and --terminal",
				},
				&cli.BoolFlag{
					Name: flagFailed,
					Usage: "If set, will delete events with their worker in a FAILED " +
						"phase; mutually exclusive with --any-phase and --terminal",
				},
				&cli.BoolFlag{
					Name: flagPending,
					Usage: "If set, will delete events with their worker in a PENDING " +
						"phase; mutually exclusive with --any-phase and --terminal",
				},
				&cli.StringFlag{
					Name:     flagProject,
					Aliases:  []string{"p"},
					Usage:    "Delete events for the specified project only (required)",
					Required: true,
				},
				&cli.BoolFlag{
					Name: flagRunning,
					Usage: "If set, will abort and delete events with their worker in " +
						"a RUNNING phase; mutually exclusive with --any-phase and " +
						"--terminal",
				},
				&cli.BoolFlag{
					Name: flagSucceeded,
					Usage: "If set, will delete events with their worker in a " +
						"SUCCEEDED phase; mutually exclusive with --any-phase and " +
						"--terminal",
				},
				&cli.BoolFlag{
					Name: flagTerminal,
					Usage: "If set, will delete events with their worker in any " +
						"terminal phase; mutually exclusive with all other phase flags",
				},
				&cli.BoolFlag{
					Name: flagTimedOut,
					Usage: "If set, will delete events with their worker in a " +
						"TIMED_OUT phase; mutually exclusive with --any-phase and " +
						"--terminal",
				},
				&cli.BoolFlag{
					Name: flagUnknown,
					Usage: "If set, will delete events with their worker in an UNKNOWN " +
						"phase; mutually exclusive with --any-phase and --terminal",
				},
				&cli.BoolFlag{
					Name:    flagYes,
					Aliases: []string{"y"},
					Usage:   "Non-interactively confirm deletion",
				},
			},
			Action: eventDeleteMany,
		},
		{
			Name:  "get",
			Usage: "Retrieve an event",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i"},
					Usage:    "Retrieve the specified event (required)",
					Required: true,
				},
				cliFlagOutput,
			},
			Action: eventGet,
		},
		{
			Name:        "list",
			Aliases:     []string{"ls"},
			Usage:       "List events",
			Description: "Retrieves all events unless specific criteria are provided",
			Flags: []cli.Flag{
				cliFlagOutput,
				&cli.BoolFlag{
					Name: flagAborted,
					Usage: "If set, will retrieve events with their worker in an " +
						"ABORTED phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
				&cli.BoolFlag{
					Name: flagCanceled,
					Usage: "If set, will retrieve events with their worker in a " +
						"CANCELED phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
				&cli.BoolFlag{
					Name: flagFailed,
					Usage: "If set, will retrieve events with their worker in a FAILED " +
						"phase; mutually exclusive with  --terminal and --non-terminal",
				},
				&cli.BoolFlag{
					Name: flagNonTerminal,
					Usage: "If set, will retrieve events with their worker in any " +
						"non-terminal phase; mutually exclusive with all other phase flags",
				},
				&cli.BoolFlag{
					Name: flagPending,
					Usage: "If set, will retrieve events with their worker in a " +
						"PENDING phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
				&cli.StringFlag{
					Name:    flagProject,
					Aliases: []string{"p"},
					Usage: "If set, will retrieve events only for the specified " +
						"project",
				},
				&cli.BoolFlag{
					Name: flagRunning,
					Usage: "If set, will retrieve events with their worker in RUNNING " +
						"phase; mutually exclusive with --terminal and --non-terminal",
				},
				&cli.BoolFlag{
					Name: flagSucceeded,
					Usage: "If set, will retrieve events with their worker in a " +
						"SUCCEEDED phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
				&cli.BoolFlag{
					Name: flagTerminal,
					Usage: "If set, will retrieve events with their worker in any " +
						"terminal phase; mutually exclusive with all other phase flags",
				},
				&cli.BoolFlag{
					Name: flagTimedOut,
					Usage: "If set, will retrieve events with their worker in a " +
						"TIMED_OUT phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
				&cli.BoolFlag{
					Name: flagUnknown,
					Usage: "If set, will retrieve events with their worker in an " +
						"UNKNOWN phase; mutually exclusive with --terminal and " +
						"--non-terminal",
				},
			},
			Action: eventList,
		},
	},
}

func eventCreate(c *cli.Context) error {
	payload := c.String(flagPayload)
	payloadFile := c.String(flagPayloadFile)
	projectID := c.String(flagProject)
	source := c.String(flagSource)
	eventType := c.String(flagType)

	if payload != "" && payloadFile != "" {
		return errors.New(
			"only one of --payload or --payload-file may be specified",
		)
	}
	if payloadFile != "" {
		exists, err := file.Exists(payloadFile)
		if err != nil {
			return errors.Errorf(
				"error finding configuration file at %s",
				payloadFile,
			)
		}
		if !exists {
			return errors.Errorf("no event payload was found at %s", payloadFile)
		}
		payloadBytes, err := ioutil.ReadFile(payloadFile)
		if err != nil {
			return errors.Wrapf(
				err,
				"error reading event payload from %s",
				payloadFile,
			)
		}
		payload = string(payloadBytes)
	}

	event := core.Event{
		ProjectID: projectID,
		Source:    source,
		Type:      eventType,
		Payload:   payload,
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	events, err := client.Core().Events().Create(c.Context, event)
	if err != nil {
		return err
	}
	fmt.Printf("Created event %q.\n\n", events.Items[0].ID)

	return nil
}

// nolint: gocyclo
func eventList(c *cli.Context) error {
	output := c.String(flagOutput)
	projectID := c.String(flagProject)

	workerPhases := []core.WorkerPhase{}

	if c.Bool(flagAborted) {
		workerPhases = append(workerPhases, core.WorkerPhaseAborted)
	}
	if c.Bool(flagCanceled) {
		workerPhases = append(workerPhases, core.WorkerPhaseCanceled)
	}
	if c.Bool(flagFailed) {
		workerPhases = append(workerPhases, core.WorkerPhaseFailed)
	}
	if c.Bool(flagPending) {
		workerPhases = append(workerPhases, core.WorkerPhasePending)
	}
	if c.Bool(flagRunning) {
		workerPhases = append(workerPhases, core.WorkerPhaseRunning)
	}
	if c.Bool(flagSucceeded) {
		workerPhases = append(workerPhases, core.WorkerPhaseSucceeded)
	}
	if c.Bool(flagTimedOut) {
		workerPhases = append(workerPhases, core.WorkerPhaseTimedOut)
	}
	if c.Bool(flagUnknown) {
		workerPhases = append(workerPhases, core.WorkerPhaseUnknown)
	}

	if c.Bool(flagTerminal) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--terminal is mutually exclusive with all other phase flags",
			)
		}
		workerPhases = core.WorkerPhasesTerminal()
	}

	if c.Bool(flagNonTerminal) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--non-terminal is mutually exclusive with all other phase flags",
			)
		}
		workerPhases = core.WorkerPhasesNonTerminal()
	}

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	selector := core.EventsSelector{
		ProjectID:    projectID,
		WorkerPhases: workerPhases,
	}
	opts := meta.ListOptions{}

	for {
		events, err := client.Core().Events().List(c.Context, &selector, &opts)
		if err != nil {
			return err
		}

		if len(events.Items) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		switch strings.ToLower(output) {
		case flagOutputTable:
			table := uitable.New()
			table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "WORKER PHASE")
			for _, event := range events.Items {
				table.AddRow(
					event.ID,
					event.ProjectID,
					event.Source,
					event.Type,
					duration.ShortHumanDuration(time.Since(*event.Created)),
					event.Worker.Status.Phase,
				)
			}
			fmt.Println(table)

		case flagOutputYAML:
			yamlBytes, err := yaml.Marshal(events)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get events operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case flagOutputJSON:
			prettyJSON, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from get events operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if events.RemainingItemCount < 1 || events.Continue == "" {
			break
		}

		// Exit after one page of output if this isn't a terminal
		if !terminal.IsTerminal(int(os.Stdout.Fd())) {
			break
		}

		if shouldContinue, err :=
			shouldContinue(events.RemainingItemCount); err != nil {
			return err
		} else if !shouldContinue {
			break
		}

		opts.Continue = events.Continue
	}

	return nil
}

func eventGet(c *cli.Context) error {
	id := c.String(flagID)
	output := c.String(flagOutput)

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	event, err := client.Core().Events().Get(c.Context, id)
	if err != nil {
		return err
	}

	switch strings.ToLower(output) {
	case flagOutputTable:
		table := uitable.New()
		table.AddRow("ID", "PROJECT", "SOURCE", "TYPE", "AGE", "WORKER PHASE")
		var age string
		if event.Created != nil {
			age = duration.ShortHumanDuration(time.Since(*event.Created))
		}
		table.AddRow(
			event.ID,
			event.ProjectID,
			event.Source,
			event.Type,
			age,
			event.Worker.Status.Phase,
		)
		fmt.Println(table)

		if len(event.Worker.Jobs) > 0 {
			fmt.Printf("\nEvent %q jobs:\n\n", event.ID)
			table = uitable.New()
			table.AddRow("NAME", "STARTED", "ENDED", "PHASE")
			for jobName, job := range event.Worker.Jobs {
				jobStatus := job.Status
				var started, ended string
				if jobStatus.Started != nil {
					started =
						duration.ShortHumanDuration(time.Since(*jobStatus.Started))
				}
				if jobStatus.Ended != nil {
					ended =
						duration.ShortHumanDuration(time.Since(*jobStatus.Ended))
				}
				table.AddRow(
					jobName,
					started,
					ended,
					jobStatus.Phase,
				)
			}
			fmt.Println(table)
		}

	case flagOutputYAML:
		yamlBytes, err := yaml.Marshal(event)
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(yamlBytes))

	case flagOutputJSON:
		prettyJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			return errors.Wrap(
				err,
				"error formatting output from get event operation",
			)
		}
		fmt.Println(string(prettyJSON))
	}

	return nil
}

func eventCancel(c *cli.Context) error {
	id := c.String(flagID)

	confirmed, err := confirmed(c)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	if err = client.Core().Events().Cancel(c.Context, id); err != nil {
		return err
	}
	fmt.Printf("Event %q canceled.\n", id)

	return nil
}

func eventCancelMany(c *cli.Context) error {
	projectID := c.String(flagProject)
	cancelRunning := c.Bool(flagRunning)

	confirmed, err := confirmed(c)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	selector := core.EventsSelector{
		ProjectID:    projectID,
		WorkerPhases: []core.WorkerPhase{core.WorkerPhasePending},
	}
	if cancelRunning {
		selector.WorkerPhases =
			append(selector.WorkerPhases, core.WorkerPhaseRunning)
	}

	events, err := client.Core().Events().CancelMany(c.Context, selector)
	if err != nil {
		return err
	}
	fmt.Printf("Canceled %d events.\n", events.Count)

	return nil
}

func eventDelete(c *cli.Context) error {
	id := c.String(flagID)

	confirmed, err := confirmed(c)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	if err = client.Core().Events().Delete(c.Context, id); err != nil {
		return err
	}
	fmt.Printf("Event %q deleted.\n", id)

	return nil
}

func eventDeleteMany(c *cli.Context) error {
	projectID := c.String(flagProject)
	workerPhases := []core.WorkerPhase{}

	if c.Bool(flagAborted) {
		workerPhases = append(workerPhases, core.WorkerPhaseAborted)
	}
	if c.Bool(flagCanceled) {
		workerPhases = append(workerPhases, core.WorkerPhaseCanceled)
	}
	if c.Bool(flagFailed) {
		workerPhases = append(workerPhases, core.WorkerPhaseFailed)
	}
	if c.Bool(flagPending) {
		workerPhases = append(workerPhases, core.WorkerPhasePending)
	}
	if c.Bool(flagRunning) {
		workerPhases = append(workerPhases, core.WorkerPhaseRunning)
	}
	if c.Bool(flagSucceeded) {
		workerPhases = append(workerPhases, core.WorkerPhaseSucceeded)
	}
	if c.Bool(flagTimedOut) {
		workerPhases = append(workerPhases, core.WorkerPhaseTimedOut)
	}
	if c.Bool(flagUnknown) {
		workerPhases = append(workerPhases, core.WorkerPhaseUnknown)
	}

	if c.Bool(flagAnyPhase) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--any-phase is mutually exclusive with all other phase flags",
			)
		}
		workerPhases = core.WorkerPhasesAll()
	}

	if c.Bool(flagTerminal) {
		if len(workerPhases) > 0 {
			return errors.New(
				"--terminal is mutually exclusive with all other phase flags",
			)
		}
		workerPhases = core.WorkerPhasesTerminal()
	}

	confirmed, err := confirmed(c)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	client, err := getClient(c)
	if err != nil {
		return err
	}

	selector := core.EventsSelector{
		ProjectID:    projectID,
		WorkerPhases: workerPhases,
	}

	events, err := client.Core().Events().DeleteMany(c.Context, selector)
	if err != nil {
		return err
	}
	fmt.Printf("Deleted %d events.\n", events.Count)

	return nil
}
