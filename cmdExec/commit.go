package cmdExec

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

func CommitContainer(containerName, imagesName string) {
	// todo read containerInfo from config file
	mergeURL := "/var/lib/dockergsh/"
	imageTarURL := "/var/lib/dockergsh"
	if _, err := exec.Command("tar", "-cvf", imageTarURL, "-C", mergeURL, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mergeURL, err)
	}
}