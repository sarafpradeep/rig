package cluster

import (
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/cluster"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func (c Cmd) get(cmd *cobra.Command, args []string) error {
	ctx := c.Ctx
	resp, err := c.Rig.Cluster().GetConfig(ctx, connect.NewRequest(&cluster.GetConfigRequest{}))
	if err != nil {
		return err
	}
	config := resp.Msg
	bytes, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	return nil
}
