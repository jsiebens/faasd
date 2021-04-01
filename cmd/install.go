package cmd

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/openfaas/faasd/pkg/assets"
	systemd "github.com/openfaas/faasd/pkg/systemd"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install faasd",
	RunE:  runInstall,
}

const workingDirectoryPermission = 0644

const faasdwd = "/var/lib/faasd"

const faasdProviderWd = "/var/lib/faasd-provider"

func runInstall(_ *cobra.Command, _ []string) error {

	if err := ensureWorkingDir(path.Join(faasdwd, "secrets")); err != nil {
		return err
	}

	if err := ensureWorkingDir(faasdProviderWd); err != nil {
		return err
	}

	if basicAuthErr := makeBasicAuthFiles(path.Join(faasdwd, "secrets")); basicAuthErr != nil {
		return errors.Wrap(basicAuthErr, "cannot create basic-auth-* files")
	}

	if err := cp(assets.DockerCompose, path.Join(faasdwd, "docker-compose.yaml")); err != nil {
		return err
	}

	if err := cp(assets.PrometheusConfig, path.Join(faasdwd, "prometheus.yml")); err != nil {
		return err
	}

	if err := cp(assets.ResolvConf, path.Join(faasdwd, "resolv.conf")); err != nil {
		return err
	}

	err := binExists("/usr/local/bin/", "faasd")
	if err != nil {
		return err
	}

	err = systemd.InstallUnit(assets.FaasdProviderService, "faasd-provider", map[string]string{
		"Cwd":             faasdProviderWd,
		"SecretMountPath": path.Join(faasdwd, "secrets")})

	if err != nil {
		return err
	}

	err = systemd.InstallUnit(assets.FaasdService, "faasd", map[string]string{"Cwd": faasdwd})
	if err != nil {
		return err
	}

	err = systemd.DaemonReload()
	if err != nil {
		return err
	}

	err = systemd.Enable("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Enable("faasd")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd-provider")
	if err != nil {
		return err
	}

	err = systemd.Start("faasd")
	if err != nil {
		return err
	}

	fmt.Println(`Check status with:
  sudo journalctl -u faasd --lines 100 -f

Login with:
  sudo cat /var/lib/faasd/secrets/basic-auth-password | faas-cli login -s`)

	return nil
}

func binExists(folder, name string) error {
	findPath := path.Join(folder, name)
	if _, err := os.Stat(findPath); err != nil {
		return fmt.Errorf("unable to stat %s, install this binary before continuing", findPath)
	}
	return nil
}

func ensureWorkingDir(folder string) error {
	if _, err := os.Stat(folder); err != nil {
		err = os.MkdirAll(folder, workingDirectoryPermission)
		if err != nil {
			return err
		}
	}

	return nil
}

func cp(content []byte, destination string) error {
	return ioutil.WriteFile(destination, content, 0644)
}
