package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("==> Executing: %s\n", strings.Join(cmd.Args, " "))
}

func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}
func move(x, y string) {
	fmt.Println(x)
	fmt.Println(y)
	q := fmt.Sprintf("xdotool mousemove_relative --  %[1]s %[2]s ", x, y)
	fmt.Println(q)
	cmd := exec.Command("bash", "-c", q)
	var waitStatus syscall.WaitStatus
	if err := cmd.Run(); err != nil {
		fmt.Println("There was an error")
		fmt.Println(err)
		printError(err)
		// Did the command fail because of an unsuccessful exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
			printOutput([]byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
		}
	} else {
		// Command was successful
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
		printOutput([]byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
	}
}

//JobInterface something
type JobInterface interface {
	run() string
}

//Job struct something
type Job struct {
	command string
	param   string
}

//SmallJob struct ish
type SmallJob struct{ Job }

//LargeJob struct ish
type LargeJob struct{ Job }

//InvalidJob struct ish
type InvalidJob struct{ Job }

func (job SmallJob) run() string {
	command := job.command

	param := job.param
	switch command {
	case "right":
		move(param, "0")
	case "up":
		move("0", " -"+param)
	case "down":
		move("0", param)
	case "left":
		move("-"+param, "0")
	}

	return "done with param = " + job.param
}

func (job LargeJob) run() string {
	time.Sleep(5 * time.Second)
	return "Completed in 5 second with param = " + job.param
}

func (job InvalidJob) run() string {
	return "Invalid command is specified"
}

func jobRunner(job JobInterface, out chan string) {
	out <- job.run() + "\n"
}

func jobFactory(input string) JobInterface {
	array := strings.Split(input, " ")
	if len(array) >= 2 {
		command := array[0]
		param := array[1]
		switch command {
		case "left", "right", "up", "down":
			return SmallJob{Job{
				param:   param,
				command: command,
			}}
		}
	}
	return InvalidJob{Job{param: ""}}

}
func requestHandler(conn net.Conn, out chan string) {
	defer close(out)

	for {
		line, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			return
		}

		job := jobFactory(strings.TrimRight(string(line), "\n"))
		go jobRunner(job, out)
	}
}

func sendData(conn net.Conn, in <-chan string) {
	defer conn.Close()

	for {
		message := <-in
		log.Print(message)
		io.Copy(conn, bytes.NewBufferString(message))
	}
}

func main() {
	psock, err := net.Listen("tcp", ":5000")
	if err != nil {
		return
	}

	for {
		conn, err := psock.Accept()
		if err != nil {
			return
		}

		channel := make(chan string)
		go requestHandler(conn, channel)
		go sendData(conn, channel)
	}
}
