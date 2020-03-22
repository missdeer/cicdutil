package main

type Appveyor struct {
	username   string
	token      string
	repository string
	keep       string
	download   string
}

func (av *Appveyor) ListArtifacts() {

}

func (av *Appveyor) DeleteArtifacts() {

}

func (av *Appveyor) DownloadArtifacts() {

}
