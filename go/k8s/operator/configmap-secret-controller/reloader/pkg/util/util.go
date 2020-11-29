package util

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"strings"
)

type List []string

func (l List) Contains(s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}

// GenerateSHA generates SHA from string
func GenerateSHA(data string) string {
	hasher := sha1.New()
	_, err := io.WriteString(hasher, data)
	if err != nil {
		logrus.Errorf("Unable to write data in hash writer %v", err)
	}
	sha := hasher.Sum(nil)
	return fmt.Sprintf("%x", sha)
}

// InterfaceSlice converts an interface to an interface array
func InterfaceSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		logrus.Errorf("InterfaceSlice() given a non-slice type")
	}

	ret := make([]interface{}, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}

	return ret
}

// ConvertToEnvVarName converts the given text into a usable env var
// removing any special chars with '_' and transforming text to upper case
func ConvertToEnvVarName(text string) string {
	var buffer bytes.Buffer
	upper := strings.ToUpper(text)
	lastCharValid := false
	for i := 0; i < len(upper); i++ {
		ch := upper[i]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			buffer.WriteString(string(ch))
			lastCharValid = true
		} else {
			if lastCharValid {
				buffer.WriteString("_")
			}
			lastCharValid = false
		}
	}
	return buffer.String()
}

type ObjectMeta struct {
	metav1.ObjectMeta
}

func ToObjectMeta(kubernetesObject interface{}) ObjectMeta {
	objectValue := reflect.ValueOf(kubernetesObject)
	fieldName := reflect.TypeOf((*metav1.ObjectMeta)(nil)).Elem().Name()
	field := objectValue.FieldByName(fieldName).Interface().(metav1.ObjectMeta)

	return ObjectMeta{
		ObjectMeta: field,
	}
}
