module github.com/7thzero/InferDockerRegistryHash

go 1.14

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/containerd/containerd v1.3.5 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	//
	// Use this in your go.mod file to pull the docker API (It auto-resolves to a commit ID):
	//
	//  require (
	//	  github.com/docker/docker master
	//  )
	//
	// Per: https://github.com/moby/moby/issues/40185#issuecomment-571023162
	//
	// Also of note: https://github.com/moby/moby/issues/39302#issuecomment-639640386
	//
	//  Pay attention to the git hash as opposed to the 'pseudo version' that gets populated.
	//      For example, when I ran this on 2020-06-29 I was given this:
	//
	//          require github.com/docker/docker v17.12.0-ce-rc1.0.20200629163842-a70842f9c84f+incompatible
	//
	//      When you look at the git hash (a70842f9c84f) It points to the latest commit on master branch
	//      (https://github.com/moby/moby/commit/a70842f9c84f92741877301f33651353eddb44bb)
	github.com/docker/docker v17.12.0-ce-rc1.0.20200629163842-a70842f9c84f+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.6.0 // indirect
	google.golang.org/grpc v1.30.0 // indirect
)
