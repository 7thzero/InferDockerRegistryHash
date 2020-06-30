# About

This project shows an approach to generate a docker registry hash for a given docker image.

# Why use this?
If you find yourself supporting older versions of Quay enterprise, you may be in a position where the older docker registry format prevents you from easily comparing images in other registries (like AWS ECR). 

I wrote this to help me on a security project where base images are stored in an older quay repository yet derived app images live in AWS ECR and we needed to know if app images were derived/built from approved base images.

# Any notes?
Yes, please see the note in `go.mod` for notes on how to pull the docker client libraries. Moby/docker require a little massaging to pull the right API version.

A basic authentication scheme is implemented yet commented out if you'd like to try this against a repository that uses basic authentication.

# How do I use this?
Ideally you'd step through the code to see how it works. You can also run it like this:
  `./InferDockerRegistryHash -ImageName redis:latest` (you can specify any dockerhub image out of the box) 
