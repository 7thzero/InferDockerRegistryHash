//
// Copyright 2020 7thzero
// Licensed under Apache2
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/7thzero/InferDockerRegistryHash/imagehashconverter"
	"github.com/docker/docker/api/types"
	"io"
	"log"
	"os"
	"time"
	"github.com/docker/docker/client"
)

//
// Shows the use of the Registry hash Inferrer
func main() {
	imageToPull := flag.String("ImageName", "redis", "Specify the docker image you want to pull. Defaults to something in hub.docker.com")
	flag.Parse()
	//
	// Get a connection to the local docker server
	dockerClient, dockerClientErr := client.NewClientWithOpts(client.WithAPIVersionNegotiation(), client.WithTimeout(300 * time.Second))

	if dockerClientErr != nil{
		log.Println("Unable to establish connection to docker server. Err:", dockerClientErr)
		return
	}

	//
	// Pull an image from hub.docker.com
	//		You can pull from any docker image registry, doesn't have to be the 'official' repo
	pErr := pullDockerImage(dockerClient, *imageToPull)
	if pErr != nil{
		log.Println("Failed to pull image. Err: ", pErr)
		return
	}

	//
	// Get a list of image layer hashes that are converted to the format used when layers are stored in docker image
	// registry
	layersRegistry, layersRegistryErr := imagehashconverter.ExtractRegistryV2LayerHashes(dockerClient, *imageToPull)
	if layersRegistryErr != nil{
		log.Println("Unable to convert local docker image layers to registry format and extract registry hashes. Err:", layersRegistryErr)
		return
	}

	fmt.Println("Successfully converted docker image layer hashes to registry format for image", imageToPull, "Layers:")
	for _, layer :=range layersRegistry{
		fmt.Println("\t"+layer)
	}
}

//
// Pulls an image
func pullDockerImage(dClient *client.Client, img string) error{
	//
	// Pull an image
	//	Commented out authentication bits for simple example
	//	Pulls an amd64 image

	//aConfig, _ := encodeCredentials("your_username_here", "your_s3cure_password_here")
	imagePullOpts := types.ImagePullOptions{
		All: false,
		//RegistryAuth: aConfig
		PrivilegeFunc: nil,
		Platform: "amd64",
	}

	//
	// Pull the target image
	//	Defer closing of the ReadCloser so we can copy the status logs to a file
	ctx := context.Background()
	pulledImgOutput, pulledImgErr := dClient.ImagePull(ctx, img, imagePullOpts)
	if pulledImgErr != nil{
		log.Println("Unable to pull image. Err:", pulledImgErr)
		return pulledImgErr
	}
	defer pulledImgOutput.Close()
	pullErr := writeDockerDaemonLogfile(pulledImgOutput, "docker-daemon-pull.log")
	if pullErr != nil{
		log.Println("Unable to write docker daemon status to logfile. Err:", pullErr)
		return pullErr
	}

	return nil
}

//
// Encodes a simple username/password for authenticating
//		Returns the base64 encoding of the JSON username/password of Docker's types.AuthConfig
func encodeCredentials(username, password string) (string, error){
	authenticationConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}

	authJSON, encodedAuthErr := json.Marshal(authenticationConfig)
	if encodedAuthErr != nil{
		return "", encodedAuthErr
	}

	encodedAuth := base64.URLEncoding.EncodeToString(authJSON)

	return encodedAuth, nil
}

//
// Write out the status messages from the docker daemon
//		Could just as easily stream to stdout or just ignore if you don't need it
func writeDockerDaemonLogfile(stream io.ReadCloser, logFile string) error{
	oFile, oFileErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY, 0660)
	if oFileErr != nil{
		return oFileErr
	}

	io.Copy(oFile, stream)
	oFile.Close()

	return nil
}