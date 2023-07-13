package igin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/locales"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

const (
	TranslatorLocaleEN = "en"
	TranslatorLocaleZH = "zh"
)

var (
	Translator   *translator
	UtTranslator ut.Translator
)

type (
	LocaleRegisterDefaultTranslations func(locale string, v *validator.Validate, UtTranslator ut.Translator)
	translator                        struct {
		SupportedLocales                  []locales.Translator
		LocaleRegisterDefaultTranslations LocaleRegisterDefaultTranslations
	}
)

func DefaultTranslator() *translator {
	return NewTranslator([]locales.Translator{en.New(), zh.New()}, func(locale string, v *validator.Validate, UtTranslator ut.Translator) {
		switch locale {
		case TranslatorLocaleEN:
			_ = enTranslations.RegisterDefaultTranslations(v, UtTranslator)
		case TranslatorLocaleZH:
			_ = zhTranslations.RegisterDefaultTranslations(v, UtTranslator)
		default:
			_ = enTranslations.RegisterDefaultTranslations(v, UtTranslator)
		}
	})
}
func NewTranslator(supportedLocales []locales.Translator, localeRegisterDefaultTranslations LocaleRegisterDefaultTranslations) *translator {
	once := &sync.Once{}
	if Translator == nil {
		once.Do(func() {
			Translator = &translator{
				SupportedLocales:                  supportedLocales,
				LocaleRegisterDefaultTranslations: localeRegisterDefaultTranslations,
			}
		})
	}
	return Translator
}

func (t translator) UtTranslator(v *validator.Validate, locale string) (ut.Translator, error) {
	uni := ut.New(t.SupportedLocales[0], t.SupportedLocales...)
	var ok bool
	UtTranslator, ok = uni.GetTranslator(locale)
	if !ok {
		return nil, fmt.Errorf("uni.GetTranslator(%s) failed", locale)
	}
	t.LocaleRegisterDefaultTranslations(locale, v, UtTranslator)
	return UtTranslator, nil
}

func (t translator) ValidateError(err error) error {
	if err != nil {
		if UtTranslator != nil {
			var errs ValidateErrors
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for key, value := range validationErrors.Translate(UtTranslator) {
					errs = append(errs, &ValidateError{
						Key:     key,
						Message: value,
					})
				}
			} else {
				errs = append(errs, &ValidateError{
					Key:     "",
					Message: err.Error(),
				})
			}
			err = errs
		}
		return err
	}
	return nil
}

type ValidateErrors []*ValidateError

type ValidateError struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

func (v *ValidateError) Error() string {
	return v.Message
}

func (v ValidateErrors) Error() string {
	return strings.Join(v.Errors(), ",")
}

func (v ValidateErrors) Errors() []string {
	var errs []string
	for _, err := range v {
		errs = append(errs, err.Error())
	}
	return errs
}
