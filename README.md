# tmux-pomodoro [![Build Status](https://travis-ci.org/justincampbell/tmux-pomodoro.svg?branch=conversion)](https://travis-ci.org/justincampbell/tmux-pomodoro)

## Installation

1. Download the latest package for your platform from the [Releases page](https://github.com/justincampbell/tmux-pomodoro/releases/latest).
2. Untar the package with `tar -zxvf tmux-pomodoro*.tar.gz`.
3. Move the extracted `pomodoro` file to a directory in your `$PATH` (for most systems, this will be `/usr/local/bin/`).

Or, if you have a [Go development environment](https://golang.org/doc/install):

```
go get github.com/justincampbell/tmux-pomodoro
```

## Usage

### Tmux Configuration

```tmux
# Place the current pomodoro status on the right side of your status bar
set -g status-right '#(pomodoro status)'

# Map a key to start a timer
bind-key p run-shell 'pomodoro start'
```

### Commands

* `start` Start a timer for 25 minutes
* `status` Show the remaining time, or an exclamation point if done
* `clear` Clear the timer

`start` and `clear` also call `tmux refresh-client -S`, which will instantly update your tmux status bar shell commands.

### For Development: Explain how this program works
+ "start"  command: this program will:
                        ->set end time base on get current time,  then store this end time in first line of ~/.pomodoro file
                        +> kill running beeper process (this program, but started with "beep" commands)
                        +> start another beeper process

+ "beep"   command:   - this is internal command, user can not use this command
                      - This command will be trigger when user use "start" command
                      - When program execute with this command, this process will wait until 30m, then it will
                        -> stop timer
                        +> notify :"Pomodoro done, take a break!"
                        +> increase number of done pomodo timer by 1, save this number to ~/.pomodoro_done
                      - then program wait for 5m (break time) then get into infinte loop to remind "It too late, Please start a new pomodoro!"
                      - this beeper program is killed only when user use command "start" or "clear"

+ "status" command: this program will:
                        -> show number of done pomodoro today by reading info from ~/.pomodoro_done
                        +> show status of current pomodoro timer
                            -> read the information when the timer will be stopedd from ~/.pomodoro file
                            +> get the current time
                            +> remain = end time - current time
                            +>if (remain < 30min)
                                -> status = work
                            +>if (-5min <= remain <= 0min)
                                -> status = break
                            +>if (remain < -5min)
                                -> status = excess

+ "stop"  command: this command will delete everything in ~/.pomodoro file
                   then kill running beeper process (this program, but started with "beep" commands)

+ "reset"  command: this command will delete everything in ~/.pomodoro_done then execute what "stop" command do
+ "cancel" command: this command will mark previous expired pomodoro as failure ->
                        decrease current number of done pomodoro today by 1
                        then execute what "stop" command do

### TODO:
    - create a "play" command to relax after had finished 4 or 5 pomodoro timer
    - show the total number of continuous successul pomodoro timer
