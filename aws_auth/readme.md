### AWS Auth
A simple util to get a session token from AWS, using your MFA code and user ARN
and storing that against and existing AWS profile.  For use with the AWS SDK 
standard tools.

## Installation
```shell
go install github.com/SeventhBeam/golang_tools/aws_auth
```

## Usage

```shell
aws_auth
```

Follow the prompts to enter:
Profile
User ARN (Serial)
Then MFA Code