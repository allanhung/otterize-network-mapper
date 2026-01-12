package mysqlstore

import (
	"github.com/otterize/network-mapper/src/mapper/pkg/config"
	"github.com/spf13/viper"
)

type Config struct {
	ClientIgnoreListByName      string
	ClientIgnoreListByNamespace string
	Cluster                     string
	DbHost                      string
	DbUsername                  string
	DbPassword                  string
	DbPort                      string
	DbDatabase                  string
	GhaDispatchEnabled          bool
	GhaToken                    string
	GhaUrl                      string
	GhaOwner                    string
	GhaRepo                     string
	GhaEventType                string
	RetentionDays               int
}

func ConfigFromViper() Config {
	return Config{
		ClientIgnoreListByName:      viper.GetString(config.ClientIgnoreListByNameKey),
		ClientIgnoreListByNamespace: viper.GetString(config.ClientIgnoreListByNamespaceKey),
		Cluster:                     viper.GetString(config.ClusterKey),
		DbHost:                      viper.GetString(config.DbHostKey),
		DbUsername:                  viper.GetString(config.DbUsernameKey),
		DbPassword:                  viper.GetString(config.DbPasswordKey),
		DbPort:                      viper.GetString(config.DbPortKey),
		DbDatabase:                  viper.GetString(config.DbDatabaseKey),
		GhaDispatchEnabled:          viper.GetBool(config.GhaDispatchEnabledKey),
		GhaToken:                    viper.GetString(config.GhaTokenKey),
		GhaUrl:                      viper.GetString(config.GhaUrlKey),
		GhaOwner:                    viper.GetString(config.GhaOwnerKey),
		GhaRepo:                     viper.GetString(config.GhaRepoKey),
		GhaEventType:                viper.GetString(config.GhaEventTypeKey),
		RetentionDays:               viper.GetInt(config.ExternalIntentsRetentionDaysKey),
	}
}
