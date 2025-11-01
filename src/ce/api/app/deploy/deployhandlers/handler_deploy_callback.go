package deployhandlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stormkit-io/stormkit-io/src/ce/api/admin"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deploy/deployhooks"
	"github.com/stormkit-io/stormkit-io/src/lib/config"
	"github.com/stormkit-io/stormkit-io/src/lib/integrations"
	"github.com/stormkit-io/stormkit-io/src/lib/shttp"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
	"gopkg.in/guregu/null.v3"
)

const OutcomeSuccess = "success"
const OutcomeFailed = "failed"
const OutcomeSkipped = "skipped"
const OutcomeCancelled = "cancelled"

type deployCallbackRequest struct {
	DeployID string `json:"deployId"`

	// Commit related information
	RunID  string            `json:"runId"`
	Commit deploy.CommitInfo `json:"commit"`

	// Logs related information
	Logs string `json:"logs"`

	// Deployment result related information
	Result          integrations.UploadResult `json:"result"`  // The upload result
	UploadError     string                    `json:"error"`   // This field is generated when an error occurs during the upload
	Outcome         string                    `json:"outcome"` // Possible values: success | failure | cancelled | skipped
	Manifest        *deploy.BuildManifest     `json:"manifest"`
	HasStatusChecks bool                      `json:"hasStatusChecks"`

	// Final call
	Lock bool `json:"lock"`

	deployment *deploy.Deployment
}

// handlerDeployCallback updates the deployment based on the request.
func handlerDeployCallback(req *shttp.RequestContext) *shttp.Response {
	data := deployCallbackRequest{}

	if err := req.Post(&data); err != nil {
		return shttp.Error(err)
	}

	deployID, err := utils.DecryptID(data.DeployID)

	if err != nil {
		return shttp.NotAllowed()
	}

	depl, err := deploy.NewStore().MyDeployment(req.Context(), &deploy.DeploymentsQueryFilters{
		DeploymentID: deployID,
		IncludeLogs:  aws.Bool(true),
	})

	if err != nil {
		return shttp.Error(err)
	}

	if depl == nil {
		return shttp.NotFound()
	}

	data.deployment = depl

	if data.deployment.Status() != "running" {
		return &shttp.Response{
			Status: http.StatusConflict,
		}
	}

	if data.deployment.IsLocked() {
		return shttp.NoContent()
	}

	if data.Lock {
		return lockDeployment(req, data)
	}

	if data.Commit.ID.ValueOrZero() != "" || data.RunID != "" {
		return updateCommit(req, data)
	}

	if data.Logs != "" {
		if data.deployment.ExitCode.Valid {
			return updateStatusCheckLogs(req, data)
		}

		return updateLogs(req, data)
	}

	if data.Outcome != "" {
		return updateExit(req, data)
	}

	return shttp.OK()
}

func updateLogs(req *shttp.RequestContext, data deployCallbackRequest) *shttp.Response {
	store := deploy.NewStore()
	ctx := req.Context()

	if err := store.UpdateLogs(ctx, data.deployment.ID, data.Logs); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}

func updateStatusCheckLogs(req *shttp.RequestContext, data deployCallbackRequest) *shttp.Response {
	store := deploy.NewStore()
	ctx := req.Context()

	if err := store.UpdateStatusChecks(ctx, data.deployment.ID, data.Logs); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}

func updateCommit(req *shttp.RequestContext, data deployCallbackRequest) *shttp.Response {
	store := deploy.NewStore()
	ctx := req.Context()

	if data.Commit.ID.ValueOrZero() != "" {
		if err := store.UpdateCommitInfo(ctx, data.deployment.ID, data.Commit); err != nil {
			return shttp.Error(err)
		}
	}

	if data.RunID != "" {
		if err := store.UpdateGithubRunID(ctx, data.deployment.ID, utils.StringToID(data.RunID)); err != nil {
			return shttp.Error(err)
		}
	}

	return shttp.OK()
}

func updateExit(req *shttp.RequestContext, data deployCallbackRequest) *shttp.Response {
	// TODO: Retry job perhaps?
	if data.UploadError != "" {
		data.deployment.Error = null.NewString(
			fmt.Sprintf("Error: %s", data.UploadError),
			true,
		)
	}

	data.deployment.BuildManifest = data.Manifest

	switch data.Outcome {
	case OutcomeSuccess:
		data.deployment.ExitCode = null.NewInt(0, true)
	case OutcomeCancelled, OutcomeSkipped:
		data.deployment.ExitCode = null.NewInt(-1, true)
	default:
		data.deployment.ExitCode = null.NewInt(1, true)
	}

	if err := deploy.NewStore().UpdateDeploymentResult(req.Context(), data.deployment, data.Result); err != nil {
		return shttp.Error(err, fmt.Sprintf("error while updating deployment result: %s", err.Error()))
	}

	if config.IsSelfHosted() {
		ctx := req.Context()
		cnf, err := admin.Store().Config(ctx)

		if err != nil {
			return shttp.Error(err, fmt.Sprintf("error while getting instance config: %s", err.Error()))
		}

		slog.Debug(slog.LogOpts{
			Msg:   fmt.Sprintf("auto install runtimes after deployment: %v", cnf.SystemConfig == nil || cnf.SystemConfig.AutoInstall),
			Level: slog.DL2,
		})

		if cnf.SystemConfig == nil || cnf.SystemConfig.AutoInstall {
			if err := admin.AddRuntimes(context.Background(), data.deployment.BuildManifest.Runtimes); err != nil {
				return shttp.Error(err, fmt.Sprintf("error while adding runtime configs to the instance: %s", err.Error()))
			}
		}
	}

	deployhooks.Exec(req.Context(), data.deployment)

	if !data.HasStatusChecks {
		return lockDeployment(req, data)
	}

	return shttp.OK()
}

// lockDeployment is called when the deployment is complete. A deployment is complete
// when all status checks have completed. If there are no status checks, it is complete
// when the main deployment process is complete.
//
// @since Runner v1.6.17
func lockDeployment(req *shttp.RequestContext, d deployCallbackRequest) *shttp.Response {
	isSuccess := d.Outcome == OutcomeSuccess

	if isSuccess {
		if err := deploy.AutoPublishIfNecessary(req.Context(), d.deployment); err != nil {
			return shttp.Error(err)
		}
	}

	var statusChecksPassed null.Bool

	if d.HasStatusChecks {
		statusChecksPassed = null.BoolFrom(isSuccess)
	}

	if err := deploy.NewStore().LockDeployment(req.Context(), d.deployment.ID, statusChecksPassed); err != nil {
		return shttp.Error(err)
	}

	return shttp.OK()
}
