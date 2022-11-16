package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {

	var (
		instanceId string
		err        error
	)

	ctx := context.Background()
	if instanceId, err = createEC2(ctx, "us-east-1"); err != nil {
		fmt.Printf("createEC2 error: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Instace ID: %s\n", instanceId)
}

func createEC2(ctx context.Context, region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
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
