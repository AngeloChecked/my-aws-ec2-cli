package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	exec "os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func main() {
	openSsh := flag.Bool("ssh", false, "enter in ssh")
	flag.Parse()

	var command string
	var instanceName string

	if len(os.Args) >= 2 {
		command = os.Args[1]
		instanceName = os.Args[2]
	}
	if len(os.Args) >= 3 {
		command = os.Args[2]
		instanceName = os.Args[2]
	}

	fmt.Printf("instance=%s command=%s\n", instanceName, command)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	client := ec2.NewFromConfig(cfg)

	output, err := client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		log.Fatalf("failed to describe instances, %v", err)
	}

	fmt.Println("EC2 Instances:")
	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			printInstance(instance)
			if instanceName == *instance.KeyName && command == "start" {
				fmt.Printf(">>>>> starting instance: %s\n", *instance.KeyName)
				params := ec2.StartInstancesInput{
					InstanceIds: []string{*instance.InstanceId},
				}
				_, err := client.StartInstances(context.TODO(), &params)
				if err != nil {
					fmt.Printf("error: %s", err.Error())
				}
			}
			if instanceName == *instance.KeyName && command == "stop" {
				fmt.Printf(">>>>> stopping instance: %s\n", *instance.KeyName)
				params := ec2.StopInstancesInput{
					InstanceIds: []string{*instance.InstanceId},
				}
				_, err := client.StopInstances(context.TODO(), &params)
				if err != nil {
					fmt.Printf("error: %s", err.Error())
				}
			}
			if *openSsh && instanceName == *instance.KeyName {
				fmt.Printf("ssh %s\n", instanceName)

				home, _ := os.UserHomeDir()
				sshHost := fmt.Sprintf("admin@%s", aws.ToString(instance.PublicIpAddress))
				certFilePath := fmt.Sprintf("%s/vtt2.pem", home)
				cmd := exec.Command("ssh", "-i", certFilePath, sshHost)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Printf("error: %s\n", err.Error())
				}
			}
		}
	}

}

// Helper function to print EC2 instance details
func printInstance(instance types.Instance) {
	fmt.Printf("Keyname: %s\n", aws.ToString(instance.KeyName))
	fmt.Printf("Instance ID: %s\n", aws.ToString(instance.InstanceId))
	fmt.Printf("State: %s\n", instance.State.Name)
	fmt.Printf("Type: %s\n", instance.InstanceType)
	if instance.PublicIpAddress != nil {
		fmt.Printf("Public Dns Name: %s\n", aws.ToString(instance.PublicDnsName))
		fmt.Printf("Public IP: %s\n", aws.ToString(instance.PublicIpAddress))
	}
	if instance.PrivateIpAddress != nil {
		fmt.Printf("Private IP: %s\n", aws.ToString(instance.PrivateIpAddress))
	}
	fmt.Println("---")

	fmt.Println("connect with:")
	fmt.Printf("ssh -i \"vtt2.pem\" admin@%s\n", aws.ToString(instance.PublicDnsName))
	fmt.Printf("http://%s:30000\n", aws.ToString(instance.PublicIpAddress))
}
