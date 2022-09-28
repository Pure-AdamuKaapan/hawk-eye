package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	svc *s3.S3
)

const (
	BUCKET_NAME = ""
	REGION      = "eu-central-1"
)

func init() {
	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("YOUR-ACCESSKEYID", "YOUR-SECRETACCESSKEY", ""),
		Region:           aws.String(REGION),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	})
	// Create S3 service client
	svc = s3.New(sess)
}

func listBuckets() (resp *s3.ListBucketsOutput) {
	resp, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		panic(err)
	}

	return resp
}

func createBucket() (resp *s3.CreateBucketOutput) {
	resp, err := svc.CreateBucket(&s3.CreateBucketInput{
		// ACL: aws.String(s3.BucketCannedACLPrivate),
		// ACL: aws.String(s3.BucketCannedACLPublicRead),
		Bucket: aws.String(BUCKET_NAME),
		// CreateBucketConfiguration: &s3.CreateBucketConfiguration{
		// 	LocationConstraint: aws.String(REGION),
		// },
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println("Bucket name already in use!")
				panic(err)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println("Bucket exists and is owned by you!")
			default:
				panic(err)
			}
		}
	}

	return resp
}

func deleteBucket() error {
	// Delete all the objects deleting the bucket
	// Setup BatchDeleteIterator to iterate through a list of objects.
	iter := s3manager.NewDeleteListIterator(svc, &s3.ListObjectsInput{
		Bucket: aws.String(BUCKET_NAME),
	})

	// Traverse iterator deleting each object
	err := s3manager.NewBatchDeleteWithClient(svc).Delete(aws.BackgroundContext(), iter)
	if err != nil {
		fmt.Printf("Unable to delete objects from bucket %q  %v", BUCKET_NAME, err)
		return err
	}
	fmt.Printf("Deleted object(s) from bucket: %s", BUCKET_NAME)

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(BUCKET_NAME),
	})
	if err != nil {
		return err
	}

	err = svc.WaitUntilBucketNotExists(&s3.HeadBucketInput{
		Bucket: aws.String(BUCKET_NAME),
	})
	if err != nil {
		return err
	}

	return nil
}

func uploadObject(filename string) (resp *s3.PutObjectOutput) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	fmt.Println("Uploading:", filename)
	resp, err = svc.PutObject(&s3.PutObjectInput{
		Body:   f,
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(filename),
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	})

	if err != nil {
		panic(err)
	}

	return resp
}

func listObjects() (resp *s3.ListObjectsV2Output) {
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(BUCKET_NAME),
	})

	if err != nil {
		panic(err)
	}

	return resp
}

func getObject(filename string) {
	fmt.Println("Downloading: ", filename)

	resp, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(filename),
	})

	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	err = ioutil.WriteFile(filename, body, 0644)
	if err != nil {
		panic(err)
	}
}

func deleteObject(filename string) (resp *s3.DeleteObjectOutput) {
	fmt.Println("Deleting: ", filename)
	resp, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(filename),
	})

	if err != nil {
		panic(err)
	}

	return resp
}

func main() {
	op := os.Args[1]
	fmt.Printf("Operation : %s", op)
	fmt.Println("")

	inputFile := os.Args[2]
	fmt.Printf("Input File : %s", inputFile)
	fmt.Println("")

	if op == "upload" {
		uploadObject(inputFile)
		return
	}

	if op == "download" {
		getObject(inputFile)
		return
	}

	fmt.Printf("No operation was executed")
	/*fmt.Println("List :")
	fmt.Println(listBuckets()))

	fmt.Println("Creating bucket :")
	fmt.Println(createBucket())
	fmt.Println("List :")
	fmt.Println(listBuckets())*/

	// fmt.Println("Deleting bucket :")
	// fmt.Println(deleteBucket())
	// fmt.Println("List :")
	// fmt.Println(listBuckets()

	// fmt.Println("Preparing to upload files")
	// // fmt.Println(s3session.ListBuckets(&s3.ListBucketsInput{}))

	// folder := "files"

	// files, _ := ioutil.ReadDir(folder)
	// fmt.Println(files)
	// for _, file := range files {
	// 	if file.IsDir() {
	// 		continue
	// 	} else {
	// 		uploadObject(folder + "/" + file.Name())
	// 	}
	// }

	// fmt.Println(listObjects())

	// for _, object := range listObjects().Contents {
	// 	getObject(*object.Key)
	// 	deleteObject(*object.Key)
	// }

	// fmt.Println(listObjects())

	// fmt.Println("Deleting bucket :")
	// fmt.Println(deleteBucket())
	// fmt.Println("List :")
	// fmt.Println(listBuckets())

}
