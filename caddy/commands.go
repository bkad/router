package caddy

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Start Caddy.
func Start() error {
	log.Println("INFO: Starting caddy...")
	cmd := exec.Command("caddy", "--pidfile=/var/run/caddy.pid", "--conf=/opt/router/Caddyfile", "--log=stdout", "--agree=true", "--port=80")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	log.Println("INFO: caddy started.")
	return nil
}

// Reload caddy configuration.
func Reload() error {
	log.Println("INFO: Reloading caddy...")
	contents, err := ioutil.ReadFile("/var/run/caddy.pid")
	if err != nil {
		log.Println("ERROR: Failed to read caddy's pid file.")
		log.Println(err)
		return err
	}
	if len(contents) > 0 {
		pid, err := strconv.Atoi(strings.TrimSpace(string(contents)))
		if err != nil {
			log.Println("ERROR: Failed to convert caddy pid file to integer.")
			log.Println(err)
			return err
		}
		err = syscall.Kill(pid, syscall.SIGUSR1)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	contents, err = ioutil.ReadFile("/opt/router/Caddyfile")
	if err != nil {
		log.Println("ERROR: Failed to read caddyfile.")
		log.Println(err)
		return err
	}
	if len(contents) > 0 {
		log.Println(string(contents))
	}

	log.Println("INFO: caddy reloaded.")
	return nil
}
