package provider

import (
	"errors"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	"app/shared/config"

	"app/constants"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// getAmazonCredentials call AWS servers for auth
func getAmazonCredentials() (*aws.Config, error) {
	clientCredentials := credentials.
		NewStaticCredentials(config.AWSS3().AWSAccessKey, config.AWSS3().AWSSecretKey, "")

	_, err := clientCredentials.Get()
	if err != nil {
		return nil, err
	}

	cfg := aws.NewConfig().WithRegion(config.AWSS3().AWSLocation).WithCredentials(clientCredentials)
	if cfg == nil {
		return nil, errors.New("bad amazon credentials")
	}

	return cfg, nil
}

// getAmazonSession set amazon config for easy file uploading
func getAmazonSession() (*session.Session, error) {
	amazonConfig, err := getAmazonCredentials()
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession(amazonConfig)
	if err != nil {
		return nil, err
	}

	sess = session.Must(sess, err)
	if sess == nil {
		return nil, errors.New("AWS session is nil")
	}

	return sess, nil
}

// uploadFileWithAWS send file to AWSS3
func uploadFileWithAWS(file *multipart.File, fileID, userName string) (string, error) {
	if file == nil {
		return "", errors.New("error while upload files with amazon: multipart.File is nil")
	}

	if fileID == "" || fileID == "0" {
		return "", errors.New("error while upload files with amazon: file id is empty or zero")
	}

	sess, err := getAmazonSession()
	if err != nil {
		return "", err
	}

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		// Bucket: aws.String(config.AWSS3().AWSBucketName),
		// Key:    aws.String(fileID),
		// Body:   *file,
		// ACL:    aws.String(config.AWSS3().AWSAccessType),

		// SSECustomerAlgorithm: aws.String("AES256"),
		// SSECustomerKey:       aws.String(CreateUsersFileKey(userName)),
		// SSECustomerKeyMD5:    aws.String(GetMD5Hash(CreateUsersFileKey(userName))),
		Bucket:               aws.String(config.AWSS3().AWSBucketName),
		Key:                  aws.String(fileID),
		Body:                 *file,
		ACL:                  aws.String(config.AWSS3().AWSAccessType),
		ServerSideEncryption: aws.String(constants.ServerSideEncryptionType),
		SSEKMSKeyId:          aws.String(constants.SSEKMSKeyID),
	})

	if err != nil {
		return "", err
	}

	if result == nil {
		return "", errors.New("nil uploaded result from AWS s3")
	}

	if result.Location == "" {
		return "", errors.New("void path to uploaded file from AWS s3")
	}

	log.Printf("file uploaded to, %s\n", aws.StringValue(&result.Location))
	return result.Location, nil
}

func getPresignLinkFromAWS(key string) (string, error) {
	if key == "" {
		return "", errors.New("error while getting presign link: bad file key")
	}

	sess, err := getAmazonSession()
	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		// Bucket:               aws.String(config.AWSS3().AWSBucketName),
		// Key:                  aws.String(key),
		// SSECustomerAlgorithm: aws.String("AES256"),
		Bucket: aws.String(config.AWSS3().AWSBucketName),
		Key:    aws.String(key),
	})

	t, err := strconv.Atoi("10")
	if err != nil {
		return "", errors.New("error while parsing presign link life period: " + err.Error())
	}

	urlStr, _, err := req.PresignRequest(time.Duration(t) * time.Minute)

	if err != nil {
		return "", errors.New("error while getting presign link: " + err.Error())
	}

	return urlStr, nil
}

// RemoveFileWithAWS can remove file with fileID (Key).
func RemoveFileWithAWS(fileID string) error {
	if fileID == "" || fileID == "0" {
		return errors.New("error while remove file with amazon: file id is empty or zero")
	}

	sess, err := getAmazonSession()
	if err != nil {
		return err
	}

	scv := s3.New(sess, sess.Config)
	_, err = scv.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.AWSS3().AWSBucketName),
		Key:    aws.String(fileID),
	})

	if err != nil {
		return err
	}

	return nil
}
