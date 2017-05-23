package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gosuri/uiprogress"
)

func countKeys(svc *s3.S3, bucketName string, prefix string) (int, error) {
	var keyCount int = 0
	input := s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}
	err := svc.ListObjectsV2Pages(&input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			keyCount += len(page.Contents)
			return true
		})
	return keyCount, err
}

func listKeysToChannel(svc *s3.S3, bucketName string, prefix string, c chan<- string) error {
	input := s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}
	err := svc.ListObjectsV2Pages(&input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				c <- *obj.Key
			}
			return true
		})
	close(c)
	return err
}

type Grants struct {
	FullControl []string
	Read []string
	// TODO: ReadACP, WriteACP
}

func idList(ids []string) string {
	var buffer bytes.Buffer
	buffer.WriteString("id=")
	buffer.WriteString(ids[0])
	for _, id := range ids[1:] {
		buffer.WriteString(", ")
		buffer.WriteString("id=")
		buffer.WriteString(id)
	}
	return buffer.String()
}

func createPutAclInput(bucketName string, key string, grants Grants) *s3.PutObjectAclInput {
	putInput := s3.PutObjectAclInput {
		Bucket: aws.String(bucketName),
		Key: aws.String(key),
	}
	if len(grants.FullControl) > 0 {
		putInput.GrantFullControl = aws.String(idList(grants.FullControl))
	}
	if len(grants.Read) > 0 {
		putInput.GrantRead = aws.String(idList(grants.Read))
	}
	return &putInput
}


func putAclOnKeysFromChannel(svc *s3.S3, bucketName string, g Grants,
	c <-chan string, wg *sync.WaitGroup, bar *uiprogress.Bar) error {

	defer wg.Done()
	for key := range c {
		_, err := svc.PutObjectAcl(createPutAclInput(bucketName, key, g))
		if err != nil {
			         if awsErr, ok := err.(awserr.Error); ok {
					log.Printf("Key: %s, Error: %s", key, awsErr.Code())
			         } else {
					 log.Printf("Key: %s, Error: %s", key, err.Error())
			         }
		}
		bar.Incr()
	}
	return nil
}


func createProgressBar(keyCount int) (bar *uiprogress.Bar) {
	bar = uiprogress.AddBar(keyCount)
	bar.AppendCompleted()
	bar.PrependElapsed()
	return bar
}

/*
	PRO TIP: when giving rights to other accounts, always include all the existing
	rights for yourself or otherwise you'll not get back at the data!
 */
func putAclRecursive(bucketName string, prefix string, g Grants) error {
	f, err := os.OpenFile("put-acl.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	config := aws.Config{Region: aws.String("eu-west-1")}
	sess := session.Must(session.NewSession(&config))
	svc := s3.New(sess)

	keyCount, err := countKeys(svc, bucketName, prefix)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return err
	} else {
		fmt.Printf("number of keys: %d\n", keyCount)
	}

	uiprogress.Start()
	bar := createProgressBar(keyCount)

	c := make(chan string)
	wg := new(sync.WaitGroup)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go putAclOnKeysFromChannel(svc, bucketName, g, c, wg, bar)
	}
	listKeysToChannel(svc, bucketName, prefix, c)
	wg.Wait()
	time.Sleep(100 * time.Millisecond)  // for progress bar to finish drawing itself to get full 100% at the end.
	return nil
}
