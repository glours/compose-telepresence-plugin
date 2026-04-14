package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Down uninstalls telepresence helm chart and disconnects from the cluster
func Down(options PluginOptions) error {
	// Detect if we need docker mode by trying to reach the daemon
	conn := detectConnection(options)

	if err := deleteIntercept(options, conn); err != nil {
		return err
	}

	if err := uninstallTelepresenceChart(options, conn); err != nil {
		_ = sendErrorf("failed to uninstall telepresence chart: %v", err)
		return err
	}

	if err := disconnectFromCluster(conn); err != nil {
		_ = sendErrorf("failed to disconnect: %v", err)
		return err
	}

	return nil
}

// detectConnection checks if a docker-based daemon is running and returns
// the appropriate connectionInfo.
func detectConnection(opts PluginOptions) connectionInfo {
	// Check if a docker daemon container exists for this context/namespace
	daemonName, err := findDaemonName(opts)
	if err != nil {
		return connectionInfo{}
	}

	cmd := exec.Command("docker", "ps", "-q", "--filter", fmt.Sprintf("name=%s", daemonName))
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil || strings.TrimSpace(stdout.String()) == "" {
		return connectionInfo{}
	}
	_ = sendDebug(fmt.Sprintf("detected docker daemon: %s", daemonName))
	return connectionInfo{daemonUse: daemonName}
}

func deleteIntercept(options PluginOptions, conn connectionInfo) error {
	_ = sendInfo("Removing intercept...\n")
	args := []string{"leave", options.Name}
	args = append(args, conn.useArgs()...)

	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Tolerate errors when the intercept is already gone or when
		// we can't reach the daemon (stale state from a failed up).
		if strings.Contains(stderrStr, "traffic manager not found") ||
			strings.Contains(stderrStr, "Found no ") ||
			strings.Contains(stderrStr, "failed to connect to root daemon") ||
			strings.Contains(stderrStr, "Not connected") {
			_ = sendInfo(fmt.Sprintf("Intercept %q already removed\n", options.Name))
			return nil
		}
		_ = sendErrorf("failed to remove intercept %q: %s: %s", options.Name, err, stderrStr)
		return err
	}

	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	_ = sendInfo(fmt.Sprintf("Intercept %q removed successfully\n", options.Name))
	return nil
}

func uninstallTelepresenceChart(options PluginOptions, conn connectionInfo) error {
	args := []string{"helm", "uninstall"}
	args = append(args, conn.useArgs()...)
	if options.Namespace != "" {
		args = append(args, "--namespace", options.Namespace)
	}
	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	_ = sendDebug(fmt.Sprintf("uninstall command: telepresence %s", strings.Join(args, " ")))
	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Tolerate errors when we can't reach the daemon (stale state)
		if strings.Contains(stderrStr, "failed to connect to root daemon") ||
			strings.Contains(stderrStr, "Not connected") {
			_ = sendInfo("Telepresence chart already uninstalled or unreachable\n")
			return nil
		}
		_ = sendErrorf("telepresence helm uninstall failed: %s: %s", err, stderrStr)
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return nil
}

// quitDaemon silently disconnects a daemon, ignoring errors.
// Used to clean up a stale docker daemon before retrying a fresh connection.
func quitDaemon(conn connectionInfo) error {
	args := []string{"quit"}
	args = append(args, conn.useArgs()...)
	cmd := exec.Command("telepresence", args...)
	_ = cmd.Run()
	return nil
}

func disconnectFromCluster(conn connectionInfo) error {
	args := []string{"quit"}
	args = append(args, conn.useArgs()...)

	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	_ = sendDebug(fmt.Sprintf("disconnect command: telepresence %s", strings.Join(args, " ")))
	if err := cmd.Run(); err != nil {
		stderrStr := stderr.String()
		// Tolerate errors when there's nothing to disconnect
		if strings.Contains(stderrStr, "failed to connect to root daemon") ||
			strings.Contains(stderrStr, "Not connected") {
			_ = sendInfo("Already disconnected\n")
			return nil
		}
		_ = sendErrorf("%s: %s", err, stderrStr)
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return nil
}
