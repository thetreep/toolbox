{{- /* This template generates StreamN functions from n=2 to maxStreams */ -}}
{{- $maxStreams := 40 -}}
package streamwork

/* GENERATED FILE, DO NOT EDIT BY HAND */

import "context"

{{- range $i := iterate 2 $maxStreams }}
func Stream{{ $i }}[
{{- range $j := iterate 1 $i -}}
{{- if eq $j 1 -}}T1 any{{- else -}}, T{{- $j }} any{{- end -}}
{{- end -}}
](
  ctx context.Context,
  source Source[T1],
{{- range $j := iterate 1 (sub $i 1) }}
  worker{{ $j }} Worker[T{{ $j }}, T{{ add $j 1 }}],
{{- end }}
  options ...StreamOption,
) ([]T{{ $i }}, error) {
  return stream(
    func(cfg streamConfig) <-chan T{{ $i }} {
      {{- /* Generate nested worker calls */ -}}
      {{- $depth := 1 -}}
      {{- $nestedCalls := "source(ctx, cfg)" -}}
      {{- range $j := iterate 2 $i -}}
      {{- $nestedCalls = printf "worker%d(ctx, %s, cfg)" (sub $j 1) $nestedCalls -}}
      {{- end -}}

      return {{ $nestedCalls -}}
    }, options...,
  )
}
{{- end }}
