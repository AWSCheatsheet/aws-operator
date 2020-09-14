package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Finalizer Cloud Formation Stack.
Resources:
  {{ template "record_sets" . }}
{{ end }}
`
