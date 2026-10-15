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

## Documents embarqués

Certaines API portent un document entier dans une valeur string (un élément SOAP dont le texte est du XML échappé en entités, ou une string JSON contenant du JSON). `Policy.Embedded` liste les clés (même matching que `Rules.Keys`, chemins compris) dont la valeur est redactée comme un document imbriqué avec la même policy, plutôt que masquée : la valeur est décodée (entités et CDATA en XML, échappements en JSON), le format est détecté (`<` pour XML, `{` ou `[` pour JSON), puis le document est redacté et ré-échappé, sur 3 niveaux d'imbrication au plus. Tout ce qui n'est pas un document bien formé (texte, JSON ou XML invalide, entité inconnue, niveau 4, attribut XML, élément avec des enfants ou un commentaire, scalaire non null) est masqué en entier : fail closed. Une valeur vide, blanche ou `null` est conservée, et un objet ou tableau JSON déjà structuré est redacté en place avec les règles normales (ses strings ne sont pas traitées comme des documents). Dans un body form, la valeur est masquée en entier. Les règles `Secrets` gagnent si la même clé y figure, `PII` aussi (sauf avec `WithoutPII()`, qui s'applique aussi au document imbriqué), `Volatile` aussi avec `WithVolatile()`. Les chemins repartent de la racine du document imbriqué : `$.x` y matche aussi.

```go
redact.Policy{PII: redact.Rules{Keys: []string{"lastName"}}, Embedded: []string{"fileContent"}}
// <r><fileContent>&lt;o&gt;&lt;lastName&gt;Dupont&lt;/lastName&gt;&lt;ref&gt;A1&lt;/ref&gt;&lt;/o&gt;</fileContent></r>
// -> <r><fileContent>&lt;o&gt;&lt;lastName&gt;REDACTED&lt;/lastName&gt;&lt;ref&gt;A1&lt;/ref&gt;&lt;/o&gt;</fileContent></r>
```
