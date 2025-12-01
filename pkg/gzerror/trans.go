package gzerror

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant_TW"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

func Trans(err error) string {
	return TransErr(err).Error()
}

func TransErr(err error) error {
	locale := "zh"
	var ret []string
	if validationErrors, ok := err.(validator.ValidationErrors); !ok {
		return err
	} else {
		for _, e := range validationErrors {
			ret = append(ret, e.Translate(getTranslator(locale)))
		}
	}

	return errors.New(strings.Join(ret, ";"))
}

func getTranslator(locale string) ut.Translator {
	uni := ut.New(en.New(), zh.New(), zh_Hant_TW.New())
	trans, _ := uni.GetTranslator(locale)
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		switch locale {
		case "zh":
			_ = zh_translations.RegisterDefaultTranslations(v, trans)
			break
		case "en":
			_ = en_translations.RegisterDefaultTranslations(v, trans)
			break
		default:
			_ = zh_translations.RegisterDefaultTranslations(v, trans)
			break
		}
	}

	return trans
}
