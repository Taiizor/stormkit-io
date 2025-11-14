package jobs

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/stormkit-io/stormkit-io/src/ce/api/app/deployservice"
	"github.com/stormkit-io/stormkit-io/src/lib/errors"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"github.com/stormkit-io/stormkit-io/src/lib/utils"
)

// HandleDeploymentStart sends the deployment request to the deployer service.
func HandleDeploymentStart(ctx context.Context, t *asynq.Task) error {
	payload := t.Payload()
	message, err := deployservice.FromEncrypted(string(payload))

	if err != nil {
		err = errors.Wrapf(err, errors.ErrorTypeInternal, "failed to decrypt deployment message")
		slog.Errorf("cannot retrieve deployment message: %v", err)
		return err
	}

	args := deployservice.SendPayloadArgs{
		DeploymentID: utils.StringToID(message.Build.DeploymentID),
		EncryptedMsg: string(payload),
	}

	if err := deployservice.Service().SendPayload(args); err != nil {
		// Stopped deployment - so no need to continue
		if err.Error() == "exit status 128" {
			return nil
		}

		err = errors.Wrapf(err, errors.ErrorTypeInternal, "failed to send payload to deploy service for deployment_id=%s", message.Build.DeploymentID)
		slog.Errorf("deploy service error: %v", err)
		return err
	}

	return nil
}
