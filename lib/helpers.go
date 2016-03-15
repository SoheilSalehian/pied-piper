package piedpiper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"

	log "github.com/Sirupsen/logrus"
)

func (a *Ad) Download() error {
	url := fmt.Sprintf("http://s3.amazonaws.com/ownlocal.adforge.production/ads/%s/original_pdfs.pdf", a.id)
	output, err := os.Create(a.fileName)
	defer output.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			fmt.Println("req", req)
			fmt.Println("via", via)
			fmt.Println("headers", req.Header)
			return nil
		},
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("Error while creating cookie jar", url, "-", err)
		return err
	}
	client.Jar = jar

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("GET failed", err)
		return err
	}

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while downloading ", url, "-", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Info(response.Status)
		return err
	}

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Fatal("Error while copying response", url, "-", err)
		return err
	}

	log.Println("Completed download of ad with id: ", a.id)
	return nil
}

func getAdIDsAPI(date string, limit int) error {

	var ids []string

	log.Info("Requesting adIDs from the API")

	url := fmt.Sprintf("http://api.ownlocal.com/ads?created_at=%s&size=%s", date, limit)

	response, err := http.Get(url)
	if err != nil {
		log.Fatal("GET failed", err)
		return err
	}

	body, err := ioutil.ReadAll(response.Body)

	var data Ad
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal("%T\n%s\n%#v\n", err, err, err)
	}
	log.Info(data)

	for id := range data.id {
		ids = append(ids, string(id))
	}
	if len(ids) == 0 {
		log.Fatal("Ad IDs are were not retrived from the API")
	}
	log.Info(ids)
	return nil
}
