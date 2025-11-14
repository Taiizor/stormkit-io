package integrations

import (
	"strings"

	"github.com/stormkit-io/stormkit-io/src/lib/errors"
)

// GetFile uses AWS SDK under the hood to return the uploaded file.
func (a AlibabaClient) GetFile(args GetFileArgs) (*GetFileResult, error) {
	args.Location = strings.TrimPrefix(args.Location, "alibaba:")
	result, err := a.awsClient.GetFile(args)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to get file from Alibaba OSS").
			WithContext("location", args.Location).
			WithContext("file_name", args.FileName).
			WithContext("deployment_id", args.DeploymentID.String())
	}
	return result, nil
}
