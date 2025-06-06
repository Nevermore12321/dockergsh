package client

import "io"

func (cli *DockerGshClient) Stream(method, path string, in io.Reader, out io.Writer, headers map[string][]string) error {
	return cli.streamHelper(method, path, in, out, headers)
}
