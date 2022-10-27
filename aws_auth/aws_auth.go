package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

func main() {
	Execute()
}

var (
	rootCmd = &cobra.Command{
		Use:   "login",
		Short: "A Tool to configure a session token for AWS",
		Long:  `A Tool to configure a session token for AWS`,

		Run: func(cmd *cobra.Command, args []string) {
			home := os.Getenv("HOME")
			profile := "default"
			profile = promptGetInput(promptContent{
				errorMsg: "AWS Profile",
				label:    fmt.Sprintf("Profile [%s]", profile),
				fallback: &profile,
			})

			ser, err := os.ReadFile(home + "/.aws/.serial")
			serial := string(ser)
			serial = promptGetInput(promptContent{
				errorMsg: "AWS User Serial Number is required",
				label:    fmt.Sprintf("AWS User Serial [%s]", serial),
				fallback: &serial,
			})

			err = os.WriteFile(home+"/.aws/.serial", []byte(serial), os.ModePerm)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "unable to store serial")
			}

			mfs := promptGetInput(promptContent{
				errorMsg: "MFA Token is required",
				label:    "AWS MFA Token",
			})

			mySession := session.Must(session.NewSession())

			svc := sts.New(mySession)
			ti := sts.GetSessionTokenInput{
				SerialNumber: &serial,
				TokenCode:    &mfs,
			}
			tok, err := svc.GetSessionToken(&ti)
			if err != nil {
				panic("unable to get session token")
			}

			log.Default().Printf("KeyId:%s\nSession token:\n%s\n", *tok.Credentials.AccessKeyId, *tok.Credentials.SessionToken)

			file, err := os.OpenFile(home+"/.aws/config", os.O_RDWR, os.ModeAppend)

			if err != nil {
				panic(fmt.Sprintf("can't open config file %s", err))
			}
			scanner := bufio.NewScanner(file)
			scanner.Split(bufio.ScanLines)

			var text []string

			isProf := false
			for scanner.Scan() {
				next := scanner.Text()
				if isProf {
					if strings.Contains(next, "region") {
						text = append(text, next)
					}
					if strings.Contains(next, "[") {
						text = append(text, fmt.Sprintf("aws_access_key_id=%s", *tok.Credentials.AccessKeyId))
						text = append(text, fmt.Sprintf("aws_secret_access_key=%s", *tok.Credentials.SecretAccessKey))
						text = append(text, fmt.Sprintf("aws_session_token=%s", *tok.Credentials.SessionToken))
						text = append(text, "")
						text = append(text, next)
						isProf = false
					}
				} else {
					text = append(text, next)
				}
				if strings.Contains(next, profile) {
					isProf = true
				}
			}
			if isProf {
				text = append(text, fmt.Sprintf("aws_access_key_id=%s", *tok.Credentials.AccessKeyId))
				text = append(text, fmt.Sprintf("aws_secret_access_key=%s", *tok.Credentials.SecretAccessKey))
				text = append(text, fmt.Sprintf("aws_session_token=%s", *tok.Credentials.SessionToken))
				text = append(text, "")
			}

			err = file.Truncate(0)
			_, err = file.Seek(0, 0)
			if err != nil {
				panic(err)
			}

			dataWriter := bufio.NewWriter(file)

			for _, data := range text {
				_, err = dataWriter.WriteString(data + "\n")
				if err != nil {
					panic(err)
				}
			}
			err = dataWriter.Flush()
			if err != nil {
				panic(err)
			}
			err = file.Close()
			if err != nil {
				panic(err)
			}
		},
	}
)

func init() {
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func promptGetInput(pc promptContent) string {
	validate := func(input string) error {
		if pc.fallback == nil && len(input) <= 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	templates := &promptui.PromptTemplates{
		Prompt:  "{{ . }} ",
		Valid:   "{{ . | green }} ",
		Invalid: "{{ . | red }} ",
		Success: "{{ . | bold }} ",
	}

	prompt := promptui.Prompt{
		Label:     pc.label,
		Templates: templates,
		Validate:  validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	if len(result) > 0 {
		return result
	}
	return *pc.fallback
}

type promptContent struct {
	errorMsg string
	label    string
	fallback *string
}
