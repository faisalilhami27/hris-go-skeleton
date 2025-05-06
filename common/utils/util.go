package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"html/template"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func GenerateSHA256(inputString string) string {
	hash := sha256.New()
	hash.Write([]byte(inputString))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func BindFromJSON(dest any, filename, path string) error {
	v := viper.New()

	v.SetConfigType("json")
	v.AddConfigPath(path)
	v.SetConfigName(filename)

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal config env: %+v\n", err)
		return err
	}

	return nil
}

func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New()

	v.SetConfigType("json")
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		return err
	}

	err = v.ReadRemoteConfig()
	if err != nil {
		return err
	}

	logrus.Infof("using config from consul: %s/%s.\n", endPoint, path)

	err = v.Unmarshal(dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal config dest: %+v\n", err)
		return err
	}

	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("failed to set env from consul: %+v\n", err)
		return err
	}

	return nil
}

func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("failed to unmarshal config env: %+v\n", err)
		return err
	}

	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool())
		}

		err = os.Setenv(k, val)
		if err != nil {
			return err
		}
	}

	return nil
}

func FormatCompanyName(name string) string {
	businessEntities := []string{"PT", "CV", "LTD", "PTE", "Corp", "Inc"}
	re := regexp.MustCompile(`[^a-zA-Z\s]`)
	cleaned := re.ReplaceAllString(name, "")
	words := strings.Fields(cleaned)

	var filteredWords []string
	for _, word := range words {
		if !contains(businessEntities, word) {
			filteredWords = append(filteredWords, word)
		}
	}

	formatted := strings.ToLower(strings.Join(filteredWords, "-"))
	return formatted
}

func contains(slice []string, item string) bool {
	for _, str := range slice {
		if strings.EqualFold(str, item) {
			return true
		}
	}
	return false
}

func RupiahFormat(amount *float64) string {
	stringValue := "0"
	if amount != nil {
		humanizeValue := humanize.CommafWithDigits(*amount, 0)
		stringValue = strings.ReplaceAll(humanizeValue, ",", ".")
	}
	return fmt.Sprintf("Rp. %s", stringValue)
}

func Recover() {
	if r := recover(); r != nil {
		logrus.SetLevel(logrus.ErrorLevel)
		logrus.Errorf("recovered from panic: %v", r)
	}
}

func add1(a int) int {
	return a + 1
}

func GeneratePDFFromHTML(htmlTemplate string, data any) ([]byte, error) {
	funcMap := template.FuncMap{
		"add1": add1,
	}

	template, err := template.New("htmlTemplate").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	var filledTemplate bytes.Buffer
	if err := template.Execute(&filledTemplate, data); err != nil {
		return nil, err
	}
	htmlContent := filledTemplate.String()

	pdfGenerator, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		logrus.Errorf("failed to create pdf generator: %v", err)
		return nil, err
	}

	pdfGenerator.Dpi.Set(600)
	pdfGenerator.NoCollate.Set(false)
	pdfGenerator.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	pdfGenerator.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfGenerator.Grayscale.Set(false)
	pdfGenerator.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(htmlContent)))

	err = pdfGenerator.Create()
	if err != nil {
		logrus.Errorf("failed to create pdf: %v", err)
		return nil, err
	}

	return pdfGenerator.Bytes(), err
}
