package main

import (
	"log"

	flag "github.com/spf13/pflag"
)

var (
	provider           string
	action             string
	project            string
	username           string
	token              string
	downloadToolPath   string
	keep               string
	download           string
	saveDir            string
	downloadNameFilter string
)

type IProvider interface {
	ListArtifacts()
	DeleteArtifacts()
	DownloadArtifacts()
}

func main() {
	flag.StringVarP(&username, "username", "u", "", "Username of the project")
	flag.StringVarP(&token, "token", "t", "", "Token of the service provider")
	flag.StringVarP(&project, "project", "r", "", "Name of project to be operated on")
	flag.StringVarP(&action, "action", "a", "list", "Take action, candidate: list, delete, download")
	flag.StringVarP(&provider, "provider", "p", "github", "Service provider, candidate: gihtub, appveyor")
	flag.StringVarP(&downloadToolPath, "downloader", "d", "", "Donwload tool path, supports aria2, curl, wget")
	flag.StringVarP(&download, "download", "", "", "Download artifacts, can be count number or today")
	flag.StringVarP(&keep, "keep", "k", "0", "Keep artifacts that won't be deleted, can be count number or today")
	flag.StringVarP(&saveDir, "download-directory", "", ".", "Download file to the specified directory")
	flag.StringVarP(&downloadNameFilter, "download-name-filter", "f", "", "Download name filter, it supports regular expression pattern")
	flag.Parse()

	var handler IProvider
	switch provider {
	case "github":
		handler = &Github{}
	case "appveyor":
		handler = &Appveyor{}
	default:
		log.Fatal("unsupported provider")
	}
	switch action {
	case "list":
		handler.ListArtifacts()
	case "delete":
		handler.DeleteArtifacts()
	case "download":
		handler.DownloadArtifacts()
	default:
		log.Fatal("unsupported action")
	}
}
