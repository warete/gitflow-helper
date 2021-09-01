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

	"github.com/fatih/color"
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
		lastTag = ""
	}

	lastTagParts := strings.Split(lastTag, ".")

	currentDate := time.Now().Format("2006.01.02")

	nextReleaseNumber := 0
	if len(lastTagParts) > 1 && currentDate == strings.Join(lastTagParts[:3], ".") {
		if lastTagParts[3][0] == '0' {
			nextReleaseNumber, _ = strconv.Atoi(string(lastTagParts[3][1]))
		} else {
			nextReleaseNumber, _ = strconv.Atoi(lastTagParts[3][:len(lastTagParts[3])-1])
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

func startGitflowReleaseAction(action string) (string, string, error) {
	newVersion, err := getNextReleaseNumber()
	if err != nil {
		return "", "", err
	}
	gitflowCommandBuffer := fmt.Sprintf("git flow %s start %s", action, newVersion)
	gitflowResult, err := execCommand(gitflowCommandBuffer)
	if err != nil {
		return "", "", err
	}

	return gitflowResult, newVersion, nil
}

func finishGitflowAction(action string, version string) (string, error) {
	gitflowCommandBuffer := fmt.Sprintf("git flow %s finish %s -m \"Tagging version %s\"", action, version, version)
	gitflowResult, err := execCommand(gitflowCommandBuffer)
	if err != nil {
		return "", err
	}

	return gitflowResult, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("empty gitflow action")
	}

	log.SetOutput(color.Output)
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	_, err := execCommand("git fetch origin")
	if err != nil {
		log.Println(red(err))
	}

	switch os.Args[1] {
	case "hotfix":
		fallthrough
	case "release":
		gitflowResult, _, err := startGitflowReleaseAction(os.Args[1])
		if err != nil {
			log.Fatal(red(err))
		}
		log.Println(green(gitflowResult))
		break
	case "fast_release":
		gitflowResult, newVersion, err := startGitflowReleaseAction("hotfix")
		if err != nil {
			log.Fatal(red(err), err.Error())
		}
		log.Println(green(gitflowResult))
		gitflowResult, err = finishGitflowAction("hotfix", newVersion)
		log.Println(green(gitflowResult))
		gitflowResult, newVersion, err = startGitflowReleaseAction("release")
		if err != nil {
			log.Fatal(red(err))
		}
		log.Println(green(gitflowResult))
		gitflowResult, err = finishGitflowAction("release", newVersion)
		log.Println(green(gitflowResult))
	case "merge_cur_to_stage":
		log.Println("start check exists feature/stage branch")
		checkBranchResult, err := execCommand("git branch --list feature/stage")
		if err != nil {
			log.Fatal(red(err))
		}
		if len(checkBranchResult) == 0 {
			log.Fatal(red("branch feature/stage does not exist"))
		} else {
			log.Println(green("branch feature/stage exists"))
		}

		log.Println("start getting current branch name")
		currentBranchName, err := execCommand("git branch --show-current")
		if err != nil {
			log.Fatal(red(err))
		}
		currentBranchName = currentBranchName[:len(currentBranchName)-1]
		log.Println(green("current branch name: " + currentBranchName))
		log.Println("git checkout feature/stage")
		_, err = execCommand("git checkout feature/stage")
		if err != nil {
			log.Fatal(red(err))
		}
		log.Println(fmt.Sprintf("git merge %s", currentBranchName))
		mergeResult, err := execCommand(fmt.Sprintf("git merge %s", currentBranchName))
		if err != nil {
			_, err = execCommand("git checkout " + currentBranchName)
			log.Fatal(red(err))
		}
		log.Println(green(mergeResult))
		_, err = execCommand("git checkout " + currentBranchName)
	default:
		log.Fatal("unknown gitflow action")
	}
}
