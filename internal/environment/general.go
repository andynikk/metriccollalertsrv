package environment

import (
	"bytes"
	"net"
	"os/exec"
	"strings"

	"github.com/andynikk/metriccollalertsrv/internal/constants"
)

func isOSWindows() bool {

	var stderr bytes.Buffer
	defer stderr.Reset()

	var out bytes.Buffer
	defer out.Reset()

	cmd := exec.Command("cmd", "ver")
	cmd.Stdin = strings.NewReader("some input")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false
	}
	myOS := out.String()
	return strings.Contains(myOS, "Microsoft Windows")
}

func ParseConfigBytes(res []byte) bytes.Buffer {

	var out bytes.Buffer
	configLines := strings.Split(string(res), "\n")
	for i := 0; i < len(configLines); i++ {

		if configLines[i] != "" {
			var strs string
			splitStr := strings.SplitAfterN(configLines[i], "// ", -1)
			if len(splitStr) != 0 {
				strs = strings.Replace(splitStr[0], "// ", "\n", -1)
				out.WriteString(strs)
			}
		}
	}
	return out
}

func getLocalIPAddress(address string) (string, error) {

	conn, err := net.Dial("udp4", address)
	if err != nil {
		constants.Logger.ErrorLog(err)
		return "", err
	}
	defer conn.Close()

	IPAddress := strings.Split(conn.LocalAddr().String(), ":")[0]
	return IPAddress, nil

}
