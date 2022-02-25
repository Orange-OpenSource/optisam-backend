package dgraph

import (
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
	"text/template"
)

const opsEquipTemplate templateType = "opsEquipmentTemplate"

const equipProcCalTmpl = `
{  
	{{- template "equipUIDFromID" .EquipID }}
	{{ $clOPS := .CalLevelOPS}}
	{{- template "procCalculation" .CalLevelOPS }}
	{{- if ceilRequired $.Met $clOPS.Current $clOPS.Parent}}
		{{- template "licensesEquipCeil" .EqType }}
	{{- else}}
		{{- template "licensesEquip" .EqType }}
	{{- end}}	
}
`

const licensesEquipTmpl = `
{{- define "licensesEquip" }}
     Licenses(){
		 Licenses: sum(val({{$}}_t_{{$}}))
	 }
{{- end}}
`
const licensesEquipCeilTmpl = `
{{- define "licensesEquipCeil" }}
     Licenses(){
		 Licenses: sum(val({{$}}_t_{{$}}_ceil))
		 LicensesNoCeil: sum(val({{$}}_t_{{$}}))
	 }
{{- end}}
`

const equipUIDFromIDTmpl = `
{{- define "equipUIDFromID" }}
var(func:eq(equipment.id,"{{$}}")){
	ID as uid
}
{{- end}}`

const procCalTmpl = `
{{- define "procCalculations"}}
	{{- $baseIndex := baseElementIndex $ -}}
	{{- $length := len $.EqTypeTree}}
	{{- range $index,$_ := seq (sub_int $length $baseIndex ) }}
		{{- $startIndex := (sub_int $length $index 1 ) }}
		{{ if ge $startIndex $baseIndex }}
			{{ template "procCalculation" getCalLevels $ $startIndex $startIndex }}
		{{ end }}
	{{- end}}
{{- end}}

{{- define "procCalculation"}}
var(func:uid({{getProcCalFilter .}})){
	{{- $PL := index $.Mat.EqTypeTree $.Parent}}
	{{- $CL := index $.Mat.EqTypeTree $.Current}}
	{{- if eq $PL.Type $.Mat.BaseType.Type }}
		{{- template "baseCal" getCalLevels $.Mat $.Parent $.Current }}
	{{- else}}
		{{- template "getChild" getCalLevels  $.Mat $.Parent (dec $.Current)}}
		{{- $CL_CHILD := index $.Mat.EqTypeTree (dec $.Current)}}
		{{$CL.Type}}_t_{{$PL.Type}} as sum(val({{$CL_CHILD.Type}}_t_{{$PL.Type}}{{- if  ceilLevel $.Mat (dec $.Current)}}_ceil{{- end}}))
		{{- if ceilRequired $.Mat $.Current $.Parent}}
			{{$CL.Type}}_t_{{$PL.Type}}_ceil as math(ceil {{$CL.Type}}_t_{{$PL.Type}}) 
		{{- end}}
	{{- end}}
}

{{- end}}

{{ define "getChild" -}}
{{- $CL := index $.Mat.EqTypeTree $.Current}}
{{- $PL := index $.Mat.EqTypeTree $.Parent}}
{{- $length := len $.Mat.EqTypeTree}}
~equipment.parent{
	{{- if  eq $CL.Type $.Mat.BaseType.Type -}}
	{{ template "baseCal" getCalLevels $.Mat $.Parent $.Current }}
	{{- else }}
		{{- $CL_CHILD := index $.Mat.EqTypeTree (dec $.Current)}}
		{{- template "getChild" getCalLevels $.Mat $.Parent (dec $.Current)}}
		{{$CL.Type}}_t_{{$PL.Type}} as sum(val({{$CL_CHILD.Type}}_t_{{$PL.Type}}{{- if  ceilLevel $.Mat (dec $.Current)}}_ceil{{- end}}))
		{{- if ceilRequired $.Mat $.Current $.Parent}}
		{{$CL.Type}}_t_{{$PL.Type}}_ceil as math(ceil {{$CL.Type}}_t_{{$PL.Type}})   
		{{- end}}
	{{- end }}
	}
{{- end }}

{{ define "baseCal"}}
{{- $CL := index $.Mat.EqTypeTree $.Current}}
{{- $PL := index $.Mat.EqTypeTree $.Parent}}
{{- $length := len $.Mat.EqTypeTree }}
	{{- if $.Mat.NumCPUAttr.IsSimulated }}
		cpu_{{$PL.Type}} as math({{$.Mat.NumCPUAttr.Val}})
 	{{- else }}
	 	cpu_{{$PL.Type}} as equipment.{{$CL.Type}}.{{$.Mat.NumCPUAttr.Name}}
	{{- end}}
	{{- if $.Mat.CoreFactorAttr.IsSimulated }}
	   coreFactor_{{$PL.Type}} as math({{$.Mat.CoreFactorAttr.Val}})
 	{{- else }}
	 	coreFactor_{{$PL.Type}} as equipment.{{$CL.Type}}.{{$.Mat.CoreFactorAttr.Name}}
	{{- end}}
	{{- if $.Mat.NumCoresAttr.IsSimulated }}
		cores_{{$PL.Type}} as math({{$.Mat.NumCoresAttr.Val}})
 	{{- else }}
	 	cores_{{$PL.Type}} as equipment.{{$CL.Type}}.{{$.Mat.NumCoresAttr.Name}}
 	{{- end}} 
{{$CL.Type}}_t_{{$PL.Type}} as math(cpu_{{$PL.Type}}*cores_{{$PL.Type}}*coreFactor_{{$PL.Type}})
{{- if ceilRequired $.Mat $.Current $.Parent}}
{{ $CL.Type}}_t_{{$PL.Type}}_ceil as math(ceil {{$CL.Type}}_t_{{$PL.Type}})   
{{- end}}
{{- end }}
`

