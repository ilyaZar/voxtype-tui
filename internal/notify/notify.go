package notify

import (
	"os/exec"
)

func Send(urgency string, timeout string, title string, message string) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return
	}
	_ = exec.Command("notify-send", "-u", urgency, "-t", timeout, title, message).Run()
}
