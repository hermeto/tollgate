{{- define "tollgate.name" -}}{{ .Chart.Name }}{{- end -}}
{{- define "tollgate.fullname" -}}{{ .Release.Name }}-{{ include "tollgate.name" . }}{{- end -}}
