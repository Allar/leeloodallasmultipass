package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"log"
	"os"
	"os/exec"
	"strings"
)

type multipassEntry struct {
	Hostname string
	IP       string
	State    string
	Image    string
	Raw      string
}

func getMultipassNodes() []multipassEntry {
	out, err := exec.Command("multipass", "list").Output()
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")[1:]
	lines = lines[:len(lines)-1]

	var entries []multipassEntry = nil
	for _, line := range lines {
		entry := strings.Fields(line)
		log.Println(entry)
		entries = append(entries, multipassEntry{
			Hostname: entry[0],
			IP:       entry[2],
			State:    entry[1],
			Image:    entry[3],
			Raw:      line,
		})
	}

	return entries
}

func performNodeAction(node multipassEntry) {

	var options []string

	// #TODO: Learn all the states Multipass nodes can be in
	switch node.State {
	case "Running":
		options = []string{"shell", "stop", "restart", "delete"}
		break
	case "Stopped":
		options = []string{"start", "delete"}
		break
	case "Deleted":
		options = []string{"recover", "purge"}
		break
	default:
		options = []string{"delete"}
	}

	options = append(options, "exit")

	var option string
	prompt := &survey.Select{
		Message: "What to do with " + node.Hostname + "?",
		Options: options,
	}
	err := survey.AskOne(prompt, &option)
	if err == terminal.InterruptErr {
		os.Exit(1)
	}

	switch option {
	case "shell":
		cmd := exec.Command("multipass", "shell", node.Hostname)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		cmd.Wait()
		os.Exit(0)
		return
	case "purge":
		_, err := exec.Command("multipass", option).Output()
		if err != nil {
			log.Fatal(err)
		}
		mainMenu()
	case "exit":
		mainMenu()
		return
	default:
		_, err := exec.Command("multipass", option, node.Hostname).Output()
		if err != nil {
			log.Fatal(err)
		}
		mainMenu()
		return
	}

}

func mainMenu() {
	nodes := getMultipassNodes()

	var options []string
	for _, node := range nodes {
		options = append(options, node.Raw)
	}
	options = append(options, "launch a new instance", "exit")

	var optionIndex int
	prompt := &survey.Select{
		Message: "Choose a Multipass Node:",
		Options: options,
	}
	err := survey.AskOne(prompt, &optionIndex)
	if err == terminal.InterruptErr {
		os.Exit(1)
	}

	// exit
	if optionIndex == len(options)-1 {
		os.Exit(0)
	}

	// launch
	// #TODO create a nice launch menu full of options like cpu core count
	if optionIndex == len(options)-2 {
		_, err := exec.Command("multipass", "launch").Output()
		if err != nil {
			log.Fatal(err)
		}
		mainMenu()
	}

	performNodeAction(nodes[optionIndex])
}

func main() {
	mainMenu()

}
