package hasura

import "i3s-service/internal/configs"

func postAuthUserMetadata(config *configs.Config) error {
	meta := &HasuraMetadata{}

	meta.Type = "pg_track_table"
	meta.Args.Source = config.Hasura.Source

	meta.Args.Table.Schema = "auth"
	meta.Args.Table.Name = "users"
	meta.Args.Configuration.CustomName = "users"

	customRootFields := &meta.Args.Configuration.CustomRootFields
	customRootFields.Select = "users"
	customRootFields.SelectByPk = "user"
	customRootFields.SelectAggregate = "users_aggregate"
	customRootFields.Insert = "insert_users"
	customRootFields.InsertOne = "insert_user"
	customRootFields.Update = "update_users"
	customRootFields.UpdateByPk = "update_user"
	customRootFields.Delete = "delete_users"
	customRootFields.DeleteByPk = "delete_user"

	return PostHasuraMetadata(config.Hasura.URL, config.Hasura.Secret, meta)
}
