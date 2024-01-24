{{ define "name" }}{{ .Release.Name | default "ephemeral" | lower }}-{{ .Chart.Name | lower }}{{ end -}}
{{ define "image" }}{{ .registry }}/{{ .repository }}:{{ .tag }}{{ end -}}
