package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	typesS3 "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const bucketName = "aws-demo-test-bucket-22dez"
const regionName = "us-east-1"

func main() {

	var (
		instanceId string
		err        error
		s3Client   *s3.Client
	)

	ctx := context.Background()
	if instanceId, err = createEC2(ctx); err != nil {
		fmt.Printf("createEC2 error: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Instace ID: %s\n", instanceId)

	if s3Client, err = initS3Client(ctx); err != nil {
		fmt.Printf("initS3Client error: %s", err)
		os.Exit(1)
	}

	if err = createS3Bucket(ctx, s3Client); err != nil {
		fmt.Printf("createS3Bucket error: %s", err)
		os.Exit(1)
	}

	if err = uploadToS3Bucket(ctx, s3Client); err != nil {
		fmt.Printf("uploadToS3Bucket error: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Upload complete successfully.\n")
}

func createEC2(ctx context.Context) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(regionName))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}
	ec2_client := ec2.NewFromConfig(cfg)

	key_pairs, err := ec2_client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{"go-aws-demo"},
	})

	if err != nil && !strings.Contains(err.Error(), "InvalidKeyPair.NotFound") {
		return "", fmt.Errorf("DescribeKeyPairs error: %s", err)
	}

	if key_pairs == nil || len(key_pairs.KeyPairs) == 0 {
		key_pair, err := ec2_client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
			KeyName: aws.String("go-aws-demo"),
		})

		if err != nil {
			return "", fmt.Errorf("DescribeKeyPairs error: %s", err)
		}

		err = os.WriteFile("go-aws-ec2.pem", []byte(*key_pair.KeyMaterial), 0600)

		if err != nil {
			return "", fmt.Errorf("WriteFile error: %s", err)
		}
	}

	_, err = ec2_client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{
		KeyName: aws.String("go-aws-demo"),
	})

	if err != nil {
		return "", fmt.Errorf("CreateKeyPair error: %s", err)
	}

	images_output, err := ec2_client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{"ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"},
			},
			{
				Name:   aws.String("virtualization-type"),
				Values: []string{"hvm"},
			},
		},
		Owners: []string{"099720109477"},
	})

	if err != nil {
		return "", fmt.Errorf("DescribeImages error: %s", err)
	}

	if len(images_output.Images) == 0 {
		return "", fmt.Errorf("images_output.Images is of 0 lenght.")
	}

	instance, err := ec2_client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      images_output.Images[0].ImageId,
		KeyName:      aws.String("go-aws-demo"),
		InstanceType: types.InstanceTypeT3Micro,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	})

	if err != nil {
		return "", fmt.Errorf("RunInstances error: %s", err)
	}
	//images_output.Images[0].ImageId

	if len(instance.Instances) == 0 {
		return "", fmt.Errorf("instance.Instances is of 0 lenght.")
	}

	return *instance.Instances[0].InstanceId, nil
}

func initS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(regionName))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	return s3.NewFromConfig(cfg), nil
}

func createS3Bucket(ctx context.Context, s3Client *s3.Client) error {
	allBuckets, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})

	if err != nil {
		return fmt.Errorf("ListBuckets error: %s", err)
	}

	found := false

	for _, bucket := range allBuckets.Buckets {
		if *bucket.Name == bucketName {
			found = true
		}
	}

	if !found {
		_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &typesS3.CreateBucketConfiguration{
				LocationConstraint: regionName,
			},
		})
		if err != nil {
			return fmt.Errorf("CreateBucket error: %s", err)
		}
	}
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &typesS3.CreateBucketConfiguration{
			LocationConstraint: regionName,
		},
	})
	if err != nil {
		return fmt.Errorf("CreateBucket error: %s", err)
	}
	return nil
}

func uploadToS3Bucket(ctx context.Context, s3Client *s3.Client) error {

	uploader := manager.NewUploader(s3Client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("test.txt"),
		Body:   strings.NewReader("hello world"),
	})

	if err != nil {
		return fmt.Errorf("Upload error: %s", err)
	}
	return nil
}
