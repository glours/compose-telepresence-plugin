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

// Up orchestrates the telepresence connection and setup
func Up(options PluginOptions) error {
	if err := checkTelepresenceInstalled(); err != nil {
		return err
	}

	if err := connectToCluster(options); err != nil {
		if strings.Contains(err.Error(), "traffic manager not found") {
			_ = sendInfo("Traffic manager not found. Installing Telepresence helm chart...\n")
			if err := installTelepresenceChart(options); err != nil {
				return err
			}
			// Retry connection after installation
			if err := connectToCluster(options); err != nil {
				return err
			}
		} else {
			return sendErrorf("failed to connect: %v", err)
		}
	}

	if err := createIntercept(options); err != nil {
		return err
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

func connectToCluster(opts PluginOptions) error {
	_ = sendInfo("Connecting to the cluster...\n")
	args := []string{"connect"}
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	cmd := exec.Command("telepresence", args...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	_ = sendDebug(fmt.Sprintf("connexion command: telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	if err != nil {
		if strings.Contains(stderr.String(), "traffic manager not found") {
			return fmt.Errorf("%s: %s", err, stderr.String())
		}
		_ = sendErrorf("%s: %s", err, stderr.String())
		return err
	}
	if stdout.String() != "" {
		_ = sendInfo(stdout.String())
	}
	return nil
}

func installTelepresenceChart(opts PluginOptions) error {
	args := []string{"helm", "install"}
	if opts.Namespace != "" {
		args = append(args, "--namespace", opts.Namespace)
	}
	cmd := exec.Command("telepresence", args...)
	var stderr, stout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stout
	_ = sendDebug(fmt.Sprintf("intallation command: telepresence %s", strings.Join(args, " ")))
	err := cmd.Run()
	if err != nil {
		_ = sendErrorf("telepresence helm install failed: %s: %s", err, stderr.String())
		return err
	}
	_ = sendInfo(stout.String())
	return nil
}

func createIntercept(opts PluginOptions) error {
	_ = sendInfo(fmt.Sprintf("Creating intercept %q ...\n", opts.Name))
	args := []string{"intercept", opts.Name}

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
