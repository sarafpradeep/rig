package scale

import (
	"fmt"
	"math"

	"github.com/bufbuild/connect-go"
	"github.com/rigdev/rig-go-api/api/v1/capsule"
	"github.com/rigdev/rig/cmd/common"
	capsule_cmd "github.com/rigdev/rig/cmd/rig/cmd/capsule"
	"github.com/rigdev/rig/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
)

func (r Cmd) vertical(cmd *cobra.Command, args []string) error {
	ctx := r.Ctx
	container, _, err := capsule_cmd.GetCurrentContainerResources(ctx, r.Rig)
	if err != nil {
		return nil
	}
	if container == nil {
		fmt.Println("Capsule has no rollouts yet")
		return nil
	}

	if allFlagsEmpty() {
		err = setResourcesInteractive(container.Resources)
	} else {
		err = setResourcesFromFlags(container.Resources)
	}
	if err != nil {
		return err
	}

	req := connect.NewRequest(&capsule.DeployRequest{
		CapsuleId: capsule_cmd.CapsuleID,
		Changes: []*capsule.Change{
			{
				Field: &capsule.Change_ContainerSettings{
					ContainerSettings: container,
				},
			},
		},
	})

	_, err = r.Rig.Capsule().Deploy(ctx, req)
	if errors.IsFailedPrecondition(err) && errors.MessageOf(err) == "rollout already in progress" {
		_, err = capsule_cmd.AbortAndDeploy(ctx, capsule_cmd.CapsuleID, r.Rig, req)
	}
	if err != nil {
		return err
	}

	return nil
}

func setResourcesInteractive(curResources *capsule.Resources) error {
	for {
		i, _, err := common.PromptSelect("What to update", []string{"Requests", "Limits", "Done"})
		if err != nil {
			return err
		}

		var curR *capsule.ResourceList
		var name string
		done := false
		switch i {
		case 0:
			curR = curResources.Requests
			name = "request"
		case 1:
			curR = curResources.Limits
			name = "limit"
		default:
			done = true
		}
		if done {
			break
		}

		isLimit := i == 1
		var cpu string
		var mem string
		for {
			i, _, err := common.PromptSelect(fmt.Sprintf("Which resource %s to update", name), []string{"CPU", "Memory", "Done"})
			if err != nil {
				return err
			}
			var name string
			var current string
			var resourceString *string
			done := false
			switch i {
			case 0:
				name = "CPU"
				if curR.GetCpuMillis() == 0 && isLimit {
					current = "-"
				} else {
					current = milliIntToString(uint64(curR.GetCpuMillis()))
				}
				resourceString = &cpu
			case 1:
				name = "memory"
				if curR.GetMemoryBytes() == 0 && isLimit {
					current = "-"
				} else {
					current = intToByteString(curR.GetMemoryBytes())
				}
				resourceString = &mem
			default:
				done = true
			}
			if done {
				break
			}

			label := fmt.Sprintf("New %s (current is %s):", name, current)
			*resourceString, err = common.PromptInput(label, common.ValidateQuantityOpt)
			if err != nil {
				return err
			}
		}
		if err := updateResources(curR, cpu, mem); err != nil {
			return err
		}
	}

	return nil
}

func milliIntToString(millis uint64) string {
	return fmt.Sprintf("%v", float64(millis)/1000)
}

func intToByteString(i uint64) string {
	return common.FormatIntToSI(i, 3) + "B"
}

func setResourcesFromFlags(curResources *capsule.Resources) error {
	if err := updateResources(curResources.Requests, requestCPU, requestMemory); err != nil {
		return err
	}

	if err := updateResources(curResources.Limits, limitCPU, limitMemory); err != nil {
		return err
	}

	return nil
}

func updateResources(resources *capsule.ResourceList, cpu, mem string) error {
	if cpu != "" {
		milliCPU, err := parseMilli(cpu)
		if err != nil {
			return err
		}
		resources.CpuMillis = milliCPU
	}

	if mem != "" {
		mem, err := parseBytes(mem)
		if err != nil {
			return nil
		}
		resources.MemoryBytes = mem
	}

	return nil
}

func parseMilli(s string) (uint32, error) {
	sQ, err := resource.ParseQuantity(s)
	if err != nil {
		return 0, err
	}
	sF := sQ.AsApproximateFloat64()
	return uint32(math.Round(sF * 1000)), nil
}

func parseBytes(s string) (uint64, error) {
	lastChar := s[len(s)-1]
	if lastChar == 'B' || lastChar == 'b' {
		s = s[:len(s)-1]
	}
	bytesQ, err := resource.ParseQuantity(s)
	if err != nil {
		return 0, err
	}
	bytesF := bytesQ.AsApproximateFloat64()
	return uint64(math.Round(bytesF)), nil
}

func allFlagsEmpty() bool {
	return requestCPU == "" && requestMemory == "" && limitCPU == "" && limitMemory == ""
}
