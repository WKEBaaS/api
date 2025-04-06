package hasura

type HasuraTrackFunctionMetadata struct {
	Type string `json:"type"`
	Args struct {
		Source   string `json:"source"`
		Function struct {
			Schema string `json:"schema"`
			Name   string `json:"name"`
		} `json:"function"`
		Configuration struct {
			SessionArgument *string `json:"session_argument"`
			ExposedAs       *string `json:"exposed_as"`
		} `json:"configuration"`
		Comment *string `json:"comment"`
	} `json:"args"`
}

func (s *HasuraService) TrackFunction(schema string, functionName string, sessionArgument *string, exposedAs *string, comment *string) error {
	meta := &HasuraTrackFunctionMetadata{}

	meta.Type = "pg_track_function"
	meta.Args.Source = s.config.Hasura.Source

	meta.Args.Function.Schema = schema
	meta.Args.Function.Name = functionName
	meta.Args.Configuration.SessionArgument = sessionArgument
	meta.Args.Configuration.ExposedAs = exposedAs
	meta.Args.Comment = comment

	if err := s.PostMetadata(meta); err != nil {
		return err
	}

	return nil
}
