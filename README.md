# Rainbow Storage

This is the storage component of Rainbow :rainbow:, an implementation of Amazon S3 for local development.

## Questions

* What's Rainbow?

A future project for local implementation of Amazon AWS services.

* Why not LocalStack?

LocalStack appears to be still using the moto implementation of Amazon S3, which doesn't seem to work well with large files. Perhaps this was solved w/ the paid version,
but I wanted something lighter weight and not so tightly-coupled to the rest of the LocalStack services (e.g. specify a different targets of Event Notifications).

* Why not just minio?

Minio seems great, but doesn't seem like supports Event Notifications to a local implementation of an Amazon Lambda Functions. It's also not clear how much can be
configured with Terraform.
