package runcomfyserverless

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/url"
)

// ConvertImageToB64URL prepares a golang Image to an URL encoded representation usable in Overrides.
func ConvertImageToB64URL(img image.Image) (URLencoded string, err error) {
	var data bytes.Buffer
	if err = png.Encode(&data, img); err != nil {
		err = fmt.Errorf("failed to encode image as PNG: %w", err)
		return
	}
	URLencoded = ConvertPNGToB64URL(data.Bytes())
	return
}

// ConvertPNGToB64URL takes a png encoded data and returns its URL encoded representation usable in Overrides.
func ConvertPNGToB64URL(data []byte) (URLencoded string) {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
}

// ConvertJPGToB64URL takes a jpg encoded data and returns its URL encoded representation usable in Overrides.
func ConvertJPGToB64URL(data []byte) (URLencoded string) {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)
}

// Outputs contains the output nodes results.
// first key is the output node id, second key is output type, value depends on output type.
// Ex:
// map[
//
//		 	38:42:map[
//				text:[['Transform to an elephant, wearing a pearl tiara and lace collar, while maintaining the same refined quality and soft color tones.']]
//		 	]
//			58:map[
//		 		images:[
//					map[
//						filename:ComfyUI_00010_.png
//						subfolder:
//						type:output
//						url:https://serverless-api-storage.runcomfy.net/output/..../ComfyUI_00010_.png
//					]
//				]
//			]
//	 ]
type Outputs map[string]map[string]any

// ExtractImagesResults loop thru a result output and extract any image output contained within the outputs collection.
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

// ImageOutput contains all the informations about an image output from am end node in the workflow.
type ImageOutput struct {
	FromNodeID string
	FileName   string
	SubFolder  string
	Type       string
	URL        *url.URL
}
