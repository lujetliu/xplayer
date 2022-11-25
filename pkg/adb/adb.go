package adb

import (
	"fmt"
	"fyne.io/fyne/v2"
	"go-scrcpy/pkg/log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func Push(serial, local, remote string) error {
	return Exec(serial, "push", local, remote)
}

func OpenApp(serial, appInfo string) error {
	return Exec(serial, "shell", "am", "start", "-c", "android.intent.category.LAUNCHER",
		"-a",
		"android.intent.action.MAIN ", fmt.Sprintf("'%s'", appInfo))
}

func InputText(serial, text string) error {
	return Exec(serial, "shell", "am", "broadcast", "-a", "ADB_INPUT_TEXT", "--es",
		"msg", fmt.Sprintf("'%s'", text))
}

func Devices() ([]byte, error) {
	cmd := exec.Command(adbCmd, "devices")
	return cmd.Output()
}

func Install(serial, local string) error {
	return Exec(serial, "install", "-r", local)
}

func RemovePath(serial, path string) error {
	return Exec(serial, "shell", "rm", "-rf", path)
}

func Reverse(serial, sockName string, localPort int) error {
	return Exec(serial, "reverse",
		fmt.Sprintf("localabstract:%s", sockName),
		fmt.Sprintf("tcp:%d", localPort))
}

func ReverseRemove(serial, sockName string) error {
	return Exec(serial, "reverse", "--remove",
		fmt.Sprintf("localabstract:%s", sockName))
}

func Forward(serial string, localPort int, sockName string) error {
	return Exec(serial, "forward",
		fmt.Sprintf("tcp:%d", localPort),
		fmt.Sprintf("localabstract:%s", sockName))
}

func ForwardRemove(serial string, localPort int) error {
	return Exec(serial, "forward", "--remove",
		fmt.Sprintf("tcp:%d", localPort))
}

func Exec(serial string, params ...string) error {
	if cmd, err := ExecAsync(serial, params...); err != nil {
		return err
	} else {
		return cmd.Wait()
	}
}

type logWriter struct {
}

func (*logWriter) Write(p []byte) (n int, err error) {
	log.Info("%s", string(p))
	return len(p), nil
}

func ExecAsync(serial string, params ...string) (*exec.Cmd, error) {
	args := make([]string, 0, 8)
	if len(serial) > 0 {
		args = append(args, "-s", serial)
	}
	args = append(args, params...)

	adbCmdOnce.Do(getAdbCommand)
	log.Debug("执行 %s %s\n", adbCmd, strings.Join(args, " "))
	cmd := exec.Command(adbCmd, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

var adbCmd = "adb"
var adbCmdOnce sync.Once

func getAdbCommand() {
	adbEnv := fyne.CurrentApp().Preferences().StringWithFallback("adb_path", "")
	log.Info("adb env %s", adbEnv)
	if len(adbEnv) > 0 {
		adbCmd = adbEnv
	} else {
		adbCmd = "adb.exe"
	}
}
