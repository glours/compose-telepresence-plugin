package pkg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type PluginOptions struct {
	Name      string
	Service   string
	Namespace string
	Port      string
}

// connectionInfo holds state about the telepresence connection established during Up.
type connectionInfo struct {
	// daemonUse is the --use flag value for targeting the docker-based daemon.
	// Empty when connected without docker mode.
	daemonUse string
}

// Up orchestrates the telepresence connection and setup
func Up(options PluginOptions) error {
	if err := checkTelepresenceInstalled(); err != nil {
		return err
	}

	conn, err := connectToCluster(options)
	if err != nil {
		if strings.Contains(err.Error(), "traffic manager not found") {
			// Quit the stale daemon before installing, so we can start fresh after
			if conn.daemonUse != "" {
				_ = quitDaemon(conn)
			}
			_ = sendInfo("Traffic manager not found. Installing Telepresence helm chart...\n")
			if err := installTelepresenceChart(options, conn); err != nil {
				return err
			}
			// Fresh connection now that traffic manager is installed
			conn, err = connectToCluster(options)
			if err != nil {
				return err
			}
		} else {
			return sendErrorf("failed to connect: %v", err)
		}
	}

	if err := createIntercept(options, conn); err != nil {
		return err
	}

	// In docker mode, the daemon container forwards intercepted traffic to 127.0.0.1
	// which is the container's own loopback. Set up a relay to the host.
	if conn.daemonUse != "" && options.Port != "" {
		if err := setupPortRelay(options, conn); err != nil {
			_ = sendDebug(fmt.Sprintf("port relay setup failed: %v", err))
		}
	}

	return nil
}

func checkTelepresenceInstalled() error {
	_, err := exec.LookPath("telepresence")
	if err != nil {
		_ = sendErrorf("telepresence is not installed or not in PATH: %v", err)
		return err
	}
	return nil
}

