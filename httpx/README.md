# httpx

Logging, redaction, retry et VCR pour les appels HTTP sortants, partagés entre pelican, zeus et les autres services.

## Le problème

La v1 vivait dans pelican et avait trois défauts qui font mal en prod :

- **Fuite de PII dans les logs** : la denylist était par client et opt-in, les options *remplaçaient* les défauts (le client IMA perdait `Cookie`), et les regex ratent les valeurs imbriquées, les guillemets échappés, les booléens, les attributs XML.
- **Entrées de log > 256 KB** : Cloud Logging tronque, le JSON ne parse plus, et le reporting perd l'appel. Fréquent avec PAO.
- **Retry caché dans le transport de logs** : un POST non idempotent était rejoué sur 5xx (risque de double réservation).

## Les décisions

- `redact` : une policy déclarative (`Secrets`, `PII`, `Volatile`), composable par `Merge` (append-only, `Default()` toujours inclus). Un seul passage tokenizer sur les octets, pas de unmarshal/remarshal : ~260 MB/s contre ~9 MB/s pour N regex. Des détecteurs filet de sécurité (email, téléphone, PAN+Luhn, IBAN, JWT, Basic/Bearer) tournent sur les valeurs string.
- `LoggingTransport` : ne mute jamais la requête, capture la réponse en streaming (borné), log à la fermeture du body. Le budget est calculé sur l'entrée sérialisée, JSON valide loggé comme objet, sinon preview tronquée sur une frontière de rune. Le schéma du log reste celui lu par le reporting.
- `bodysink` : au-delà du budget, le body complet (redacté) part dans GCS en async, le log garde la preview et `bodyRef`.
- `httpretry` : opt-in, séparé. Rejoue seulement GET/HEAD/OPTIONS/PUT/DELETE ou une requête avec `Idempotency-Key`, respecte `Retry-After`.
- `httpvcr` : toujours go-vcr v4 et les mêmes cassettes `.k7.yaml`. La cassette est redactée à l'enregistrement, le matcher compare des copies normalisées.

## Usage

```go
policy := redact.Policy{
	Secrets: redact.Rules{Keys: []string{"cardNumber"}},
	PII:     redact.Rules{Keys: []string{"firstName", "lastName"}},
}

logging := httpx.NewLoggingTransport("pao", httpx.NewTransport(), policy,
	httpx.WithParentRequestID(requestid.Get),
	httpx.WithUserAgent("imaas"),
	httpx.WithBodySink(sink), // optionnel, voir bodysink/gcs
)

// L'ordre compte : retry au-dessus, chaque tentative est loggée séparément.
client := &http.Client{Transport: httpretry.NewTransport(logging)}
```

En test :

```go
client, recording := httpvcr.NewClient(t, "pao", policy, nil)
// ...
httpvcr.AssertNoPII(t, "testdata") // dans un test dédié, échoue si une cassette contient de la PII
```

`CI` force le replay, `K7_RECORD=1` force l'enregistrement.

Le sink se ferme au shutdown : `defer sink.Close(ctx)`.
