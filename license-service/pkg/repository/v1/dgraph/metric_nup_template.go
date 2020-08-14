// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
	"text/template"
)

const nupTemplate templateType = "oracle_nup"

const templ = `
{
{{- template "getHighestParents" . }}
{{- template "procCalculations" . }}
{{- template "nupCals" . }}
{{- template "Licenses" . }}
}	

{{- define "getHighestParents" }}
	{{- $baseIndex := baseElementIndex . -}}
	{{- $length := len $.EqTypeTree}}
	{{- range $childIndex , $_ :=  seq (add_int $baseIndex 1) -}}
		{{- range $parentIndex , $_ :=  seq (sub_int $length  $baseIndex ) -}}
			{{- $endIndex := (sub_int $length (add_int  $parentIndex 1))  }}
			{{- template "getHightestParent" getLevel $ $childIndex $endIndex $childIndex $endIndex -}}
		{{- end}}
	{{- end}}
{{- end }}
  
{{- define "getHightestParent"}}
	var (func:uid($ID))@cascade{
		{{- $START := index $.Mat.EqTypeTree $.First}}
	    {{- $LAST  := index $.Mat.EqTypeTree $.Last}}
		user_{{$START.Type}}_{{$LAST.Type}} as product.users {{getFilter $}}{
			{{$START.Type}}_{{$START.Type}}_{{$LAST.Type}} as ~equipment.users  @filter(eq(equipment.type,"{{$START.Type}}")) {
			{{- if lt $.First $.Last}}
				{{template "getParent" getLevel $.Mat $.First  $.Last $.First  $.Last }} 
			{{- end}}
			}
		}
	}
{{- end}}

{{define "getParent"}}
	{{- $begin := inc $.Begin}}
	{{- $FIRST := index $.Mat.EqTypeTree $.First}}
	{{- $LAST  := index $.Mat.EqTypeTree $.Last}}
	{{- $BEGIN := index $.Mat.EqTypeTree $begin -}}
	{{- if eq $begin $.End -}}
	equipment.parent {
		{{ $FIRST.Type}}_{{$BEGIN.Type}}_{{$LAST.Type}}  as uid
		}
	{{- else}}
		{{- $FIRST.Type -}}_{{$BEGIN.Type}}_{{$LAST.Type}} as equipment.parent{
		{{template "getParent" getLevel $.Mat $.First $.Last $begin $.End}}
	    }
	{{- end}}
{{- end}}

{{- define "nupCals" }}
{{- $baseIndex := baseElementIndex $ -}}
{{- $length := len $.EqTypeTree}}
{{- range $index,$_ := seq (sub_int $length $baseIndex ) }}
	{{- $startIndex := (sub_int $length $index 1) }}
	{{ if ge $startIndex $baseIndex }}
		{{- template "nupCal" getCalLevels $ $startIndex $startIndex }}
	{{ end }}
{{- end}}
{{- end}}

{{ define "nupCal"}}
{{- $CL := index $.Mat.EqTypeTree $.Current}}
{{- $PL := index $.Mat.EqTypeTree $.Parent}}
{{- if ceilRequired $.Mat $.Current $.Parent}}
	var(func :uid({{$CL.Type}}_t_{{$CL.Type}}_ceil)){
		v_{{$CL.Type}} as math({{$.Mat.NumOfUsers}}*{{$CL.Type}}_t_{{$CL.Type}}_ceil)
{{- else}}
	var(func :uid({{$CL.Type}}_t_{{$CL.Type}})){
		v_{{$CL.Type}} as math({{$.Mat.NumOfUsers}}*{{$CL.Type}}_t_{{$CL.Type}})
{{- end}}
		{{- if eq $.Mat.BaseType.Type $CL.Type}}
		equipment.users @filter(uid(user_{{$CL.Type}}_{{$PL.Type}})){
			c_{{$CL.Type}}_{{$PL.Type}} as users.count
			{{$CL.Type}}_t_{{$PL.Type}}_users as math(max(v_{{$PL.Type}},c_{{$CL.Type}}_{{$PL.Type}}))
		    }
			{{- if gt $.Current 0 }}
				{{- template "getChildNUP" getCalLevels $.Mat $.Parent (dec $.Current) }}
			{{- end}}
		{{else }}
			{{- template "getChildNUP" getCalLevels $.Mat $.Parent (dec $.Current) }}
		{{- end}}
	}
{{- end}}

{{ define "getChildNUP"}}
	{{- $CL := index $.Mat.EqTypeTree $.Current}}
	{{- $PL := index $.Mat.EqTypeTree $.Parent}}
	{{- $baseIDX := baseElementIndex $.Mat}}
	~equipment.parent @filter(uid({{getNupCalChildFilter $}})) {
		{{- if le $.Current $baseIDX}}
		equipment.users @filter(uid(user_{{$CL.Type}}_{{$PL.Type}})){
			c_{{$CL.Type}}_{{$PL.Type}} as users.count
			{{$CL.Type}}_t_{{$PL.Type}}_users as math(max(v_{{$PL.Type}},c_{{$CL.Type}}_{{$PL.Type}}))
		}
			{{- if gt $.Current 0 }}
			{{- template "getChildNUP" getCalLevels $.Mat $.Parent (dec $.Current) }}
			{{- end}}
		{{- else }}
		{{- template "getChildNUP" getCalLevels $.Mat $.Parent (dec $.Current) }}
		{{- end}}
	}
{{- end}}

{{ define "Licenses"}}
Licenses()@normalize{
	{{getLicensesSum $}}
}
{{- end}}
`

