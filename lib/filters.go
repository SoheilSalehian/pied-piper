package piedpiper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type Ad struct {
	id         string
	created_at string
	fileName   string
	processed  bool
	deleted    bool
	PdfMetrics
}

const NUM_WORKERS = 8

type PdfMetrics struct {
	TextLayer bool
}

func NewAd(id string) *Ad {
	a := new(Ad)
	a.id = id
	a.fileName = fmt.Sprintf("downloads/ad-%s.pdf", a.id)
	a.processed = false
	a.deleted = false
	return a
}

type DownloadAll struct{}

func (dl DownloadAll) Process(in chan *Ad) chan *Ad {
	out := make(chan *Ad, 10000)
	fmt.Println(len(in))
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go func() {
				for ad := range in {
					ad.Download()
					out <- ad
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(out)
	}()
	fmt.Printf("Finished all goroutines")
	return out
}

type CheckPDFText struct{}

func (cpt CheckPDFText) Process(in chan *Ad) chan *Ad {
	result := make(chan *Ad, 10000)
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go func() {
				for ad := range in {
					out, err := exec.Command("pdftotext", ad.fileName, "-").Output()
					if err != nil {
						log.Errorf("Checking the PDF text layer failed for file: ", ad.fileName)
						log.Info(err)
					}
					if len(string(out)) > 0 {
						ad.PdfMetrics.TextLayer = true
					} else {
						ad.PdfMetrics.TextLayer = false
					}
					ad.processed = true
					result <- ad
					log.Info("Finished processing ad: ", ad.id)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(result)
	}()
	return result

}

// TODO: make in memory calls to not have to create/download/delete temp files
type DeletePDF struct{}

func (dtp DeletePDF) Process(in chan *Ad) chan *Ad {
	del := make(chan *Ad, 10000)
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)

			go func() {
				for ad := range in {
					err := os.Remove(ad.fileName)
					if err != nil {
						log.Errorf("Failed in deleting the PDF file: ", ad.fileName)
						log.Info(err)
					}

					ad.deleted = true
					del <- ad
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(del)
	}()
	return del
}

type UploadAllGoogle struct{}

func (dl UploadAllGoogle) Process(in chan *Ad) chan *Ad {
	out := make(chan *Ad, 10000)
	fmt.Println(len(in))
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go func() {
				for ad := range in {
					ad.Download()
					out <- ad
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(out)
	}()
	fmt.Printf("Finished all goroutines")
	return out
}

type ConvertToPng struct{}

func (ctp ConvertToPng) Process(in chan *Ad) chan *Ad {
	result := make(chan *Ad, 10000)
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go func() {
				for ad := range in {
					cmd := "pdftoppm"
					args := []string{"-rx", "300", "-ry", "300", "-png", fmt.Sprintf("./%s", ad.fileName), fmt.Sprintf("./downloads/ad-%s", ad.id)}

					_, err := exec.Command(cmd, args...).Output()
					if err != nil {
						log.Errorf("Converrting to PNG failed for file: ", ad.fileName)
						log.Info(err)
					}

					ad.processed = true
					result <- ad
					log.Info("Finished processing ad: ", ad.id)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(result)
	}()
	return result

}

type CallOcrClient struct{}

func (coc CallOcrClient) Process(in chan *Ad) chan *Ad {
	result := make(chan *Ad, 10000)
	go func() {
		wg := new(sync.WaitGroup)
		for i := 0; i < NUM_WORKERS; i++ {
			wg.Add(1)
			go func() {
				for ad := range in {

					ocrclientReq := fmt.Sprintf("{\"image_path\":\"%s/downloads/ad-%s-1.png\"}", os.Getenv("PWD"), ad.id)
					jsonStr := []byte(ocrclientReq)

					req, err := http.NewRequest("POST", "http://localhost:8082/ocrreport", bytes.NewBuffer(jsonStr))
					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						log.Error(err)
					}
					defer resp.Body.Close()

					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Println("response: ", string(body))

					ad.processed = true
					result <- ad
					log.Info("Finished processing ad: ", ad.id)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		close(result)
	}()
	return result

}
