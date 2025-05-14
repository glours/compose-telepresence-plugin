package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Down uninstalls telepresence helm chart and disconnects from the cluster
func Down(options PluginOptions) error {
	if err := deleteIntercept(options); err != nil {
		return err
	}

	if err := uninstallTelepresenceChart(options); err != nil {
		_ = sendErrorf("failed to uninstall telepresence chart: %v", err)
		return err
	}

	if err := disconnectFromCluster(); err != nil {
		_ = sendErrorf("failed to disconnect: %v", err)
		return err
	}

	return nil
}

func deleteIntercept(options PluginOptions) error {
	_ = sendInfo("Removing intercept...\n")
	args := []string{"leave", options.Name}

	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "traffic manager not found") || strings.Contains(stderr.String(), "Found no ") {
			_ = sendInfo(fmt.Sprintf("Intercept %q already removed\n", options.Name))
			return nil
		}
		_ = sendErrorf("failed to remove intercept %q: %s: %s", options.Name, err, stderr.String())
		return err
	}

	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	_ = sendInfo(fmt.Sprintf("Intercept %q removed successfully\n", options.Name))
	return nil
}

func uninstallTelepresenceChart(options PluginOptions) error {
	args := []string{"helm", "uninstall"}
	if options.Namespace != "" {
		args = append(args, "--namespace", options.Namespace)
	}
	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		_ = sendErrorf("telepresence helm uninstall failed: %s: %s", err, stderr.String())
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return nil
}

func disconnectFromCluster() error {
	cmd := exec.Command("telepresence", "quit")
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		_ = sendErrorf("%s: %s", err, stderr.String())
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return nil
}
