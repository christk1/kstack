{{- define "example-app.name" -}}
{{- default .Chart.Name .Values.nameOverride -}}
{{- end -}}

{{- define "example-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "example-app.fullname" -}}
{{- printf "%s-%s" (include "example-app.name" .) .Release.Name -}}
{{- end -}}
