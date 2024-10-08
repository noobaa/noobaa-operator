[NooBaa Operator](../README.md) /

# S3 Compatibility in NooBaa
S3 (also known as Simple Storage Service) is an object storage service provided by Amazon. However, S3 is often colloquially used to refer to the S3 API - the RESTful interface for interaction with AWS S3. Over time, the S3 API has reached a point where many consider it the de facto standard API for object storage, and is supported by many cloud providers and storage vendors - even ones like Microsoft Azure and Google Cloud Platform, which also offer their own APIs alongside S3 compatibility.

## API Compatibility
Due to the wide adoption of the S3 API, NooBaa has been designed to be S3 compatible and adherent. NooBaa buckets and objects can be managed with most S3 clients without a need for proprietary tools or workarounds. All a user needs in order to interact with NooBaa through an average S3 client is the endpoint of the NooBaa system, and a set of fitting credentials.

## Utilization of S3 Compatible Storage Services
NooBaa can also be used as a gateway to other storage services that are S3 compatible. This means that NooBaa can be used to store and manage data in certain storage services (as long as they provide an S3 compatible API), even if they are not natively supported by the product. This is done by creating a [backingstore](backing-store-crd.md)/[namespacestore](namespace-store-crd.md) of type `s3-compatible` and providing the the appropriate endpoint.