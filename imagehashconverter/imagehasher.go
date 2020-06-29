//
// Copyright 2020 7thzero
// Licensed under Apache2
package imagehashconverter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
	"path"
)

//
// For a given docker image, extract the image layers as they would be stored/represented
// in a docker image registry conforming to V2 specifications
func ExtractRegistryV2LayerHashes(dClient *client.Client, img string, tempFiles ...string) ([]string, error){
	//
	// If you have a temp directory for the calculation work, use it!
	// Otherwise, just use the working directory
	tempExportFile := "image.tar"
	tempExportLayerGz := "layer.tar"
	if len(tempFiles) > 0{
		tempExportFile = path.Join(tempFiles[0], tempExportFile)
		tempExportLayerGz = path.Join(tempFiles[0], tempExportLayerGz)
	}

	//
	// Export image using docker daemon API
	eiErr := exportImageDockerApi(dClient, img, tempExportFile)
	if eiErr != nil{
		return nil, eiErr
	}

	//
	// Find the manifest file that determines layer order
	//		There may be more than one manifest listed if you export multiple images
	manifestBytes, manifestBytesErr := getFileContentFromTar(tempExportFile, "manifest.json")
	if manifestBytesErr != nil{
		log.Println("Unable to get manifest file from exported docker image tar archive. Err:", manifestBytesErr)
		return nil, manifestBytesErr
	}

	//
	// Unmarshal the manifest
	var imgManifests []DockerImgExportManifest
	unmarshalErr := json.Unmarshal(manifestBytes, &imgManifests)
	if unmarshalErr != nil{
		log.Println("Unable to unmarshal docker image manifest(s) from tar archive. Err:", unmarshalErr)
		return nil, unmarshalErr
	}

	imgManifest := imgManifests[0]	// This example only deals with exporting a SINGLE image

	registryLayerHashes, registryLayerHashesErr := getRegistryHashPerLayer(imgManifest, tempExportFile, tempExportLayerGz)
	if registryLayerHashesErr != nil{
		log.Println("Unable to get the registry layer hash values for the image. Err:", registryLayerHashesErr)
		return nil, registryLayerHashesErr
	}

	return registryLayerHashes, nil
}

//
// Iterates over the layers in the image manifest, returns an in-order list of registry hashes that
// correspond to each layer in the docker image
func getRegistryHashPerLayer(imgManifest DockerImgExportManifest, tarFile string, gzExportFileName string) ([]string, error){
	var layerHashesRegistry []string
	for _, imgLayer := range imgManifest.Layers{
		//
		// Extract the layer
		extractLayerErr := extractFileFromTar(tarFile, imgLayer, gzExportFileName)
		if extractLayerErr != nil{
			log.Println("Unable to extract layer", imgLayer, "Err: ", extractLayerErr)
			return nil, extractLayerErr
		}

		//
		// Compress the layer. This is required to obtain a matching registry hash
		gzExportFileNameRes := gzExportFileName+".gz"
		compressErr := gzipFileContents(gzExportFileName, gzExportFileNameRes)
		if compressErr != nil{
			log.Println("Unable to compress image layer", imgLayer, "Err: ", compressErr)
			return nil, compressErr
		}

		//
		// Get the SHA256 of the layer. This seems to be the only way I can calculate this and have it work correctly
		sha256Hash, sha256HashErr := getSha256FromFile(gzExportFileNameRes)
		if sha256HashErr != nil{
			log.Println("Unable to get sha256 hash for compressed image layer", gzExportFileName, " Err:", sha256HashErr)
			return nil, sha256HashErr
		}

		sha256Sum := fmt.Sprintf("%x", sha256Hash)
		layerHashesRegistry = append(layerHashesRegistry, sha256Sum)
	}

	return layerHashesRegistry, nil
}

//
// Get the golang sha256 sum of the file. Returns an empty hash and an error if there is a failure
func getSha256FromFile(fileName string) ([]byte, error) {
	var emptyHash []byte

	fileF, fileFErr := os.Open(fileName)
	if fileFErr != nil{
		log.Println("Unable to get file handle to file", fileName, " Err:", fileFErr)
		return emptyHash, fileFErr
	}
	defer fileF.Close()

	//
	// Use stream processing to calculate the hash
	hasher256 := sha256.New()
	if _, hashErr := io.Copy(hasher256, fileF); hashErr != nil{
		log.Println("Error while hashing file contents for file", fileName, " Err:", hashErr)
		return emptyHash, hashErr
	}

	//
	// Get the hash
	sha256Sum := hasher256.Sum(nil)
	//log.Printf("%x", sha256sum)

	return sha256Sum, nil
}

