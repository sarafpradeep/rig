package mount

import (
	"os"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
)

func (c Cmd) set(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	var err error
	if srcPath == "" {
		srcPath, err = common.PromptInput("Source path", common.ValidateFilePathOpt)
		if err != nil {
			return err
		}
	}

	bytes, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}

	if dstPath == "" {
		dstPath, err = common.PromptInput("Destination path", common.ValidateAbsPathOpt)
		if err != nil {
			return err
		}
	}

	cf := &capsule.Change_SetConfigFile{
		SetConfigFile: &capsule.Change_ConfigFile{
			Path:    dstPath,
			Content: bytes,
		},
	}

	req := &connect.Request[capsule.DeployRequest]{
		Msg: &capsule.DeployRequest{
			CapsuleId: capsule_cmd.CapsuleID,
			Changes: []*capsule.Change{
				{
					Field: cf,
				},
			},
		},
	}

	_, err = c.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.AbortAndDeploy(ctx, capsule_cmd.CapsuleID, c.Rig, req)
	}
	if err != nil {
		return err
	}
	return nil
}
