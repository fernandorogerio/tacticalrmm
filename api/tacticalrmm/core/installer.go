package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	Inno        string
	Api         string
	Client      string
	Site        string
	Atype       string
	Power       string
	Rdp         string
	Ping        string
	Token       string
	DownloadUrl string
)

func downloadAgent(filepath string) (err error) {

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(DownloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bad response: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	debugLog := flag.String("log", "", "Verbose output")
	localSalt := flag.String("local-salt", "", "Use local salt minion")
	localMesh := flag.String("local-mesh", "", "Use local mesh agent")
	flag.Parse()

	var debug bool = false

	if strings.TrimSpace(*debugLog) == "DEBUG" {
		debug = true
	}

	agentBinary := filepath.Join(os.Getenv("windir"), "Temp", Inno)
	tacrmm := filepath.Join(os.Getenv("PROGRAMFILES"), "TacticalAgent", "tacticalrmm.exe")

	cmdArgs := []string{
		"/C", tacrmm,
		"-m", "install", "--api", Api, "--client-id",
		Client, "--site-id", Site, "--agent-type", Atype,
		"--auth", Token,
	}

	if debug {
		cmdArgs = append(cmdArgs, "--log", "DEBUG")
	}

	if len(strings.TrimSpace(*localSalt)) != 0 {
		cmdArgs = append(cmdArgs, "--local-salt", *localSalt)
	}

	if len(strings.TrimSpace(*localMesh)) != 0 {
		cmdArgs = append(cmdArgs, "--local-mesh", *localMesh)
	}

	if Rdp == "1" {
		cmdArgs = append(cmdArgs, "--rdp")
	}

	if Ping == "1" {
		cmdArgs = append(cmdArgs, "--ping")
	}

	if Power == "1" {
		cmdArgs = append(cmdArgs, "--power")
	}

	if debug {
		fmt.Println("Installer:", agentBinary)
		fmt.Println("Tactical Agent:", tacrmm)
		fmt.Println("Download URL:", DownloadUrl)
		fmt.Println("Install command:", "cmd.exe", strings.Join(cmdArgs, " "))
	}

	fmt.Println("Downloading agent...")
	dl := downloadAgent(agentBinary)
	if dl != nil {
		fmt.Println("ERROR: unable to download agent from", DownloadUrl)
		fmt.Println(dl)
		os.Exit(1)
	}

	fmt.Println("Extracting files...")
	winagentCmd := exec.Command(agentBinary, "/VERYSILENT", "/SUPPRESSMSGBOXES")
	err := winagentCmd.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(20 * time.Second)

	fmt.Println("Installation starting.")
	cmd := exec.Command("cmd.exe", cmdArgs...)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	e := os.Remove(agentBinary)
	if e != nil {
		fmt.Println(e)
	}
}
