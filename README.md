# Rainbow Storage

This is the storage component of Rainbow :rainbow:, an implementation of Amazon S3 for local development.

## Getting Started

```shell
$ docker pull atenderholt/rainbow-storage
$ docker network create rainbow
$ docker run --rm --name rainbow-storage --network rainbow -p 9000:9000 -v $PWD/data:/data -v /var/run/docker.sock:/var/run/docker.sock atenderholt/rainbow-storage:1.0.0
```

Notes:
* Since rainbow-storage is using Docker to start another container (i.e. minio), it needs the following:
  * `--name rainbow-storage` to correctly bind-mount the data directory from the host into the other container
  * `--network rainbow` to reference other containers by name
  * `-v /var/run/docker.sock:/var/run/docker.sock` to control Docker from inside the running container
* Buckets, objects, etc. are persisted to `/data`.

## Questions

* What's Rainbow?

A future project for local implementation of Amazon AWS services.

* Why not LocalStack?

LocalStack appears to be still using the moto implementation of Amazon S3, which doesn't seem to work well with large files. Perhaps this was solved w/ the paid version,
but I wanted something lighter weight and not so tightly-coupled to the rest of the LocalStack services (e.g. specify a different targets of Event Notifications).

* Why not just minio?

Minio seems great, but doesn't seem like supports Event Notifications to a local implementation of an Amazon Lambda Functions. It's also not clear how much can be
configured with Terraform.