//
// Gzips a files contents and writes to a new file
func gzipFileContents(extractedLayerF string, compressedLayerF string) error {
	//
	// Get a file handle to the extracted layer tar
	layerF, layerFErr := os.Open(extractedLayerF)
	if layerFErr != nil{
		log.Println("Unable to open extracted layer for processing", extractedLayerF, "Err: ", layerFErr)
		return layerFErr
	}
	defer layerF.Close()

	//
	// Get a file handle to the target/output gzipped file
	exportF, exportFErr := os.Create(compressedLayerF)
	if exportFErr != nil{
		log.Println("Unable to create output file for gzipped layer", compressedLayerF, " Err: ", exportFErr)
		return exportFErr
	}
	defer exportF.Close()

	//
	// Get a golang GZIP Writer (pegged to the output file)
	gzipW := gzip.NewWriter(exportF)

	//
	// Compress the file
	_, compressErr := io.Copy(gzipW, layerF)
	if compressErr != nil{
		log.Println("Error encountered while compressing layer tar to GZIP", compressedLayerF, " Err: ", compressErr)
		return compressErr
	}

	return nil
}

//
// Extract a file stored within a specified tar archive
// similar to the 'getFileContentsFromTar' function, except that it io.copies straight to a filesystem file
// (this prevents us having to cache a file in memory)
func extractFileFromTar(tarArchive string, tarFileName string, extractedFileName string) (error) {
	tarF, tarFErr := os.Open(tarArchive)
	if tarFErr != nil{
		return tarFErr
	}
	defer tarF.Close()

	tReader := tar.NewReader(tarF)
	var iterationErr error
	for {
		tarEntry, tarEntryErr := tReader.Next()
		if tarEntryErr != nil{
			//
			// Check if we're done processing the file. When finished, quit
			if tarEntryErr == io.EOF{
				break
			}

			//
			// Somethine else went wrong: note an error!
			log.Println("Unexpected error occurred while processing tar archive. Err:", tarEntryErr)
			iterationErr = tarEntryErr
			break
		}

		if tarEntry.Name == tarFileName {
			eFile, eFileErr := os.Create(extractedFileName)
			if eFileErr != nil{
				log.Println("Unable to create extraction file. Err:", eFileErr)
				iterationErr = eFileErr
				break
			}
			defer eFile.Close()

			byteCount, copyErr := io.Copy(eFile, tReader)	// Need to read from the READER and not the HEADER/Entry
			if copyErr != nil{
				log.Println("Unable to copy file contents to output buffer:", tarFileName, "Byte Count:", byteCount, "Err:", copyErr)
				iterationErr = copyErr
				break
			}
		}
	}

	return iterationErr
}


//
// Pull bytes from a file stored within a specified tar archive
func getFileContentFromTar(tarArchive string, fileName string) ([]byte, error) {
	tarF, tarFErr := os.Open(tarArchive)
	if tarFErr != nil{
		return nil, tarFErr
	}
	defer tarF.Close()

	tReader := tar.NewReader(tarF)
	var fileBytes bytes.Buffer
	var iterationErr error
	for {
		tarEntry, tarEntryErr := tReader.Next()
		if tarEntryErr != nil{
			//
			// Check if we're done processing the file. When finished, quit
			if tarEntryErr == io.EOF{
				break
			}

			//
			// Somethine else went wrong: note an error!
			log.Println("Unexpected error occurred while processing tar archive. Err:", tarEntryErr)
			iterationErr = tarEntryErr
			break
		}

		if tarEntry.Name == fileName{
			byteCount, copyErr := io.Copy(&fileBytes, tReader)		// Need to read from the READER and not the HEADER/Entry
			if copyErr != nil{
				log.Println("Unable to copy file contents to output buffer:", fileName, "Byte Count:", byteCount, "Err:", copyErr)
				iterationErr = copyErr
				break
			}
		}
	}

	return fileBytes.Bytes(), iterationErr
}

//
// Export an image to TAR format on the HD
func exportImageDockerApi(dClient *client.Client, dockerImage string, exportFile string) error{
	dockerImageStream, dockerImageStreamErr := dClient.ImageSave(context.Background(), []string{dockerImage})
	if dockerImageStreamErr != nil{
		//
		// Note: you may need a retry here depending on how busy the docker socket is

		log.Print("Unable to export docker image. Err:", dockerImageStreamErr)
		return dockerImageStreamErr
	}
	defer dockerImageStream.Close()		// Ensure we close this!

	//
	// Write out the saved image stream to disk
	savedImgF, savedImgFErr := os.OpenFile(exportFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if savedImgFErr != nil{
		log.Println("Unable to open a file for exporting the docker image. Err:", savedImgFErr)
		return savedImgFErr
	}
	defer savedImgF.Close()

	if _, copyErr := io.Copy(savedImgF, dockerImageStream); copyErr != nil{
		log.Println("Unable to copy/save docker image to TAR archive. Err:", copyErr)
		return copyErr
	}
	savedImgF.Close()

	return nil
}