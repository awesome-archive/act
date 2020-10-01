package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/nektos/act/pkg/common"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// NewDockerPullExecutorInput the input for the NewDockerPullExecutor function
type NewDockerPullExecutorInput struct {
	Image     string
	ForcePull bool
}

// NewDockerPullExecutor function to create a run executor for the container
func NewDockerPullExecutor(input NewDockerPullExecutorInput) common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		logger.Debugf("%sdocker pull %v", logPrefix, input.Image)

		if common.Dryrun(ctx) {
			return nil
		}

		pull := input.ForcePull
		if !pull {
			imageExists, err := ImageExistsLocally(ctx, input.Image)
			log.Debugf("Image exists? %v", imageExists)
			if err != nil {
				return errors.WithMessagef(err, "unable to determine if image already exists for image %q", input.Image)
			}

			if !imageExists {
				pull = true
			}
		}

		if !pull {
			return nil
		}

		imageRef := cleanImage(input.Image)
		logger.Debugf("pulling image '%v'", imageRef)

		cli, err := GetDockerClient(ctx)
		if err != nil {
			return err
		}

		reader, err := cli.ImagePull(ctx, imageRef, types.ImagePullOptions{})
		_ = logDockerResponse(logger, reader, err != nil)
		if err != nil {
			return err
		}
		return nil

	}

}

func cleanImage(image string) string {
	imageParts := len(strings.Split(image, "/"))
	if imageParts == 1 {
		image = fmt.Sprintf("docker.io/library/%s", image)
	} else if imageParts == 2 {
		image = fmt.Sprintf("docker.io/%s", image)
	}

	return image
}
