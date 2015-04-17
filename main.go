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

func launch(c string) {
	var q string
	switch c {
	case "browser":
		q = "google-chrome-stable"
	case "music":
		q = ""
	}
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

func openURL(URL string) {
	q := fmt.Sprintf("xdg-open http://%s", URL)
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

func moveAngle(x, y string) {
	fmt.Println(x)
	fmt.Println(y)
	q := fmt.Sprintf("xdotool mousemove_relative --polar  %[1]s %[2]s ", x, y)
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
	param1  string
	param2  string
}

//SmallJob struct ish
type SmallJob struct{ Job }

//LargeJob struct ish
type LargeJob struct{ Job }

//InvalidJob struct ish
type InvalidJob struct{ Job }

func (job SmallJob) run() string {
	command := job.command

	switch command {
	case "move":
		switch job.param1 {
		case "right":
			move(job.param2, "0")
		case "up":
			move("0", " -"+job.param2)
		case "down":
			move("0", job.param2)
		case "left":
			move("-"+job.param2, "0")
		}
	case "angle":
		/*	x, err := strconv.Atoi(job.param2)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(job.param2)
			fmt.Println(x)
			for i := 1; i <= x; i++ {
				moveAngle(job.param1, "1")
				fmt.Println(i)
			}
		*/
		moveAngle(job.param1, job.param2)
	case "open":
		openURL(job.param1)
	case "launch":
		launch(job.param1)
	}

	return "done with param = " + job.param1
}

func (job LargeJob) run() string {
	time.Sleep(5 * time.Second)
	return "Completed in 5 second with param = " + job.param1
}

func (job InvalidJob) run() string {
	return "Invalid command is specified"
}

func jobRunner(job JobInterface, out chan string) {
	out <- job.run() + "\n"
}

func jobFactory(input string) JobInterface {
	array := strings.Split(input, ",")
	if len(array) >= 2 {
		command := array[0]
		param1 := array[1]
		param2 := strings.TrimSuffix(array[2], "\r")
		switch command {
		case "move", "angle", "open", "launch":
			return SmallJob{Job{
				param1:  param1,
				param2:  param2,
				command: command,
			}}
		}
	}
	return InvalidJob{Job{param1: ""}}

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
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
		os.Exit(1)
	}
	fmt.Println("Your System IP address(es) is(are): ")
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				os.Stdout.WriteString(ipnet.IP.String() + "\n")
			}
		}
	}
	fmt.Println("While the port is 5000")
	psock, err := net.Listen("tcp", ":5000")

	if err != nil {
		return
	}

	for {
		conn, err := psock.Accept()
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Print("conection from : ")
		fmt.Print(conn.RemoteAddr())
		fmt.Println(conn.LocalAddr())
		channel := make(chan string)
		go requestHandler(conn, channel)
		go sendData(conn, channel)
	}
}
