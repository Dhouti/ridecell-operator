{{ define "componentName" }}web{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "target"}}
    apiVersion: "apps/v1"
    kind: Deployment
    name: {{ .Instance.Name }}-web
{{- end }}
{{ define "minReplicas" }}{{ .Instance.Spec.Replicas.WebAuto.Min }}{{ end }}
{{ define "maxReplicas" }}{{ .Instance.Spec.Replicas.WebAuto.Max }}{{ end }}
{{ define "metric" }}
        name: ridecell:rabbitmq_summon_web_queue_scaler
        selector:
          matchLabels: 
            vhost: {{ .Instance.Name | quote }}
{{- end }}
{{ define "mTarget" }}
        type: Value
        value: 1
{{- end }}
{{ template "hpa" . }}