func connectToCluster(opts PluginOptions) (connectionInfo, error) {
	_ = sendInfo("Connecting to the cluster...\n")
	args := []string{"connect"}
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	// Try normal connect first
	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	_ = sendDebug(fmt.Sprintf("connect command: telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	if err == nil {
		if stdout.String() != "" {
			_ = sendInfo(stdout.String())
		}
		return connectionInfo{}, nil
	}

	// If the root daemon can't be launched or has issues (e.g., subnet conflicts
	// from stale sessions), fall back to docker mode which bypasses the root daemon.
	stderrStr := stderr.String()
	if strings.Contains(stderrStr, "failed to launch the daemon service") ||
		strings.Contains(stderrStr, "failed to connect to root daemon") {
		_ = sendDebug("Root daemon not available or unhealthy, falling back to docker mode...")
		return connectWithDocker(opts)
	}

	if strings.Contains(stderr.String(), "traffic manager not found") {
		return connectionInfo{}, fmt.Errorf("%s: %s", err, stderr.String())
	}
	_ = sendErrorf("%s: %s", err, stderr.String())
	return connectionInfo{}, err
}

func connectWithDocker(opts PluginOptions) (connectionInfo, error) {
	// Clean up any stale daemon container from a previous session (e.g., after cluster reset)
	cleanupStaleDaemon(opts)

	args := []string{"connect", "--docker", "--insecure-skip-tls-verify"}
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}

	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	_ = sendDebug(fmt.Sprintf("connect command (docker mode): telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	// Even on failure, the docker daemon may have been launched.
	// Resolve the daemon name so subsequent commands (helm install, retry connect) can use --use.
	daemonUse, nameErr := findDaemonName(opts)
	conn := connectionInfo{}
	if nameErr == nil {
		conn = connectionInfo{daemonUse: daemonUse}
		_ = sendDebug(fmt.Sprintf("using docker daemon: %s", daemonUse))
	}

	if err != nil {
		if strings.Contains(stderr.String(), "traffic manager not found") {
			return conn, fmt.Errorf("%s: %s", err, stderr.String())
		}
		_ = sendErrorf("docker connect failed: %s: %s", err, stderr.String())
		return conn, err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return conn, nil
}

func findDaemonName(opts PluginOptions) (string, error) {
	// Get current k8s context
	cmd := exec.Command("kubectl", "config", "current-context")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get current k8s context: %v", err)
	}
	context := strings.TrimSpace(stdout.String())
	ns := opts.Namespace
	if ns == "" {
		ns = "default"
	}
	return fmt.Sprintf("%s-%s-cn", context, ns), nil
}

// useArgs returns --use flags if a docker daemon is active.
func (c connectionInfo) useArgs() []string {
	if c.daemonUse != "" {
		return []string{"--use", c.daemonUse}
	}
	return nil
}

func installTelepresenceChart(opts PluginOptions, conn connectionInfo) error {
	args := []string{"helm", "install"}
	args = append(args, conn.useArgs()...)
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	cmd := exec.Command("telepresence", args...)
	var stderr, stout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stout
	_ = sendDebug(fmt.Sprintf("installation command: telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	if err != nil {
		_ = sendErrorf("telepresence helm install failed: %s: %s", err, stderr.String())
		return err
	}
	_ = sendInfo(stout.String())
	return nil
}

// cleanupStaleDaemon removes a leftover daemon container from a previous session.
// This handles the case where the k8s cluster was reset but the container survived.
func cleanupStaleDaemon(opts PluginOptions) {
	daemonName, err := findDaemonName(opts)
	if err != nil {
		return
	}
	// Check if container exists (running or stopped)
	cmd := exec.Command("docker", "ps", "-aq", "--filter", fmt.Sprintf("name=^%s$", daemonName))
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil || strings.TrimSpace(stdout.String()) == "" {
		return
	}
	_ = sendDebug(fmt.Sprintf("removing stale daemon container: %s", daemonName))
	rmCmd := exec.Command("docker", "rm", "-f", daemonName)
	_ = rmCmd.Run()
}

// setupPortRelay installs socat in the daemon container and starts a relay from
// 127.0.0.1:<localPort> to host.docker.internal:<localPort>. This is needed because
// the telepresence daemon runs inside a Docker container where 127.0.0.1 is the
// container's own loopback, not the host.
func setupPortRelay(opts PluginOptions, conn connectionInfo) error {
	localPort := extractLocalPort(opts.Port)
	if localPort == "" {
		return fmt.Errorf("could not extract local port from %q", opts.Port)
	}

	_ = sendDebug(fmt.Sprintf("setting up port relay: 127.0.0.1:%s -> host.docker.internal:%s", localPort, localPort))

	// Install socat if not present
	installCmd := exec.Command("docker", "exec", conn.daemonUse, "sh", "-c",
		"which socat >/dev/null 2>&1 || apk add --no-cache socat >/dev/null 2>&1")
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install socat in daemon container: %v", err)
	}

	// Start socat relay in background
	relayCmd := exec.Command("docker", "exec", "-d", conn.daemonUse,
		"socat", fmt.Sprintf("TCP-LISTEN:%s,fork,reuseaddr", localPort),
		fmt.Sprintf("TCP:host.docker.internal:%s", localPort))
	if err := relayCmd.Run(); err != nil {
		return fmt.Errorf("failed to start port relay: %v", err)
	}

	_ = sendInfo(fmt.Sprintf("Port relay active: container:%s -> host:%s\n", localPort, localPort))
	return nil
}

// extractLocalPort gets the local port number from the port option.
// Formats: "5732:api-80" -> "5732", "5732" -> "5732"
func extractLocalPort(port string) string {
	if i := strings.Index(port, ":"); i > 0 {
		return port[:i]
	}
	return port
}

func createIntercept(opts PluginOptions, conn connectionInfo) error {
	_ = sendInfo(fmt.Sprintf("Creating intercept %q ...\n", opts.Name))
	args := []string{"intercept", opts.Name}
	args = append(args, conn.useArgs()...)

	if opts.Port != "" {
		args = append(args, "--port", opts.Port)
	}

	if opts.Service != "" {
		args = append(args, "--service", opts.Service)
	}

	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	_ = sendDebug(fmt.Sprintf("intercept command: telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	if err != nil {
		if strings.Contains(stderr.String(), "already exists") {
			_ = sendInfo(fmt.Sprintf("Intercept %q already exists\n", opts.Name))
			return nil
		}
		_ = sendErrorf("failed to create intercept %q: %s: %s", opts.Name, err, stderr.String())
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}

	_ = sendInfo(fmt.Sprintf("Intercept %q created successfully\n", opts.Name))
	return nil
}
