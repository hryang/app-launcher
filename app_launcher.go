package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetParentAndChildPids(parentId int) []int {
	cmd := exec.Command("pstree", "-p", fmt.Sprintf("%d", parentId))
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to run pstree. Error: %v\n", err)
	}
	str := string(output)
	log.Printf("Process tree: \n%s", str)

	procs := strings.Split(strings.TrimSpace(str), "\n")
	ret := make([]int, 0, len(procs))
	isChild := false
	for _, pstr := range procs {
		a := strings.Fields(pstr)
		pid, _ := strconv.Atoi(a[1])
		if pid == parentId {
			isChild = true
		}
		if isChild {
			ret = append(ret, pid)
		}
	}
	return ret
}

func Launch(appCmd string, n time.Duration) {
	app := exec.Command(os.Getenv("SHELL"), "-c", appCmd)
	err := app.Start()
	if err != nil {
		log.Printf("Error launching %s: %v\n", appCmd, err)
		return
	}
	log.Printf("Launch %s. Pid: %d\n", appCmd, app.Process.Pid)
	time.Sleep(n)
	pids := GetParentAndChildPids(app.Process.Pid)
	for _, pid := range pids {
		cmd := exec.Command("kill", "-9", fmt.Sprintf("%d", pid))
		err := cmd.Run()
		if err != nil {
			log.Printf("Error killing pid %d: %v\n", pid, err)
		} else {
			log.Printf("Kill pid %d\n", pid)
		}
	}
	app.Wait()
	alert := exec.Command("/usr/bin/osascript", "-e", "display alert \"警告: 你现在不能玩 Minecraft！\n黑发不知勤学早，老来方知读书迟！\" as critical")
	alert.Run()
}

func Quota(in time.Duration) time.Duration {
	now := time.Now()
	var quota time.Duration
	if time.Saturday == now.Weekday() || time.Sunday == now.Weekday() {
		quota = 30 * time.Minute
	} else {
		quota = 1 * time.Minute
	}

	if in < quota {
		return in
	} else {
		return quota
	}
}

func main() {
	appCmd := flag.String("app", "", "the app start command")
	duration := flag.Int("time", 1, "the max app running time, in seconds")

	flag.Parse()

	log.Printf("===============================================\n")
	log.Printf("Start launcher: %d\n", os.Getpid())

	quota := Quota(time.Duration(*duration) * time.Second)
	log.Printf("Time quota: %v\n", quota)

	Launch(*appCmd, quota)
}
