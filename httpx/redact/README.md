# redact

Redaction des secrets et de la PII dans les headers, query params et bodies (JSON, XML, formulaires).

## Clés par chemin

Dans `Rules.Keys`, une clé sans point (`name`) matche ce nom à n'importe quelle profondeur. Une clé avec des points (`customer.name`) est un chemin : elle matche `name` dont le parent direct est la valeur de la clé `customer`, à n'importe quelle profondeur. Les tableaux sont transparents (`{"customers":[{"name":..}]}` matche `customers.name`) et la casse est ignorée par segment.

Un préfixe `$.` ancre le chemin à la racine : `$.name` ne matche que le `name` de premier niveau (un tableau à la racine est transparent aussi).

En XML, le chemin porte sur les noms d'éléments, préfixes de namespace ignorés, dans le document comme dans les parents de la règle (`soap:Body.ssn` matche `<Body><ssn>`) ; le dernier segment est un élément ou un attribut de l'élément parent : `driver.last_name` matche `<driver last_name="..">` et `<driver><last_name>..</last_name></driver>`. `$.` ancre devant l'élément racine du document : `$.driver` matche l'élément racine `<driver>`, `$.driver.id` son attribut ou enfant `id`.

```go
redact.Policy{PII: redact.Rules{Keys: []string{"customer.name", "$.pin"}}}
// {"customer":{"name":"x"},"agency":{"name":"y"},"pin":1,"a":{"pin":2}}
// -> {"customer":{"name":"REDACTED"},"agency":{"name":"y"},"pin":"REDACTED","a":{"pin":2}}
```

Les règles `Volatile` ont la même sémantique. Les query params, champs de formulaire et headers restent plats : une clé avec des points y est un nom littéral. Un chemin matche aussi une clé JSON ou XML dont le nom contient littéralement les points (`{"customer.name":..}`). Une clé avec un segment vide (`a..b`, `a.`, `.a`, `$.`) fait paniquer `redact.New`.

## Sans PII

`redact.WithoutPII()` ignore les règles `PII` et les détecteurs email, téléphone, PAN et IBAN ; les secrets (règles `Secrets`, JWT, Bearer, Basic) restent redactés, et `Volatile` reste actif avec `WithVolatile()`. Prévu pour les cassettes, qui ne doivent pas fuiter de secrets mais peuvent contenir de fausses PII.
