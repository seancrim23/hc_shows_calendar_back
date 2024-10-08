package utils

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

func ReadFileFromBucket(fileName string, bucketName string) (string, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	rc, err := client.Bucket(bucketName).Object(fileName).NewReader(ctx)
	if err != nil {
		return "", err
	}
	slurp, err := ioutil.ReadAll(rc)
	rc.Close()
	if err != nil {
		return "", err
	}

	return string(slurp), nil
}
