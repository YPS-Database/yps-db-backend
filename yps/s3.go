package yps

import ypss3 "github.com/YPS-Database/yps-db-backend/yps/s3"

var TheS3 *ypss3.YPSS3

func OpenS3(bucket, uploadKeyPrefix, uploadURLPrefix string) error {
	thisS3, err := ypss3.Open(bucket, uploadKeyPrefix, uploadURLPrefix)
	if err != nil {
		return err
	}

	TheS3 = thisS3

	return nil
}
