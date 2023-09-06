package terraform

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

func WhichTerraform() (string, error) {
	return exec.LookPath("terraform")
}

func PullState() (PullStateOutput, error) {
	_, err := WhichTerraform()
	if err != nil {
		return PullStateOutput{}, err
	}

	data, err := exec.Command("terraform", "state", "pull").Output()
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to execute 'terraform state pull': %w", err)
	}

	output := PullStateOutput{}
	err = json.Unmarshal(data, &output)
	if err != nil {
		return PullStateOutput{}, fmt.Errorf("failed to parse terraform state json: %w", err)
	}

	return output, nil
}

type PullStateOutput struct {
	Version          int    `json:"version"`
	TerraformVersion string `json:"terraform_version"`
	Serial           int    `json:"serial"`
	Lineage          string `json:"lineage"`
	Outputs          struct {
	} `json:"outputs"`
	Resources []struct {
		Module    string `json:"module"`
		Mode      string `json:"mode"`
		Type      string `json:"type"`
		Name      string `json:"name"`
		Provider  string `json:"provider"`
		Instances []struct {
			SchemaVersion int `json:"schema_version"`
			Attributes    struct {
				Address                            string        `json:"address"`
				AllocatedStorage                   int           `json:"allocated_storage"`
				AllowMajorVersionUpgrade           interface{}   `json:"allow_major_version_upgrade"`
				ApplyImmediately                   interface{}   `json:"apply_immediately"`
				Arn                                string        `json:"arn"`
				AutoMinorVersionUpgrade            bool          `json:"auto_minor_version_upgrade"`
				AvailabilityZone                   string        `json:"availability_zone"`
				BackupRetentionPeriod              int           `json:"backup_retention_period"`
				BackupWindow                       string        `json:"backup_window"`
				CaCertIdentifier                   string        `json:"ca_cert_identifier"`
				CharacterSetName                   string        `json:"character_set_name"`
				CopyTagsToSnapshot                 bool          `json:"copy_tags_to_snapshot"`
				CustomIamInstanceProfile           string        `json:"custom_iam_instance_profile"`
				CustomerOwnedIPEnabled             bool          `json:"customer_owned_ip_enabled"`
				DbName                             string        `json:"db_name"`
				DbSubnetGroupName                  string        `json:"db_subnet_group_name"`
				DeleteAutomatedBackups             bool          `json:"delete_automated_backups"`
				DeletionProtection                 bool          `json:"deletion_protection"`
				Domain                             string        `json:"domain"`
				DomainIamRoleName                  string        `json:"domain_iam_role_name"`
				EnabledCloudwatchLogsExports       []string      `json:"enabled_cloudwatch_logs_exports"`
				Endpoint                           string        `json:"endpoint"`
				Engine                             string        `json:"engine"`
				EngineVersion                      string        `json:"engine_version"`
				EngineVersionActual                string        `json:"engine_version_actual"`
				FinalSnapshotIdentifier            interface{}   `json:"final_snapshot_identifier"`
				HostedZoneID                       string        `json:"hosted_zone_id"`
				IamDatabaseAuthenticationEnabled   bool          `json:"iam_database_authentication_enabled"`
				ID                                 string        `json:"id"`
				Identifier                         string        `json:"identifier"`
				IdentifierPrefix                   string        `json:"identifier_prefix"`
				InstanceClass                      string        `json:"instance_class"`
				Iops                               int           `json:"iops"`
				KmsKeyID                           string        `json:"kms_key_id"`
				LatestRestorableTime               Timestamp     `json:"latest_restorable_time"`
				LicenseModel                       string        `json:"license_model"`
				MaintenanceWindow                  string        `json:"maintenance_window"`
				MaxAllocatedStorage                int           `json:"max_allocated_storage"`
				MonitoringInterval                 int           `json:"monitoring_interval"`
				MonitoringRoleArn                  string        `json:"monitoring_role_arn"`
				MultiAz                            bool          `json:"multi_az"`
				Name                               string        `json:"name"`
				NcharCharacterSetName              string        `json:"nchar_character_set_name"`
				NetworkType                        string        `json:"network_type"`
				OptionGroupName                    string        `json:"option_group_name"`
				ParameterGroupName                 string        `json:"parameter_group_name"`
				Password                           string        `json:"password"`
				PerformanceInsightsEnabled         bool          `json:"performance_insights_enabled"`
				PerformanceInsightsKmsKeyID        string        `json:"performance_insights_kms_key_id"`
				PerformanceInsightsRetentionPeriod int           `json:"performance_insights_retention_period"`
				Port                               int           `json:"port"`
				PubliclyAccessible                 bool          `json:"publicly_accessible"`
				ReplicaMode                        string        `json:"replica_mode"`
				Replicas                           []string      `json:"replicas"`
				ReplicateSourceDb                  string        `json:"replicate_source_db"`
				ResourceID                         string        `json:"resource_id"`
				RestoreToPointInTime               []interface{} `json:"restore_to_point_in_time"`
				S3Import                           []interface{} `json:"s3_import"`
				SecurityGroupNames                 []interface{} `json:"security_group_names"`
				SkipFinalSnapshot                  bool          `json:"skip_final_snapshot"`
				SnapshotIdentifier                 interface{}   `json:"snapshot_identifier"`
				Status                             string        `json:"status"`
				StorageEncrypted                   bool          `json:"storage_encrypted"`
				StorageType                        string        `json:"storage_type"`
				Tags                               struct {
				} `json:"tags"`
				TagsAll struct {
				} `json:"tags_all"`
				Timeouts struct {
					Create interface{} `json:"create"`
					Delete interface{} `json:"delete"`
					Update interface{} `json:"update"`
				} `json:"timeouts"`
				Timezone            string   `json:"timezone"`
				Username            string   `json:"username"`
				VpcSecurityGroupIds []string `json:"vpc_security_group_ids"`
			} `json:"attributes"`
			Private      string   `json:"private"`
			Dependencies []string `json:"dependencies"`
		} `json:"instances"`
	} `json:"resources"`
	CheckResults interface{} `json:"check_results"`
}

type Timestamp time.Time

func (m *Timestamp) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	return json.Unmarshal(data, (*time.Time)(m))
}
