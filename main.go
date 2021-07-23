package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func checkIsBranchExists(branchName string) bool {
	var result bytes.Buffer
	cmd := exec.Command("git", "branch", "--list", branchName)
	cmd.Stdout = &result
	_ = cmd.Run()
	return len(result.String()) > 0
}

func execCommand(command string) (string, error) {
	var stdOutBuff, stdErrBuff bytes.Buffer
	cmdParts := strings.Split(command, " ")
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = &stdOutBuff
	cmd.Stderr = &stdErrBuff
	err := cmd.Run()
	if err != nil {
		return "", errors.New(stdErrBuff.String())
	}

	return stdOutBuff.String(), nil
}

func getLastVersion() (string, error) {
	lastTag, err := execCommand("git describe --abbrev=0")
	if err != nil {
		return "", err
	}
	return lastTag, nil
}

func getNextReleaseNumber() (string, error) {
	lastTag, err := getLastVersion()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	lastTagParts := strings.Split(lastTag, ".")

	currentDate := time.Now().Format("2006.01.02")

	nextReleaseNumber := 0
	if currentDate == strings.Join(lastTagParts[:3], ".") {
		if lastTagParts[3][0] == '0' {
			nextReleaseNumber, _ = strconv.Atoi(string(lastTagParts[3][1]))
		} else {
			nextReleaseNumber, _ = strconv.Atoi(lastTagParts[3])
		}
	}
	var newVersion string
	for {
		nextReleaseNumber += 1
		newVersion = fmt.Sprintf("%s.%02d", currentDate, nextReleaseNumber)
		if !checkIsBranchExists("release/"+newVersion) && !checkIsBranchExists("hotfix/"+newVersion) {
			break
		}
	}

	return newVersion, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("empty gitflow action")
	}
	switch os.Args[1] {
	case "hotfix":
		fallthrough
	case "release":
		newVersion, err := getNextReleaseNumber()
		if err != nil {
			log.Fatal(err)
		}
		gitflowCommandBuffer := fmt.Sprintf("git flow %s start %s", os.Args[1], newVersion)
		gitflowResult, err := execCommand(gitflowCommandBuffer)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(gitflowResult)
		break
	default:
		log.Fatal("unknown gitflow action")
	}
}
