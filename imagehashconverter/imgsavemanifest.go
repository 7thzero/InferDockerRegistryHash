//
// Copyright 2020 7thzero
// Licensed under Apache2

package imagehashconverter

//
// When you perform a 'docker save img:tag > image.tar' operation, this manifest will be generated
// and saved as a file within the tar archive titled `manifest.json`.
//
// Note that if you export multiple docker image/tag pairs at once there will be multiple manifests in the JSON array
type DockerImgExportManifest struct{
	Config string
	RepoTags []string
	Layers []string
}
