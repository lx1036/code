package log

import (
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	kubeapi "k8s.io/api/core/v1"
	"os"
	"reflect"
)

/**
@see https://github.com/kubernetes/kubectl
*/
func Table(data []kubeapi.Event) {
	if len(data) == 0 {
		fmt.Println("empty: 0")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	v := reflect.ValueOf(data[0])
	typeOfV := v.Type()
	var headerRow table.Row
	for i := 0; i < v.NumField(); i++ {
		headerRow = append(headerRow, typeOfV.Field(i).Name)
	}
	t.AppendHeader(headerRow)

	fmt.Println(headerRow)

	var dataRow table.Row
	for _, value := range data {
		v := reflect.ValueOf(value)
		dataRow = table.Row{}
		for i := 0; i < v.NumField(); i++ {
			dataRow = append(dataRow, v.Field(i).Interface())
		}
		t.AppendRow(dataRow)
	}

	t.Render()
}
