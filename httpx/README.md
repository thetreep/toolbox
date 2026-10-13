# httpx

Logging, redaction, retry et VCR pour les appels HTTP sortants, partagés entre nos services.

## Le problème

La v1 vivait dans un seul service et avait trois défauts :

- **Redaction fragile** : la denylist était par client et opt-in, les options *remplaçaient* les défauts (un client perdait `Cookie`), et les regex ratent les valeurs imbriquées, les guillemets échappés, les booléens, les attributs XML.
- **Entrées de log > 256 KB** : Cloud Logging tronque, le JSON ne parse plus, et l'appel devient inexploitable. Fréquent sur les grosses réponses.
- **Retry caché dans le transport de logs** : un POST non idempotent était rejoué sur 5xx (risque de double réservation).

## Les décisions

- `redact` : une policy déclarative (`Secrets`, `PII`, `Volatile`), composable par `Merge` (append-only, `Default()` toujours inclus). Un seul passage tokenizer sur les octets, pas de unmarshal/remarshal : un ordre de grandeur plus rapide que N regex (voir `redact/bench_test.go`). Des détecteurs filet de sécurité (email, téléphone, PAN+Luhn, IBAN, JWT, Basic/Bearer) tournent sur les valeurs string.
- `LoggingTransport` : ne mute jamais la requête, capture la réponse en streaming (borné), log à la fermeture du body. `duration` mesure le temps jusqu'aux headers de réponse, pas la lecture du body par l'appelant. Le budget est calculé sur l'entrée sérialisée, JSON valide loggé comme objet, sinon preview tronquée sur une frontière de rune. Le schéma du log reste compatible avec la v1.
- `bodysink` : au-delà du budget, le body complet (redacté) part dans GCS en async, le log garde la preview et `bodyRef` (`bodyRefPartial: true` si la capture a été coupée avant d'être envoyée au sink). `WithBodySinkAlways` envoie tous les bodies non binaires au sink, pas seulement ceux qui dépassent le budget (JSON brut plus simple à copier que l'UI Cloud Logging) : le body reste loggé inline quand il tient, avec `bodyRef` en plus.
- `httpretry` : opt-in, séparé. Rejoue seulement GET/HEAD/OPTIONS/PUT/DELETE ou une requête avec `Idempotency-Key` (garde-fou toujours actif, `WithShouldRetry` choisit seulement quels échecs rejouer), respecte `Retry-After`.
- `httpvcr` : toujours go-vcr v4 et les mêmes cassettes `.k7.yaml`. La cassette est redactée à l'enregistrement, le matcher compare des copies normalisées.

## Usage

```go
policy := redact.Policy{
	Secrets: redact.Rules{Keys: []string{"cardNumber"}},
	PII:     redact.Rules{Keys: []string{"firstName", "lastName"}},
}

logging := httpx.NewLoggingTransport("example-api", httpx.NewTransport(), policy,
	httpx.WithParentRequestID(requestid.Get),
	httpx.WithUserAgent("my-service"),
	httpx.WithBodySink(sink), // optionnel, voir bodysink/gcs
)

// L'ordre compte : retry au-dessus, chaque tentative est loggée séparément.
client := &http.Client{Transport: httpretry.NewTransport(logging)}
```

En test :

```go
client, recording := httpvcr.NewClient(t, "example-api", policy, nil)
// ...
httpvcr.AssertNoPII(t, "testdata") // dans un test dédié, échoue si une cassette contient de la PII
```

`CI` force le replay, `K7_RECORD=1` force le réenregistrement, sinon la cassette est enregistrée si absente.

Le sink se ferme au shutdown : `defer sink.Close(ctx)`.
