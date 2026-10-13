# httpx

Logging et redaction pour les appels HTTP sortants, partagés entre nos services.

## Le problème

La v1 vivait dans un seul service et avait deux défauts :

- **Redaction fragile** : la denylist était par client et opt-in, les options *remplaçaient* les défauts (un client perdait `Cookie`), et les regex ratent les valeurs imbriquées, les guillemets échappés, les booléens, les attributs XML.
- **Entrées de log > 256 KB** : Cloud Logging tronque, le JSON ne parse plus, et l'appel devient inexploitable. Fréquent sur les grosses réponses.

## Les décisions

- `redact` : une policy déclarative (`Secrets`, `PII`, `Volatile`), composable par `Merge` (append-only, `Default()` toujours inclus). Un seul passage tokenizer sur les octets, pas de unmarshal/remarshal : un ordre de grandeur plus rapide que N regex (voir `redact/bench_test.go`). Des détecteurs filet de sécurité (email, téléphone, PAN+Luhn, IBAN, JWT, Basic/Bearer) tournent sur les valeurs string.
- `LoggingTransport` : ne mute jamais la requête, capture la réponse en streaming (borné), log à la fermeture du body. `duration` mesure le temps jusqu'aux headers de réponse, pas la lecture du body par l'appelant. Le budget est calculé sur l'entrée sérialisée, JSON valide loggé comme objet, sinon preview tronquée sur une frontière de rune. Le schéma du log reste compatible avec la v1.
- `BodySink` : interface optionnelle (`WithBodySink`) pour stocker hors du log le body complet (redacté) qui dépasse le budget ; le log garde la preview et `bodyRef` (`bodyRefPartial: true` si la capture a été coupée avant d'être envoyée au sink). `WithBodySinkAlways` envoie tous les bodies non binaires au sink, pas seulement ceux qui dépassent le budget (JSON brut plus simple à copier que l'UI Cloud Logging) : le body reste loggé inline quand il tient, avec `bodyRef` en plus.

## Usage

```go
policy := redact.Policy{
	Secrets: redact.Rules{Keys: []string{"cardNumber"}},
	PII:     redact.Rules{Keys: []string{"firstName", "lastName"}},
}

client := &http.Client{
	Transport: httpx.NewLoggingTransport("example-api", httpx.NewTransport(), policy,
		httpx.WithParentRequestID(requestid.Get),
		httpx.WithUserAgent("my-service"),
	),
}
```