func templateNup() (*template.Template, error) {
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"dec": func(i int) int {
			return i - 1
		},
		"inc": func(i int) int {
			return i + 1
		},
		"add_int": func(params ...int) int {
			sum := 0
			for _, p := range params {
				sum += p
			}
			return sum
		},
		"sub_int": func(params ...int) int {
			sub := 0
			for i, p := range params {
				if i == 0 {
					sub = p
				} else {
					sub -= p
				}

			}
			return sub
		},
		"getLevel":         getLevel,
		"baseElementIndex": baseElementIndex,
		"mul_int": func(params ...int) int {
			result := 1
			for _, p := range params {
				result *= p
			}
			return result
		},
		"seq": func(i int) []int {
			//fmt.Println(i)
			return make([]int, i)
		},
		"getFilter":            getFilter,
		"getProcCalFilter":     getProcCalFilter,
		"getCalLevels":         getCalLevels,
		"getNupCalChildFilter": getNupCalChildFilter,
		"getLicensesSum":       getLicensesSum,
		"ceilRequired":         ceilRequired,
		"ceilLevel":            ceilLevel,
	}

	return template.New("cal_orac_NUP").Funcs(funcMap).Parse(templ + "\n" + procCalTmpl)
	// buf := &bytes.Buffer{}
	// err = tmpl.Execute(buf, mat)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(formatter(buf.String()))
}

func ceilRequired(mat *v1.MetricNUPComputed, currentIdx, parentIdx int) bool {
	aggIdx := aggregateElementIndex(mat)
	return aggIdx == currentIdx || (currentIdx <= aggIdx && parentIdx == currentIdx)
}

func ceilLevel(mat *v1.MetricNUPComputed, currentIdx int) bool {
	aggIdx := aggregateElementIndex(mat)
	return aggIdx == currentIdx
}

func getLicensesSum(mat *v1.MetricNUPComputed) string {
	idx := baseElementIndex(mat)
	if idx < 0 {
		return ""
	}
	comps := []string{}
	totals := []string{}
	for i := 0; i <= idx; i++ {
		for j := idx; j < len(mat.EqTypeTree); j++ {
			totals = append(totals, "l_"+mat.EqTypeTree[i].Type+"_"+mat.EqTypeTree[j].Type)
			comps = append(comps, "l_"+mat.EqTypeTree[i].Type+"_"+mat.EqTypeTree[j].Type+" as sum(val("+mat.EqTypeTree[i].Type+"_t_"+mat.EqTypeTree[j].Type+"_users))")
		}
	}
	if len(comps) == 0 {
		return ""
	}
	comps = append(comps, "Licenses:math("+strings.Join(totals, "+")+")")
	return strings.Join(comps, "\n")
}

func getNupCalChildFilter(l *CalLevel) string {
	bi := baseElementIndex(l.Mat)
	filter := make([]string, 0, bi+1)
	if l.Current < bi {
		bi = l.Current
	}
	for i := 0; i <= bi; i++ {
		//	fmt.Println("getProcCalFilter", i, l.Parent)
		elem := l.Mat.EqTypeTree[i].Type + "_" + l.Mat.EqTypeTree[l.Current].Type + "_" + l.Mat.EqTypeTree[l.Parent].Type
		filter = append(filter, elem)
	}
	return strings.Join(filter, ",")
}

func getProcCalFilter(l *CalLevel) string {
	bi := baseElementIndex(l.Mat)
	filter := make([]string, bi+1)
	for i := 0; i <= bi; i++ {
		//	fmt.Println("getProcCalFilter", i, l.Parent)
		filter[i] = l.Mat.EqTypeTree[i].Type + "_" + l.Mat.EqTypeTree[l.Current].Type + "_" + l.Mat.EqTypeTree[l.Parent].Type
	}
	return strings.Join(filter, ",")
}

// Level ...
type Level struct {
	First int
	Last  int
	Begin int
	End   int
	Mat   *v1.MetricNUPComputed
}

func getFilter(l *Level) string {
	if l.First == 0 && l.Last == len(l.Mat.EqTypeTree)-1 {
		return ""
	}
	bi := baseElementIndex(l.Mat)
	filters := []string{}
	for i := 0; i <= l.First; i++ {
		if i < l.First {
			for j := bi; j < len(l.Mat.EqTypeTree); j++ {
				filters = append(filters, "user_"+l.Mat.EqTypeTree[i].Type+"_"+l.Mat.EqTypeTree[j].Type)
			}
			continue
		}

		for j := l.Last + 1; j < len(l.Mat.EqTypeTree); j++ {
			filters = append(filters, "user_"+l.Mat.EqTypeTree[i].Type+"_"+l.Mat.EqTypeTree[j].Type)
		}
	}
	return "@filter(NOT uid(" + strings.Join(filters, ",") + "))"
}

func baseElementIndex(mat *v1.MetricNUPComputed) int {
	for i, eqType := range mat.EqTypeTree {
		if eqType.Type == mat.BaseType.Type {
			return i
		}
	}
	return 0
}

func aggregateElementIndex(mat *v1.MetricNUPComputed) int {
	for i, eqType := range mat.EqTypeTree {
		if eqType.Type == mat.AggregateLevel.Type {
			return i
		}
	}
	return 0
}

func getLevel(mat *v1.MetricNUPComputed, first, last, begin, end int) *Level {
	// fmt.Println(first, last, begin, end)
	return &Level{
		First: first,
		Last:  last,
		Begin: begin,
		End:   end,
		Mat:   mat,
	}
}

// CalLevel ...
type CalLevel struct {
	Current int
	Parent  int
	Mat     *v1.MetricNUPComputed
}

func getCalLevels(mat *v1.MetricNUPComputed, parent, current int) *CalLevel {
	//	fmt.Println(parent, current)
	return &CalLevel{
		Parent:  parent,
		Current: current,
		Mat:     mat,
	}
}
