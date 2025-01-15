# Translator

You can use go template for gender need:

```yaml
surroundedByFriends:
    hash: sha1-eb6b8ad3d8b687bad425c01c011eddaae00b86f9
    other: '{{if eq .Gender "female"}}entouré d''amie{{else if eq .Gender "male"}}entouré d''ami{{else}}entouré d''ami.e{{end}}'
```


```go
svc.Translate(ctx, messageID, map[string]string{"Gender": "neutral"})
```

you can check `translation_test.go` for more use cases
