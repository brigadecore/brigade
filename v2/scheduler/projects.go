package main

import (
	"context"
	"log"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/pkg/errors"
)

func (s *scheduler) manageProjects(ctx context.Context) {
	// Maintain a map of functions for canceling the loops for each known Project
	loopCancelFns := map[string]func(){}

	ticker := time.NewTicker(s.config.addAndRemoveProjectsInterval)
	defer ticker.Stop()

	for {

		// Build a set of current projects. This makes it a little faster and easier
		// to search for projects later in this algorithm.
		currentProjects := map[string]struct{}{}
		listOpts := &meta.ListOptions{Limit: 100}
		for {
			projects, err := s.projectsClient.List(ctx, nil, listOpts)
			if err != nil {
				select {
				case s.errCh <- errors.Wrap(err, "error listing projects"):
				case <-ctx.Done():
				}
				return
			}
			for _, project := range projects.Items {
				currentProjects[project.ID] = struct{}{}
			}
			if projects.RemainingItemCount > 0 {
				listOpts.Continue = projects.Continue
			} else {
				break
			}
		}

		// Reconcile differences between projects we knew about already and the
		// current set of projects...

		// Stop Worker and Job scheduling loops for projects that have been deleted
		for projectID, cancelFn := range loopCancelFns {
			if _, stillExists := currentProjects[projectID]; !stillExists {
				log.Printf(
					"stopping worker and job scheduling loops for project %q",
					projectID,
				)
				cancelFn()
				delete(loopCancelFns, projectID)
			}
		}

		// Start Worker and Job scheduling loops for any projects that have been
		// added
		for projectID := range currentProjects {
			if _, known := loopCancelFns[projectID]; !known {
				loopCtx, loopCtxCancelFn := context.WithCancel(ctx)
				loopCancelFns[projectID] = loopCtxCancelFn
				log.Printf(
					"starting worker and job scheduling loops for project %q",
					projectID,
				)
				go s.runWorkerLoopFn(loopCtx, projectID)
				go s.runJobLoopFn(loopCtx, projectID)
			}
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}

}
