package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	rcsl "github.com/iguanesolutions/go-runcomfyserverless"
)

const (
	apiToken    = "<Your-User-API-Key>"
	deployID    = "<Your-Deployment-ID>"
	jpgFilePath = "/path/to/rabbit.jpg"
)

func main() {
	// Init
	deploy := rcsl.LinkToDeployment(rcsl.Config{
		UserAPIToken: apiToken,
		DeploymentID: deployID,
	})
	// Prepare for cancellation
	runCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// Start the process
	imgData, err := loadImage()
	if err != nil {
		panic(fmt.Sprintf("failed to load image: %s", err))
	}
	reqID, err := deploy.Start(context.TODO(),
		[]rcsl.Override{
			// Prompt
			{
				NodeID: "31", // relative to your saved workflow, export the api workflow to retreive the infos
				Inputs: map[string]any{
					"value": "Using this elegant style, create a portrait of an elephant wearing a pearl tiara and lace collar, maintaining the same refined quality and soft color tones.",
				},
			},
			// Image
			{
				NodeID: "16", // relative to your saved workflow, export the api workflow to retreive the infos
				Inputs: map[string]any{
					"image": imgData,
				},
			},
			// Seed
			{
				NodeID: "27", // relative to your saved workflow, export the api workflow to retreive the infos
				Inputs: map[string]any{
					"value": int64(rand.Uint64()),
				},
			},
		},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start inference: %s", err))
	}
	fmt.Println("Job sent:", reqID)
	// Follow
	start := time.Now()
	queued := true
	ticker := time.NewTicker(10 * time.Second) // ideally set a wait time that is half of a single run duration (once warm)
	defer ticker.Stop()
	for {
		select {
		case <-runCtx.Done():
			fmt.Println(" Cancelling request")
			cancelStatus, err := deploy.Cancel(context.TODO(), reqID)
			if err != nil {
				panic(err)
			}
			fmt.Println("Cancel request status:", cancelStatus)
			runCtx = context.Background() // reset context to avoid looping into cancel
		case <-ticker.C:
			status, err := deploy.Status(context.TODO(), reqID)
			if err != nil {
				panic(fmt.Sprintf("failed to get job status: %s", err))
			}
			switch status.Status {
			case rcsl.RequestStatusInQueue:
				fmt.Println("in queue:", status.QueuePosition)
			case rcsl.RequestStatusInProgress:
				if queued {
					fmt.Printf("started ! (queued for %s)\n", time.Since(start))
					queued = false
					start = time.Now()
				} else {
					fmt.Println("still in progress")
				}
			case rcsl.RequestStatusCompleted:
				stop := time.Now()
				result, err := deploy.Result(context.TODO(), reqID)
				if err != nil {
					panic(fmt.Sprintf("failed to get result: %s", err))
				}
				switch result.Status {
				case rcsl.RequestStatusSucceeded:
					fmt.Printf("Done (generation took %s -not counting wait time-):\n", stop.Sub(start))
					for _, imageOutput := range result.Outputs.ExtractImagesResults() {
						if err = dwldFileToDisk(imageOutput.FileName, imageOutput.URL); err != nil {
							panic(err)
						}
						fmt.Printf("\t%s\n", imageOutput.FileName)
					}
				case rcsl.RequestStatusFailed:
					fmt.Printf("Job failed: %+v\n", result.Error)
				case rcsl.RequestStatusInQueue, rcsl.RequestStatusInProgress, rcsl.RequestStatusCanceled:
					// should not happen with this code as we check status first
					// but should be handled if Result() is called without verifying Status() first
				default:
					panic(fmt.Sprintf("unknown resultat status %s", result.Status))
				}
				return
			case rcsl.RequestStatusCanceled:
				fmt.Println("canceled")
				return
			default:
				panic(fmt.Sprintf("unknown status: %s", status.Status))
			}
		}
	}
}

func loadImage() (urlEncoded string, err error) {
	fd, err := os.Open(jpgFilePath)
	if err != nil {
		return
	}
	defer fd.Close()
	data, err := io.ReadAll(fd)
	if err != nil {
		return
	}
	urlEncoded = rcsl.ConvertJPGToB64URL(data)
	return
}

func dwldFileToDisk(filePath string, url *url.URL) (err error) {
	resp, err := http.Get(url.String())
	if err != nil {
		err = fmt.Errorf("failed to get URL: %w", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("download failed: %s", resp.Status)
		return
	}
	fd, err := os.Create(filePath)
	if err != nil {
		err = fmt.Errorf("failed create destination file: %w", err)
		return
	}
	defer fd.Close()
	if _, err = io.Copy(fd, resp.Body); err != nil {
		err = fmt.Errorf("failed to copy file data to destination file: %w", err)
		return
	}
	return
}
