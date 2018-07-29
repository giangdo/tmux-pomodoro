package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
	"github.com/giangdo/tmux-pomodoro/tmux"
)

const timeFormat = time.RFC3339

var duration, _ = time.ParseDuration("30m")
var playDuration, _ = time.ParseDuration("15m")
var noTime time.Time
var notify *notificator.Notificator

const usage = `
github.com/giangdo/tmux-pomodoro

  pomodoro start   Start a timer for 30 minutes
  pomodoro status  Show number of done pomodoro and status of current pomodoro timer
  pomodoro clear   Clear the current pomodoro timer
  pomodoro reset   Reset number of done pomodoro to 0
`
const version = "v1.2.1"

// State is the state of the world passed through the functions to determine
// side-effects.
type State struct {
	endTime time.Time
	now     time.Time
}

// Output has fields for functions to set/append when they intend to output to
// the user.
type Output struct {
	text       string
	returnCode int
}

func init() {
	flag.Usage = func() {
		fmt.Printf("tmux-pomodoro %s\n", version)
		fmt.Printf("%s\n", strings.TrimSpace(usage))
	}

	flag.Parse()
}

func main() {
	state := State{
		endTime: readExistingTime(),
		now:     time.Now(),
	}

	args := flag.Args()
	var command string
	if len(args) == 0 {
		command = ""
	} else {
		command = args[0]
	}

	notify = notificator.New(notificator.Options{
		AppName: "tmux-pomodoro",
	})

	newState, output := parseCommand(state, command)

	if newState.endTime != state.endTime {
		writeTime(newState.endTime)
	}

	if output.text != "" {
		fmt.Println(output.text)
	}

	if output.returnCode != 0 {
		os.Exit(output.returnCode)
	}
}

func refreshTmux() {
	_ = tmux.RefreshClient("-S")
}

func parseCommand(state State, command string) (newState State, output Output) {
	newState = state

	switch command {
	case "start":
		newState.endTime = state.now.Add(duration)
		output.text = "Timer started, 30 minutes remaining"
		notify.Push("Pomodoro", output.text, "", notificator.UR_NORMAL)
		_, _ = exec.Command("say", output.text).Output()

		killRunningBeepers()
		_ = startBeeper()
		refreshTmux()
	case "status":
		if state.endTime == noTime {
			return
		}
		output.text = formatRemainingTime(state.endTime, state.now)
	case "clear":
		newState.endTime = noTime
		output.text = "Pomodoro cleared!"
		killRunningBeepers()
		refreshTmux()
	case "beep":
		<-time.NewTicker(duration).C
		var message = "Pomodoro done, take a break!"
		_ = tmux.DisplayMessage(message)
		notify.Push("Pomodoro", message, "", notificator.UR_NORMAL)
		_, _ = exec.Command("say", message).Output()

		logPomoDone()
		refreshTmux()
	case "reset":
		cleanPomoDone()
	case "":
		flag.Usage()
	default:
		flag.Usage()
		output.returnCode = 1
	}

	return
}

func cleanPomoDone() {
	bytes := []byte("0")
	err := ioutil.WriteFile(fileDonePath(), bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func logPomoDone() {
	num, err := getPomodoDone()
	if err != nil {
		panic(err)
	}

	num++

	str := strconv.Itoa(num)
	bytes := []byte(str)
	err = ioutil.WriteFile(fileDonePath(), bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func getPomodoDone() (num int, err error) {
	bytes, err := ioutil.ReadFile(fileDonePath())
	if err != nil {
		bytes = []byte("0")
		err = ioutil.WriteFile(fileDonePath(), bytes, 0644)
		if err != nil {
			return
		}
	}

	str := string(bytes)
	re := regexp.MustCompile(`\r?\n`)
	str = re.ReplaceAllString(str, " ")
	num, err = strconv.Atoi(str)
	return
}

func startBeeper() (err error) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	command := exec.Command(ex, "beep")
	err = command.Start()
	if err != nil {
		log.Println(err)
		return
	}
	bytes := []byte(strconv.Itoa(command.Process.Pid))
	err = ioutil.WriteFile(pidFilePath(), bytes, 0644)
	if err != nil {
		log.Println(err)
	}

	return
}

func killRunningBeepers() {
	bytes, err := ioutil.ReadFile(pidFilePath())
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(string(bytes[:]))
	if err != nil {
		return
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}
	_ = process.Kill()
}

func formatRemainingTime(existingTime time.Time, now time.Time) string {
	remaining := existingTime.Sub(now)
	remainingMinutes := remaining.Minutes()

	num, err := getPomodoDone()
	if err != nil {
		panic(err)
	}
	out := strconv.Itoa(num) + "|"
	if remainingMinutes >= 0 {
		return out + "▼ " + strconv.FormatFloat(remainingMinutes, 'f', 0, 64) + " work "
	} else {
		excess := now.Sub(existingTime)
		excessMinutes := excess.Minutes()
		if excessMinutes <= 5 {
			// display remain break time in total 5min
			return out + "▼ " + strconv.FormatFloat(5-excessMinutes, 'f', 0, 64) + " break"
		} else {
			// display the time excess after finish 5min break time
			return out + "▲ " + strconv.FormatFloat(excessMinutes-5, 'f', 0, 64) + " !!!"
		}
	}
}

func writeTime(t time.Time) {
	var bytes []byte
	if t != noTime {
		bytes = []byte(t.Format(timeFormat))
	}
	err := ioutil.WriteFile(filePath(), bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readExistingTime() time.Time {
	bytes, err := ioutil.ReadFile(filePath())
	if err != nil {
		return noTime
	}

	contents := string(bytes[:])
	contents = strings.TrimSpace(contents)

	result, err := time.Parse(timeFormat, contents)
	if err != nil {
		return noTime
	}

	return result
}

func filePath() string {
	return homeDir() + "/.pomodoro"
}

func fileDonePath() string {
	return homeDir() + "/.pomodoro_done"
}
func pidFilePath() string {
	return homeDir() + "/.pomodoro.pid"
}

func homeDir() string {
	return os.Getenv("HOME")
}
