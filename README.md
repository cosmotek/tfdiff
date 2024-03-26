# tfdiff
Generate reports for your migration from ClickOps to Terraform.

<img width="855" alt="image" src="https://github.com/cosmotek/tfdiff/assets/54327825/a100f92e-0583-4e18-81be-977b1d025eab">

Example Report:
```sh
tfdiff completed in 2m10.645885167s.

final report:
managed (77/2656 - 2.899096%)
unmanaged (2579/2656 - 97.100904%)

unmanaged asset breakdown:
	region us-east-1 (2540/2579 - 98.487786%):
		ecs:task 1000
		ecs:task-definition 998
		ssm:parameter 154
		rds:snapshot 113
		ec2:network-interface 69
		logs:log-group 26
		ec2:security-group-rule 26
		events:rule 19
		elasticache:parametergroup 15
		s3:bucket 11
		ec2:volume 10
		kms:key 8
		cloudformation:stack 8
		ec2:instance 6
		ecs:container-instance 6
		ec2:vpc-endpoint 5
		rds:pg 5
		rds:og 4
		ecs:service 4
		ec2:network-insights-path 4
		cloudwatch:alarm 4
		lambda:function 3
		memorydb:parametergroup 3
		ecr:repository 3
		ec2:key-pair 2
		resource-explorer-2:view 2
		rds:cluster-pg 2
		sns:topic 2
		ec2:security-group 2
		ecs:cluster 2
		rds:auto-backup 2
		ec2:dhcp-options 1
		resource-explorer-2:index 1
		elasticloadbalancing:listener-rule/app 1
		rds:secgrp 1
		elasticloadbalancing:targetgroup 1
		elasticloadbalancing:loadbalancer/app 1
		memorydb:user 1
		backup:backup-plan 1
		ec2:network-acl 1
		ec2:route-table 1
		athena:workgroup 1
		s3:storage-lens 1
		elasticache:user 1
		elasticfilesystem:file-system 1
		events:event-bus 1
		ec2:elastic-ip 1
		rds:cluster-snapshot 1
		elasticloadbalancing:listener/app 1
		ec2:internet-gateway 1
		athena:datacatalog 1
		states:stateMachine 1
		ec2:natgateway 1
	region us-east-2 (39/2579 - 1.512214%):
		elasticache:parametergroup 14
		ec2:subnet 3
		memorydb:parametergroup 3
		ec2:security-group-rule 2
		events:rule 2
		rds:secgrp 1
		memorydb:user 1
		ec2:dhcp-options 1
		ec2:internet-gateway 1
		ec2:security-group 1
		ec2:vpc 1
		cloudformation:stack 1
		athena:datacatalog 1
		resource-explorer-2:index 1
		cloudformation:stackset 1
		ec2:network-acl 1
		ec2:route-table 1
		elasticache:user 1
		events:event-bus 1
		athena:workgroup 1
```

## Features
- [x] Output list of unmanaged resources to CSV
- [x] CLI reporting with asset breakdown by region and resource type
- [x] Multi-region scan support
- [x] Support for AWS SSO managed credentials
- [x] Resource type exclusion filtering

## Installation

This program may be installed by downloading the latest executable from the releases page, moving it into your path, and making it executable. See the example below for Unix-based environments:
```sh
wget https://github.com/cosmotek/tfdiff/releases/download/v1.1.0-rc/tfdiff-linux-amd64.zip
unzip tfdiff-linux-amd64.zip
mv tfdiff /usr/local/bin/tfdiff
chmod +x /usr/local/bin/tfdiff
```

## Usage

Before running tfdiff, you will need to have the following:
- An AWS account with Resource Explorer 2 enabled
- A valid AWS credentials file with a profile for the account you want to diff
- Terraform installed on your machine

Once you have all the requirements, you may run tfdiff like so:
```sh
# open your terraform project
cd my-terraform-project

# select the terraform workspace you want to diff (assuming you have one)
terraform workspace select development

# select an AWS profile with credentials for the target environment
export AWS_PROFILE=development

# run tfdiff against two regions (instead of defaulting to all regions), outputing the list of unmanaged resources to a csv file
tfdiff aws --regions=us-east-1,us-east-2 --output-file unmanaged_resources.csv
```

### Ignore Files

In order to ignore one or many resources when scanning the system, create a file named `.tfdiff_ignore` located within the Terraform project directory you've been operating in. Just specify one ARN or glob per a line, save, and run tfdiff. Ignore files are automatically detected and validated before each scan.

Here's a little example:
```
arn:aws:rds:us-east-1:0123456789:*
arn:aws:ecs:us-east-1:0123456789:*
arn:aws:ecs:us-east-1:0123456789:task-definition/amazing-api:123
```

For more configuration options, run `tfdiff aws --help`.

## Known issues & limitations

- AWS Inventory Truncation:
This tool uses the AWS Resource Explorer 2 API to list asset inventory in the target environment. This API has a max page size of 1000, with no pagination support. Tfdiff scans each region and each resource type individually in order avoid to hitting this limit, but it's possible that regions/resource types with many assets may be truncated at 1000. We are currently exploring other workarounds. For now Tfdiff will output a warning for any region/resource type combo that returns exactly 1000 resources.
- AWS Service Quotas:
Given Tfdiff makes `num_target_regions * num_resource_types` queries for each diff, the AWS services quotas may be exceeded with many monthly executions. Hitting a quota will cause this tool to error out completely. You may request a quota adjustment by AWS in the Services Quota Console.

## Planned features

- [x] Resource Type Filters
- [x] The ability to ignore resources by ARN/Identifier via .tfdiff_ignore files
- [ ] Support for GCP, Azure & DigitalOcean
- [ ] (Possible) Scan Caching
- [ ] Support for multiple Terraform projects and workspaces at once
- [ ] Automated config drift detection
