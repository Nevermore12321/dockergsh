package client

import (
	"fmt"
	"github.com/Nevermore12321/dockergsh/service"
)

func GetHelpUsage(method string) string {
	var usages = map[string]string{
		"attach":  "Attach to a running container",
		"build":   "Build an image from a Dockerfile",
		"commit":  "Create a new image from a container's changes",
		"cp":      "Copy files/folders from a container's filesystem to the host path",
		"diff":    "Insp:ct changes on a container's filesystem",
		"events":  "Get real time events from the server",
		"export":  "Stream the contents of a container as a tar archive",
		"history": "Show the history of an image",
		"images":  "List images",
		"import":  "Create a new filesystem image from the contents of a tarball",
		"info":    "Display system-wide information",
		"inspect": "Return low-level information on a container",
		"kill":    "Kill a running container",
		"load":    "Load an image from a tar archive",
		"login":   "Register or log in to a Docker registry server",
		"logout":  "Log out from a Docker registry server",
		"logs":    "Fetch the logs of a container",
		"port":    "Lookup the public-facing port that is NAT-ed to PRIVATE_PORT",
		"pause":   "Pause all processes within a container",
		"ps":      "List containers",
		"pull":    "Pull an image or a repository from a Docker registry server",
		"push":    "Push an image or a repository to a Docker registry server",
		"restart": "Restart a running container",
		"rm":      "Remove one or more containers",
		"rmi":     "Remove one or more images",
		"run":     "Run a command in a new container",
		"save":    "Save an image to a tar archive",
		"search":  "Search for an image on the Docker Hub",
		"start":   "Start a stopped container",
		"stop":    "Stop a running container",
		"tag":     "Tag an image into a repository",
		"top":     "Lookup the running processes of a container",
		"unpause": "Unpause a paused container",
		"version": "Show the Docker version information",
		"wait":    "Block until a container stops, then print its exit code",
	}
	if method != "" {
		cmdHelp, exist := usages[method]
		if exist {
			return cmdHelp
		}
	}
	help := fmt.Sprintf("dockergsh [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", service.DEFAULTUNIXSOCKET)
	return help
}
