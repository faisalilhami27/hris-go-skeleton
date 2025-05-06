package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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

func Recover() {
	if r := recover(); r != nil {
		logrus.SetLevel(logrus.ErrorLevel)
		logrus.Errorf("recovered from panic: %v", r)
	}
}

func RupiahFormat(amount *float64) string {
	stringValue := "0"
	if amount != nil {
		humanizeValue := humanize.CommafWithDigits(*amount, 0)
		stringValue = strings.ReplaceAll(humanizeValue, ",", ".")
	}
	return fmt.Sprintf("Rp. %s", stringValue)
}

func IndonesianDateFormat(t time.Time, isYear bool) string {
	monthID := map[time.Month]string{
		1: "Jan", 2: "Feb", 3: "Mar", 4: "Apr", 5: "Mei", 6: "Jun",
		7: "Jul", 8: "Agu", 9: "Sep", 10: "Okt", 11: "Nov", 12: "Des",
	}

	day := t.Day()
	month := monthID[t.Month()]
	year := t.Year()

	if isYear == false {
		return fmt.Sprintf("%d %s", day, month)
	}
	return fmt.Sprintf("%d %s %d", day, month, year)
}

func IndonesianMonthFormat(month int) string {
	monthID := map[int]string{
		1:  "Januari",
		2:  "Februari",
		3:  "Maret",
		4:  "April",
		5:  "Mei",
		6:  "Juni",
		7:  "Juli",
		8:  "Agustus",
		9:  "September",
		10: "Oktober",
		11: "November",
		12: "Desember",
	}

	return monthID[month]
}
