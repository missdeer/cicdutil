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
	"time"
)

type Github struct {
}

type GithubActionsArtifacts struct {
	TotalCount int `json:"total_count'`
	Artifacts  []struct {
		ID                 int       `json:"id"`
		NodeID             string    `json:"node_id"`
		Name               string    `json:"name"`
		SizeInBytes        int       `json:"size_in_bytes"`
		URL                string    `json:"url"`
		ArchiveDownloadURL string    `json:"archive_download_url"`
		Expired            bool      `json:"expired"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
	} `json:"artifacts"`
}

func (gh *Github) list() ([]byte, error) {
	u := fmt.Sprintf(`https://api.github.com/repos/%s/%s/actions/artifacts`, username, project)
	client := &http.Client{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse artifacts list request:", err)
		return nil, err
	}

	req.SetBasicAuth(username, token)

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

func (gh *Github) deleteWorkflowRun(id int) error {
	u := fmt.Sprintf(`https://api.github.com/repos/%s/%s/actions/runs/%d`, username, project, id)
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		log.Println("Could not parse delete workflow runs request:", err)
		return err
	}

	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send delete workflow runs request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Println("delete workflow runs response not 204:", resp.Status)
		return fmt.Errorf("delete workflow runs response not 204:", resp.Status)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading delete workflow runs response failed", err)
		return err
	}
	log.Println("workflow runs", id, "has been deleted")
	return nil
}

func (gh *Github) listFailedWorkflowRuns() ([]int, error) {
	u := fmt.Sprintf(`https://api.github.com/repos/%s/%s/actions/runs`, username, project)
	client := &http.Client{}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println("Could not parse workflow runs list request:", err)
		return nil, err
	}

	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send workflow runs list request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("workflow runs list response not 200:", resp.Status)
		return nil, fmt.Errorf("workflow runs list response not 200:", resp.Status)
	}

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading workflow runs list failed", err)
		return nil, err
	}

	type WorkflowRuns struct {
		ID         int    `json:"id"`
		Conclusion string `json:"conclusion"`
	}
	var res []WorkflowRuns
	if err = json.Unmarshal(c, &res); err != nil {
		log.Println("can't unmarshal result", err)
		return nil, err
	}
	ids := make([]int, len(res))
	for _, r := range res {
		ids = append(ids, r.ID)
	}
	return ids, nil
}

func (gh *Github) delete(id int) error {
	u := fmt.Sprintf(`https://api.github.com/repos/%s/%s/actions/artifacts/%d`, username, project, id)
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		log.Println("Could not parse delete artifact request:", err)
		return err
	}

	req.SetBasicAuth(username, token)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Could not send delete artifact request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		log.Println("delete artifact response not 204:", resp.Status)
		return fmt.Errorf("delete artifact response not 204:", resp.Status)
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading delete artifact response failed", err)
		return err
	}
	log.Println("artifact", id, "has been deleted")
	return nil
}

func (gh *Github) ListArtifacts() {
	c, err := gh.list()
	if err == nil {
		fmt.Println(string(c))
	}
}

func (gh *Github) DeleteArtifacts() {
	for {
		c, err := gh.list()
		if err != nil {
			return
		}
		var artifacts GithubActionsArtifacts
		err = json.Unmarshal(c, &artifacts)
		if err != nil {
			log.Println(err)
			return
		}
		if keep == "today" {
			for _, artifact := range artifacts.Artifacts {
				if artifact.CreatedAt.Year() != time.Now().UTC().Year() ||
					artifact.CreatedAt.Month() != time.Now().UTC().Month() ||
					artifact.CreatedAt.Day() != time.Now().UTC().Day() {
					gh.delete(artifact.ID)
				}
			}
			return
		}

		count, err := strconv.Atoi(keep)
		if err != nil {
			log.Println(err)
			return
		}

		if len(artifacts.Artifacts) <= count {
			log.Println(len(artifacts.Artifacts), "artifacts found, nothing to delete")
			return
		}

		for i := count; i < len(artifacts.Artifacts); i++ {
			artifact := artifacts.Artifacts[i]
			gh.delete(artifact.ID)
		}
	}
}

func (gh *Github) DownloadArtifacts() {
	c, err := gh.list()
	if err != nil {
		return
	}
	var artifacts GithubActionsArtifacts
	err = json.Unmarshal(c, &artifacts)
	if err != nil {
		log.Println(err)
		return
	}

	r := regexp.MustCompile(downloadNameFilter)
	if download == "today" {
		for _, artifact := range artifacts.Artifacts {
			if r != nil && !r.MatchString(artifact.Name) {
				continue
			}
			if artifact.CreatedAt.Year() != time.Now().UTC().Year() ||
				artifact.CreatedAt.Month() != time.Now().UTC().Month() ||
				artifact.CreatedAt.Day() != time.Now().UTC().Day() {
				continue
			}
			Download(downloadToolPath, artifact.ArchiveDownloadURL, filepath.Join(saveDir, fmt.Sprintf("%s.zip", artifact.Name)))
		}
		return
	}

	count, err := strconv.Atoi(download)
	if err != nil {
		log.Println(err)
		return
	}

	for i := 0; i < count; i++ {
		artifact := artifacts.Artifacts[i]
		if r != nil && !r.MatchString(artifact.Name) {
			continue
		}
		Download(downloadToolPath, artifact.ArchiveDownloadURL, filepath.Join(saveDir, fmt.Sprintf("%s.zip", artifact.Name)))
	}
}

func (gh *Github) Build() {

}

func (gh *Github) DeleteFailedWorkflowRuns() {

}
