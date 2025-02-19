package openbsdvmm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepOutDir struct {
	outputPath string
	name       string
	force      bool
}

func (step *stepOutDir) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Check if the output directory exists.
	if _, err := os.Stat(step.outputPath); !os.IsNotExist(err) {
		ui.Say("Output directory already exists, skipping")
	} else {
		// Create the output directory.
		if err := os.MkdirAll(step.outputPath, 0755); err != nil {
			state.Put("error", fmt.Errorf("output %s", step.outputPath))
			return multistep.ActionHalt
		}
	}

	// Check if output image exists
	if _, err := os.Stat(
		filepath.Join(step.outputPath, step.name)); !os.IsNotExist(err) {
		// If the build isn't forced, error out here.
		if !step.force {
			state.Put("error", fmt.Errorf("image already exists: %s", step.name))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (step *stepOutDir) Cleanup(state multistep.StateBag) {}
