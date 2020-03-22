package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

func curl(path string, u string, output string) error {
	cmd := exec.Command(path, "-L", "--user", fmt.Sprintf("%s:%s", username, token), "-o", output, u)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("start downloading", u, "by curl, save to", output)
	return nil
}

func wget(path string, u string, output string) error {
	cmd := exec.Command(path, fmt.Sprintf("--http-user=%s", username), fmt.Sprintf("--http-password=%s", token), "-O", output, u)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("start downloading", u, "by wget, save to", output)
	return nil
}

func aria2(path string, u string, output string) error {
	cmd := exec.Command(path, fmt.Sprintf("--http-user=%s", username), fmt.Sprintf("--http-passwd=%s", token), "-o", output, u)
	err := cmd.Start()
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("start downloading", u, "by aria2, save to", output)
	return nil
}

func Download(tool string, u string, output string) error {
	switch filepath.Base(tool) {
	case "aria2c", "aria2c.exe":
		return aria2(tool, u, output)
	case "curl", "curl.exe":
		return curl(tool, u, output)
	case "wget", "wget.exe":
		return wget(tool, u, output)
	default:
		return fmt.Errorf("not supported tool: %s", tool)
	}
	return nil
}
