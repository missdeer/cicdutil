package main

import (
	"log"

	flag "github.com/spf13/pflag"
)

var (
	provider         string
	action           string
	repository       string
	username         string
	token            string
	downloadToolPath string
	keep             string
	download         string
)

type IProvider interface {
	ListArtifacts()
	DeleteArtifacts()
	DownloadArtifacts()
}

func main() {
	flag.StringVarP(&username, "username", "u", "", "Username of the repository")
	flag.StringVarP(&token, "token", "t", "", "Token of the service provider")
	flag.StringVarP(&repository, "repository", "r", "", "Name of repository to be operated on")
	flag.StringVarP(&action, "action", "a", "", "Take action, candidate: list, delete, download")
	flag.StringVarP(&provider, "provider", "p", "", "Service provider, candidate: gihtub, appveyor")
	flag.StringVarP(&downloadToolPath, "downloader", "d", "", "Donwload tool path, supports aria2, curl, wget")
	flag.StringVarP(&download, "download", "", "", "Download artifacts, can be count number or today")
	flag.StringVarP(&keep, "keep", "k", "", "Keep artifacts that won't be deleted, can be count number or today")
	flag.Parse()

	var handler IProvider
	switch provider {
	case "github":
		handler = &Github{
			username:   username,
			token:      token,
			repository: repository,
			download:   download,
			keep:       keep,
		}
	case "appveyor":
		handler = &Appveyor{
			username:   username,
			token:      token,
			repository: repository,
			download:   download,
			keep:       keep,
		}
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