// EquipProcCal ...
type EquipProcCal struct {
	EqType  string
	EquipID string
	Met     *v1.MetricOPSComputed
}

// CalLevelOPS  ...
func (e *EquipProcCal) CalLevelOPS() *CalLevelOPS {
	for i := range e.Met.EqTypeTree {
		if e.EqType == e.Met.EqTypeTree[i].Type {
			return &CalLevelOPS{
				Parent:  i,
				Current: i,
				Mat:     e.Met,
			}
		}
	}
	return &CalLevelOPS{
		Parent:  -1,
		Current: -1,
		Mat:     e.Met,
	}
}

func ceilRequiredOPS(mat *v1.MetricOPSComputed, currentIdx, parentIdx int) bool {
	aggIdx := aggregateElementIndexOPS(mat)
	return aggIdx == currentIdx || (currentIdx <= aggIdx && parentIdx == currentIdx)
}

func ceilLevelOPS(mat *v1.MetricOPSComputed, currentIdx int) bool {
	aggIdx := aggregateElementIndexOPS(mat)
	return aggIdx == currentIdx
}

func aggregateElementIndexOPS(mat *v1.MetricOPSComputed) int {
	for i, eqType := range mat.EqTypeTree {
		if eqType.Type == mat.AggregateLevel.Type {
			return i
		}
	}
	return 0
}

type CalLevelOPS struct {
	Current int
	Parent  int
	Mat     *v1.MetricOPSComputed
}

func getLevelsOPS(mat *v1.MetricOPSComputed, parent, current int) *CalLevelOPS {
	return &CalLevelOPS{
		Parent:  parent,
		Current: current,
		Mat:     mat,
	}
}

// func getProcCalFilterOPSIndividualEuipment(l *CalLevelOPS) string {
// 	bi := baseElementIndexOPS(l.Mat)
// 	filter := make([]string, bi+1)
// 	for i := 0; i <= bi; i++ {
// 		//	fmt.Println("getProcCalFilter", i, l.Parent)
// 		filter[i] = l.Mat.EqTypeTree[i].Type + "_" + l.Mat.EqTypeTree[l.Current].Type + "_" + l.Mat.EqTypeTree[l.Parent].Type
// 	}
// 	return strings.Join(filter, ",")
// }

func baseElementIndexOPS(mat *v1.MetricOPSComputed) int {
	for i, eqType := range mat.EqTypeTree {
		if eqType.Type == mat.BaseType.Type {
			return i
		}
	}
	return 0
}

func templEquipOPS() (*template.Template, error) {
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
		"baseElementIndex": baseElementIndexOPS,
		"mul_int": func(params ...int) int {
			result := 1
			for _, p := range params {
				result *= p
			}
			return result
		},
		"seq": func(i int) []int {
			return make([]int, i)
		},
		"getProcCalFilter": func(l *CalLevelOPS) string { return "ID" },
		"getCalLevels":     getLevelsOPS,
		"ceilRequired":     ceilRequiredOPS,
		"ceilLevel":        ceilLevelOPS,
	}

	templates := []string{
		equipProcCalTmpl,
		equipUIDFromIDTmpl,
		procCalTmpl,
		licensesEquipTmpl,
		licensesEquipCeilTmpl,
	}

	tmplStr := strings.Join(templates, "\n")
	// fmt.Println(tmplStr)

	return template.New("proctempl").Funcs(funcMap).Parse(tmplStr)

}
