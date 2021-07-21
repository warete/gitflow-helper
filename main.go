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

func getLastVersion() (string, error) {
	var lastTag, errBuffer bytes.Buffer
	cmd := exec.Command("git", "describe", "--abbrev=0")
	cmd.Stdout = &lastTag
	cmd.Stderr = &errBuffer
	err := cmd.Run()
	if err != nil {
		return "", errors.New(errBuffer.String())
	}
	return lastTag.String(), nil
}

func getNextReleaseNumber() (string, error) {
	lastTag, err := getLastVersion()
	if err != nil {
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

	var errBuffer bytes.Buffer

	newVersion, err := getNextReleaseNumber()
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "hotfix":
		fallthrough
	case "release":
		gitflowCommandBuffer := fmt.Sprintf("git flow %s start %s", os.Args[1], newVersion)
		parts := strings.Split(gitflowCommandBuffer, " ")

		var gitflowResult bytes.Buffer
		gitflowCmd := exec.Command(parts[0], parts[1:]...)
		gitflowCmd.Stdout = &gitflowResult
		gitflowCmd.Stderr = &errBuffer
		err = gitflowCmd.Run()
		if err != nil {
			log.Fatal(errBuffer.String())
		}
		fmt.Println(gitflowResult.String())
		break
	default:
		log.Fatal("unknown gitflow action")
	}
}
