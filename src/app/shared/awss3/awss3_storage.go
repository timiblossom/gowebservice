package awss3

// AWSS3Config represents Amazon web services S3 storage configuration
type AWSS3Config struct {
	AWSAccessKey            string
	AWSSecretKey            string
	AWSLocation             string
	AWSBucketName           string
	AWSAccessType           string
	AWSPresignLifePeriodMin uint
}

