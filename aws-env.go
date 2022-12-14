package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const (
	formatExports          = "exports"
	formatDotenv           = "dotenv"
	formatDotenvNoQuotes   = "dotenvnoquotes"
	formatIgnoreBreakLines = "ignorebreaklines"
	onlyValue              = "onlyvalue"
)

func main() {
	if os.Getenv("AWS_ENV_PATH") == "" {
		log.Println("aws-env running locally, without AWS_ENV_PATH")
		return
	}

	recursivePtr := flag.Bool("recursive", false, "recursively process parameters on path")
	format := flag.String("format", formatExports, "output format")
	flag.Parse()

	if *format == formatExports || *format == formatDotenv || *format == formatDotenvNoQuotes || *format == formatIgnoreBreakLines || *format == onlyValue {
	} else {
		log.Fatal("Unsupported format option. Must be 'exports' or 'dotenv' or 'dotenvnoquotes' or 'ignorebreaklines'")
	}

	sess := CreateSession()
	client := CreateClient(sess)

	ExportVariables(client, os.Getenv("AWS_ENV_PATH"), *recursivePtr, *format, "")
}

func CreateSession() *session.Session {
	return session.Must(session.NewSession())
}

func CreateClient(sess *session.Session) *ssm.SSM {
	return ssm.New(sess)
}

func ExportVariables(client *ssm.SSM, path string, recursive bool, format string, nextToken string) {
	input := &ssm.GetParametersByPathInput{
		Path:           &path,
		WithDecryption: aws.Bool(true),
		Recursive:      aws.Bool(recursive),
	}

	if nextToken != "" {
		input.SetNextToken(nextToken)
	}

	output, err := client.GetParametersByPath(input)

	if err != nil {
		log.Panic(err)
	}

	for _, element := range output.Parameters {
		OutputParameter(path, element, format)
	}

	if output.NextToken != nil {
		ExportVariables(client, path, recursive, format, *output.NextToken)
	}
}

func OutputParameter(path string, parameter *ssm.Parameter, format string) {
	name := *parameter.Name
	value := *parameter.Value

	env := strings.Replace(strings.Trim(name[len(path):], "/"), "/", "_", -1)
	rawValue := value
	value = strings.Replace(value, "\n", "\\n", -1)

	switch format {
	case formatExports:
		fmt.Printf("export %s=$'%s'\n", env, value)
	case formatDotenv:
		fmt.Printf("%s=\"%s\"\n", env, value)
	case formatDotenvNoQuotes:
		fmt.Printf("%s=%s\n", env, value)
	case onlyValue:
		fmt.Printf("%s", value)
	case formatIgnoreBreakLines:
		fmt.Printf("%s=%s\n", env, rawValue)
	}
}
