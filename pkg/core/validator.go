package core

import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/marmotedu/errors"
	"reflect"
	"saas_service/internal/pkg/code"
	"strings"
)

var (
	uni      *ut.UniversalTranslator
	Validate *validator.Validate
	trans    ut.Translator
)

func FormatCustomValidationErr(err error) error {
	if err != nil {

		var msg []string
		validationErrors := err.(validator.ValidationErrors)
		for _, e := range validationErrors.Translate(trans) {
			msg = append(msg, e)
		}

		err = errors.WithCode(code.ErrValidationCustom, strings.Join(msg, "###"))

		return err

	}
	return err
}

func ValidateStruct(s interface{}) error {
	err := FormatCustomValidationErr(Validate.Struct(s))
	return err
}

func InitValidator() {
	// todo 此处可以再进行扩展，做i18n
	Validate = validator.New()
	zhl := zh.New()
	uni = ut.New(zhl)

	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})
	trans, _ = uni.GetTranslator("zh")
	zh_translations.RegisterDefaultTranslations(Validate, trans)
}
