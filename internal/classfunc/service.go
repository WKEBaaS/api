package classfunc

import (
	"baas-api/internal/dto"
	"baas-api/internal/pgrest"
	"baas-api/internal/usersdb"
	"bytes"
	"context"
	_ "embed"
	"reflect"
	"text/template"

	"github.com/samber/do/v2"
	"gorm.io/gorm"
)

type CreateClassFunctionData struct {
	Name          string
	Params        []FunctionParam
	Volatility    bool
	Security      string
	Variables     []FunctionVariable
	Authenticated bool
	RootNode      dto.RootNode
	Nodes         dto.Node
}

type FunctionParam struct {
	Name string
	Type string
}

type FunctionVariable struct {
	Name string
	Type string
}

type Service interface {
	CreateClassAPIFunction(ctx context.Context, jwt string, in *dto.CreateClassFunctionInput) error
	DeleteClassAPIFunction(ctx context.Context, jwt string, in *dto.DeleteClassFunctionInput) error
}

type service struct {
	createClassFuncTmpl *template.Template

	// dependencies
	pgrest  pgrest.Service
	usersdb usersdb.Service
}

var _ Service = (*service)(nil)

//go:embed create_class_func.gotmpl
var createClassFuncTmplStr string

func NewService(i do.Injector) (*service, error) {
	createClassFuncTmpl, err := template.New("create_class_function").Parse(createClassFuncTmplStr)
	if err != nil {
		return nil, err
	}

	return &service{
		createClassFuncTmpl: createClassFuncTmpl,
		pgrest:              do.MustInvokeAs[pgrest.Service](i),
		usersdb:             do.MustInvokeAs[usersdb.Service](i),
	}, nil
}

func (s *service) CreateClassAPIFunction(ctx context.Context, jwt string, in *dto.CreateClassFunctionInput) error {
	db, err := s.usersdb.GetDB(ctx, jwt, in.Body.ProjectRef, "superuser")
	if err != nil {
		return err
	}

	err = s.pgrest.CreateClassFunction(ctx, jwt, in)
	if err != nil {
		return err
	}

	funcData, err := NewCreateClassFunctionData(in)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = s.createClassFuncTmpl.Execute(&buf, funcData)
	if err != nil {
		return err
	}

	dropFuncSQL := generateDropFunctionSQL("api", in.Body.Name)
	createFuncSQL := buf.String()

	// Execute SQL
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(dropFuncSQL).Error; err != nil {
			return err
		}
		if err := tx.Exec(createFuncSQL).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteClassAPIFunction(ctx context.Context, jwt string, in *dto.DeleteClassFunctionInput) error {
	db, err := s.usersdb.GetDB(ctx, jwt, in.Body.ProjectRef, "superuser")
	if err != nil {
		return err
	}

	err = s.pgrest.DeleteClassFunction(ctx, jwt, in)
	if err != nil {
		return err
	}

	dropFuncSQL := generateDropFunctionSQL("api", in.Body.Name)

	// Execute SQL
	if err := db.Exec(dropFuncSQL).Error; err != nil {
		return err
	}

	return nil
}

func NewCreateClassFunctionData(in *dto.CreateClassFunctionInput) (*CreateClassFunctionData, error) {
	f := &CreateClassFunctionData{}
	f.Name = generateFunctionName("api", in.Body.Name)
	f.Volatility = true
	f.Security = "DEFINER"
	f.Authenticated = in.Body.Authenticated

	params := []FunctionParam{}
	f.collectNodeParams(in.Body.Node, &params)
	f.Params = params

	// Nodes
	f.RootNode = in.Body.RootNode
	f.Nodes = in.Body.Node
	f.Nodes.Top = true

	return f, nil
}

func (f *CreateClassFunctionData) collectNodeParams(node dto.Node, params *[]FunctionParam) {
	// 1. Get the reflection Value and Type of the Fields struct
	val := reflect.ValueOf(node.Fields)
	typ := val.Type()

	// 2. Iterate over every field in the NodeFields struct
	for i := 0; i < val.NumField(); i++ {
		fieldValue := val.Field(i)
		structField := typ.Field(i)

		// Safety check: Skip if the field pointer is nil
		if fieldValue.IsNil() {
			continue
		}

		// 3. Type assertion: Convert the reflect value back to *FieldConfig
		// (Assuming FieldConfig is inside package 'dto')
		config, ok := fieldValue.Interface().(*dto.FieldConfig)
		if !ok || config.ParamName == nil {
			continue
		}

		// 4. Switch on the Struct Field Name to determine SQL Type
		var sqlType string
		switch structField.Name {
		case "ChineseName", "EnglishName":
			sqlType = "varchar(256)"
		case "ChineseDescription", "EnglishDescription":
			sqlType = "varchar(4000)"
		case "EntityID":
			sqlType = "integer"
		default:
			// Skip fields that aren't in our list
			continue
		}

		// 5. Append to the slice pointer
		*params = append(*params, FunctionParam{
			Name: *config.ParamName,
			Type: sqlType,
		})
	}

	// 6. Recursion for children (unchanged)
	for _, child := range node.Children {
		f.collectNodeParams(child, params)
	}
}
