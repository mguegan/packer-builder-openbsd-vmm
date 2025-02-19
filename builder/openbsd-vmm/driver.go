package openbsdvmm

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/packer"
)

type Driver interface {
	bootcommand.BCDriver
	VmctlCmd(bool, ...string) error
	Start(...string) error
	Stop(string) error
	GetTapIPAddress(string) (string, error)
	GetVMId(string) string
	//Flush() error
}

type vmmDriver struct {
	doas    string
	vmctl   string
	logfile string
	tty     io.WriteCloser
	console int
	ui      packer.Ui
}

func (d *vmmDriver) GetVMId(name string) string {
	var stdout bytes.Buffer
	cmd := exec.Command("vmctl", "status", name)
	cmd.Stdout = &stdout
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("vmctl status error")
	}
	stdoutString := strings.TrimSpace(stdout.String())
	vmctl := regexp.MustCompile(`(\d+)`)
	resultarr := vmctl.FindAllStringSubmatch(stdoutString, -1)
	if resultarr == nil {
		return "VMAWOL"
	}
	return resultarr[0][1]
}

func (d *vmmDriver) VmctlCmd(usedoas bool, args ...string) error {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd
	if usedoas {
		args = append([]string{d.vmctl}, args...)
		//args = append([]string{"ktrace"}, args...)
		log.Printf("Executing doas: %#v", args)
		cmd = exec.Command(d.doas, args...)
	} else {
		//args = append([]string{"vmctl"}, args...)
		//cmd = exec.Command("ktrace", args...)
		log.Printf("Executing vmctl: %#v", args)
		cmd = exec.Command(d.vmctl, args...)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("vmctl error")
	}
	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)
	return err
}

// Start the VM and create a pipe to insert commands into the VM. (from packer-builder-vmm)
func (d *vmmDriver) Start(args ...string) error {
	//d.ui.Message("Logging console output to " + d.logfile)
	logFile, err := os.Create(d.logfile)
	if err != nil {
		return err
	}

	args = append([]string{d.vmctl}, args...)
	//d.ui.Message("Executing " + d.doas + " " + strings.Join(args, " "))

	cmd := exec.Command(d.doas, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Create an stdin pipe that is used to issue commands.
	if d.tty, err = cmd.StdinPipe(); err != nil {
		return err
	}

	// Write the console output to the log file.
	go func() {
		defer stdout.Close()
		defer logFile.Close()

		_, _ = io.Copy(logFile, stdout)
	}()

	// Start up the VM.
	if err := cmd.Start(); err != nil {
		return err
	}

	// Give the VM a bit of time to start up.
	time.Sleep(3 * time.Second)
	return nil
}

func (d *vmmDriver) Stop(name string) error {
	cmd := exec.Command(d.doas, d.vmctl, "stop", name)
	//err := cmd.Run()
	cmd.Run()
	return nil
}

func (d *vmmDriver) GetTapIPAddress(id string) (string, error) {
	var stdout bytes.Buffer
	tapn, _ := strconv.Atoi(id)
	tapn--
	tapname := fmt.Sprintf("tap%d", tapn)
	log.Printf("tapname: %s", tapname)
	cmd := exec.Command("ifconfig", tapname)
	cmd.Stdout = &stdout
	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("ifconfig error")
	}
	stdoutString := strings.TrimSpace(stdout.String())
	log.Printf("ifconfig: %s", stdoutString)
	vmctl := regexp.MustCompile(`inet (\d+\.\d+\.\d+\.\d+) `)
	resultarr := vmctl.FindAllStringSubmatch(stdoutString, -1)
	return resultarr[0][1], err
}

//// interface Seq requires the following, not using it so far
// SendKey sends a key press.
func (d *vmmDriver) SendKey(key rune, action bootcommand.KeyAction) error {
	data := []byte{byte(key)}

	_, err := d.tty.Write(data)
	return err
}

// SendSpecial sends a special character.
func (d *vmmDriver) SendSpecial(special string, action bootcommand.KeyAction) error {
	var data []byte

	switch special {
	case "enter":
		data = []byte("\n")
	}

	if len(data) != 0 {
		if _, err := d.tty.Write(data); err != nil {
			return err
		}
	}

	return nil
}
func (driver *vmmDriver) Flush() error {
	return nil
}
