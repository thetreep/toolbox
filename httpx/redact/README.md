# redact

Redaction des secrets et de la PII dans les headers, query params et bodies (JSON, XML, formulaires).

## Clés par chemin

Dans `Rules.Keys`, une clé sans point (`name`) matche ce nom à n'importe quelle profondeur. Une clé avec des points (`customer.name`) est un chemin : elle matche `name` dont le parent direct est la valeur de la clé `customer`, à n'importe quelle profondeur. Les tableaux sont transparents (`{"customers":[{"name":..}]}` matche `customers.name`) et la casse est ignorée par segment.

Un préfixe `$.` ancre le chemin à la racine : `$.name` ne matche que le `name` de premier niveau (un tableau à la racine est transparent aussi).

En XML, le chemin porte sur les noms d'éléments, préfixes de namespace ignorés ; le dernier segment est un élément ou un attribut de l'élément parent : `driver.last_name` matche `<driver last_name="..">` et `<driver><last_name>..</last_name></driver>`. `$.` ancre devant l'élément racine du document : `$.driver` matche l'élément racine `<driver>`, `$.driver.id` son attribut ou enfant `id`.

```go
redact.Policy{PII: redact.Rules{Keys: []string{"customer.name", "$.pin"}}}
// {"customer":{"name":"x"},"agency":{"name":"y"},"pin":1,"a":{"pin":2}}
// -> {"customer":{"name":"REDACTED"},"agency":{"name":"y"},"pin":"REDACTED","a":{"pin":2}}
```

Les règles `Volatile` ont la même sémantique. Les query params, champs de formulaire et headers restent plats : une clé avec des points y est un nom littéral. Une clé avec un segment vide (`a..b`, `a.`, `$.`) n'est pas un chemin et reste un nom littéral. En JSON, un chemin ne matche pas une clé plate qui contiendrait elle-même des points (`{"customer.name":..}`).
