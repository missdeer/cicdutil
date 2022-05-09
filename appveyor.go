package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Appveyor struct {
}

type AppveyorBuild struct {
	Project struct {
		ProjectID   int    `json:"projectId"`
		AccountID   int    `json:"accountId"`
		AccountName string `json:"accountName"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
	} `json:"project"`
	Build struct {
		BuildNumber int       `json:"buildNumber"`
		Version     string    `json:"version"`
		Status      string    `json:"status"`
		Created     time.Time `json:"created"`
		Jobs        []struct {
			JobID          string    `json:"jobId"`
			Name           string    `json:"name"`
			ArtifactsCount int       `json:"artifactsCount"`
			Status         string    `json:"status"`
			Created        time.Time `json:"created"`
		} `json:"jobs"`
	} `json:"build"`
}

type AppveyorArtifact struct {
	FileName string    `json:"fileName"`
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Size     int       `json:"size"`
	Created  time.Time `json:"created"`
}

func (av *Appveyor) list() ([]byte, error) {
	u := fmt.Sprintf(`https://ci.appveyor.com/api/projects/%s/%s`, username, project)
	client := &http.Client{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse artifacts list request:", err)
		return nil, err
	}
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + token},
		"Content-type":  []string{"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send artifacts list request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("artifacts list response not 200:", resp.Status)
		return nil, fmt.Errorf("artifacts list response not 200:", resp.Status)
	}

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading artifacts list failed", err)
		return nil, err
	}
	return c, nil
}

func (av *Appveyor) getJobArtifacts(jobID string) ([]byte, error) {
	u := fmt.Sprintf(`https://ci.appveyor.com/api/buildjobs/%s/artifacts`, jobID)
	client := &http.Client{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse artifacts list request:", err)
		return nil, err
	}
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + token},
		"Content-type":  []string{"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send artifacts list request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("artifacts list response not 200:", resp.Status)
		return nil, fmt.Errorf("artifacts list response not 200:", resp.Status)
	}

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading artifacts list failed", err)
		return nil, err
	}
	return c, nil
}

func (av *Appveyor) ListArtifacts() {
	c, err := av.list()
	if err == nil {
		var build AppveyorBuild
		err = json.Unmarshal(c, &build)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(build)
	}
}

func (av *Appveyor) DeleteArtifacts() {

}

func (av *Appveyor) Build() {
	u := `https://ci.appveyor.com/api/builds`
	client := &http.Client{}
	body := fmt.Sprintf(`{ "accountName": "%s", "projectSlug": "%s",  "branch": "master"	}`, username, project)
	req, err := http.NewRequest("POST", u, strings.NewReader(body))
	if err != nil {
		log.Println("Could not parse artifacts list request:", err)
		return
	}
	req.Header = http.Header{
		"Authorization": []string{"Bearer " + token},
		"Content-type":  []string{"application/json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send artifacts list request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("artifacts list response not 200:", resp.Status)
		return
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading artifacts list failed", err)
	}

	return
}

func (av *Appveyor) DownloadArtifacts() {
	c, err := av.list()
	if err != nil {
		return
	}

	var build AppveyorBuild
	err = json.Unmarshal(c, &build)
	if err != nil {
		log.Println(err)
		return
	}

	r := regexp.MustCompile(downloadNameFilter)

	if download == "today" {
		for _, job := range build.Build.Jobs {
			c, err := av.getJobArtifacts(job.JobID)
			if err != nil {
				continue
			}
			var artifacts []AppveyorArtifact
			if err = json.Unmarshal(c, &artifacts); err != nil {
				log.Println(err)
				continue
			}
			for _, artifact := range artifacts {
				if artifact.Created.Year() != time.Now().UTC().Year() ||
					artifact.Created.Month() != time.Now().UTC().Month() ||
					artifact.Created.Day() != time.Now().UTC().Day() {
					continue
				}
				if r == nil || r.MatchString(artifact.Name) {
					Download(downloadToolPath,
						fmt.Sprintf(` https://ci.appveyor.com/api/buildjobs/%s/artifacts/%s`, job.JobID, artifact.FileName),
						filepath.Join(saveDir, filepath.Base(artifact.FileName)))
				}
			}
		}
		return
	}

	count, err := strconv.Atoi(download)
	if err != nil {
		log.Println(err)
		return
	}

	for i, job := range build.Build.Jobs {
		if i >= count {
			break
		}
		c, err := av.getJobArtifacts(job.JobID)
		if err != nil {
			continue
		}
		var artifacts []AppveyorArtifact
		if err = json.Unmarshal(c, &artifacts); err != nil {
			log.Println(err)
			continue
		}
		for _, artifact := range artifacts {
			if r == nil || r.MatchString(artifact.Name) {
				Download(downloadToolPath,
					fmt.Sprintf(` https://ci.appveyor.com/api/buildjobs/%s/artifacts/%s`, job.JobID, artifact.FileName),
					filepath.Join(saveDir, filepath.Base(artifact.FileName)))
			}
		}
	}
}

func (av *Appveyor) DeleteFailedWorkflowRuns() {

}
