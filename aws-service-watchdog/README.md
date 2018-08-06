# New Relic Infrastructure Integration for aws-service-watchdog

Watches over a set of application resources monitoring their ports.

## Requirements

This integration depends on AWS SDK, go fetch one by

    go get github.com/aws/aws-sdk-go

## Configuration

Cross compilation can be achieved by prepending platform directives, e.g.
    
    env GOOS=linux GOARCH=amd64 make compile-only

## Installation

1. Binary files goes in */var/db/newrelic-infra/custom-integrations/*
    - sudo cp acme-aws-service-watchdog /var/db/newrelic-infra/custom-integrations/
2. Definition file goes in */var/db/newrelic-infra/custom-integrations/acme-aws-service-watchdog-definition.yml*,
    - sudo cp acme-aws-service-watchdog-definition.yml /var/db/newrelic-infra/custom-integrations/acme-aws-service-watchdog-definition.yml
3. Config file goes in */etc/newrelic-infra/integrations.d/*
    - sudo cp acme-aws-service-watchdog-config.yml /etc/newrelic-infra/integrations.d/
4. Restart the new relic infra service

## Usage

The plugin first filters AWS EC2 instances by the selected tag name and then pings the application on the selected port.
Multiple services,instances and ports are selected so one can target many application at a time.

AWS credentials are chained in the following order: shared credentials, environment variables credentials, ec2 role credentials.
If you want to use shared credentials then profile parameter must be defined.

Supported arguments:
- tagname: name of the tag to filter EC2 instances on, mandatory
- region: AWS region to operate in, mandatory
- profile: AWS profile to use while getting EC2 metadata, optional
- timeout: timeout interval, in seconds, to wait for the port become available, optional
- servers: application list in the JSON form, e.g. {\"services\":[{\"name\":\"servicename\",\"tag\":\"tagvalue\",\"port\":\"12345\"}]}

## Compatibility

* Supported OS: Darwin, Linux
* aws-service-watchdog versions: 0.1.0

