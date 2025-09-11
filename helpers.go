package runcomfyserverless

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/url"
)

func ConvertImageToB64URL(img image.Image) (URLencoded string, err error) {
	var data bytes.Buffer
	if err = png.Encode(&data, img); err != nil {
		err = fmt.Errorf("failed to encode image as PNG: %w", err)
		return
	}
	URLencoded = ConvertPNGToB64URL(data.Bytes())
	return
}

func ConvertPNGToB64URL(data []byte) (URLencoded string) {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
}

func ConvertJPGToB64URL(data []byte) (URLencoded string) {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
}

type Outputs map[string]map[string]any

func (o Outputs) ExtractImagesResults() (images []ImageOutput) {
	images = make([]ImageOutput, 0, 1) // 99% of the time we will be getting one image from the remote workflow
	for nodeID, nodeOutput := range o {
		for outputType, outputPayload := range nodeOutput {
			if outputType == "images" {
				if imagesList, ok := outputPayload.(map[string]any); ok {
					output := ImageOutput{
						FromNodeID: nodeID,
					}
					for key, value := range imagesList {
						switch key {
						case "filename":
							output.FileName = value.(string)
						case "subfolder":
							output.SubFolder = value.(string)
						case "type":
							output.Type = value.(string)
						case "url":
							var err error
							if output.URL, err = url.Parse(value.(string)); err != nil {
								output.URL = nil // just to be sure
							}
						}
					}
					if output.URL != nil {
						images = append(images, output)
					}
				}
			}
		}
	}
	return
}

type ImageOutput struct {
	FromNodeID string
	FileName   string
	SubFolder  string
	Type       string
	URL        *url.URL
}
