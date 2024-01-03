package data

import (
	"log"
	"reflect"
	"regexp"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	ch_translations "github.com/go-playground/validator/v10/translations/zh"
)

// 全局Validate数据校验实列
var Validate *validator.Validate

// 全局翻译器
var Trans ut.Translator

func IntiValidate() {
	chinese := zh.New()
	uni := ut.New(chinese, chinese)
	trans, _ := uni.GetTranslator("zh")
	Trans = trans
	Validate = validator.New()
	_ = ch_translations.RegisterDefaultTranslations(Validate, Trans)
	_ = Validate.RegisterValidation("checkAtomComposeName", atomComposeNameValidation)
	_ = Validate.RegisterValidation("checkParamName", paramNameValidation)
}

func atomComposeNameValidation(fl validator.FieldLevel) bool {
	verificationRole := "^[a-zA-Z\u4e00-\u9fa5][.\\-_a-zA-Z0-9\u4e00-\u9fa5]{0,49}$"
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		re, err := regexp.Compile(verificationRole)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		return re.MatchString(field.String())
	default:
		return false
	}
}

func paramNameValidation(fl validator.FieldLevel) bool {
	verificationRole := "^[a-zA-Z][_a-zA-Z0-9]{0,49}$"
	field := fl.Field()
	switch field.Kind() {
	case reflect.String:
		re, err := regexp.Compile(verificationRole)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		return re.MatchString(field.String())
	default:
		return false
	}
}
